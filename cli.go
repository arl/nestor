package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/kong"

	"nestor/emu/log"
)

type CLIConfig struct {
	CPUProfile string     `name:"cpuprofile" help:"${cpuprofile_help}" type:"path"`
	RomPath    string     `arg:"" name:"/path/to/rom" optional:"" help:"${rompath_help}" type:"existingfile"`
	Log        logModMask `help:"${log_help}" placeholder:"mod0,mod1,..."`
}

var vars = kong.Vars{
	"rompath_help":    "Run the ROM directly, skip the graphical user interface.",
	"cpuprofile_help": "Write CPU profile to file. (only when running a ROM)",
	"log_help":        "Enable logging for specified modules (comma separated).",
}

func (cfg CLIConfig) check() {
	if cfg.RomPath == "" {
		// gui mode
		if cfg.CPUProfile != "" {
			fatalf("cpuprofile option is only available when running a ROM")
		}
	}
}

type logModMask log.ModuleMask

// implements kong.MapperValue interface.
func (lm logModMask) Decode(ctx *kong.DecodeContext) error {
	tok := ctx.Scan.Pop()
	nolog := false
	allLogs := false

	for _, v := range strings.Split(tok.Value.(string), ",") {
		switch v {
		case "all":
			allLogs = true
		case "no":
			nolog = true
		default:
			mod, ok := log.ModuleByName(v)
			if !ok {
				return fmt.Errorf("unknown log module %s", v)
			}
			lm |= logModMask(mod.Mask())
		}
	}

	if nolog {
		if allLogs {
			return fmt.Errorf("cannot use 'all' and 'no' together")
		}
		if lm != 0 {
			return fmt.Errorf("cannot combine 'no' with other log modules")
		}
		log.SetOutput(io.Discard)
		return nil
	}

	if allLogs {
		lm = logModMask(log.ModuleMaskAll)
	}

	log.EnableDebugModules(log.ModuleMask(lm))
	return nil
}

func parseArgs(args []string) CLIConfig {
	var cfg CLIConfig

	parser, err := kong.New(&cfg,
		kong.Name("nestor"),
		kong.Description("NES emulator. github.com/arl/nestor"),
		kong.UsageOnError(),
		kong.Help(printHelp),
		vars)
	if err != nil {
		panic(err)
	}

	ctx, err := parser.Parse(args)
	checkf(err, "failed to parse command line")
	checkf(ctx.Error, "failed to parse command line")
	return cfg
}

func printHelp(options kong.HelpOptions, ctx *kong.Context) error {
	err := kong.DefaultHelpPrinter(options, ctx)
	if err != nil {
		return err
	}

	// Logging help
	loggingHelp := `
Log modules:
  Accepted log modules:
  %s

  As a special case, the following values are accepted: 
    - 'no'                   Disable all logging.
    - 'all'                  Enable all logs.
`
	fmt.Fprintf(os.Stderr, loggingHelp, strings.Join(log.ModuleNames(), ", "))
	return nil
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
