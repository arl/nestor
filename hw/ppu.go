package hw

import (
	"image/color"
	"unsafe"

	"nestor/emu/log"
	"nestor/hw/hwio"
)

const (
	NumScanlines = 262 // Number of scanlines per frame.
	NumCycles    = 341 // Number of PPU cycles per scanline.

	ntscDivider = 4
)

type PPU struct {
	// The PPU addresses a 14-bit (16kB) address space, $0000-$3FFF, completely
	// separate from the CPU's address bus. It is either directly accessed by
	// the PPU itself, or via the CPU with memory mapped registers at $2006 and
	// $2007.
	Bus *hwio.Table
	CPU *CPU

	masterClock uint64
	Cycle       uint32 // Current cycle/pixel in scanline
	Scanline    int    // Current scanline being drawn
	FrameCount  uint32 // Current frame

	// $3F00-$3F1F	$0020	Palette RAM indexes
	// $3F20-$3FFF	$00E0	Mirrors of $3F00-$3F1F
	Palettes hwio.Mem `hwio:"offset=0x3F00,size=0x20,vsize=0x100,wcb"`

	PPUCTRL   ppuctrl
	PPUMASK   ppumask
	PPUSTATUS ppustatus

	oamMem     [0x100]byte
	oamAddr    byte
	oam, oam2  [8]sprite
	ppudataBuf uint8 // only used for PPUDATA reads

	framebuf []uint32 // RGBA framebuffer

	oddFrame      bool
	preventVblank bool

	// VRAM read/write
	vramAddr   loopy
	vramTmp    loopy
	writeLatch bool

	busAddr         uint16
	openBus         uint8
	openBusDecayBuf [8]uint32

	bg bgregs
}

func NewPPU() *PPU {
	p := &PPU{
		Bus: hwio.NewTable("ppu"),
		// Throwaway frame buffer for the first PPU cycles,
		// before one is provided for the frame.
		framebuf: make([]uint32, 256*240),
	}

	hwio.MustInitRegs(p)
	p.Bus.MapBank(0x0000, p, 0)

	// At power up, palette ram is pre-filled. (use Blargg's NES values).
	copy(p.Palettes.Data, []byte{
		0x09, 0x01, 0x00, 0x01, 0x00, 0x02, 0x02, 0x0D,
		0x08, 0x10, 0x08, 0x24, 0x00, 0x00, 0x04, 0x2C,
		0x09, 0x01, 0x34, 0x03, 0x00, 0x04, 0x00, 0x14,
		0x08, 0x3A, 0x00, 0x02, 0x00, 0x20, 0x2C, 0x08,
	})

	return p
}

func (p *PPU) SetFrameBuffer(framebuf []byte) {
	// We're using a RGBA8 framebuffer.
	p.framebuf = unsafe.Slice((*uint32)(unsafe.Pointer(&framebuf[0])), len(framebuf)/4)
}

func (p *PPU) Reset() {
	p.Scanline = -1
	p.FrameCount = 1
	p.Cycle = 339
	p.writeLatch = false
	p.vramAddr = 0
	p.PPUCTRL = 0
	p.PPUMASK = 0
	p.PPUSTATUS = 0
	p.oddFrame = false
	p.preventVblank = false

	for i := range p.oamMem {
		p.oamMem[i] = 0x00
	}

	p.openBus = 0
	p.openBusDecayBuf = [8]uint32{0, 0, 0, 0, 0, 0, 0, 0}
}

func (p *PPU) Run(until uint64) {
	for {
		p.Tick()
		p.masterClock += ntscDivider
		if p.masterClock+ntscDivider > until {
			break
		}
	}
}

func (p *PPU) Tick() {
	switch {
	case p.Scanline < 240:
		p.doScanline(renderMode)
	case p.Scanline == 240:
		p.doScanline(postRender)
	case p.Scanline == 241:
		p.doScanline(vblankNMI)
	case p.Scanline == 261:
		p.doScanline(preRender)
	}

	p.Cycle++
	if p.Cycle >= NumCycles {
		p.Cycle %= NumCycles
		p.Scanline++
		if p.Scanline >= NumScanlines {
			p.Scanline = 0
			p.oddFrame = !p.oddFrame
		}
	}
}

type scanlineMode int

const (
	preRender scanlineMode = iota
	renderMode
	postRender
	vblankNMI
)

