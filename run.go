package main

import (
	"fmt"
	"io"
	"os"
	"runtime/pprof"

	"github.com/veandco/go-sdl2/sdl"

	"nestor/emu"
	"nestor/emu/rpc"
	"nestor/hw/input"
	"nestor/ines"
	"nestor/ui"
)

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

		emulator, err := emu.Launch(rom, cfg.Config)
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

		if args.Port != 0 {
			fmt.Println("creating rpc server", args.Port)
			server, err := rpc.NewServer(args.Port, emulator)
			if err != nil {
				fmt.Fprintf(os.Stderr, "RPC error: %v", err)
				exitcode = 1
				return
			}
			defer server.Close()
		}

		emulator.Run()
	})
	os.Exit(exitcode)
}

func captureMain(args Capture) {
	sdl.Main(func() {
		code, err := input.StartCapture(args.Monitor, args.Button)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error capturing input: %v", err)
			os.Exit(1)
		}
		out, err := code.MarshalText()
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal text error: %v", err)
			os.Exit(1)
		}

		fmt.Printf("%s", out)
		os.Exit(0)
	})
}
