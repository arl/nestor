package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"io"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"

	"nestor/emu"
	"nestor/emu/hwio"
	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
)

func main() {
	ftraceLog := &outfile{}
	romInfos := false
	logflag := ""
	cpuprofile := ""
	resetVector := int64(-1)
	dbgAddr := ""

	flag.BoolVar(&romInfos, "rominfos", false, "print infos about the iNes rom and exit")
	flag.StringVar(&logflag, "log", "", "enable logging for specified modules (no: disable all logging)")
	flag.Var(ftraceLog, "trace", "write cpu trace log to [file|stdout|stderr] (warning: quickly gets very big)")
	flag.Int64Var(&resetVector, "reset", -1, "overwrite CPU reset vector with (-1: rom-defined)")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&dbgAddr, "dbg", "", "connect to debugger at [host]:port (default: disabled)")

	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
		return
	}

	path := flag.Arg(0)
	rom, err := ines.ReadRom(path)
	checkf(err, "failed to open rom")
	if rom.IsNES20() {
		fatalf("nes 2.0 roms are not supported yet")
	}

	if romInfos {
		rom.PrintInfos(os.Stdout)
		return
	}

	if logflag == "no" {
		log.SetOutput(io.Discard)
	} else if logflag != "" {
		var modmask log.ModuleMask
		for _, modname := range strings.Split(logflag, ",") {
			if modname == "all" {
				modmask |= log.ModuleMaskAll
			} else if m, found := log.ModuleByName(modname); found {
				modmask |= m.Mask()
			} else {
				log.ModEmu.FatalZ("invalid module name").String("name", modname).End()
			}
		}
		log.EnableDebugModules(modmask)
	}

	nes, err := emu.PowerUp(rom, dbgAddr)
	checkf(err, "error during power up")

	emulator := emu.NewEmulator(nes)

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		checkf(err, "failed to create cpu profile file")
		checkf(pprof.StartCPUProfile(f), "failed to start cpu profile")
		emulator.Defer(func() {
			pprof.StopCPUProfile()
			f.Close()
			fmt.Println("CPU profile written to", cpuprofile)
		})
	}

	if resetVector != -1 {
		hwio.Write16(nes.CPU.Bus, hw.ResetVector, uint16(resetVector))
	}

	// Input setup
	pads := emu.StdControllerPair{
		Pad1Connected: true,
	}
	emulator.ConnectInputDevice(&pads)

	// Output setup
	nes.Frames = make(chan image.RGBA)
	out := hw.NewOutput(hw.OutputConfig{
		Width:           256,
		Height:          240,
		NumVideoBuffers: 2,
		FrameOutCh:      nes.Frames,
	})

	go func() {
		if ftraceLog.w != nil {
			emulator.Defer(func() { ftraceLog.Close() })
			nes.CPU.SetTraceOutput(ftraceLog)
		}
		nes.Run(out)
	}()

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	emulator.Run(ctx, nes)
}

func checkf(err error, format string, args ...any) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "fatal error:")
	fmt.Fprintf(os.Stderr, "\n\t%s: %s\n", fmt.Sprintf(format, args...), err)
	os.Exit(1)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fatal error:")
	fmt.Fprintf(os.Stderr, "\n\t%s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}

type outfile struct {
	w    io.Writer
	name string
}

func (f *outfile) Set(s string) error {
	f.name = s
	switch s {
	case "stdout":
		f.w = os.Stdout
	case "stderr":
		f.w = os.Stderr
	default:
		fd, err := os.Create(s)
		if err != nil {
			return err
		}
		f.w = fd
	}
	return nil
}

func (f *outfile) String() string              { return f.name }
func (f *outfile) Write(p []byte) (int, error) { return f.w.Write(p) }
func (f *outfile) Close() error {
	if f.name == "stdout" || f.name == "stderr" {
		return nil
	}
	return f.w.(io.Closer).Close()
}