func (p *PPU) doScanline(sm scanlineMode) {
	switch sm {
	case vblankNMI:
		if p.Cycle == 1 {
			if !p.preventVblank {
				p.PPUSTATUS.setVblank(true)

				if p.PPUCTRL.nmi() {
					p.CPU.setNMIflag()
					log.ModPPU.DebugZ("Set NMI flag").String("src", "vblank").End()
				}
			}
			p.preventVblank = false
		}
	case postRender:
		switch p.Cycle {
		case 1:
			// At the start of vblank, the bus address is set back
			// to VRAM address.
			p.busAddr = p.vramAddr.addr()
		case 340:
			p.FrameCount++
		}

	case preRender, renderMode:
		switch p.Cycle {
		case 1:
			p.clearOAM()
			if sm == preRender {
				p.PPUSTATUS.setSpriteHit(false)
				p.PPUSTATUS.setSpriteOverflow(false)
			}
		case 257:
			p.evalSprites()
		case 321:
			p.loadSprites()
		}

		switch {
		case p.Cycle >= 2 && p.Cycle <= 255,
			p.Cycle >= 322 && p.Cycle <= 337:
			p.renderPixel()

			switch p.Cycle & 0b111 {

			// nametable
			case 1:
				p.bg.addrLatch = p.ntAddr()
				p.refillShifters()
			case 2:
				p.bg.nt = p.ReadVRAM(p.bg.addrLatch)

			// attribute table
			case 3:
				p.bg.addrLatch = p.atAddr()
			case 4:
				p.bg.at = p.ReadVRAM(p.bg.addrLatch)
				if p.vramAddr.coarsey()&2 != 0 {
					p.bg.at >>= 4
				}
				if p.vramAddr.coarsex()&2 != 0 {
					p.bg.at >>= 2
				}

			// low background byte
			case 5:
				p.bg.addrLatch = p.bgAddr()
			case 6:
				p.bg.bglo = p.ReadVRAM(p.bg.addrLatch)

			// high background byte
			case 7:
				p.bg.addrLatch += 8
			case 0:
				p.bg.bghi = p.ReadVRAM(p.bg.addrLatch)
				p.horzScroll()
			}

		case p.Cycle == 256:
			p.renderPixel()
			p.bg.bghi = p.ReadVRAM(p.bg.addrLatch)
			p.vertScroll()
		case p.Cycle == 257:
			p.renderPixel()
			p.refillShifters()
			p.horzUpdate()
		case p.Cycle >= 280 && p.Cycle <= 304:
			if sm == preRender {
				p.vertUpdate()
			}

		// shifters aren't refilled
		case p.Cycle == 1:
			p.bg.addrLatch = p.ntAddr()
			if sm == preRender {
				p.PPUSTATUS.setVblank(false)
				p.CPU.clearNMIflag()

				if p.oamAddr >= 0x08 && p.isRenderingEnabled() {
					// This should only be done if rendering is enabled
					// (otherwise oam_stress test fails immediately).
					//
					// If OAMADDR is not less than eight when rendering starts,
					// the eight bytes starting at OAMADDR & 0xF8 are copied to
					// the first eight bytes of OAM".
					p.writeOAM(p.readOAM())
				}
			}
		case p.Cycle == 321, p.Cycle == 339:
			p.bg.addrLatch = p.ntAddr()

		// 'garbage' fetches
		case p.Cycle == 338:
			p.bg.nt = p.ReadVRAM(p.bg.addrLatch)
		case p.Cycle == 340:
			p.bg.nt = p.ReadVRAM(p.bg.addrLatch)
			if sm == preRender && p.isRenderingEnabled() && p.oddFrame {
				p.Cycle++
			}
		}
	}
}

// background registers
type bgregs struct {
	// temporary address latch storing the address for next cycle,
	// since fetches takes 2 cycles.
	addrLatch uint16

	// 3-bit 'fine x scroll' register.
	finex uint8

	// latches for background rendering.
	nt, at, bglo, bghi uint8

	// shift registers/latches.
	bgShiftlo, bgShifthi uint16
	atShiftlo, atShifthi uint8
	atLatchlo, atLatchhi bool
}

func (p *PPU) ntAddr() uint16 {
	return 0x2000 | p.vramAddr.val()&0xfff
}

func (p *PPU) atAddr() uint16 {
	return 0x23C0 | p.vramAddr.nametable()<<10 | p.vramAddr.coarsey()/4<<3 | p.vramAddr.coarsex()/4
}

