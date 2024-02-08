package hw

import (
	"fmt"
	"image"
	"image/color"

	"nestor/emu/hwio"
	"nestor/emu/log"
)

const (
	NumScanlines = 262 // Number of scanlines per frame.
	NumCycles    = 341 // Number of PPU cycles per scanline.
)

type PPU struct {
	Bus *hwio.Table // PPU bus
	CPU *CPU

	Cycle    int // Current cycle/pixel in scanline
	Scanline int // Current scanline being drawn

	//	$0000-$0FFF	$1000	Pattern table 0
	//	$1000-$1FFF	$1000	Pattern table 1
	PatternTables hwio.Mem `hwio:"offset=0x0000,size=0x2000,wcb"`

	// $2000-$23FF	$0400	Nametable 0
	// $2400-$27FF	$0400	Nametable 1
	// $2800-$2BFF	$0400	Nametable 2
	// $2C00-$2FFF	$0400	Nametable 3
	// $3000-$3EFF	$0F00	Mirrors of $2000-$2EFF
	NameTables hwio.Mem `hwio:"offset=0x2000,size=0x1000,vsize=0x1F00,wcb"`

	// $3F00-$3F1F	$0020	Palette RAM indexes
	// $3F20-$3FFF	$00E0	Mirrors of $3F00-$3F1F
	Palettes hwio.Mem `hwio:"offset=0x3F00,size=0x20,vsize=0x100,wcb"`

	// CPU-exposed memory-mapped PPU registers
	// mapped from $2000 to $2007, mirrored up to $3fff
	PPUCTRL   hwio.Reg8 `hwio:"bank=1,offset=0x0,writeonly,wcb"`
	PPUMASK   hwio.Reg8 `hwio:"bank=1,offset=0x1,writeonly,wcb"`
	PPUSTATUS hwio.Reg8 `hwio:"bank=1,offset=0x2,readonly,rcb"`
	// OAMADDR   hwio.Reg8 `hwio:"bank=1,offset=0x3,writeonly,wcb"`
	// OAMDATA   hwio.Reg8 `hwio:"bank=1,offset=0x4"`
	PPUSCROLL hwio.Reg8 `hwio:"bank=1,offset=0x5,writeonly,wcb"`
	PPUADDR   hwio.Reg8 `hwio:"bank=1,offset=0x6,writeonly,wcb"`
	PPUDATA   hwio.Reg8 `hwio:"bank=1,offset=0x7,rcb,wcb,"`

	// OAMDMA hwio.Reg8 `hwio:"bank=2,writeonly,wcb"`

	screen image.RGBA

	// VRAM read/write
	vramAddr   loopy
	vramTmp    loopy
	writeLatch bool

	ppuDataRbuf uint8 // only used for PPUDATA reads

	bg bgregs
}

func NewPPU() *PPU {
	return &PPU{
		Bus:    hwio.NewTable("ppu"),
		screen: *image.NewRGBA(image.Rect(0, 0, 256, 224)),
	}
}

func (p *PPU) Output() *image.RGBA {
	return &p.screen
}

func (p *PPU) InitBus() {
	hwio.MustInitRegs(p)
	p.Bus.MapBank(0x0000, p, 0)
}

