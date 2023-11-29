package ppu

import (
	"nestor/emu/hwio"
)

type PPU struct {
	PPUCTRL   hwio.Reg8 `reg:""`
	PPUMASK   hwio.Reg8 `reg:""`
	PPUSTATUS hwio.Reg8 `reg:""`
}

func New() *PPU {
	ppu := &PPU{}
	ppu.Reset()
	return ppu
}

func (p *PPU) Reset() {

}

func (p *PPU) Tick() {

}
