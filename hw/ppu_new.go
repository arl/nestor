package hw

import (
	"fmt"
	"os"
	"unsafe"

	"nestor/hw/hwio"
	"nestor/hw/snapshot"
)

type ConsoleRegion int

const (
	Ntsc ConsoleRegion = iota
	Pal
	Dendy
	Auto
)

type ppureg uint16

const (
	PPUCTRL   ppureg = 0x00
	PPUMASK   ppureg = 0x01
	PPUSTATUS ppureg = 0x02
	OAMADDR   ppureg = 0x03
	OAMDATA   ppureg = 0x04
	PPUSCROLL ppureg = 0x05
	PPUADDR   ppureg = 0x06
	PPUDATA   ppureg = 0x07

	OAMDMA ppureg = 0x4014
)

type ppuControl struct {
	VerticalWrite         bool
	SpritePatternAddr     uint16
	BackgroundPatternAddr uint16
	LargeSprites          bool
	SecondaryPpu          bool
	NmiOnVerticalBlank    bool
}

type ppuMask struct {
	Grayscale         bool
	BackgroundMask    bool
	SpriteMask        bool
	BackgroundEnabled bool
	SpritesEnabled    bool
	IntensifyRed      bool
	IntensifyGreen    bool
	IntensifyBlue     bool
}

type ppuStatus struct {
	SpriteOverflow bool
	Sprite0Hit     bool
	VerticalBlank  bool
}

type tileInfo struct {
	LowByte       uint8
	HighByte      uint8
	PaletteOffset uint8
	TileAddr      uint16
}

type spriteInfo struct {
	SpriteX            uint8
	LowByte            uint8
	HighByte           uint8
	PaletteOffset      uint8
	HorizontalMirror   bool
	BackgroundPriority bool
}

type PPU struct {
	CPU *CPU

	// Clocking
	masterClock uint64
	Cycle       uint32
	Scanline    int16

	masterClockDivider uint8

	// RGBA Frame Buffer
	framebuf []uint32

	// Memory
	// The PPU addresses a 14-bit (16kB) address space, $0000-$3FFF, completely
	// separate from the CPU's address bus. It is either directly accessed by
	// the PPU itself, or via the CPU with memory mapped registers at $2006 and
	// $2007.
	Bus *hwio.Table

	// $3F00-$3F1F	$0020	Palette RAM indexes
	// $3F20-$3FFF	$00E0	Mirrors of $3F00-$3F1F
	Palette            hwio.Mem `hwio:"offset=0x3F00,size=0x20,vsize=0x100,wcb"`
	spriteRam          [0x100]uint8
	secondarySpriteRam [0x20]uint8
	memoryReadBuffer   uint8

	// PPU Registers state
	control           ppuControl
	mask              ppuMask
	statusFlags       ppuStatus
	videoRamAddr      uint16 // 'v'
	tmpVideoRamAddr   uint16 // 't'
	xScroll           uint8  // fine x
	writeToggle       bool
	spriteRamAddr     uint8
	ppuBusAddress     uint16
	openBus           uint8
	openBusDecayStamp [8]int32

	// Rendering state
	FrameCount                     uint32
	renderingEnabled               bool
	prevRenderingEnabled           bool
	lowBitShift                    uint16
	highBitShift                   uint16
	tile                           tileInfo
	currentTilePalette             uint8
	previousTilePalette            uint8
	needStateUpdate                bool
	needVideoRamIncrement          bool
	ignoreVramRead                 int
	updateVramAddrDelay            uint8
	updateVramAddr                 uint16
	preventVblFlag                 bool
	allowFullPpuAccess             bool
	intensifyColorBits             uint16
	paletteRamMask                 uint16
	minimumDrawBgCycle             uint16
	minimumDrawSpriteCycle         uint16
	minimumDrawSpriteStandardCycle uint16

	// Sprite evaluation state
	spriteTiles            [64]spriteInfo // max possible sprites
	hasSprite              [257]bool
	spriteCount            int
	secondaryOamAddr       uint8
	spriteIndex            int
	spriteInRange          bool
	oamCopybuffer          uint8
	oamCopyDone            bool
	spriteAddrH            uint8
	spriteAddrL            uint8
	sprite0Added           bool
	sprite0Visible         bool
	lastSprite             *spriteInfo
	overflowBugCounter     uint8
	firstVisibleSpriteAddr uint32
	lastVisibleSpriteAddr  uint32

	// OAM Corruption
	corruptOamRow [32]bool

	// Region-specific timings
	region                ConsoleRegion
	nmiScanline           int16
	vblankEnd             int16
	standardNmiScanline   int16
	standardVblankEnd     int16
	palSpriteEvalScanline int16

	// Misc
	lastUpdatedPixel int32
}

// NewPPU creates and initializes a new NesPpu instance.
func NewPPU() *PPU {
	p := &PPU{
		Bus: hwio.NewTable("ppu"),
		// Throwaway frame buffer for the first PPU cycles,
		// before one is provided for the frame.
		framebuf: make([]uint32, 256*240),
	}

	hwio.MustInitRegs(p)
	p.Bus.MapBank(0x0000, p, 0)

	p.masterClock = 0
	p.masterClockDivider = 4

	copy(p.Palette.Data[:], []byte{
		0x09, 0x01, 0x00, 0x01, 0x00, 0x02, 0x02, 0x0D, 0x08, 0x10, 0x08, 0x24, 0x00, 0x00, 0x04, 0x2C,
		0x09, 0x01, 0x34, 0x03, 0x00, 0x04, 0x00, 0x14, 0x08, 0x3A, 0x00, 0x02, 0x00, 0x20, 0x2C, 0x08,
	})

	p.videoRamAddr = 0

	p.UpdateTimings(Ntsc)
	p.Reset(false)

	return p
}

func (p *PPU) BeginFrame(framebuf []byte) {
	// Received frame buffer is RGBA8.
	p.framebuf = unsafe.Slice((*uint32)(unsafe.Pointer(&framebuf[0])), len(framebuf)/4)
}

