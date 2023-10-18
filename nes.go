package main

import (
	"log"
	"nestor/ines"
)

type NES struct {
	CPU *CPU
}

func powerUp(rom *ines.Rom) (*NES, error) {
	cpubus := newCpuBus("cpu")
	cpubus.MapMemory()

	nes := &NES{
		CPU: NewCPU(cpubus),
	}

	// Only handle mapper 000 (NROM) for now.
	if rom.Mapper() != 0 {
		log.Fatalf("only mapper 000 supported")
	}
	loadMapper000(rom, cpubus)

	if disasmOn {

	}

	nes.CPU.reset()
	nes.CPU.Run(512) // debug: run 512 cycles

	return nes, nil
}
