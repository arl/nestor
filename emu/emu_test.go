package emu

import (
	"flag"
	"path/filepath"
	"testing"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
	"nestor/tests"
)

var romPath = flag.String("rom", "", "ROM file to load for BenchmarkCPUSpeed")

func loadEmulator(b *testing.B, romPath string) *Emulator {
	log.Disable()
	b.ReportAllocs()

	rom, err := ines.ReadRom(romPath)
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
	return &e
}

func BenchmarkCPUSpeed(b *testing.B) {
	if *romPath == "" {
		b.Fatal("missing -rom flag")
	}

	e := loadEmulator(b, *romPath)
	frame := e.out.BeginFrame()

	for b.Loop() {
		for range 300 {
			e.NES.RunOneFrame(frame)
		}
	}
}

func BenchmarkSaveState(b *testing.B) {
	romPath := filepath.Join(tests.RomsPath(b), "spritecans-2011", "spritecans.nes")
	e := loadEmulator(b, romPath)

	frame := e.out.BeginFrame()
	e.NES.RunOneFrame(frame)

	b.ResetTimer()
	for b.Loop() {
		_, _ = e.NES.SaveSnapshot()
	}
}
