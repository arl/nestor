package emu

import (
	"flag"
	"testing"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
)

var romName = flag.String("rom", "", "ROM file to load for BenchmarkCPUSpeed")

func BenchmarkCPUSpeed(b *testing.B) {
	log.Disable()

	if *romName == "" {
		b.Fatal("missing -rom flag")
	}

	rom, err := ines.ReadRom(*romName)
	if err != nil {
		b.Fatal(err)
	}

	nes, err := powerUp(rom)
	if err != nil {
		b.Fatal(err)
	}

	cfg := TestingOutputConfig{
		Height: hw.NTSCHeight,
		Width:  hw.NTSCWidth,
	}
	e := Emulator{
		NES: nes,
		out: newTestingOutput(cfg),
	}
	frame := e.out.BeginFrame()

	for b.Loop() {
		for range 300 {
			e.NES.RunOneFrame(frame)
		}
	}
}
