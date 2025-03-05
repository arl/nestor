package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"runtime/pprof"
	"slices"

	"github.com/veandco/go-sdl2/sdl"

	"nestor/emu"
	"nestor/ines"
	"nestor/ui"
)

func main() {
	args := parseArgs(os.Args[1:])

	cfg := ui.LoadConfigOrDefault()

	switch args.mode {
	case guiMode:
		ui.RunApp(&cfg)
	case romInfosMode:
		romInfosMain(args.RomInfos.RomPath)
	case runMode:
		emuMain(args.Run, &cfg)
	case versionMode:
		versionMain()
	}
}

// emuMain runs the emulator directly with the given rom.
func emuMain(args Run, cfg *ui.Config) {
	var exitcode int
	sdl.Main(func() {
		rom, err := ines.ReadRom(args.RomPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading ROM: %s", err)
			exitcode = 1
			return
		}

		var traceout io.WriteCloser
		if args.Trace != nil {
			traceout = args.Trace
			defer traceout.Close()
		}

		cfg.TraceOut = traceout
		cfg.Video.Monitor = args.Monitor

		nes, err := emu.Launch(rom, cfg.Config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to start emulator: %v\n", err)
			exitcode = 1
			return
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
	})
	os.Exit(exitcode)
}

func romInfosMain(romPath string) {
	rom, err := ines.ReadRom(romPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading ROM: %s", err)
		os.Exit(1)
	}
	rom.PrintInfos(os.Stdout)
}

func versionMain() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Fprintf(os.Stderr, "no build info")
		os.Exit(1)
	}

	key := func(key string) func(s debug.BuildSetting) bool {
		return func(s debug.BuildSetting) bool {
			return s.Key == key
		}
	}

	irev := slices.IndexFunc(info.Settings, key("vcs.revision"))
	itime := slices.IndexFunc(info.Settings, key("vcs.time"))
	if irev == -1 || itime == -1 {
		fmt.Println("dev")
		return
	}
	rev := info.Settings[irev].Value
	time := info.Settings[itime].Value[:10]
	if len(rev) > 7 {
		rev = rev[:7]
	}
	fmt.Printf("%s - %s\n", rev, time)
}
