package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"slices"

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
	case captureMode:
		captureMain(args.Capture)
	case versionMode:
		versionMain()
	}
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
