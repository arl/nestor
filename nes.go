package main

import (
	"log"
	"nestor/ines"
)

type NES struct {
	CPU *CPU
}

func bootNES(rom *ines.Rom) (*NES, error) {
	cpuBus := newcpuBus("cpu")
	cpuBus.MapMemory()

	nes := &NES{
		CPU: NewCPU(cpuBus),
	}

	// Only handle mapper 000 (NROM) for now.
	if rom.Mapper() != 0 {
		log.Fatalf("only mapper 000 supported")
	}
	loadMapper000(rom, nes)

	if disasmOn {

	}

	nes.CPU.reset()
	nes.CPU.Run(256) // debug: run 128 cycles

	return nes, nil
}
