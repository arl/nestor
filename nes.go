package main

import (
	"fmt"
	"nestor/ines"
)

type NES struct {
	CPU *CPU
}

func (nes *NES) Boot(rom *ines.Rom) error {
	cpubus := newCpuBus("cpu")
	cpubus.MapMemory()

	nes.CPU = NewCPU(cpubus, defDisasm)
	if rom.Mapper() != 0 {
		// Only handle mapper 000 (NROM) for now.
		return fmt.Errorf("unsupported mapper: %d", rom.Mapper())
	}

	if err := loadMapper000(rom, cpubus); err != nil {
		return fmt.Errorf("failed to load mapper %03d: %s", rom.Mapper(), err)
	}
	return nil
}

func (nes *NES) Run() {
	nes.CPU.reset()
	nes.CPU.Run(512) // debug: run 512 cycles
}
