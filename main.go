//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"slices"

	"nestor/cli"
	"nestor/ines"
	"nestor/ui"
)

func main() {
	args, err := cli.ParseCmdLineArgs(os.Args[1:])
	checkf(err, "failed to parse command line")

	cfg := ui.LoadConfigOrDefault()

	switch args.Mode {
	case cli.GUIMode:
		ui.RunApp(&cfg)
	case cli.ROMInfosMode:
		romInfosMain(args.RomInfos.RomPath)
	case cli.RunMode:
		emuMain(args.Run, &cfg)
	case cli.CaptureMode:
		captureMain(args.Capture)
	case cli.VersionMode:
		versionMain()
	}
}

func romInfosMain(romPath string) {
	rom, err := ines.ReadROM(romPath)
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
