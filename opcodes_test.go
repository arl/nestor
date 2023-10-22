package main

import (
	"testing"
)

func TestOpcodeLDASTA(t *testing.T) {
	dump := `0600: a9 01 8d 00 02 a9 05 8d 01 02 a9 08 8d 02 02`
	cpu := loadCPUWith(t, dump)
	cpu.PC = 0x0600
	cpu.Run(21)

	wantCPUState(t, cpu,
		"A", 0x08,
		"Pb", 1,
		"PC", 0x060F,
		"SP", 0xfd,
	)
}

func TestEOR(t *testing.T) {
	t.Run("zeropage", func(t *testing.T) {
		dump := `
0000: 06
0100: 45 00`
		cpu := loadCPUWith(t, dump)
		cpu.PC = 0x0100
		cpu.A = 0x80
		cpu.Run(3)

		wantCPUState(t, cpu,
			"A", 0x86,
			"Pn", 1,
			"Pz", 0,
		)
	})
}

func TestROR(t *testing.T) {
	t.Run("zeropage", func(t *testing.T) {
		dump := `
0000: 55
0100: 66 00`
		cpu := loadCPUWith(t, dump)
		cpu.PC = 0x0100
		cpu.A = 0x80
		cpu.P.writeBit(pbitC, true)

		cpu.Run(5)

		wantMem8(t, cpu, 0x0000, 0xAA)
		wantCPUState(t, cpu,
			"Pn", 1,
			"Pc", 1,
			"Pz", 0,
		)
	})
}
