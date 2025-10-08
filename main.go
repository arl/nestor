package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"runtime/pprof"
	"slices"
	"strings"

	"github.com/arl/statsviz"

	"nestor/config"
	"nestor/emu/log"
	"nestor/ines"
	"nestor/ui"
)

func usage() {
	const help = `NES emulator. github.com/arl/nestor

Usage:

    nestor [options] [/path/to/rom]

Run Nestor:

    nestor                    gui mode.

    nestor <path/to/rom>      starts ROM directly. 

Options:
    -version           print version and exit.
    -rom-infos FILE    print information about the ROM and exit.

    -ramfile FILE      read 'save ram' from file [WIP/TODO].

    -cpuprofile FILE   write cpu profile to FILE.
    -monitor HOST:PORT expose Go runtime real-time on HOST:PORT
    -trace FILE        write cpu trace log to FILE,
                       also accepts: stdout or stderr
    -v, --verbose      set log level to info (default: warning).
    -log               comma separated list of log modules to enable in debug mode, from:
                       %s
                       accepts also 'all' or 'no' (disable log entirely).

`
	fmt.Fprintf(os.Stderr, help, strings.Join(log.ModuleNames(), ", "))
}

func main() {
	var (
		// general options
		version  bool
		rominfos string

		// cli options
		cpuprofile string
		trace      outfile
		ramfile    existingFile
		logModules logModMask
		verbose    bool
		monitor    string
	)

	fs := flag.NewFlagSet("nestor", flag.ContinueOnError)

	fs.BoolVar(&version, "version", false, "print version and exit")
	fs.StringVar(&rominfos, "rom-infos", "", "print ROM information and exit")

	fs.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	fs.StringVar(&monitor, "monitor", "", "expose Go runtime real-time on HOST:PORT")
	fs.Var(&trace, "trace", "write cpu trace log to FILE|stdout|stderr")
	fs.Var(&ramfile, "ramfile", "Read 'save ram' from file [WIP/TODO].")
	fs.Var(&logModules, "log", "comma separated list of log modules to enable (or 'all' or 'no').")
	fs.BoolVar(&verbose, "v", false, "enable verbose logging (set info log level)")
	fs.BoolVar(&verbose, "verbose", false, "enable verbose logging (set info log level)")
	fs.Usage = usage
	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return
		}
		os.Exit(1)
	}

	if version {
		printVersion()
		return
	}

	if verbose {
		log.SetLevel(log.InfoLevel)
	}


	if trace.name != "" {
		cfg.TraceOut = &trace
		defer trace.Close()
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		checkf(err, "failed to create cpu profile file")
		checkf(pprof.StartCPUProfile(f), "failed to start cpu profile")
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
			fmt.Fprintf(os.Stderr, "CPU profile written to %q", cpuprofile)
		}()
	}

	if monitor != "" {
		ss, err := statsviz.NewServer()
		checkf(err, "statsviz")
		ss.Register(http.DefaultServeMux)

		fmt.Println("statsviz UI: point your browser to", prettyAddr(monitor)+"/debug/statsviz")
		go http.ListenAndServe(monitor, nil)
	}

	if rominfos != "" {
		if err := printRomInfos(rominfos); err != nil {
			fatalf("failed to print ROM infos: %v", err)
		}
		return
	cfg := config.LoadOrDefault()
	}

	installStacktraceHandler()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	switch fs.NArg() {
	case 0:
		checkf(ui.StartUI(ctx, cfg), "failed to start ui")
	case 1:
		checkf(ui.StartROM(ctx, cfg, fs.Arg(0)), "failed to start rom")
	default:
		usage()
		os.Exit(1)
	}

	// if args.RAMFile != "" {
	// 	saveram, err := os.ReadFile(args.RAMFile)
	// 	if err != nil {
	// 		fmt.Fprintf(os.Stderr, "failed to load 'save ram' file: %v", err)
	// 		exitcode = 1
	// 		return
	// 	}
	// 	if err := emulator.NES.Mapper.SetBatteryPackedRAM(saveram); err != nil {
	// 		fmt.Fprintf(os.Stderr, "failed to assign 'save ram' to ROM: %v", err)
	// 		exitcode = 1
	// 		return
	// 	}
	// }

	// tmpdir, err := os.MkdirTemp("", "nestor.out.*")
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to create nestor out temp directory: %v", err)
	// 	exitcode = 1
	// 	return
	// }
	// emulator.SetTempDir(tmpdir)

	// if args.CPUProfile != "" {
}

func printRomInfos(romPath string) error {
	rom, err := ines.ReadROM(romPath)
	if err != nil {
		return fmt.Errorf("error reading ROM: %s", err)
	}

	rom.PrintInfos(os.Stdout)
	return nil
}

func printVersion() {
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

func checkf(err error, format string, args ...any) {
	if err == nil {
		return
	}
	fatalf(format+".\n"+err.Error(), args...)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fatal error:")
	fmt.Fprintf(os.Stderr, "\n\t%s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}

func prettyAddr(addr string) string {
	if addr == "" {
		return "localhost"
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	if host == "" {
		host = "localhost"
	}
	return fmt.Sprintf("http://%s:%s", host, port)
}