// Reset resets the PPU to a startup or post-reset state.
func (p *PPU) Reset(softReset bool) {
	Log("reset ppu %d\n", boolToUint8(softReset))

	p.masterClock = 0

	p.preventVblFlag = false
	p.needStateUpdate = false
	p.prevRenderingEnabled = false
	p.renderingEnabled = false
	p.ignoreVramRead = 0
	p.openBus = 0
	p.openBusDecayStamp = [8]int32{}
	p.tmpVideoRamAddr = 0
	p.highBitShift = 0
	p.lowBitShift = 0
	p.spriteRamAddr = 0
	p.xScroll = 0
	p.writeToggle = false
	p.control = ppuControl{}
	p.mask = ppuMask{}

	if !softReset {
		p.statusFlags = ppuStatus{}
		p.statusFlags.VerticalBlank = false
	}

	p.tile = tileInfo{}
	p.currentTilePalette = 0
	p.previousTilePalette = 0
	p.ppuBusAddress = 0
	p.intensifyColorBits = 0
	p.paletteRamMask = 0x3F
	p.lastUpdatedPixel = -1
	p.lastSprite = nil
	p.oamCopybuffer = 0
	p.spriteInRange = false
	p.sprite0Added = false
	p.spriteAddrH = 0
	p.spriteAddrL = 0
	p.oamCopyDone = false
	p.hasSprite = [257]bool{}
	p.spriteTiles = [64]spriteInfo{}
	p.spriteCount = 0
	p.secondaryOamAddr = 0
	p.sprite0Visible = false
	p.spriteIndex = 0
	p.Scanline = -1
	p.Cycle = 340
	p.FrameCount = 1
	p.memoryReadBuffer = 0
	p.overflowBugCounter = 0
	p.updateVramAddrDelay = 0
	p.updateVramAddr = 0
	p.firstVisibleSpriteAddr = 0
	p.lastVisibleSpriteAddr = 0
	p.allowFullPpuAccess = false

	p.updateMinimumDrawCycles()
}

func (p *PPU) updateMinimumDrawCycles() {
	switch {
	case !p.mask.BackgroundEnabled:
		p.minimumDrawBgCycle = 300
	case !p.mask.BackgroundMask:
		p.minimumDrawBgCycle = 8
	default:
		p.minimumDrawBgCycle = 0
	}

	switch {
	case !p.mask.SpritesEnabled:
		p.minimumDrawSpriteCycle = 300
		p.minimumDrawSpriteStandardCycle = 300
	case !p.mask.SpriteMask:
		p.minimumDrawSpriteCycle = 8
		p.minimumDrawSpriteStandardCycle = 8
	default:
		p.minimumDrawSpriteCycle = 0
		p.minimumDrawSpriteStandardCycle = 0
	}
}

// UpdateTimings sets PPU timing constants based on the console region.
func (p *PPU) UpdateTimings(region ConsoleRegion) {
	p.region = region

	switch p.region {
	case Ntsc:
		p.nmiScanline = 241
		p.vblankEnd = 260
		p.standardNmiScanline = 241
		p.standardVblankEnd = 260
		p.masterClockDivider = 4
	case Pal:
		p.nmiScanline = 241
		p.vblankEnd = 310
		p.standardNmiScanline = 241
		p.standardVblankEnd = 310
		p.masterClockDivider = 5
	case Dendy:
		p.nmiScanline = 291
		p.vblankEnd = 310
		p.standardNmiScanline = 291
		p.standardVblankEnd = 310
		p.masterClockDivider = 5
	default:
		panic("nes region should be set here")
	case Auto:
		panic("nes region shouldn't be auto here")
	}

	p.palSpriteEvalScanline = p.nmiScanline + 24
}

func (p *PPU) readPalette(addr uint16) uint8 {
	addr &= 0x1F
	if addr == 0x10 || addr == 0x14 || addr == 0x18 || addr == 0x1C {
		addr &^= 0x10
	}
	return p.Palette.Data[addr]
}

func (p *PPU) WritePALETTE(addr uint16, val uint8) {
	val &= 0x3F
	addr &= 0x1F
	switch addr {
	case 0x00, 0x10:
		p.Palette.Data[0x00] = val
		p.Palette.Data[0x10] = val
	case 0x04, 0x14:
		p.Palette.Data[0x04] = val
		p.Palette.Data[0x14] = val
	case 0x08, 0x18:
		p.Palette.Data[0x08] = val
		p.Palette.Data[0x18] = val
	case 0x0C, 0x1C:
		p.Palette.Data[0x0C] = val
		p.Palette.Data[0x1C] = val
	default:
		p.Palette.Data[addr] = val
	}
}

// Peek8 reads a PPU register without causing side-effects.
func (p *PPU) Peek8(addr uint16) uint8 {
	openbusMask := uint8(0xFF)
	var ret uint8

	switch getRegisterID(addr) {
	case PPUSTATUS:
		ret = boolToUint8(p.statusFlags.SpriteOverflow)<<5 | boolToUint8(p.statusFlags.Sprite0Hit)<<6 | boolToUint8(p.statusFlags.VerticalBlank)<<7
		if p.Scanline == int16(p.nmiScanline) && p.Cycle < 3 {
			ret &^= 0x80 // Clear vertical blank flag
		}
		openbusMask = 0x1F

	case OAMDATA:
		if p.Scanline <= 239 && p.isRenderingEnabled() {
			if p.Cycle >= 257 && p.Cycle <= 320 {
				step := (uint8(p.Cycle-257) % 8)
				if step > 3 {
					step = 3
				}
				oamAddr := uint8(p.Cycle-257)/8*4 + step
				ret = p.secondarySpriteRam[oamAddr]
			} else {
				ret = p.oamCopybuffer
			}
		} else {
			ret = p.spriteRam[p.spriteRamAddr]
		}
		openbusMask = 0x00

	case PPUDATA:
		ret = p.memoryReadBuffer
		if (p.videoRamAddr & 0x3FFF) >= 0x3F00 {
			ret = p.readPalette(p.videoRamAddr)&uint8(p.paletteRamMask) | (p.openBus & 0xC0)
			openbusMask = 0xC0
		} else {
			openbusMask = 0x00
		}
	}
	return ret | (p.openBus & openbusMask)
}

