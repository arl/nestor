package main

import (
	"fmt"
	"io"
	"os"
	"runtime/pprof"

	"github.com/veandco/go-sdl2/sdl"

	"nestor/emu"
	"nestor/ines"
	"nestor/ui"
)

func main() {
	sdl.Main(main1)
}

func main1() {
	args := parseArgs(os.Args[1:])

	switch args.mode {
	case guiMode:
		ui.RunApp()
	case romInfosMode:
		romInfosMain(args.RomInfos.RomPath)
	case runMode:
		emuMain(args.Run)
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

// emuMain runs the emulator directly with the given rom.
func emuMain(args Run) {
	rom, err := ines.ReadRom(args.RomPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading ROM: %s", err)
		os.Exit(1)
	}

	var traceout io.WriteCloser
	if args.Trace != nil {
		traceout = args.Trace
		defer traceout.Close()
	}

	cfg := ui.LoadConfigOrDefault()
	cfg.TraceOut = traceout
	nes, err := emu.Launch(rom, cfg.Config, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start emulator: %v\n", err)
		os.Exit(1)
	}

	if args.CPUProfile != "" {
		f, err := os.Create(args.CPUProfile)
		checkf(err, "failed to create cpu profile file")
		checkf(pprof.StartCPUProfile(f), "failed to start cpu profile")
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
			fmt.Println("CPU profile written to", args.CPUProfile)
		}()
	}

	nes.Run()
}