func (p *PPU) bgAddr() uint16 {
	return p.PPUCTRL.bgTable()*0x1000 + (uint16(p.bg.nt) * 16) + p.vramAddr.finey()
}

func (p *PPU) refillShifters() {
	p.bg.bgShiftlo = (p.bg.bgShiftlo & 0xFF00) | uint16(p.bg.bglo)
	p.bg.bgShifthi = (p.bg.bgShifthi & 0xFF00) | uint16(p.bg.bghi)

	p.bg.atLatchlo = u8tob(p.bg.at & 1)
	p.bg.atLatchhi = u8tob(p.bg.at & 2)
}

func (p *PPU) horzScroll() {
	if !p.isRenderingEnabled() {
		return
	}
	if p.vramAddr.coarsex() == 31 {
		p.vramAddr.setVal(p.vramAddr.val() ^ 0x41F)
	} else {
		p.vramAddr.setCoarsex(p.vramAddr.coarsex() + 1)
	}
}

func (p *PPU) vertScroll() {
	if !p.isRenderingEnabled() {
		return
	}
	if finey := p.vramAddr.finey(); finey < 7 {
		p.vramAddr.setFiney(finey + 1)
		return
	}

	p.vramAddr.setFiney(0)
	if p.vramAddr.coarsey() == 31 {
		p.vramAddr.setCoarsey(0)
	} else if p.vramAddr.coarsey() == 29 {
		p.vramAddr.setCoarsey(0)
		p.vramAddr.setNametable(p.vramAddr.nametable() ^ 0b10)
	} else {
		p.vramAddr.setCoarsey(p.vramAddr.coarsey() + 1)
	}
}

func (p *PPU) horzUpdate() {
	if !p.isRenderingEnabled() {
		return
	}
	p.vramAddr.setVal((p.vramAddr.val() & ^uint16(0x041F)) | (p.vramTmp.val() & 0x041F))
}

func (p *PPU) vertUpdate() {
	if !p.isRenderingEnabled() {
		return
	}
	p.vramAddr.setVal((p.vramAddr.val() & ^uint16(0x7BE0)) | (p.vramTmp.val() & 0x7BE0))
}

func (p *PPU) isRenderingEnabled() bool {
	return p.PPUMASK.bg() || p.PPUMASK.sprites()
}

func (p *PPU) spriteHeight() int {
	if p.PPUCTRL.spriteSize() {
		return 16
	}
	return 8
}

func (p *PPU) renderPixel() {
	var palette uint8
	var objPalette uint8
	objPriority := false
	var x = int(p.Cycle) - 2

	// PPUMASK value determines if all pixels are rendered, or if we skip the
	// first 8, or if we skip all of them. We do this for background and sprites
	// separately.
	if p.Scanline < 240 && x >= 0 && x < 256 {
		// Background.
		if p.PPUMASK.bg() && (p.PPUMASK.bgLeft() || x >= 8) {
			palette = uint8(nthbit16(p.bg.bgShifthi, 15-p.bg.finex)<<1 |
				nthbit16(p.bg.bgShiftlo, 15-p.bg.finex))
			if palette != 0 {
				palette |= (nthbit8(p.bg.atShifthi, 7-p.bg.finex)<<1 |
					nthbit8(p.bg.atShiftlo, 7-p.bg.finex)) << 2
			}
		}

		// Sprites
		if p.PPUMASK.sprites() && (p.PPUMASK.spriteLeft() || x >= 8) {
			for i := 7; i >= 0; i-- {
				if p.oam[i].id == 64 {
					continue // Void entry.
				}
				sprX := x - int(p.oam[i].x)
				if sprX >= 8 || sprX < 0 {
					continue // Not in range.
				}
				if p.oam[i].attr&0x40 != 0 {
					sprX ^= 7 // Horizontal flip.
				}

				sprPalette := (nthbit8(p.oam[i].dataH, uint8(7-sprX)) << 1) |
					nthbit8(p.oam[i].dataL, uint8(7-sprX))
				if sprPalette == 0 {
					continue // Transparent pixel.
				}

				if p.oam[i].id == 0 && palette != 0 && x != 255 {
					p.PPUSTATUS.setSpriteHit(true)
				}
				sprPalette |= (p.oam[i].attr & 3) << 2
				objPalette = sprPalette + 16
				objPriority = (p.oam[i].attr & 0x20) != 0
			}
		}

		// Sprites priority
		if objPalette != 0 && (palette == 0 || !objPriority) {
			palette = objPalette
		}

		var paddr uint16
		if p.isRenderingEnabled() {
			paddr += uint16(palette)
		}
		pidx := p.ReadVRAM(0x3F00 + paddr)
		colu32 := nesPalette[pidx]

		// TODO: emphasis not tested yet.
		// const m = 0x80 | 0x40 | 0x20
		// colu32 = emphasis(byte(mask&m), colu32)
		p.framebuf[p.Scanline*256+x] = colu32
	}

	// Perform background shifts.
	p.bg.bgShiftlo <<= 1
	p.bg.bgShifthi <<= 1
	p.bg.atShiftlo = (p.bg.atShiftlo << 1) | b2u8(p.bg.atLatchlo)
	p.bg.atShifthi = (p.bg.atShifthi << 1) | b2u8(p.bg.atLatchhi)
}