// Read8 reads from a PPU register, with side-effects.
func (p *PPU) Read8(addr uint16) uint8 {
	openbusMask := uint8(0xFF)
	var returnValue uint8

	switch getRegisterID(addr) {
	case PPUSTATUS:
		p.writeToggle = false
		returnValue = boolToUint8(p.statusFlags.SpriteOverflow)<<5 | boolToUint8(p.statusFlags.Sprite0Hit)<<6 | boolToUint8(p.statusFlags.VerticalBlank)<<7
		// TODO: C++ line 417: UpdateStatusFlag();
		p.updateStatusFlag()
		openbusMask = 0x1F
	case OAMDATA:
		if p.Scanline <= 239 && p.isRenderingEnabled() {
			if p.Cycle >= 257 && p.Cycle <= 320 {
				step := (uint8(p.Cycle-257) % 8)
				if step > 3 {
					step = 3
				}
				p.secondaryOamAddr = uint8(p.Cycle-257)/8*4 + step
				p.oamCopybuffer = p.secondarySpriteRam[p.secondaryOamAddr]
			}
			returnValue = p.oamCopybuffer
		} else {
			returnValue = p.readSpriteRam(p.spriteRamAddr)
		}
		openbusMask = 0x00

	case PPUDATA:
		if p.ignoreVramRead != 0 {
			// 2 reads to $2007 in quick succession (2 consecutive CPU
			// cycles) causes the 2nd read to be ignored (normally depends
			// on PPU/CPU timing, but this is the simplest solution)
			openbusMask = 0xFF
		} else {
			returnValue = p.memoryReadBuffer
			p.memoryReadBuffer = p.readVram(p.ppuBusAddress & 0x3FFF)

			if (p.ppuBusAddress & 0x3FFF) >= 0x3F00 {
				returnValue = (p.readPalette(p.ppuBusAddress) & uint8(p.paletteRamMask)) | (p.openBus & 0xC0)
				openbusMask = 0xC0
			} else {
				openbusMask = 0x00
			}

			p.ignoreVramRead = 6
			p.needStateUpdate = true
			p.needVideoRamIncrement = true
		}
	}
	return p.applyOpenBus(openbusMask, returnValue)
}

// Write8 writes to a PPU register.
func (p *PPU) Write8(addr uint16, value uint8) {
	if addr != 0x4014 {
		p.setOpenBus(0xFF, value)
	}

	switch getRegisterID(addr) {
	case PPUCTRL:
		p.setControlRegister(value)
	case PPUMASK:
		p.setMaskRegister(value)
	case OAMADDR:
		p.spriteRamAddr = value
	case OAMDATA:
		if (p.Scanline >= 240 && (p.region != Pal || p.Scanline < int16(p.palSpriteEvalScanline))) || !p.isRenderingEnabled() {
			if (p.spriteRamAddr & 0x03) == 0x02 {
				value &= 0xE3
			}
			p.writeSpriteRam(p.spriteRamAddr, value)
			p.spriteRamAddr = (p.spriteRamAddr + 1) & 0xFF
		} else {
			p.spriteRamAddr = (p.spriteRamAddr + 4) & 0xFF
		}
	case PPUSCROLL:
		if p.writeToggle {
			p.tmpVideoRamAddr = (p.tmpVideoRamAddr & ^uint16(0x73E0)) | (uint16(value&0xF8) << 2) | (uint16(value&0x07) << 12)
		} else {
			p.xScroll = value & 0x07
			newAddr := (p.tmpVideoRamAddr & ^uint16(0x001F)) | (uint16(value) >> 3)
			p.tmpVideoRamAddr = newAddr
		}
		p.writeToggle = !p.writeToggle
	case PPUADDR:
		if p.writeToggle {
			p.tmpVideoRamAddr = (p.tmpVideoRamAddr & ^uint16(0x00FF)) | uint16(value)
			p.needStateUpdate = true
			p.updateVramAddrDelay = 3
			p.updateVramAddr = p.tmpVideoRamAddr
		} else {
			newAddr := (p.tmpVideoRamAddr & ^uint16(0xFF00)) | (uint16(value&0x3F) << 8)
			p.tmpVideoRamAddr = newAddr
		}
		p.writeToggle = !p.writeToggle
	case PPUDATA:
		if (p.ppuBusAddress & 0x3FFF) >= 0x3F00 {
			p.WritePALETTE(p.ppuBusAddress, value)
		} else {
			if p.Scanline >= 240 || !p.isRenderingEnabled() {
				p.writeVram(p.ppuBusAddress&0x3FFF, value)
			} else {
				// During rendering, the value written is ignored, and instead the address' LSB is used (not confirmed, based on Visual NES)
				p.writeVram(p.ppuBusAddress&0x3FFF, uint8(p.ppuBusAddress))
			}
		}
		p.needStateUpdate = true
		p.needVideoRamIncrement = true
	}
}

func (p *PPU) writeVram(addr uint16, val uint8) {
	p.Bus.Write8(addr, val)
}

func (p *PPU) Run(until uint64) {
	for {
		// Always need to run at least once, check condition at the end of the
		// loop (slightly faster).
		p.Tick()
		p.masterClock += uint64(p.masterClockDivider)
		if p.masterClock+uint64(p.masterClockDivider) > until {
			break
		}
	}
}

// Tick executes a single PPU cycle.
func (p *PPU) Tick() {
	if p.Cycle < 340 {
		p.Cycle++
		if p.Scanline < 240 {
			p.ProcessScanline()
		} else if p.Cycle == 1 && p.Scanline == p.nmiScanline {
			if !p.preventVblFlag {
				p.statusFlags.VerticalBlank = true
				p.beginVBlank()
			}
			p.preventVblFlag = false
		} else if p.region == Pal && p.Scanline >= p.palSpriteEvalScanline {
			if p.Cycle <= 256 {
				p.processSpriteEvaluation()
			} else if p.Cycle >= 257 && p.Cycle < 320 {
				p.spriteRamAddr = 0
			}
		}
	} else {
		p.processScanlineFirstCycle()
	}

	if p.needStateUpdate {
		p.updateState()
	}
}

