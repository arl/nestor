package main

import (
	"fmt"
	"nestor/ines"
)

type NES struct {
	CPU *CPU
}

func bootNES(rom *ines.Rom) (*NES, error) {
	cpubus := newCpuBus("cpu")
	cpubus.MapMemory()

	nes := &NES{
		CPU: NewCPU(cpubus),
	}

	if rom.Mapper() != 0 {
		// Only handle mapper 000 (NROM) for now.
		return nil, fmt.Errorf("unsupported mapper: %d", rom.Mapper())
	}

	err := loadMapper000(rom, cpubus)
	return nes, err
}

func (nes *NES) Run() {
	nes.CPU.reset()
	nes.CPU.Run(512) // debug: run 512 cycles
}
