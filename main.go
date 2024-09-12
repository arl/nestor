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
	cfg.validate() // fails if necessary

	switch cfg.mode {
	case guiMode:
		guiMain()
	case runMode:
		emuMain(cfg.Run)
	case mapInputMode:
		hw.MapInputMain(cfg.MapInput.Button)
	}
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
func emuMain(cfg RunConfig) {
	rom, err := ines.ReadRom(cfg.RomPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading ROM: %s", err)
		os.Exit(1)
	}

	if cfg.Infos {
		rom.PrintInfos(os.Stdout)
		os.Exit(0)
	}

	var (
		traceout io.WriteCloser
	)

	if cfg.Trace != nil {
		traceout = cfg.Trace
	}

	emucfg := emu.Config{
		TraceOut: traceout,
	}

	nes, err := emu.PowerUp(rom, emucfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start emulator: %v", err)
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
