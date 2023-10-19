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
	cartridge, err := ines.ReadRom(path)
	checkf(err, "failed to open rom %s", path)

	nes, err := bootNES(cartridge)
	checkf(err, "failed to boot nes")

	nes.Run()
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "fatal error:\n")
	fmt.Fprintf(os.Stderr, "\n%s\n", err)
	os.Exit(1)
}

func checkf(err error, format string, args ...any) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "fatal error:\n")
	fmt.Fprintf(os.Stderr, "\n%s: %s\n", fmt.Sprintf(format, args...), err)
	os.Exit(1)
}
