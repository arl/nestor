package main

import (
	"nestor/ines"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNestest(t *testing.T) {
	// t.Skip("skip for now that we don't implement all opcodes")
	nes := new(NES)
	cartridge, err := ines.ReadRom("testdata/nes-test-roms/other/nestest.nes")
	tcheck(t, err)
	tcheck(t, nes.PowerUp(cartridge))

	flog, err := os.CreateTemp("", "nestor.nestet.*.log")
	tcheck(t, err)

	t.Cleanup(func() {
		flog.Close()
		t.Logf("log saved to %s", flog.Name())
		content, err := os.ReadFile(flog.Name())
		tcheck(t, err)

		want, err := os.ReadFile("testdata/nes-test-roms/other/nestest.log")
		tcheck(t, err)
		diff := cmp.Diff(want, content)
		if diff != "" {
			t.Errorf("nestest.log mismatch")
		}
		if testing.Verbose() {
			t.Logf("[%s]\n\n%s\n", flog.Name(), content)
		}
	})

	nes.CPU.setDisasm(flog, true)

	// For some reason the nestest.log shows an execution starting from 0xC000, at which
	// the CPU has already executed 7 cycles.
	nes.CPU.PC = 0xC000
	nes.CPU.Clock = 7
	nes.CPU.P = P(0b00100100)

	nes.CPU.Run(26554)
}
