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

	"github.com/veandco/go-sdl2/sdl"

	"nestor/emu"
	"nestor/emu/hwio"
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
	cfg.check() // fails if necessary

	if cfg.RomPath == "" {
		guiMain()
	} else {
		emuMain(cfg)
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
func emuMain(cfg CLIConfig) {
	rom, err := ines.ReadRom(cfg.RomPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read ROM: %s", err)
		os.Exit(1)
	}

	runEmulator, err := emu.Start(rom)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start emulator: %v", err)
		os.Exit(1)
	}
	if runEmulator != nil {
		if cfg.ProfileCPU != "" {
			f, err := os.Create(cfg.ProfileCPU)
			checkf(err, "failed to create cpu profile file")
			checkf(pprof.StartCPUProfile(f), "failed to start cpu profile")
			defer func() {
				pprof.StopCPUProfile()
				f.Close()
				fmt.Println("CPU profile written to", cfg.ProfileCPU)
			}()
		}

		runEmulator()
	}
}

func mainOld() {
	ftraceLog := &outfile{}
	romInfos := false
	logflag := ""
	nologflag := false
	cpuprofile := ""
	resetVector := int64(-1)

	flag.BoolVar(&romInfos, "rominfos", false, "print infos about the iNes rom and exit")
	flag.StringVar(&logflag, "log", "", "enable logging for specified modules")
	// TODO(arl) replace with log=no
	flag.BoolVar(&nologflag, "nolog", false, "disable all logging")
	flag.Var(ftraceLog, "trace", "write cpu trace log to [file|stdout|stderr] (warning: quickly gets very big)")
	flag.Int64Var(&resetVector, "reset", -1, "overwrite CPU reset vector with (default: rom-defined)")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")

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

	if nologflag {
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

	nes, err := emu.PowerUp(rom)
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
	})
	if err := out.EnableVideo(true); err != nil {
		fatalf("failed to show emulator window: %v", err)
	}

	go func() {
		if ftraceLog.w != nil {
			emulator.Defer(func() { ftraceLog.Close() })
			nes.CPU.SetTraceOutput(ftraceLog)
		}
		nes.Run(out)
	}()

	// TODO: gtk3
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	if err := ui.ShowMainWindow(); err != nil {
		fatalf("failed to show main window: %v", err)
	}
	_ = ctx
	// emulator.Run(ctx, nes)
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
