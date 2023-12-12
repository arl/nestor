package ppu

import (
	"fmt"

	"nestor/emu/hwio"
)

const (
	NumScanlines = 262 // Number of scanlines per frame.
	NumCycles    = 341 // Number of PPU cycles per scanline.
)

type PPU struct {
	Bus  *hwio.Table // PPU bus
	Regs *Regs       // PPU registers

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
}

func New(bus *hwio.Table) *PPU {
	ppu := &PPU{
		Bus:  bus,
		Regs: NewRegs(),
	}

	hwio.MustInitRegs(ppu)
	ppu.Reset()
	return ppu
}

func (p *PPU) InitBus() {
	p.Bus.MapBank(0x0000, p, 0)
}

func (p *PPU) Reset() {

}

func (p *PPU) Tick() {
	switch {
	// Pre-render line
	case p.Scanline == 261:
		if p.Cycle == 1 {
			// Clear vblank, sprite0Hit and spriteOverflow
			p.Regs.PPUSTATUS.Value &^= 0b11100000
		}

	// Visible scanlines
	case p.Scanline >= 0 && p.Scanline <= 239:
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

	// Post-render scanline
	case p.Scanline == 240:
		break

	// VBlank
	case p.Scanline >= 241 && p.Scanline <= 260:
		if p.Scanline == 241 && p.Cycle == 1 {
			p.Regs.PPUSTATUS.Value |= 1 << vblank
		}
	}

	p.Cycle++
	if p.Cycle >= NumCycles {
		p.Cycle = 0
		p.Scanline++
		if p.Scanline >= NumScanlines {
			p.Scanline = 0
		}
	}
}

func (p *PPU) WritePATTERNTABLES(addr uint16, n int) {
	fmt.Printf("PATTERNTABLES writes %d bytes at 0x%04x -> 0x%02X\n", n, addr, p.PatternTables.Data[addr])
}

func (p *PPU) WriteNAMETABLES(addr uint16, n int) {
	memaddr := addr & 0x0FFF
	ntnum := memaddr / 0x0400
	fmt.Printf("NAMETABLES writes %d bytes at 0x%04x (nametable %d)-> 0x%02X\n", n, addr, ntnum, p.NameTables.Data[memaddr])
}

func (p *PPU) WritePALETTES(addr uint16, n int) {
	memaddr := addr & 0x01F
	fmt.Printf("PALLETTES writes %d bytes at 0x%04x -> 0x%02X\n", n, addr, p.Palettes.Data[memaddr])
}
