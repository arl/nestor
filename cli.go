package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"nestor/emu/log"
)

// logModMask implements flag.Value, providing an easy way to enable/disable log
// modules.
type logModMask log.ModuleMask

func (lm *logModMask) String() string {
	if *lm == 0 {
		return "no"
	}
	var names []string
	for i := range 64 {
		if *lm&(1<<i) != 0 {
			if int(i) <= len(log.ModuleNames()) {
				names = append(names, log.ModuleNames()[i-1])
			}
		}
	}
	return strings.Join(names, ",")
}

func (lm *logModMask) Set(val string) error {
	switch val {
	case "no":
		log.Disable()
	case "all":
		log.EnableDebugModules(log.ModuleMaskAll)
	default:
		for v := range strings.SplitSeq(val, ",") {
			mod, ok := log.ModuleByName(v)
			if !ok {
				return fmt.Errorf("unknown log module name %q", v)
			}
			*lm |= logModMask(mod.Mask())
		}

		log.EnableDebugModules(log.ModuleMask(*lm))
	}

	return nil
}

// existingFile implements flag.Value, accepting only paths to files that exist.
type existingFile string

func (v *existingFile) String() string {
	return string(*v)
}

func (v *existingFile) Set(val string) error {
	if _, err := os.Stat(val); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %q", val)
	}
	*v = existingFile(val)
	return nil
}

// outfile implements flag.Value, accepting stdout, stderr or the path to a file
// that will be open for writing and closed with Close.
type outfile struct {
	w     io.Writer
	name  string
	close func() error
}

func (f *outfile) Set(val string) error {
	if f == nil {
		f = &outfile{}
	}
	f.name = val
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

func (f *outfile) String() string {
	if f == nil {
		return ""
	}
	return f.name
}
func (f *outfile) Write(p []byte) (int, error) { return f.w.Write(p) }
func (f *outfile) Close() error                { return f.close() }

type duration time.Duration

func (d *duration) String() string {
	return time.Duration(*d).String()
}

func (d *duration) Set(val string) error {
	dur, err := time.ParseDuration(val)
	if err != nil {
		return err
	}

	*d = duration(dur)
	return nil
}