func (p *PPU) ProcessScanline() {
	//Only called for cycle 1+
	if p.Cycle <= 256 {
		p.loadTileInfo()

		if p.prevRenderingEnabled && (p.Cycle&0x07) == 0 {
			p.incHorizontalScrolling()
			if p.Cycle == 256 {
				p.incVerticalScrolling()
			}
		}

		if p.Scanline >= 0 {
			p.drawPixel()
			p.shiftTileRegisters()

			// "Secondary OAM clear and sprite evaluation do not occur on the pre-render line".
			p.processSpriteEvaluation()
		} else if p.Cycle < 9 {
			//Pre-render scanline logic
			if p.Cycle == 1 {
				p.statusFlags.VerticalBlank = false
				p.CPU.clearNMIflag()
			}
			if p.spriteRamAddr >= 0x08 && p.isRenderingEnabled() {
				// This should only be done if rendering is enabled (otherwise
				// oam_stress test fails immediately):
				//
				//  If OAMADDR is not less than eight when rendering starts, the
				//  eight bytes starting at OAMADDR & 0xF8 are copied to the
				//  first eight bytes of OAM"
				p.writeSpriteRam(uint8(p.Cycle-1), p.readSpriteRam((p.spriteRamAddr&0xF8)+uint8(p.Cycle)-1))
			}
		}
	} else if p.Cycle >= 257 && p.Cycle <= 320 {
		if p.Cycle == 257 {
			p.spriteIndex = 0
			p.hasSprite = [257]bool{}
			if p.prevRenderingEnabled {
				// Copy horizontal scrolling value from t.
				p.videoRamAddr = (p.videoRamAddr & ^uint16(0x041F)) | (p.tmpVideoRamAddr & 0x041F)
			}
		}
		if p.isRenderingEnabled() {
			//  "OAMADDR is set to 0 during each of ticks 257-320 (the sprite
			//  tile loading interval) of the pre-render and visible scanlines."
			//  (When rendering)
			p.spriteRamAddr = 0

			switch (p.Cycle - 257) % 8 {
			// Garbage NT sprite fetch (257, 265, 273, etc.) - Required for
			// proper MC-ACC IRQs (MMC3 clone)
			case 0:
				p.readVram(p.getNameTableAddr())

			// Garbage AT sprite fetch
			case 2:
				p.readVram(p.getAttributeAddr())

			// Cycle 260, 268, etc.  This is an approximation (each tile
			// is actually loaded in 8 steps (e.g from 257 to 264))
			case 4:
				p.loadSpriteTileInfo()
				break
			}

			if p.Scanline == -1 && p.Cycle >= 280 && p.Cycle <= 304 {
				// copy vertical scrolling value from t
				p.videoRamAddr = (p.videoRamAddr & ^uint16(0x7BE0)) | (p.tmpVideoRamAddr & 0x7BE0)
			}
		}
	} else if p.Cycle >= 321 && p.Cycle <= 336 {
		p.loadTileInfo()

		if p.Cycle == 321 {
			if p.isRenderingEnabled() {
				p.oamCopybuffer = p.secondarySpriteRam[0]
			}
		} else if p.prevRenderingEnabled && (p.Cycle == 328 || p.Cycle == 336) {
			p.lowBitShift <<= 8
			p.highBitShift <<= 8
			p.incHorizontalScrolling()
		}
	} else if p.Cycle == 337 || p.Cycle == 339 {
		if p.isRenderingEnabled() {
			p.tile.TileAddr = uint16(p.readVram(p.getNameTableAddr()))

			if p.Scanline == -1 && p.Cycle == 339 && (p.FrameCount&0x01) != 0 && p.region == Ntsc {
				//This behavior is NTSC-specific - PAL frames are always the same number of cycles
				//"With rendering enabled, each odd PPU frame is one PPU clock shorter than normal" (skip from 339 to 0, going over 340)
				p.Cycle = 340
			}
		}
	}
}

func (p *PPU) shiftTileRegisters() {
	p.lowBitShift <<= 1
	p.highBitShift <<= 1
}

var plog *os.File

func Log(format string, args ...any) {
	if plog == nil {
		var err error
		plog, err = os.Create("/tmp/nestor.log")
		if err != nil {
			panic(err)
		}
	}

	fmt.Fprintf(plog, format+"\n", args...)
}

func (p *PPU) loadSpriteTileInfo() {
	sprite := p.secondarySpriteRam[p.spriteIndex*4:]
	Log("LoadSpriteTileInfo Cycle %d Scanline %d spriteY %d tileIndex %d attributes %d spriteX %d", p.Cycle, p.Scanline, sprite[0], sprite[1], sprite[2], sprite[3])

	p.loadSprite(sprite[0], sprite[1], sprite[2], sprite[3])
}

func (p *PPU) loadSprite(spriteY, tileIndex, attributes, spriteX uint8) {
	backgroundPriority := (attributes & 0x20) == 0x20
	horizontalMirror := (attributes & 0x40) == 0x40
	verticalMirror := (attributes & 0x80) == 0x80

	var tileAddr uint16
	var lineOffset uint8
	if verticalMirror {
		if p.control.LargeSprites {
			lineOffset = 15 - uint8(p.Scanline-int16(spriteY))
		} else {
			lineOffset = 7 - uint8(p.Scanline-int16(spriteY))
		}
	} else {
		lineOffset = uint8(p.Scanline - int16(spriteY))
	}

	if p.control.LargeSprites {
		if (tileIndex & 0x01) != 0 {
			tileAddr = 0x1000
		} else {
			tileAddr = 0x0000
		}
		tileAddr |= uint16(tileIndex & ^uint8(0x01)) << 4
		if lineOffset >= 8 {
			tileAddr += uint16(lineOffset + 8)
		} else {
			tileAddr += uint16(lineOffset)
		}
	} else {
		tileAddr = (uint16(tileIndex)<<4 | p.control.SpritePatternAddr) + uint16(lineOffset)
	}

	fetchLastSprite := true
	if (p.spriteIndex < p.spriteCount) && spriteY < 240 {
		info := &p.spriteTiles[p.spriteIndex]
		info.BackgroundPriority = backgroundPriority
		info.HorizontalMirror = horizontalMirror
		info.PaletteOffset = ((attributes & 0x03) << 2) | 0x10
		fetchLastSprite = false
		info.LowByte = p.readVram(tileAddr)
		info.HighByte = p.readVram(tileAddr + 8)
		info.SpriteX = spriteX

		if p.Scanline >= 0 {
			// Sprites read on prerender scanline are not shown on scanline 0
			for i := 0; i < 8 && int(spriteX)+i+1 < 257; i++ {
				p.hasSprite[int(spriteX)+i+1] = true
			}
		}
	}

	if fetchLastSprite {
		// Fetches to sprite 0xFF for remaining sprites/hidden - used by MMC3 IRQ counter
		lineOffset = 0
		tileIndex = 0xFF
		if p.control.LargeSprites {
			if (tileIndex & 0x01) != 0 {
				tileAddr = 0x1000
			} else {
				tileAddr = 0x0000
			}
			tileAddr |= uint16((tileIndex & ^uint8(0x01))) << 4
			if lineOffset >= 8 {
				tileAddr += uint16(lineOffset + 8)
			} else {
				tileAddr += uint16(lineOffset)
			}
		} else {
			tileAddr = (uint16(tileIndex)<<4 | p.control.SpritePatternAddr) + uint16(lineOffset)
		}

		p.readVram(tileAddr)
		p.readVram(tileAddr + 8)
	}

	p.spriteIndex++
}

