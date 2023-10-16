package main

import (
	"log"
	"nestor/ines"
)

type NES struct {
	CPU *CPU
}

func startNES(rom *ines.Rom) (*NES, error) {
	nes := &NES{
		CPU: NewCPU(),
	}

	nes.CPU.MapMemory()

	// Only handle mapper 000 (NROM) for now.
	if rom.Mapper() != 0 {
		log.Fatalf("only mapper 000 supported")
	}
	loadMapper000(rom, nes)

	if disasm {

	}

	nes.CPU.reset()
	nes.CPU.Run(128) // debug: run 128 cycles

	return nes, nil
}
