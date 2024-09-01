package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

type CLIConfig struct {
	ProfileCPU string `name:"cpuprofile" help:"${cpuprofile_help}" type:"path"`
	RomPath    string `arg:"" name:"/path/to/rom" optional:"" help:"${rompath_help}" type:"path"`
}

func (cfg CLIConfig) check() {
	if cfg.RomPath == "" {
		// gui mode
		if cfg.ProfileCPU != "" {
			fatalf("cpuprofile option is only available when running a ROM")
		}
	}
}

func parseArgs(args []string) CLIConfig {
	var vars = kong.Vars{
		"rompath_help":    "Run the ROM directly, skip the graphical user interface.",
		"cpuprofile_help": "Write CPU profile to file. (only when running a ROM)",
	}

	var cfg CLIConfig
	parser, err := kong.New(&cfg,
		kong.Name("nestor"),
		kong.Description("NES emulator. github.com/arl/nestor"),
		kong.UsageOnError(),
		vars)
	if err != nil {
		panic(err)
	}

	ctx, err := parser.Parse(args)
	checkf(err, "failed to parse command line")
	checkf(ctx.Error, "failed to parse command line")
	return cfg
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
