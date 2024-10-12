package main

import (
	"fmt"
	"io"
	"os"
	"runtime/pprof"

	"github.com/veandco/go-sdl2/sdl"

	"nestor/emu"
	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
	"nestor/ui"
)

func main() {
	sdl.Main(main1)
}

func main1() {
	cfg := parseArgs(os.Args[1:])

	switch cfg.mode {
	case guiMode:
		guiMain()
	case romInfosMode:
		romInfosMain(cfg.RomInfos.RomPath)
	case runMode:
		emuMain(cfg.Run)
	case mapInputMode:
		hw.MapInputMain(cfg.MapInput.Button)
	}
}

func romInfosMain(romPath string) {
	rom, err := ines.ReadRom(romPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading ROM: %s", err)
		os.Exit(1)
	}
	rom.PrintInfos(os.Stdout)
}

// guiMain runs Nestor graphical user interface.
func guiMain() {
	ch := make(chan struct{}, 1)
	go func() {
		defer close(ch)
		if err := ui.ShowMainWindow(); err != nil {
			log.ModEmu.FatalZ("failed to show main window").Error("error", err).End()
		}
	}()
	<-ch
	log.ModEmu.InfoZ("Nestor exit").End()
}

// emuMain runs the emulator directly with the given rom.
func emuMain(cfg Run) {
	rom, err := ines.ReadRom(cfg.RomPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading ROM: %s", err)
		os.Exit(1)
	}

	var traceout io.WriteCloser
	if cfg.Trace != nil {
		traceout = cfg.Trace
		defer traceout.Close()
	}

	emucfg := emu.LoadConfigOrDefault()
	emucfg.TraceOut = traceout
	nes, err := emu.Launch(rom, emucfg, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start emulator: %v\n", err)
		os.Exit(1)
	}

	if cfg.CPUProfile != "" {
		f, err := os.Create(cfg.CPUProfile)
		checkf(err, "failed to create cpu profile file")
		checkf(pprof.StartCPUProfile(f), "failed to start cpu profile")
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
			fmt.Println("CPU profile written to", cfg.CPUProfile)
		}()
	}

	nes.Run()
}
