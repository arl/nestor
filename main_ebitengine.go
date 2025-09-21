package main

import (
	"fmt"
	"os"

	"nestor/ebit"
)

func main() {
	romPath, cfg, cleanup, err := parseArguments()
	defer cleanup()
	_ = cfg
	checkf(err, "failed to parse command line")

	ebit.StartROM(romPath)
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
