package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"

	"nestor/emu/hwio"
	log "nestor/emu/logger"
	"nestor/hw"
	"nestor/ines"
)

func main() {
	disasmLog := &outfile{}
	romInfos := false
	flagLogging := ""
	cpuprofile := ""
	resetVector := int64(-1)

	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.BoolVar(&romInfos, "infos", false, "print infos about the rom and exit")
	flag.StringVar(&flagLogging, "log", "", "enable logging for specified modules")
	flag.Var(disasmLog, "execlog", "write execution log to [file|stdout|stderr] (very very verbose")
	flag.Int64Var(&resetVector, "reset", -1, "overwrite reset vector (default: use reset vector from rom")

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

	if flagLogging != "" {
		var modmask log.ModuleMask
		for _, modname := range strings.Split(flagLogging, ",") {
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

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.ModEmu.FatalZ(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		if cpuprofile != "" {
			pprof.StopCPUProfile()
			log.ModEmu.Infof("CPU profile written to %s", cpuprofile)
		}
		os.Exit(1)
	}()

	var nes NES
	checkf(nes.PowerUp(rom), "error during power up")
	if resetVector != -1 {
		hwio.Write16(nes.Hw.CPU.Bus, hw.CpuResetVector, uint16(resetVector))
	}
	nes.Reset()

	go func() {
		if disasmLog.w != nil {
			defer disasmLog.Close()
			nes.RunDisasm(disasmLog)
		} else {
			nes.Run()
		}
	}()

	ui := newGUI(&nes)
	ui.run()
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "fatal error:")
	fmt.Fprintf(os.Stderr, "\n\t%s\n", err)
	os.Exit(1)
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