// ABGR format. Convenient for little endian since it has the same memory layout
// as RGBA struct.
//
// TODO: should be defined as color.RGBA and generated at either compile time or
// runtime, based on the target architecture.
var nesPalette = [...]uint32{
	0xFF7C7C7C, 0xFFFC0000, 0xFFBC0000, 0xFFBC2844, 0xFF840094, 0xFF2000A8, 0xFF0010A8, 0xFF001488,
	0xFF003050, 0xFF007800, 0xFF006800, 0xFF005800, 0xFF584000, 0xFF000000, 0xFF000000, 0xFF000000,
	0xFFBCBCBC, 0xFFF87800, 0xFFF85800, 0xFFFC4468, 0xFFCC00D8, 0xFF5800E4, 0xFF0038F8, 0xFF105CE4,
	0xFF007CAC, 0xFF00B800, 0xFF00A800, 0xFF44A800, 0xFF888800, 0xFF000000, 0xFF000000, 0xFF000000,
	0xFFF8F8F8, 0xFFFCBC3C, 0xFFFC8868, 0xFFF87898, 0xFFF878F8, 0xFF9858F8, 0xFF5878F8, 0xFF44A0FC,
	0xFF00B8F8, 0xFF18F8B8, 0xFF54D858, 0xFF98F858, 0xFFD8E800, 0xFF787878, 0xFF000000, 0xFF000000,
	0xFFFCFCFC, 0xFFFCE4A4, 0xFFF8B8B8, 0xFFF8B8D8, 0xFFF8B8F8, 0xFFC0A4F8, 0xFFB0D0F0, 0xFFA8E0FC,
	0xFF78D8F8, 0xFF78F8D8, 0xFFB8F8B8, 0xFFD8F8B8, 0xFFFCFC00, 0xFFF8D8F8, 0xFF000000, 0xFF000000,
}

func (p *PPU) drawPixel() {
	var colidx uint8
	// This is called 3.7 million times per second - needs to be as fast as possible.
	if p.isRenderingEnabled() || ((p.videoRamAddr & 0x3F00) != 0x3F00) {
		color := p.pixelColor()
		if color&0x03 == 0 {
			color = 0
		}
		colidx = p.Palette.Data[color]
	} else {
		// "If the current VRAM address points in the range $3F00-$3FFF during
		// forced blanking, the color indicated by this palette location will be
		// shown on screen instead of the backdrop color."
		colidx = p.Palette.Data[p.videoRamAddr&0x1F]
	}

	p.framebuf[(uint32(p.Scanline)<<8)+p.Cycle-1] = nesPalette[colidx]
}

func (p *PPU) pixelColor() uint8 {
	offset := p.xScroll
	backgroundColor := uint8(0)
	spriteBgColor := uint8(0)

	if p.Cycle > uint32(p.minimumDrawBgCycle) {
		// BackgroundMask = false: Hide background in leftmost 8 pixels of screen
		spriteBgColor = uint8((((p.lowBitShift << offset) & 0x8000) >> 15) | (((p.highBitShift << offset) & 0x8000) >> 14))
		backgroundColor = spriteBgColor
	}

	if p.hasSprite[p.Cycle] && p.Cycle > uint32(p.minimumDrawSpriteCycle) {
		// SpriteMask = true: Hide sprites in leftmost 8 pixels of screen
		for i := range p.spriteCount {
			shift := int32(p.Cycle) - int32(p.spriteTiles[i].SpriteX) - 1
			if shift >= 0 && shift < 8 {
				p.lastSprite = &p.spriteTiles[i]
				var spriteColor uint8
				if p.spriteTiles[i].HorizontalMirror {
					spriteColor = ((p.lastSprite.LowByte >> shift) & 0x01) | ((p.lastSprite.HighByte>>shift)&0x01)<<1
				} else {
					spriteColor = ((p.lastSprite.LowByte<<shift)&0x80)>>7 | ((p.lastSprite.HighByte<<shift)&0x80)>>6
				}

				if spriteColor != 0 {
					// First sprite without a 00 color, use it.
					if i == 0 && spriteBgColor != 0 && p.sprite0Visible && p.Cycle != 256 && p.mask.BackgroundEnabled && !p.statusFlags.Sprite0Hit && p.Cycle > uint32(p.minimumDrawSpriteStandardCycle) {
						//  "The hit condition is basically sprite zero is in
						//   range AND the first sprite output unit is outputting
						//   a non-zero pixel AND the background drawing unit is
						//   outputting a non-zero pixel."
						//  "Sprite zero hits do not register at x=255" (cycle 256)
						//  "... provided that background and sprite rendering are both enabled"
						//  "Should always miss when Y >= 239"
						p.statusFlags.Sprite0Hit = true
					}
					if backgroundColor == 0 || !p.spriteTiles[i].BackgroundPriority {
						// Check sprite priority
						return p.lastSprite.PaletteOffset + spriteColor
					}
					break
				}
			}
		}
	}

	if offset+uint8((p.Cycle-1)&0x07) < 8 {
		return p.previousTilePalette + backgroundColor
	}
	return p.currentTilePalette + backgroundColor
}

