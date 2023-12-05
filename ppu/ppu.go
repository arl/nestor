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

	PPUCTRL   hwio.Reg8 `hwio:"offset=0x0,writeonly,wcb"`
	PPUMASK   hwio.Reg8 `hwio:"offset=0x1,writeonly,wcb"`
	PPUSTATUS hwio.Reg8 `hwio:"offset=0x2,readonly,rcb"`
	// OAMADDR   hwio.Reg8 `hwio:"offset=0x3,writeonly,wcb"`
	// OAMDATA   hwio.Reg8 `hwio:"offset=0x4"`
	// PPUSCROLL hwio.Reg8 `hwio:"offset=0x5,writeonly,wcb"`
	PPUADDR hwio.Reg8 `hwio:"offset=0x6,writeonly,wcb"`
	PPUDATA hwio.Reg8 `hwio:"offset=0x7,rcb,wcb,"`

	// OAMDMA hwio.Reg8 `hwio:"bank:1,writeonly,wcb"`
}

func New(bus *hwio.Table) *PPU {
	ppu := &PPU{Bus: bus}
	hwio.MustInitRegs(ppu)
	ppu.Reset()
	return ppu
}

func (p *PPU) Reset() {

}

func (p *PPU) Tick() {

}

func (p *PPU) WritePPUCTRL(old uint8, val uint8) {
	fmt.Printf("PPUCTRL write %02x was %02x\n", val, old)
}

func (p *PPU) WritePPUMASK(old uint8, val uint8) {
	fmt.Printf("PPUMASK write %02x was %02x\n", val, old)
}

func (p *PPU) ReadPPUSTATUS(val uint8) uint8 {
	fmt.Printf("PPUSTATUS read %02x\n", val)
	return val
}

func (p *PPU) WritePPUADDR(old uint8, val uint8) {
	fmt.Printf("PPUADDR write %02x was %02x\n", val, old)
}

func (p *PPU) ReadPPUDATA(val uint8) uint8 {
	fmt.Printf("PPUDATA read %02x\n", val)
	return val
}

func (p *PPU) WritePPUDATA(old uint8, val uint8) {
	fmt.Printf("PPUDATA write %02x was %02x\n", val, old)
}
