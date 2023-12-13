package ppu

import (
	"nestor/emu/hwio"
	log "nestor/emu/logger"
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

const (
	vblank         = 7
	sprite0Hit     = 6
	spriteOverflow = 5
)

func NewRegs() *Regs {
	regs := &Regs{}
	hwio.MustInitRegs(regs)
	return regs
}

func (r *Regs) WritePPUCTRL(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUCTRL").Hex8("val", val).End()
}

func (r *Regs) WritePPUMASK(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUMASK").Hex8("val", val).End()
}

func (r *Regs) ReadPPUSTATUS(val uint8) uint8 {
	log.ModPPU.DebugZ("Read from PPUSTATUS").Hex8("val", val).End()
	return val
}

func (r *Regs) WritePPUADDR(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUADDR").Hex8("val", val).End()
}

func (r *Regs) ReadPPUDATA(val uint8) uint8 {
	log.ModPPU.DebugZ("Read from PPUDATA").Hex8("val", val).End()
	return val
}

func (r *Regs) WritePPUDATA(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUDATA").Hex8("val", val).End()
}
