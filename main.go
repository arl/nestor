package main

import (
	"flag"
	"fmt"
	"nestor/ines"
	"os"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
		return
	}

	path := flag.Arg(0)
	cart, err := ines.LoadCartridge(path)
	checkf(err, "failed to open rom %s", path)

	nes, err := startNES(cart)
	checkf(err, "failed to start nes")

	_ = nes
	println("yay!")
}

func checkf(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "fatal error:\n")
	fmt.Fprintf(os.Stderr, "\n%s: %s\n", fmt.Sprintf(format, args...), err)
	os.Exit(1)
}