func (p *PPU) loadTileInfo() {
	if p.isRenderingEnabled() {
		switch p.Cycle & 0x07 {
		case 1:
			p.previousTilePalette = p.currentTilePalette
			p.currentTilePalette = p.tile.PaletteOffset

			p.lowBitShift |= uint16(p.tile.LowByte)
			p.highBitShift |= uint16(p.tile.HighByte)

			tileIndex := p.readVram(p.getNameTableAddr())
			p.tile.TileAddr = (uint16(tileIndex) << 4) | (p.videoRamAddr >> 12) | p.control.BackgroundPatternAddr

		case 3:
			shift := ((p.videoRamAddr >> 4) & 0x04) | (p.videoRamAddr & 0x02)
			p.tile.PaletteOffset = ((p.readVram(p.getAttributeAddr()) >> shift) & 0x03) << 2

		case 5:
			p.tile.LowByte = p.readVram(p.tile.TileAddr)

		case 7:
			p.tile.HighByte = p.readVram(p.tile.TileAddr + 8)
		}
	}
}

// --- Unexported (Private/Protected) Methods ---

func getRegisterID(addr uint16) ppureg {
	if addr == 0x4014 {
		return OAMDMA
	}
	return ppureg(addr & 0x07)
}

func (p *PPU) updateVideoRamAddr() {
	if p.Scanline >= 240 || !p.isRenderingEnabled() {
		increment := uint16(1)
		if p.control.VerticalWrite {
			increment = 32
		}
		p.videoRamAddr = (p.videoRamAddr + increment) & 0x7FFF
		p.setBusAddress(p.videoRamAddr & 0x3FFF)
	} else {
		p.incHorizontalScrolling()
		p.incVerticalScrolling()
	}
}

func (p *PPU) setOpenBus(mask, value uint8) {
	if mask == 0xFF {
		p.openBus = value
		for i := range 8 {
			p.openBusDecayStamp[i] = int32(p.FrameCount)
		}
	} else {
		var openBus uint16 = uint16(p.openBus) << 8
		for i := range 8 {
			openBus >>= 1
			if (mask & 0x01) != 0 {
				if (value & 0x01) != 0 {
					openBus |= 0x80
				} else {
					openBus &^= 0x0080
				}
				p.openBusDecayStamp[i] = int32(p.FrameCount)
			} else if int32(p.FrameCount)-p.openBusDecayStamp[i] > 3 {
				openBus &^= 0x0080
			}
			value >>= 1
			mask >>= 1
		}
		p.openBus = uint8(openBus)
	}
}

func (p *PPU) applyOpenBus(mask, value uint8) uint8 {
	p.setOpenBus(^mask, value)
	return value | (p.openBus & mask)
}

func (p *PPU) setControlRegister(value uint8) {
	nameTable := uint16(value & 0x03)
	normalAddr := (p.tmpVideoRamAddr & ^uint16(0x0C00)) | (nameTable << 10)
	p.tmpVideoRamAddr = normalAddr

	p.control.VerticalWrite = (value & 0x04) == 0x04
	p.control.SpritePatternAddr = 0
	if (value & 0x08) == 0x08 {
		p.control.SpritePatternAddr = 0x1000
	}
	p.control.BackgroundPatternAddr = 0
	if (value & 0x10) == 0x10 {
		p.control.BackgroundPatternAddr = 0x1000
	}
	p.control.LargeSprites = (value & 0x20) == 0x20
	p.control.SecondaryPpu = (value & 0x40) == 0x40
	p.control.NmiOnVerticalBlank = (value & 0x80) == 0x80

	if !p.control.NmiOnVerticalBlank {
		p.CPU.clearNMIflag()
	} else if p.control.NmiOnVerticalBlank && p.statusFlags.VerticalBlank {
		p.CPU.setNMIflag()
	}
}

func (p *PPU) setMaskRegister(value uint8) {
	p.mask.Grayscale = (value & 0x01) == 0x01
	p.mask.BackgroundMask = (value & 0x02) == 0x02
	p.mask.SpriteMask = (value & 0x04) == 0x04
	p.mask.BackgroundEnabled = (value & 0x08) == 0x08
	p.mask.SpritesEnabled = (value & 0x10) == 0x10
	p.mask.IntensifyBlue = (value & 0x80) == 0x80

	if p.region == Ntsc {
		p.mask.IntensifyRed = (value & 0x20) == 0x20
		p.mask.IntensifyGreen = (value & 0x40) == 0x40
	} else if p.region == Pal || p.region == Dendy {
		p.mask.IntensifyRed = (value & 0x40) == 0x40
		p.mask.IntensifyGreen = (value & 0x20) == 0x20
	}

	if p.renderingEnabled != (p.mask.BackgroundEnabled || p.mask.SpritesEnabled) {
		p.needStateUpdate = true
	}

	p.updateMinimumDrawCycles()
	// TODO(arl)
	// p.updateGrayscaleAndIntensifyBits()
}

func (p *PPU) updateStatusFlag() {
	p.statusFlags.VerticalBlank = false
	p.CPU.clearNMIflag()

	if p.Scanline == p.nmiScanline && p.Cycle == 0 {
		p.preventVblFlag = true
	}
}

func (p *PPU) incVerticalScrolling() {
	addr := p.videoRamAddr
	if (addr & 0x7000) != 0x7000 {
		addr += 0x1000
	} else {
		addr &^= 0x7000
		y := (addr & 0x03E0) >> 5
		switch y {
		case 29:
			y = 0
			addr ^= 0x0800
		case 31:
			y = 0
		default:
			y++
		}
		addr = (addr & ^uint16(0x03E0)) | (y << 5)
	}
	p.videoRamAddr = addr
}

func (p *PPU) incHorizontalScrolling() {
	addr := p.videoRamAddr
	if addr&0x001F == 31 {
		addr = (addr & ^uint16(0x001F)) ^ 0x0400
	} else {
		addr++
	}
	p.videoRamAddr = addr
}

func (p *PPU) getNameTableAddr() uint16 {
	return 0x2000 | (p.videoRamAddr & 0x0FFF)
}

func (p *PPU) getAttributeAddr() uint16 {
	return 0x23C0 | (p.videoRamAddr & 0x0C00) | ((p.videoRamAddr >> 4) & 0x38) | ((p.videoRamAddr >> 2) & 0x07)
}

