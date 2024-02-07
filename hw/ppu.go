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
	vramAddr    loopy
	vramTmp     loopy
	writeLatch  bool
	ppuDataRbuf uint8

	bg bgregs
}

// background registers
type bgregs struct {
	// These contain the pattern table data for two tiles. Every 8 cycles,
	// the data for the next tile is loaded into the upper 8 bits of this
	// shift register. Meanwhile, the pixel to render is fetched from one of
	// the lower 8 bits.
	patternData [2]uint16

	// These contain the palette attributes for the lower 8 pixels of the
	// 16-bit shift register. These registers are fed by a latch which
	// contains the palette attribute for the next tile. Every 8 cycles, the
	// latch is loaded with the palette attribute for the next tile.
	paletteAttr [2]uint8

	finex uint8 // 3-bit 'fine x scroll' register
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

		if p.Cycle >= 1 && p.Cycle <= 256 ||
			p.Cycle >= 321 && p.Cycle <= 336 {
			switch p.Cycle & 0b111 {
			case 1:
				p.bg.patternData[0] = p.bg.patternData[1]

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
	p.PPUCTRL.Value = uint8(ppuctrl)
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
	ppustatus.setVblank(false)

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
		p.vramTmp.setFiney(val)
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
