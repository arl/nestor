package ppu

import (
	"fmt"

	"nestor/emu/hwio"
)

// CPU-exposed memory-mapped PPU registers.
// This bank is mapped at 0x2000-0x3FFF, with mirrors.
type Regs struct {
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

func NewRegs() *Regs {
	regs := &Regs{}
	hwio.MustInitRegs(regs)
	return regs
}

func (r *Regs) WritePPUCTRL(old uint8, val uint8) {
	fmt.Printf("PPUCTRL write %02x was %02x\n", val, old)
}

func (r *Regs) WritePPUMASK(old uint8, val uint8) {
	fmt.Printf("PPUMASK write %02x was %02x\n", val, old)
}

func (r *Regs) ReadPPUSTATUS(val uint8) uint8 {
	fmt.Printf("PPUSTATUS read %02x\n", val)
	return val
}

func (r *Regs) WritePPUADDR(old uint8, val uint8) {
	fmt.Printf("PPUADDR write %02x was %02x\n", val, old)
}

func (r *Regs) ReadPPUDATA(val uint8) uint8 {
	fmt.Printf("PPUDATA read %02x\n", val)
	return val
}

func (r *Regs) WritePPUDATA(old uint8, val uint8) {
	fmt.Printf("PPUDATA write %02x was %02x\n", val, old)
}
