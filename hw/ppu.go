package hw

import (
	"image"

	"nestor/emu/hwio"
	"nestor/emu/log"
)

const (
	NumScanlines = 262 // Number of scanlines per frame.
	NumCycles    = 341 // Number of PPU cycles per scanline.
)

const (
	// PPUCTRL bits
	// $2000

	// Nametable selection mask
	// (0 = $2000; 1 = $2400; 2 = $2800; 3 = $2C00)
	ntselect = 0b11

	// VRAM address increment per CPU read/write of PPUDATA
	// (0: +1 i.e. horizontal; 1: +32 i.e. vertical)
	vramIncr = 2

	// Sprite pattern table address for 8x8 sprites
	// (0: $0000; 1: $1000; ignored in 8x16 mode)
	spriteAddr = 3

	// Background pattern table address (0: $0000; 1: $1000)
	backgroundAddr = 4

	// Sprite size (0: 8x8 pixels; 1: 8x16 pixels – see byte 1 of OAM)
	spriteSize = 5

	// PPU master/slave select
	// (0: read backdrop from EXT pins; 1: output color on EXT pins)
	ppuMasterSlave = 6

	// Generate an NMI at the start of the
	// vertical blanking interval (0: off; 1: on)
	nmi = 7
)

const (
	// PPUMASK bits
	// $2001

	// Greyscale
	// (0: normal color, 1: produce a greyscale display)
	greyscale = 0

	// Show background in leftmost 8 pixels of screen
	// 1: Show, 0: Hide
	leftmostBg = 1

	// Show sprites in leftmost 8 pixels of screen
	// 1: Show, 0: Hide
	leftmostSprites = 2

	// Show background
	showBg = 3

	// Show sprites
	showSprites = 4

	// Emphasize red
	highlightRed = 5

	// Emphasize green
	highlightGreen = 6

	// Emphasize blue
	highlightBlue = 7
)

const (
	// PPUSTATUS bits
	// $2002

	// Returns stale PPU bus contents.
	openbusMask = 0b11111

	// Sprite overflow. The intent was for this flag to be set
	// whenever more than eight sprites appear on a scanline, but a
	// hardware bug causes the actual behavior to be more complicated
	// and generate false positives as well as false negatives; see
	// PPU sprite evaluation. This flag is set during sprite
	// evaluation and cleared at dot 1 (the second dot) of the
	// pre-render line.
	spriteOverflow = 5

	// Sprite 0 Hit.  Set when a nonzero pixel of sprite 0 overlaps
	// a nonzero background pixel; cleared at dot 1 of the pre-render
	// line.  Used for raster timing.
	sprite0Hit = 6

	// Vertical blank has started (0: not in vblank; 1: in vblank).
	// Set at dot 1 of line 241 (the line *after* the post-render
	// line); cleared after reading $2002 and at dot 1 of the
	// pre-render line.
	vblank = 7
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
	vramAddr    uint16
	vramTmp     uint16
	writeLatch  bool
	ppuDataRbuf uint8

	// Background registers
	bgPatternData [2]uint16
	bgPaletteAttr [2]uint8
}

func NewPPU() *PPU {
	return &PPU{
		Bus: hwio.NewTable("ppu"),
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
	case preRender:
		if p.Cycle == 1 {
			// Clear vblank, sprite0Hit and spriteOverflow
			const mask = 1<<vblank | 1<<sprite0Hit | 1<<spriteOverflow
			p.PPUSTATUS.ClearBits(mask)
		}

	case renderMode:
		switch {
		// Idle
		case p.Cycle == 0:
			break
		// Fetch data
		case p.Cycle > 0 && p.Cycle <= 256:
			break
		case p.Cycle > 256 && p.Cycle <= 320:
			break
		case p.Cycle > 321 && p.Cycle <= 336:
			break
		case p.Cycle > 337 && p.Cycle <= 340:
			break
		}

	case postRender:
		break

	case vblankNMI:
		if p.Cycle == 1 {
			p.PPUSTATUS.SetBit(vblank)
			if p.PPUCTRL.GetBit(nmi) {
				p.CPU.setNMIflag()
			}
		}
	}

}