//lint:ignore U1000 not supporting emphasis yet so unused for now.
func colorToU32(col color.RGBA) uint32 {
	// little-endian.
	return uint32(col.R)<<24 | uint32(col.G)<<16 | uint32(col.B)<<8 | 0xff
}

// TODO: use LUT or a faster way.
// Test it with game/rom that support color emphasis.
//
//lint:ignore U1000 not supporting emphasis yet so unused for now.
func emphasis(rgbmask byte, abgr uint32) uint32 {
	r := float64(abgr & 0xFF)
	g := float64((0xFF00 & abgr) >> 8)
	b := float64((0xFF0000 & abgr) >> 16)

	switch {
	case rgbmask&0x20 != 0:
		r *= 1.3
		g *= 0.8
		b *= 0.8
	case rgbmask&0x40 != 0:
		r *= 0.8
		g *= 1.3
		b *= 0.8
	case rgbmask&0x80 != 0:
		r *= 0.8
		g *= 0.8
		b *= 1.3
	}

	if r > 255 {
		r = 255
	}
	if r < 0 {
		r = 0
	}
	if g > 255 {
		g = 255
	}
	if g < 0 {
		g = 0
	}
	if b > 255 {
		b = 255
	}
	if b < 0 {
		b = 0
	}

	return uint32(r) | uint32(g)<<8 | uint32(b)<<16 | (0xFF << 24)
}

type ppureg uint16

const (
	PPUCTRL   = 0x00
	PPUMASK   = 0x01
	PPUSTATUS = 0x02
	OAMADDR   = 0x03
	OAMDATA   = 0x04
	PPUSCROLL = 0x05
	PPUADDR   = 0x06
	PPUDATA   = 0x07
	OAMDMA    = 0x4014
)

func ppuregFromAddr(addr uint16) ppureg {
	if addr == OAMDMA {
		return OAMDATA
	}
	return ppureg(addr & 0x07)
}

func (p *PPU) Peek8(addr uint16) uint8 {
	openBusMask := uint8(0xFF)
	ret := uint8(0)

	switch ppuregFromAddr(addr) {
	case PPUSTATUS:
		var tmp ppustatus
		tmp.setSpriteOverflow(p.PPUSTATUS.spriteOverflow())
		tmp.setSpriteHit(p.PPUSTATUS.spriteHit())
		tmp.setVblank(p.PPUSTATUS.vblank())
		if p.Scanline == 241 && p.Cycle < 4 {
			tmp.setVblank(false)
		}
		openBusMask = 0x1F
		ret = uint8(tmp)
	case OAMDATA:
		ret = p.oamMem[p.oamAddr]
		openBusMask = 0x00
	case PPUDATA:
		ret = p.ppudataBuf
		if (p.vramAddr.addr()) >= 0x3F00 {
			ret = (p.readPalette(p.vramAddr.addr()) & 0x3F) | (p.openBus & 0xC0)
			openBusMask = 0xC0
		} else {
			openBusMask = 0x00
		}
	}
	return ret | (p.openBus & openBusMask)
}

func (p *PPU) Read8(addr uint16) uint8 {
	switch ppuregFromAddr(addr) {
	case PPUSTATUS:
		return p.ReadPPUSTATUS()
	case OAMDATA:
		return p.ReadOAMDATA()
	case PPUDATA:
		return p.ReadPPUDATA()
	}
	const openBusMask = 0xFF
	return p.applyOpenBus(openBusMask, 0)
}

