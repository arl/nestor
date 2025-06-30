package emu

import (
	"flag"
	"testing"
	"time"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
)

var romPath = flag.String("rom", "", "ROM file to load for BenchmarkCPUSpeed")

func loadEmulator(b *testing.B, romPath string) *Emulator {
	log.Disable()
	b.ReportAllocs()

	rom, err := ines.ReadROM(romPath)
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

	const nframes = 300

	nloops := 0
	start := time.Now()

	for b.Loop() {
		for range nframes {
			e.NES.RunOneFrame(&frame)
		}
		nloops++
	}
	fps := float64(nframes*nloops) / time.Since(start).Seconds()
	b.ReportMetric(fps, "frames/s")
}

func BenchmarkSaveSnapshot(b *testing.B) {
	if *romPath == "" {
		b.Fatal("missing -rom flag")
	}
	e := loadEmulator(b, *romPath)

	frame := e.out.BeginFrame()
	e.NES.RunOneFrame(&frame)

	snapshot, err := e.NES.SaveSnapshot()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = e.NES.SaveSnapshot()
	}

	totbytes := len(snapshot) * b.N
	b.ReportMetric(float64(totbytes)/b.Elapsed().Seconds(), "bytes/s")
	b.ReportMetric(float64(len(snapshot)), "bytes")
}

func BenchmarkLoadSnapshot(b *testing.B) {
	if *romPath == "" {
		b.Fatal("missing -rom flag")
	}
	e := loadEmulator(b, *romPath)

	frame := e.out.BeginFrame()
	e.NES.RunOneFrame(&frame)

	snapshot, err := e.NES.SaveSnapshot()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		_ = e.NES.LoadSnapshot(snapshot)
	}

	totbytes := len(snapshot) * b.N
	b.ReportMetric(float64(totbytes)/b.Elapsed().Seconds(), "bytes/s")
	b.ReportMetric(float64(len(snapshot)), "bytes")
}