// render renders one pixel.
func (p *PPU) render() {

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

	nmiOnVblank := p.PPUCTRL.GetBit(nmi)

	// By toggling the nmi bit (bit 7 of PPUCTRL) during vblank without reading
	// PPUSTATUS, a program can cause /nmi to be pulled low multiple times,
	// causing multiple NMIs to be generated.
	if !nmiOnVblank {
		p.CPU.clearNMIflag()
	} else if p.PPUSTATUS.GetBit(vblank) {
		p.CPU.setNMIflag()
	}

	// Transfer the nametable bits.
	p.vramTmp &^= ntselect << 10
	p.vramTmp |= (uint16(val) & ntselect) << 10

	p.PPUCTRL.Value = val
}

// PPUMASK: $2001
func (p *PPU) WritePPUMASK(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUMASK").Hex8("val", val).End()
}

// PPUSTATUS: $2002
func (ppu *PPU) ReadPPUSTATUS(val uint8) uint8 {
	ppu.writeLatch = false
	ret := ppu.PPUSTATUS.GetBiti(spriteOverflow)<<spriteOverflow |
		ppu.PPUSTATUS.GetBiti(sprite0Hit)<<sprite0Hit |
		ppu.PPUSTATUS.GetBiti(vblank)<<vblank

	ppu.PPUSTATUS.ClearBit(vblank)
	ppu.CPU.clearNMIflag()
	// TODO: emulate open bus?
	return ret
}

// PPUSCROLL: $2005
func (p *PPU) WritePPUSCROLL(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUSCROLL").Hex8("val", val).End()

	if !p.writeLatch { // first write
		p.bg.finex = val & 0b111
		p.vramTmp &^= 0b1_1111
		p.vramTmp |= uint16(val >> 3)
	} else { // second write
		p.vramTmp &^= 0b0111_0011_1110_0000
		p.vramTmp |= uint16(val&0b111) << 12
		p.vramTmp |= uint16(val&0b1111_1000) << 2
	}

	p.writeLatch = !p.writeLatch
}

// To read/write VRAM from CPU, PPUADDR is set to the address of the operation.
// It's a 16-bit register so 2 writes are necessary.
// PPUADDR: $2006
func (p *PPU) WritePPUADDR(old, val uint8) {
	if !p.writeLatch { //first write
		p.vramTmp &^= 0b11_1111_0000_0000
		p.vramTmp |= uint16(val&0b11_1111) << 8
		p.vramTmp &^= 1 << 14 // clear z bit
	} else { // second write
		p.vramTmp &^= 0xff
		p.vramTmp |= uint16(val)
		p.vramAddr = p.vramTmp
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
		p.ppuDataRbuf = p.Bus.Read8(p.vramAddr)
		val = data
	default: // $3F00-3FFF
		// Reading palette data is immediate.
		val = p.Bus.Read8(p.vramAddr)
		// Still it overwrites the read buffer.
		p.ppuDataRbuf = val
	}

	p.incVRAMaddr()
	log.ModPPU.DebugZ("VRAM read").
		Hex16("addr", p.vramAddr).
		Hex8("val", val).
		End()
	return val
}

// PPUDATA: $2007
func (p *PPU) WritePPUDATA(old, val uint8) {
	// Mirror down address (only $000-$3fff range is valid).
	p.vramAddr &= 0x3fff
	p.Bus.Write8(p.vramAddr, val)
	p.incVRAMaddr()

	log.ModPPU.DebugZ("VRAM write").
		Hex16("addr", p.vramAddr).
		Hex8("val", val).
		End()
}

// After each i/o on PPUDATA, PPPUADDR is incremented.
func (p *PPU) incVRAMaddr() {
	if p.Scanline < 240 {
		return
	}

	// Increment VRAM address.
	incr := uint16(1)
	if p.PPUCTRL.GetBit(vramIncr) {
		p.vramAddr += 32
	}
	p.vramAddr = (p.vramAddr + incr) & 0x7fff
}
}