func (p *PPU) Write8(addr uint16, val uint8) {
	p.setOpenBus(0xFF, val)

	switch ppuregFromAddr(addr) {
	case PPUCTRL:
		p.WritePPUCTRL(val)
	case PPUMASK:
		p.WritePPUMASK(val)
	case OAMADDR:
		p.WriteOAMADDR(val)
	case OAMDATA:
		p.WriteOAMDATA(val)
	case PPUSCROLL:
		p.WritePPUSCROLL(val)
	case PPUADDR:
		p.WritePPUADDR(val)
	case PPUDATA:
		p.WritePPUDATA(val)
	}
}

// PPUCTRL: $2000
func (p *PPU) WritePPUCTRL(val uint8) {
	p.PPUCTRL = ppuctrl(val)
	log.ModPPU.DebugZ("Write to PPUCTRL").Hex8("val", val).End()

	// Transfer the nametable bits.
	p.vramTmp.setNametable(p.PPUCTRL.nametable())

	// By toggling the nmi bit (bit 7 of PPUCTRL) during vblank without reading
	// PPUSTATUS, a program can cause /nmi to be pulled low multiple times,
	// causing multiple NMIs to be generated.
	if !p.PPUCTRL.nmi() {
		p.CPU.clearNMIflag()
	} else if p.PPUSTATUS.vblank() {
		p.CPU.setNMIflag()
		log.ModPPU.DebugZ("Set NMI flag").String("src", "PPUCTRL").End()
	}
}

// PPUMASK: $2001
func (p *PPU) WritePPUMASK(val uint8) {
	p.PPUMASK = ppumask(val)
	log.ModPPU.DebugZ("Write to PPUMASK").Hex8("val", val).End()
}

// PPUSTATUS: $2002
func (p *PPU) PeekPPUSTATUS() uint8 {
	const openBusMask = 0x1F

	var ret ppustatus
	// TODO: optimize to a single OR-operation
	ret.setSpriteOverflow(p.PPUSTATUS.spriteOverflow())
	ret.setSpriteHit(p.PPUSTATUS.spriteHit())
	ret.setVblank(p.PPUSTATUS.vblank())

	if p.Scanline == 241 && p.Cycle < 3 {
		ret.setVblank(false)
	}

	return uint8(ret) | (p.openBus & openBusMask)
}

func (p *PPU) ReadPPUSTATUS() uint8 {
	p.writeLatch = false
	var ret ppustatus
	// TODO: optimize to a single OR-operation
	ret.setSpriteOverflow(p.PPUSTATUS.spriteOverflow())
	ret.setSpriteHit(p.PPUSTATUS.spriteHit())
	ret.setVblank(p.PPUSTATUS.vblank())

	p.PPUSTATUS.setVblank(false)
	p.CPU.clearNMIflag()

	if p.Scanline == 241 && p.Cycle == 1 {
		// From https://www.nesdev.org/wiki/PPU_registers#PPUSTATUS (notes):
		// Race Condition Warning: Reading PPUSTATUS within two cycles of the
		// start of vertical blank will return 0 in bit 7 but clear the latch
		// anyway, causing NMI to not occur that frame.
		p.preventVblank = true
	}
	const openBusMask = 0x1F
	return p.applyOpenBus(openBusMask, uint8(ret))
}

// PPUSCROLL: $2005
func (p *PPU) WritePPUSCROLL(val uint8) {
	log.ModPPU.DebugZ("Write to PPUSCROLL").Hex8("val", val).End()

	if !p.writeLatch { // first write
		p.bg.finex = val & 0b111
		p.vramTmp.setCoarsex(uint16(val) >> 3)
	} else { // second write
		p.vramTmp.setFiney(uint16(val))
		p.vramTmp.setCoarsey(uint16(val) >> 3)
	}

	p.writeLatch = !p.writeLatch
}

// PPUADDR: $2006
func (p *PPU) WritePPUADDR(val uint8) {
	if !p.writeLatch { //first write
		p.vramTmp.setHigh(val & 0x3f)
	} else { // second write
		p.vramTmp.setLow(val)
		p.vramAddr.setVal(uint16(p.vramTmp))
	}

	p.writeLatch = !p.writeLatch
}

// PPUDATA: $2007
func (p *PPU) PeekPPUDATA(_ uint8) uint8 {
	return p.ppudataBuf
}

