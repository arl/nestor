package main

import (
	"nestor/ines"
	"testing"
)

func TestNestest(t *testing.T) {
	t.Skip("skip for now that we don't implement all opcodes")

	nes := NES{}
	cartridge, err := ines.ReadRom("testdata/nes-test-roms/other/nestest.nes")
	tcheck(t, err)

	tcheck(t, nes.Boot(cartridge))

	// For some reason nestest.nes starts at 0xC000
	nes.CPU.PC = 0xC000
	nes.CPU.Run(8991)
}
