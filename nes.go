package main

import (
	"fmt"

	"nestor/cpu"
	"nestor/ines"
)

type NES struct {
	CPU *cpu.CPU
}

func (nes *NES) PowerUp(rom *ines.Rom) error {
	cpubus := newCpuBus("cpu")
	cpubus.MapMemory()

	nes.CPU = cpu.NewCPU(cpubus, &ticker{})
	if rom.Mapper() != 0 {
		// Only handle mapper 000 (NROM) for now.
		return fmt.Errorf("unsupported mapper: %d", rom.Mapper())
	}

	if err := loadMapper000(rom, cpubus); err != nil {
		return fmt.Errorf("failed to load mapper %03d: %s", rom.Mapper(), err)
	}
	return nil
}

// Reset forwards the reset signal to all hardware.
func (nes *NES) Reset() {
	nes.CPU.Reset()
}

func (nes *NES) Run() {
	nes.CPU.Run(29692) // random
}

type ticker struct{}

func (tt *ticker) Tick() {}