func (p *PPU) ReadPPUDATA() uint8 {
	openBusMask := uint8(0xC0)

	// Reading VRAM is too slow so the actual data
	// will be returned at the next read.
	val := p.ppudataBuf
	p.ppudataBuf = p.ReadVRAM(p.vramAddr.addr())

	if p.busAddr&0x3FFF >= 0x3F00 {
		// This is a palette read, they're immediate but they still overwrite
		// the read buffer, on which we apply mirroring (ignor bit 12 of the
		// vram address). (passes Blargg's vram_access test)
		val = (p.readPalette(p.busAddr) & 0x3F) | p.openBus&0xC0
		const mask uint16 = 1 << 12
		p.ppudataBuf = p.Bus.Read8(p.busAddr & ^mask)
	} else {
		openBusMask = 0x00
	}

	p.vramIncr()
	log.ModPPU.DebugZ("VRAM read").
		Hex16("addr", p.vramAddr.addr()).
		Hex8("val", val).
		End()
	return p.applyOpenBus(openBusMask, val)
}

// PPUDATA: $2007
func (p *PPU) WritePPUDATA(val uint8) {
	// TODO: check if this should change the bus addr or not?
	p.WriteVRAM(p.vramAddr.addr(), val)
	p.vramIncr()

	log.ModPPU.DebugZ("VRAM write").
		Hex16("addr", p.vramAddr.addr()).
		Hex8("val", val).
		End()
}

func (p *PPU) vramIncr() {
	var incr uint16 = 1
	if p.PPUCTRL.incr() {
		incr = 32 // Increment by 32 if increment mode is set.
	}
	p.vramAddr.setAddr(uint16(p.vramAddr.addr()) + incr)
}

func (p *PPU) ReadVRAM(addr uint16) uint8 {
	p.busAddr = addr
	return p.Bus.Read8(addr)
}

