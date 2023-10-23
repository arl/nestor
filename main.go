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
	rom, err := ines.ReadRom(path)
	if rom.IsNES20() {
		fatalf("nes 2.0 roms are not supported yet")
	}
	checkf(err, "failed to open rom %s", path)

	nes := new(NES)
	checkf(nes.PowerUp(rom), "failed to power up")
	nes.Reset()
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

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fatal error:\n")
	fmt.Fprintf(os.Stderr, "\n%s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}