func (p *PPU) Reset() {
	// TODO
	p.Scanline = 0
	p.Cycle = 0
	p.writeLatch = false
	p.vramAddr = 0
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
		p.Cycle -= NumCycles
		p.Scanline++
		if p.Scanline >= NumScanlines {
			p.Scanline = 0
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
	case preRender, renderMode:
		if p.Cycle == 1 {
			if sm == preRender {
				// Clear vblank, sprite0Hit and spriteOverflow
				ppustatus := ppustatus(p.PPUSTATUS.Value)
				ppustatus.setVblank(false)
				ppustatus.setSpriteHit(false)
				ppustatus.setSpriteOverflow(false)
				p.PPUSTATUS.Value = uint8(ppustatus)
			}
		}

		if p.Cycle >= 1 && p.Cycle <= 256 || p.Cycle >= 321 && p.Cycle <= 336 {
			p.renderPixel()
			switch p.Cycle & 0b111 {
			// nametable
			case 1:
				p.bg.addrLatch = p.ntAddr()
				p.refillShifters()
			case 2:
				p.bg.nt = p.Bus.Read8(p.bg.addrLatch)

			// attribute table
			case 3:
				p.bg.addrLatch = p.atAddr()
			case 4:
				p.bg.at = p.Bus.Read8(p.bg.addrLatch)

			// low background byte
			case 5:
				p.bg.addrLatch = p.bgAddr()
			case 6:
				p.bg.bglo = p.Bus.Read8(p.bg.addrLatch)

			// high background byte
			case 7:
				p.bg.addrLatch += 8
			case 0:
				p.bg.bghi = p.Bus.Read8(p.bg.addrLatch)
				p.horzScroll()
			}
		}

	case postRender:
		break

	case vblankNMI:
		if p.Cycle == 1 {
			ppustatus := ppustatus(p.PPUSTATUS.Value)
			ppustatus.setVblank(true)
			p.PPUSTATUS.Value = uint8(ppustatus)
			if ppuctrl(p.PPUCTRL.Value).nmi() {
				p.CPU.setNMIflag()
				log.ModPPU.DebugZ("Set NMI flag").End()
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

	// shift registers.
	bgShiftlo, bgShifthi uint16
}

func (p *PPU) ntAddr() uint16 {
	return 0x2000 | p.vramAddr.addr()&0xFFF
}

func (p *PPU) atAddr() uint16 {
	// TODO this is wrong since we should use coarseX and coarseY
	return 0x23C0 | uint16(p.vramAddr.nametable())<<10
}

func (p *PPU) bgAddr() uint16 {
	ppuctrl := ppuctrl(p.PPUCTRL.Value)
	return ppuctrl.bgTable()*0x1000 + (uint16(p.bg.nt) * 16) + p.vramAddr.finey()
}

func (p *PPU) refillShifters() {
	p.bg.bgShiftlo = (p.bg.bgShiftlo & 0xFF00) | uint16(p.bg.bglo)
	p.bg.bgShifthi = (p.bg.bgShifthi & 0xFF00) | uint16(p.bg.bghi)
}

func (p *PPU) horzScroll() {
	if !p.renderingEnabled() {
		return
	}
	if p.vramAddr.coarsex() == 31 {
		p.vramAddr.setVal(uint16(p.vramAddr) ^ 0x41F)
	} else {
		p.vramAddr.setCoarsex(p.vramAddr.coarsex() + 1)
	}
}

func (p *PPU) renderingEnabled() bool {
	mask := ppumask(p.PPUMASK.Value)
	return mask.bg() || mask.sprites()
}

func (p *PPU) renderPixel() {
	var palette uint8
	var x = p.Cycle - 2

	mask := ppumask(p.PPUMASK.Value)
	if p.Scanline < 240 && p.Cycle >= 0 && p.Cycle < 256 {
		if p.renderingEnabled() {
			if mask.bg() {
				hibit := uint8(p.bg.bgShifthi>>uint16(15-p.bg.finex)) & 1
				lobit := uint8(p.bg.bgShiftlo>>uint16(15-p.bg.finex)) & 1
				palette = (hibit << 1) | lobit
				fmt.Println(p.bg.bglo, p.bg.bghi, palette)
			}
			rgba := nesPalette[p.Palettes.Data[palette]]
			col := color.RGBA{R: uint8(rgba >> 16), G: uint8(rgba >> 8), B: uint8(rgba)}
			col.A = 0xFF
			p.screen.Set(x, p.Scanline, col)
		}
	}

	// Perform background shifts:
	p.bg.bgShiftlo <<= 1
	p.bg.bgShifthi <<= 1
	// atShiftL = (atShiftL << 1) | atLatchL
	// atShiftH = (atShiftH << 1) | atLatchH
}

var nesPalette = [...]uint32{
	0x7C7C7C, 0x0000FC, 0x0000BC, 0x4428BC, 0x940084, 0xA80020, 0xA81000, 0x881400,
	0x503000, 0x007800, 0x006800, 0x005800, 0x004058, 0x000000, 0x000000, 0x000000,
	0xBCBCBC, 0x0078F8, 0x0058F8, 0x6844FC, 0xD800CC, 0xE40058, 0xF83800, 0xE45C10,
	0xAC7C00, 0x00B800, 0x00A800, 0x00A844, 0x008888, 0x000000, 0x000000, 0x000000,
	0xF8F8F8, 0x3CBCFC, 0x6888FC, 0x9878F8, 0xF878F8, 0xF85898, 0xF87858, 0xFCA044,
	0xF8B800, 0xB8F818, 0x58D854, 0x58F898, 0x00E8D8, 0x787878, 0x000000, 0x000000,
	0xFCFCFC, 0xA4E4FC, 0xB8B8F8, 0xD8B8F8, 0xF8B8F8, 0xF8A4C0, 0xF0D0B0, 0xFCE0A8,
	0xF8D878, 0xD8F878, 0xB8F8B8, 0xB8F8D8, 0x00FCFC, 0xF8D8F8, 0x000000, 0x000000,
}

func (p *PPU) WritePATTERNTABLES(addr uint16, n int) {
	log.ModPPU.DebugZ("Write to PATTERNTABLES").
		Hex8("val", p.PatternTables.Data[addr]).
		Hex16("addr", addr).
		End()
}

func (p *PPU) WriteNAMETABLES(addr uint16, n int) {
	memaddr := addr & 0x0FFF
	ntnum := memaddr / 0x0400
	log.ModPPU.DebugZ("Write to NAMETABLES").
		Uint16("num", ntnum).
		Hex8("val", p.NameTables.Data[memaddr]).
		Hex16("addr", addr).
		End()
}

func (p *PPU) WritePALETTES(addr uint16, n int) {
	memaddr := addr & 0x01F
	log.ModPPU.DebugZ("Write to PALETTES").
		Hex8("val", p.Palettes.Data[memaddr]).
		Hex16("addr", addr).
		End()
}

// PPUCTRL: $2000
func (p *PPU) WritePPUCTRL(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUCTRL").Hex8("val", val).End()

	// By toggling the nmi bit (bit 7 of PPUCTRL) during vblank without reading
	// PPUSTATUS, a program can cause /nmi to be pulled low multiple times,
	// causing multiple NMIs to be generated.
	ppuctrl := ppuctrl(val)
	ppustatus := ppustatus(p.PPUSTATUS.Value)
	if !ppuctrl.nmi() {
		p.CPU.clearNMIflag()
	} else if ppustatus.vblank() {
		p.CPU.setNMIflag()
	}

	// Transfer the nametable bits.
	p.vramTmp.setNametable(ppuctrl.nametable())
}

// PPUMASK: $2001
func (p *PPU) WritePPUMASK(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUMASK").Hex8("val", val).End()
}

// PPUSTATUS: $2002
func (ppu *PPU) ReadPPUSTATUS(val uint8) uint8 {
	ppu.writeLatch = false

	ppustatus := ppustatus(val)
	ppustatus.setSpriteOverflow(true)
	ppustatus.setSpriteHit(true)
	ppustatus.setVblank(true)

	ppu.CPU.clearNMIflag()
	// TODO: emulate open bus?
	return uint8(ppustatus)
}

// PPUSCROLL: $2005
func (p *PPU) WritePPUSCROLL(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUSCROLL").Hex8("val", val).End()

	if !p.writeLatch { // first write
		p.bg.finex = val & 0b111
		p.vramTmp.setCoarsex(val >> 3)
	} else { // second write
		p.vramTmp.setFiney(uint16(val))
		p.vramTmp.setCoarsey(val >> 3)
	}

	p.writeLatch = !p.writeLatch
}

// To read/write VRAM from CPU, PPUADDR is set to the address of the operation.
// It's a 16-bit register so 2 writes are necessary.
// PPUADDR: $2006
func (p *PPU) WritePPUADDR(old, val uint8) {
	if !p.writeLatch { //first write
		p.vramTmp.setHigh(val & 0x3f)
	} else { // second write
		p.vramTmp.setLow(val)
		p.vramAddr.setVal(uint16(p.vramTmp))
	}

	p.writeLatch = !p.writeLatch
}

// PPUDATA: $2007
func (p *PPU) ReadPPUDATA(_ uint8) uint8 {
	var val uint8
	switch {
	case p.vramAddr < 0x3EFF:
		// Reading VRAM is too slow so the actual data
		// will be returned at the next read.
		data := p.ppuDataRbuf
		p.ppuDataRbuf = p.Bus.Read8(p.vramAddr.addr())
		val = data
	default: // $3F00-3FFF
		// Reading palette data is immediate.
		val = p.Bus.Read8(p.vramAddr.addr())
		// Still it overwrites the read buffer.
		p.ppuDataRbuf = val
	}

	p.incVRAMaddr()
	log.ModPPU.DebugZ("VRAM read").
		Hex16("addr", p.vramAddr.addr()).
		Hex8("val", val).
		End()
	return val
}

// PPUDATA: $2007
func (p *PPU) WritePPUDATA(old, val uint8) {
	p.Bus.Write8(p.vramAddr.addr(), val)
	p.incVRAMaddr()

	log.ModPPU.DebugZ("VRAM write").
		Hex16("addr", p.vramAddr.addr()).
		Hex8("val", val).
		End()
}

// After each i/o on PPUDATA, PPPUADDR is incremented.
func (p *PPU) incVRAMaddr() {
	if p.Scanline < 240 {
		return
	}

	// Increment VRAM address.
	ppuctrl := ppuctrl(p.PPUCTRL.Value)
	incr := uint16(1)
	if ppuctrl.incr() {
		p.vramAddr += 32
	}
	p.vramAddr.setVal(uint16(p.vramAddr) + incr)
}