func (p *PPU) WriteVRAM(addr uint16, val uint8) {
	p.busAddr = addr
	p.Bus.Write8(addr, val)
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

func (p *PPU) WritePALETTES(addr uint16, val uint8) {
	val &= 0x3F
	addr &= 0x1F
	switch addr {
	case 0x00, 0x10:
		p.Palettes.Data[0x00] = val
		p.Palettes.Data[0x10] = val
	case 0x04, 0x14:
		p.Palettes.Data[0x04] = val
		p.Palettes.Data[0x14] = val
	case 0x08, 0x18:
		p.Palettes.Data[0x08] = val
		p.Palettes.Data[0x18] = val
	case 0x0C, 0x1C:
		p.Palettes.Data[0x0C] = val
		p.Palettes.Data[0x1C] = val
	default:
		p.Palettes.Data[addr] = val
	}
}

func (p *PPU) readPalette(addr uint16) uint8 {
	addr &= 0x1F
	if addr == 0x10 || addr == 0x14 || addr == 0x18 || addr == 0x1C {
		addr &^= 0x10
	}
	return p.Palettes.Data[addr]
}

// OAM

// OAMADDR: $2003
func (p *PPU) WriteOAMADDR(val uint8) {
	log.ModPPU.DebugZ("Write to OAMADDR").Hex8("val", val).End()
	p.oamAddr = val
}

// OAMDATA: $2004
func (p *PPU) PeekOAMDATA() uint8 {
	return p.oamMem[p.oamAddr]
}

func (p *PPU) ReadOAMDATA() uint8 {
	val := p.oamMem[p.oamAddr]
	log.ModPPU.DebugZ("Read from OAMDATA").Hex8("val", val).End()

	const openBusMask = 0x00
	return p.applyOpenBus(openBusMask, val)
}

// OAMDATA: $2004
func (p *PPU) WriteOAMDATA(val uint8) {
	log.ModPPU.DebugZ("Write to OAMDATA").Hex8("val", val).End()
	if (p.Scanline >= 240 && p.Scanline < 241+24) || !p.isRenderingEnabled() {
		if (p.oamAddr & 0x03) == 0x02 {
			// The three unimplemented bits of each sprite's byte 2 do not exist
			// in the PPU and always read back as 0 on PPU revisions that allow
			// reading PPU OAM through OAMDATA ($2004)
			val &= 0xE3
		}
		p.writeOAM(val)
	} else {
		// Writes to OAMDATA during rendering (on the pre-render line and the
		// visible lines 0-239, provided either sprite or background rendering
		// is enabled) do not modify values in OAM, but do perform a glitchy
		// increment of OAMADDR, bumping only the high 6 bits"
		p.oamAddr += 4
	}
	p.writeOAM(val)
	p.oamAddr++
}

func (p *PPU) writeOAM(val uint8) {
	p.oamMem[p.oamAddr] = val
}

func (p *PPU) readOAM() uint8 {
	return p.oamMem[p.oamAddr]
}

type sprite struct {
	id    uint8 // index in OAM
	x     uint8
	y     uint8
	tile  uint8 // tile index
	attr  uint8
	dataL uint8
	dataH uint8
}

func (p *PPU) clearOAM() {
	for i := 0; i < 8; i++ {
		p.oam2[i].id = 64
		p.oam2[i].y = 0xFF
		p.oam2[i].tile = 0xFF
		p.oam2[i].attr = 0xFF
		p.oam2[i].x = 0xFF
		p.oam2[i].dataL = 0
		p.oam2[i].dataH = 0
	}
}

// Prepare sprites info in secondary OAM for next scanline
func (p *PPU) evalSprites() {
	if !p.isRenderingEnabled() {
		return
	}
	n := 0
	for i := 0; i < 64; i++ {
		line := p.Scanline
		if p.Scanline == 261 {
			line = -1
		}
		line -= int(p.oamMem[i*4+0])

		// If the sprite is in the scanline, copy its properties into secondary OAM
		if line >= 0 && line < p.spriteHeight() {
			p.oam2[n].id = uint8(i)
			p.oam2[n].y = p.oamMem[i*4+0]
			p.oam2[n].tile = p.oamMem[i*4+1]
			p.oam2[n].attr = p.oamMem[i*4+2]
			p.oam2[n].x = p.oamMem[i*4+3]

			n++
			if n >= 8 {
				p.PPUSTATUS.setSpriteOverflow(true)
				break
			}
		}
	}
}

// Load sprite info into OAM and fetch their tile data
func (p *PPU) loadSprites() {
	var addr uint16
	for i := 0; i < 8; i++ {
		p.oam[i] = p.oam2[i] // Copy secondary OAM into primary.

		// Different address modes depending on the sprite height:
		if p.spriteHeight() == 16 {
			addr = ((uint16(p.oam[i].tile) & 1) * 0x1000) + ((uint16(p.oam[i].tile) & ^uint16(1)) * 16)
		} else {
			addr = (b2u16(p.PPUCTRL.spriteTable()) * 0x1000) + (uint16(p.oam[i].tile) * 16)
		}

		if p.Scanline < 0 {
			panic("unexpected")
		}

		sprY := (p.Scanline - int(p.oam[i].y)) % p.spriteHeight() // Line inside the sprite.
		if p.oam[i].attr&0x80 != 0 {
			sprY ^= p.spriteHeight() - 1 // Vertical flip.
		}
		addr += uint16(sprY + (sprY & 8)) // Select the second tile if on 8x16.

		p.oam[i].dataL = p.Bus.Read8(addr)
		p.oam[i].dataH = p.Bus.Read8(addr + 8)
	}
}

func (p *PPU) setOpenBus(mask uint8, val uint8) {
	// Decay expired bits, set new bits and update stamps on each individual bit.
	if mask == 0xFF {
		// Shortcut when mask is 0xFF - all bits are set to the value and stamps
		// updated.
		p.openBus = val
		for i := range p.openBusDecayBuf {
			p.openBusDecayBuf[i] = p.FrameCount
		}
	} else {
		openBus := (uint16(p.openBus) << 8)
		for i := range 8 {
			openBus >>= 1
			if mask&0x01 != 0 {
				if val&0x01 != 0 {
					openBus |= 0x80
				} else {
					openBus &= 0xFF7F
				}
				p.openBusDecayBuf[i] = p.FrameCount
			} else if p.FrameCount-p.openBusDecayBuf[i] > 3 {
				// Decay bits to 0 after 3 frames. This is a very conservative
				// estimate, individual bits tend to decay much faster than
				// this.
				openBus &= 0xFF7F
			}
			val >>= 1
			mask >>= 1
		}

		p.openBus = uint8(openBus)
	}
}

func (p *PPU) applyOpenBus(mask uint8, val uint8) uint8 {
	p.setOpenBus(^mask, val)
	return val | (p.openBus & mask)
}