func (p *PPU) setBusAddress(addr uint16) {
	p.ppuBusAddress = addr
}

func (p *PPU) readVram(addr uint16) uint8 {
	p.setBusAddress(addr)
	return p.Bus.Read8(addr)
}

func (p *PPU) readSpriteRam(addr uint8) uint8 {
	return p.spriteRam[addr]
}

func (p *PPU) writeSpriteRam(addr uint8, value uint8) {
	p.spriteRam[addr] = value
}

func (p *PPU) isRenderingEnabled() bool {
	return p.mask.BackgroundEnabled || p.mask.SpritesEnabled
}

func (p *PPU) beginVBlank() { p.triggerNmi() }
func (p *PPU) triggerNmi() {
	if p.control.NmiOnVerticalBlank {
		p.CPU.setNMIflag()
	}
}

const inputScanLine = 241

// ... other unexported methods converted similarly ...
func (p *PPU) processScanlineFirstCycle() {
	p.Cycle = 0
	p.Scanline++
	if p.Scanline > p.vblankEnd {
		p.lastUpdatedPixel = -1
		p.Scanline = -1
		p.spriteCount = 0

		p.updateMinimumDrawCycles()
	}

	p.updateApuStatus()

	if p.Scanline < 240 {
		if p.Scanline == -1 {
			p.statusFlags.SpriteOverflow = false
			p.statusFlags.Sprite0Hit = false
			p.allowFullPpuAccess = true
		} else if p.prevRenderingEnabled {
			if p.Scanline > 0 || (p.FrameCount&0x01 == 0) || p.region != Ntsc {
				p.setBusAddress((p.tile.TileAddr << 4) | (p.videoRamAddr >> 12) | p.control.BackgroundPatternAddr)
			}
		}
	} else if p.Scanline == 240 {
		p.setBusAddress(p.videoRamAddr & 0x3FFF)
		p.FrameCount++
	}
}

func (p *PPU) updateState() {
	p.needStateUpdate = false

	if p.prevRenderingEnabled != p.renderingEnabled {
		p.prevRenderingEnabled = p.renderingEnabled
		if p.Scanline < 240 {
			if !p.prevRenderingEnabled {
				p.setBusAddress(p.videoRamAddr & 0x3FFF)
				if p.Cycle >= 65 && p.Cycle <= 256 {
					p.spriteRamAddr++
					p.spriteAddrH = (p.spriteRamAddr >> 2) & 0x3F
					p.spriteAddrL = p.spriteRamAddr & 0x03
				}
			}
		}
	}

	if p.renderingEnabled != (p.mask.BackgroundEnabled || p.mask.SpritesEnabled) {
		p.renderingEnabled = p.mask.BackgroundEnabled || p.mask.SpritesEnabled
		p.needStateUpdate = true
	}

	if p.updateVramAddrDelay > 0 {
		p.updateVramAddrDelay--
		if p.updateVramAddrDelay == 0 {
			p.videoRamAddr = p.updateVramAddr
			p.tmpVideoRamAddr = p.videoRamAddr
			if p.Scanline >= 240 || !p.isRenderingEnabled() {
				p.setBusAddress(p.videoRamAddr & 0x3FFF)
			}
		} else {
			p.needStateUpdate = true
		}
	}

	if p.ignoreVramRead > 0 {
		p.ignoreVramRead--
		if p.ignoreVramRead > 0 {
			p.needStateUpdate = true
		}
	}

	if p.needVideoRamIncrement {
		p.needVideoRamIncrement = false
		p.updateVideoRamAddr()
	}
}

func (p *PPU) updateApuStatus() {
	p.CPU.APU.SetEnabled(true)
	if p.Scanline <= 240 {
		return
	}

	if p.Scanline > p.standardVblankEnd {
		p.CPU.APU.SetEnabled(false)
	} else if p.Scanline >= p.standardNmiScanline && p.Scanline < p.nmiScanline {
		p.CPU.APU.SetEnabled(false)
	}
}

func (p *PPU) processSpriteEvaluation() {
	if p.isRenderingEnabled() || (p.region == Pal && p.Scanline >= p.palSpriteEvalScanline) {
		if p.Cycle < 65 {
			// Clear secondary OAM at between cycle 1 and 64.
			p.oamCopybuffer = 0xFF
			p.secondarySpriteRam[(p.Cycle-1)>>1] = 0xFF
		} else {
			if p.Cycle%2 != 0 {
				if p.Cycle == 65 {
					p.processSpriteEvaluationStart()
				}
				p.oamCopybuffer = p.readSpriteRam(p.spriteRamAddr)
			} else {
				if p.Cycle == 256 {
					p.processSpriteEvaluationEnd()
				}

				if p.oamCopyDone {
					p.spriteAddrH = (p.spriteAddrH + 1) & 0x3F
					if p.secondaryOamAddr >= 0x20 {
						p.oamCopybuffer = p.secondarySpriteRam[p.secondaryOamAddr&0x1F]
					}
				} else {
					spriteHeight := uint8(8)
					if p.control.LargeSprites {
						spriteHeight = 16
					}

					if !p.spriteInRange && p.Scanline >= int16(p.oamCopybuffer) && p.Scanline < int16(p.oamCopybuffer+spriteHeight) {
						p.spriteInRange = !p.oamCopyDone
					}

					if p.secondaryOamAddr < 0x20 {
						p.secondarySpriteRam[p.secondaryOamAddr] = p.oamCopybuffer
						if p.spriteInRange {
							if p.Cycle == 66 {
								p.sprite0Added = true
							}
							p.spriteAddrL++
							p.secondaryOamAddr++

							if p.spriteAddrL >= 4 {
								p.spriteAddrH = (p.spriteAddrH + 1) & 0x3F
								p.spriteAddrL = 0
								if p.spriteAddrH == 0 {
									p.oamCopyDone = true
								}
							}
							if (p.secondaryOamAddr & 0x03) == 0 {
								p.spriteInRange = false
								p.lastVisibleSpriteAddr = uint32(p.spriteAddrH-1) * 4

								if p.spriteAddrL != 0 {
									spriteHeight := uint8(8)
									if p.control.LargeSprites {
										spriteHeight = 16
									}
									inRange := p.Scanline >= int16(p.oamCopybuffer) && p.Scanline < int16(p.oamCopybuffer+spriteHeight)
									if !inRange {
										p.spriteAddrL = 0
									}
								}
							}
						} else {
							p.spriteAddrH = (p.spriteAddrH + 1) & 0x3F
							p.spriteAddrL = 0
							if p.spriteAddrH == 0 {
								p.oamCopyDone = true
							}
						}
					} else {
						p.oamCopybuffer = p.secondarySpriteRam[p.secondaryOamAddr&0x1F]
						if p.oamCopyDone {
							p.spriteAddrH = (p.spriteAddrH + 1) & 0x3F
							p.spriteAddrL = 0
						} else if p.spriteInRange {
							p.statusFlags.SpriteOverflow = true
							p.spriteAddrL = (p.spriteAddrL + 1)
							if p.spriteAddrL == 4 {
								p.spriteAddrH = (p.spriteAddrH + 1) & 0x3F
								p.spriteAddrL = 0
							}
							if p.overflowBugCounter == 0 {
								p.overflowBugCounter = 3
							} else if p.overflowBugCounter > 0 {
								p.overflowBugCounter--
								if p.overflowBugCounter == 0 {
									p.oamCopyDone = true
									p.spriteAddrL = 0
								}
							}
						} else {
							p.spriteAddrH = (p.spriteAddrH + 1) & 0x3F
							p.spriteAddrL = (p.spriteAddrL + 1) & 0x03
							if p.spriteAddrH == 0 {
								p.oamCopyDone = true
							}
						}
					}
				}
				p.spriteRamAddr = (p.spriteAddrL & 0x03) | (p.spriteAddrH << 2)
			}
		}
	}
}

