package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	log "nestor/emu/logger"
	"nestor/ines"
)

func main() {
	disasmLog := &outfile{}
	infosFlag := false
	flagLogging := ""

	flag.BoolVar(&infosFlag, "infos", false, "print infos about the rom and exit")
	flag.StringVar(&flagLogging, "log", "", "enable logging for specified modules")
	flag.Var(disasmLog, "execlog", "write execution log to [file|stdout|stderr] (very very verbose")

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

	if infosFlag {
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

	nes := &NES{}
	checkf(nes.PowerUp(rom), "error during power up")
	nes.Reset()

	go func() {
		if disasmLog.w != nil {
			defer disasmLog.Close()
			nes.RunDisasm(disasmLog, false)
		} else {
			nes.Run()
		}
	}()

	startScreen(nes)
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
