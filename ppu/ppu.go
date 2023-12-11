package ppu

import (
	"fmt"

	"nestor/emu/hwio"
)

/*
The PPU renders 262 scanlines per frame. Each scanline lasts for 341 PPU clock cycles (113.667 CPU clock cycles; 1 CPU cycle = 3 PPU cycles), with each clock cycle producing one pixel. The line numbers given here correspond to how the internal PPU frame counters count lines.

The information in this section is summarized in the diagram in the next section.

The timing below is for NTSC PPUs. PPUs for 50 Hz TV systems differ:

*/

type PPU struct {
	Bus  *hwio.Table
	Regs *Regs

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
