package main

import (
	"fmt"
	"io"
	"os"
	"runtime/pprof"

	"github.com/veandco/go-sdl2/sdl"

	"nestor/cli"
	"nestor/emu"
	"nestor/emu/rpc"
	"nestor/hw/input"
	"nestor/ines"
	"nestor/ui"
)

// emuMain runs the emulator directly with the given rom.
func emuMain(args cli.Run, cfg *ui.Config) {
	var exitcode int
	sdl.Main(func() {
		rom, err := ines.ReadROM(args.RomPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read rom: %s", err)
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

		if args.RAMFile != "" {
			saveram, err := os.ReadFile(args.RAMFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to load 'save ram' file: %v", err)
				exitcode = 1
				return
			}
			if err := emulator.NES.Mapper.SetBatteryPackedRAM(saveram); err != nil {
				fmt.Fprintf(os.Stderr, "failed to assign 'save ram' to ROM: %v", err)
				exitcode = 1
				return
			}
		}

		tmpdir, err := os.MkdirTemp("", "nestor.out.*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create nestor out temp directory: %v", err)
			exitcode = 1
			return
		}
		emulator.SetTempDir(tmpdir)

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

func captureMain(args cli.Capture) {
	var (
		code input.Code
		err  error
	)

	sdl.Main(func() {
		sdl.Do(func() {
			if code, err = input.StartCapture(args.Monitor, args.Button); err != nil {
				err = fmt.Errorf("error capturing input: %v", err)
			}
		})
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal text error: %v", err)
		os.Exit(1)
	}

	out, err := code.MarshalText()
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal text error: %v", err)
		os.Exit(1)
	}

	fmt.Printf("%s", out)
	os.Exit(0)
}
