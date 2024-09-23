package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/kong"

	"nestor/emu/log"
)

type mode byte

const (
	guiMode      mode = iota // Start Nestor GUI
	runMode                  // Just run a ROM
	mapInputMode             // keyboard/joystick mapping window
)

type (
	CLI struct {
		GUI      GUI      `cmd:"" help:"Run Nestor graphical user interface. The default if no commands are given." default:"1"`
		Run      Run      `cmd:"" help:"Run ROM in emulator."`
		MapInput MapInput `cmd:"" name:"map-input" hidden:""`

		mode mode
	}

	GUI      struct{}
	MapInput struct {
		Button string `name:"button" help:"NES button to map." required:""`
	}

	Run struct {
		CPUProfile string     `name:"cpuprofile" help:"${cpuprofile_help}" type:"path"`
		RomPath    string     `arg:"" name:"/path/to/rom" optional:"" help:"${rompath_help}" type:"existingfile"`
		Log        logModMask `help:"${log_help}" placeholder:"mod0,mod1,..."`
		Infos      bool       `name:"infos" help:"Show information about the ROM and exit."`
		Trace      *outfile   `name:"trace" help:"Write CPU trace log." placeholder:"FILE|stdout|stderr"`
	}
)

var vars = kong.Vars{
	"rompath_help":    "Run the ROM directly, skip the graphical user interface.",
	"cpuprofile_help": "Write CPU profile to file. (only when running a ROM)",
	"log_help":        "Enable logging for specified modules. See 'Log Modules'.",
}

func (cfg CLI) validate() {
	switch cfg.mode {
	case runMode:
		cfg := cfg.Run
		if cfg.RomPath == "" {
			// gui mode
			if cfg.CPUProfile != "" {
				fatalf("cpuprofile option is only available when running a ROM")
			}
		}
	}
}

func parseArgs(args []string) CLI {
	var cfg CLI
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

	switch ctx.Command() {
	case "gui":
		cfg.mode = guiMode
	case "map-input":
		cfg.mode = mapInputMode
	default:
		cfg.mode = runMode
	}
	return cfg
}

func printHelp(options kong.HelpOptions, ctx *kong.Context) error {
	if err := kong.DefaultHelpPrinter(options, ctx); err != nil {
		return err
	}
	if strings.HasPrefix(ctx.Command(), "run") {
		loggingHelp := `
Log modules:
  The --log flag accepts a comma-separated list of modules.

  Valid log modules are:
%s
  
  As a special case, the following values are accepted: 
    - no                     Disable all logging.
    - all                    Enable all logs.
`
		var strs []string
		for _, m := range log.ModuleNames() {
			strs = append(strs, "    - "+m)
		}

		fmt.Fprintf(os.Stderr, loggingHelp, strings.Join(strs, "\n"))
	}

	return nil
}

type logModMask log.ModuleMask

// Decode decodes a comma-separated list of module names into a module mask.
//
// Implements kong.MapperValue interface.
func (lm logModMask) Decode(ctx *kong.DecodeContext) error {
	nolog := false
	allLogs := false

	tok := ctx.Scan.Pop()
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

type outfile struct {
	w     io.Writer
	name  string
	close func() error
}

// Decode decodes FILE|stdout|stderr into an io.WriteCloser
// that writes to that file.
//
// Implements kong.MapperValue interface.
func (f *outfile) Decode(ctx *kong.DecodeContext) error {
	tok := ctx.Scan.Pop()
	f.name = tok.Value.(string)
	f.close = func() error { return nil }

	switch f.name {
	case "stdout":
		f.w = os.Stdout
	case "stderr":
		f.w = os.Stderr
	default:
		fd, err := os.Create(f.name)
		if err != nil {
			return err
		}
		f.w = fd
		f.close = fd.Close
	}
	return nil
}

func (f *outfile) String() string              { return f.name }
func (f *outfile) Write(p []byte) (int, error) { return f.w.Write(p) }
func (f *outfile) Close() error                { return f.close() }

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