func (p *PPU) processSpriteEvaluationStart() {
	p.sprite0Added = false
	p.spriteInRange = false
	p.secondaryOamAddr = 0
	p.overflowBugCounter = 0
	p.oamCopyDone = false
	p.spriteAddrH = (p.spriteRamAddr >> 2) & 0x3F
	p.spriteAddrL = p.spriteRamAddr & 0x03
	p.firstVisibleSpriteAddr = uint32(p.spriteAddrH) * 4
	p.lastVisibleSpriteAddr = p.firstVisibleSpriteAddr
}

func (p *PPU) processSpriteEvaluationEnd() {
	p.sprite0Visible = p.sprite0Added
	p.spriteCount = int((p.secondaryOamAddr + 3) >> 2)
}

func (p *PPU) State() *snapshot.PPU {
	state := &snapshot.PPU{}

	copy(state.PaletteRAM[:], p.Palette.Data[:])
	copy(state.SpriteRAM[:], p.spriteRam[:])
	copy(state.SecondarySpriteRAM[:], p.secondarySpriteRam[:])
	copy(state.OpenBusDecayStamp[:], p.openBusDecayStamp[:])

	state.SpriteRAMAddr = p.spriteRamAddr
	state.VRAMAddr = p.videoRamAddr
	state.XScroll = p.xScroll
	state.TmpVRAMAddr = p.tmpVideoRamAddr
	state.WriteToggle = p.writeToggle

	state.HighBitShift = p.highBitShift
	state.LowBitShift = p.lowBitShift

	state.PPUCTRL = snapshot.PPUCTRL(p.control)
	state.PPUMASK = snapshot.PPUMASK(p.mask)
	state.PaletteRAMMask = p.paletteRamMask
	state.IntensifyColorBits = p.intensifyColorBits

	state.PPUSTATUS = snapshot.PPUSTATUS(p.statusFlags)
	state.Scanline = int64(p.Scanline)
	state.Cycle = int64(p.Cycle)
	state.FrameCount = p.FrameCount
	state.MemoryReadBuffer = p.memoryReadBuffer
	state.Region = int(p.region)
	state.PPUBusAddress = p.ppuBusAddress
	state.MasterClock = p.masterClock

	state.CurrentTilePalette = p.currentTilePalette
	state.Tile = snapshot.PPUTileInfo(p.tile)
	state.PreviousTilePalette = p.previousTilePalette

	state.SpriteIndex = p.spriteIndex
	state.SpriteCount = p.spriteCount
	state.SpriteAddrH = p.spriteAddrH
	state.SpriteAddrL = p.spriteAddrL
	state.Sprite0Added = p.sprite0Added
	state.Sprite0Visible = p.sprite0Visible
	state.OAMCopybuffer = p.oamCopybuffer
	state.SecondaryOAMAddr = p.secondaryOamAddr
	state.SpriteInRange = p.spriteInRange
	state.RenderingEnabled = p.renderingEnabled
	state.PrevRenderingEnabled = p.prevRenderingEnabled

	state.OpenBus = p.openBus
	state.IgnoreVRAMRead = p.ignoreVramRead

	state.OAMCopyDone = p.oamCopyDone

	state.NeedStateUpdate = p.needStateUpdate
	state.NeedVideoRAMIncrement = p.needVideoRamIncrement
	state.PreventVblFlag = p.preventVblFlag
	state.OverflowBugCounter = p.overflowBugCounter
	state.UpdateVRAMAddr = p.updateVramAddr
	state.UpdateVRAMAddrDelay = p.updateVramAddrDelay
	state.AllowFullPpuAccess = p.allowFullPpuAccess

	for i := range p.spriteCount {
		state.SpriteTiles[i] = snapshot.PPUSpriteInfo(p.spriteTiles[i])
	}

	return state
}

func (p *PPU) SetState(state *snapshot.PPU) {
	panic("not implemented")
	/*
		if(!s.IsSaving()) {
				UpdateTimings(_region);
				UpdateMinimumDrawCycles();
				UpdateGrayscaleAndIntensifyBits();

				for(int i = 0; i < 0x20; i++) {
					//Set oam decay cycle to the current cycle to ensure it doesn't decay when loading a state
					_oamDecayCycles[i] = _console->GetCpu()->GetCycleCount();
				}

				memset(_corruptOamRow, 0, sizeof(_corruptOamRow));

				for(int i = 0; i < 257; i++) {
					_hasSprite[i] = true;
				}

				_lastUpdatedPixel = -1;

				UpdateApuStatus();
			}
	*/

}

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
