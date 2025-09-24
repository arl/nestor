package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"runtime/pprof"
	"slices"
	"strings"

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
    -trace FILE        write cpu trace log to FILE,
                       also accepts: stdout or stderr
    -v, --verbose      enable verbose logging (set info log level). Default is warning.
    -log               comma separated list of log modules to enable in debug mode, from:
                       %s
                       accepts also 'all' or 'no' (disable log entirely).

`
	fmt.Fprintf(os.Stderr, help, strings.Join(log.ModuleNames(), ", "))
}

func main() {
	var (
		// general options
		versionFlag  bool
		rominfosFlag string

		// cli options
		cpuprofileFlag string
		traceFlag      outfile
		ramfileFlag    existingFile
		logModules     logModMask
		verboseFlag    bool
	)

	fs := flag.NewFlagSet("nestor", flag.ContinueOnError)

	fs.BoolVar(&versionFlag, "version", false, "print version and exit")
	fs.StringVar(&rominfosFlag, "rom-infos", "", "print ROM information and exit")

	fs.StringVar(&cpuprofileFlag, "cpuprofile", "", "write cpu profile to file")
	fs.Var(&traceFlag, "trace", "write cpu trace log to FILE|stdout|stderr")
	fs.Var(&ramfileFlag, "ramfile", "Read 'save ram' from file [WIP/TODO].")
	fs.Var(&logModules, "log", "comma separated list of log modules to enable (or 'all' or 'no').")
	fs.BoolVar(&verboseFlag, "v", false, "enable verbose logging (set info log level)")
	fs.BoolVar(&verboseFlag, "verbose", false, "enable verbose logging (set info log level)")
	fs.Usage = usage
	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return
		}
		os.Exit(1)
	}

	if versionFlag {
		printVersion()
		return
	}

	if verboseFlag {
		log.SetLevel(log.InfoLevel)
	}

	cfg := config.LoadConfigOrDefault()

	if traceFlag.name != "" {
		cfg.TraceOut = &traceFlag
		defer traceFlag.Close()
	}

	if cpuprofileFlag != "" {
		f, err := os.Create(cpuprofileFlag)
		checkf(err, "failed to create cpu profile file")
		checkf(pprof.StartCPUProfile(f), "failed to start cpu profile")
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
			fmt.Fprint(os.Stderr, "CPU profile written to", cpuprofileFlag)
		}()
	}

	if rominfosFlag != "" {
		if err := printRomInfos(rominfosFlag); err != nil {
			fatalf("failed to print ROM infos: %v", err)
		}
		return
	}

	var romPath = "" // empty -> starts gui

	switch fs.NArg() {
	case 0:
	case 1:
		romPath = fs.Arg(0) // run that ROM
	default:
		usage()
		os.Exit(1)
	}

	if romPath == "" {
		checkf(ui.StartUI(cfg), "can't start ui")
	} else {
		checkf(ui.StartROM(cfg, romPath), "can't start rom")
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
