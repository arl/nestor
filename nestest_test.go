package main

import (
	"flag"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"nestor/emu"
	"nestor/ines"
)

var doNestest = flag.Bool("nestest", false, "run TestNestest")

func TestNestest(t *testing.T) {
	if !*doNestest {
		t.Skip("skip for now that we don't implement all opcodes")
	}
	nes := new(NES)
	cartridge, err := ines.ReadRom("testdata/nes-test-roms/other/nestest.nes")
	tcheck(t, err)
	tcheck(t, nes.PowerUp(cartridge))

	flog, err := os.CreateTemp("", "nestor.nestest.*.log")
	tcheck(t, err)
	t.Log(flog.Name())

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

	nes.CPU.SetDisasm(flog, true)

	// nestest.nes has an automated test mode that starts at 0xC000, with 7
	// cycles. The manual mode starting at the reset vector requires the screen.
	nes.CPU.PC = 0xC000
	nes.CPU.Clock = 7
	nes.CPU.P = emu.P(0b00100100)

	nes.CPU.RunDisasm(26560)
}
