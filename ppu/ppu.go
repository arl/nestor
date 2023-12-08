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
	Bus *hwio.Table

	PatternTables hwio.Mem `hwio:"offset=0x0000,size=0x1000,vsize=0x2000,wcb"`
	NameTables    hwio.Mem `hwio:"offset=0x2000,size=0x1000,vsize=0x1F00,wcb"`
	Palettes      hwio.Mem `hwio:"offset=0x3F00,size=0x20,vsize=0xE0,wcb"`
}

func New(bus *hwio.Table) *PPU {
	ppu := &PPU{Bus: bus}
	ppu.Reset()
	return ppu
}

func (p *PPU) Reset() {

}

func (p *PPU) Tick() {

}

func (p *PPU) WriteNAMETABLES(addr uint16, n int) {
	memaddr := addr & 0x0FFF
	fmt.Printf("NAMETABLES writes %d bytes at 0x%04x -> 0x%02X\n", n, addr, p.NameTables.Data[memaddr])
}

func (p *PPU) WritePATTERNTABLES(addr uint16, n int) {
	memaddr := addr & 0x0FFF
	fmt.Printf("PATTERNTABLES writes %d bytes at 0x%04x -> 0x%02X\n", n, addr, p.PatternTables.Data[memaddr])
}

func (p *PPU) WritePALETTES(addr uint16, n int) {
	memaddr := addr & 0x01F
	fmt.Printf("PALLETTES writes %d bytes at 0x%04x -> 0x%02X\n", n, addr, p.Palettes.Data[memaddr])
}
