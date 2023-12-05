package ppu

import (
	"fmt"
	"nestor/emu/hwio"
)

type PPU struct {
	Bus hwio.Bus

	PPUCTRL   hwio.Reg8 `hwio:"offset=0x0,writeonly,wcb"`
	PPUMASK   hwio.Reg8 `hwio:"offset=0x1,writeonly,wcb"`
	PPUSTATUS hwio.Reg8 `hwio:"offset=0x2,readonly,rcb"`
}

func New() *PPU {
	ppu := &PPU{
		Bus: &hwio.MemMap{Name: "ppu"},
	}
	hwio.MustInitRegs(ppu)
	ppu.Reset()
	return ppu
}

func (p *PPU) Reset() {

}

func (p *PPU) Tick() {

}

func (p *PPU) WritePPUCTRL(old uint8, val uint8) {
	fmt.Printf("PPUCTRL: %02x -> %02x\n", old, val)
}

func (p *PPU) WritePPUMASK(old uint8, val uint8) {
	fmt.Printf("PPUMASK: %02x -> %02x\n", old, val)
}

func (p *PPU) ReadPPUSTATUS(val uint8) uint8 {
	fmt.Printf("PPUSTATUS(%02x) read %02x\n", val, val)
	return val
}
