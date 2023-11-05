package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"nestor/emu"
	"nestor/ines"
)

func main() {
	hexbyte := hexbyte(0x34)
	flag.Var(&hexbyte, "P", "P register after first cpu reset (hex)")
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
	nes.CPU.P = emu.P(hexbyte)
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

type hexbyte byte

func (b *hexbyte) Set(s string) error {
	str := strings.TrimPrefix(s, "0x")
	v, err := strconv.ParseUint(str, 16, 8)
	if err != nil {
		return fmt.Errorf("hexbyte: can't parse %v: %s", v, err)
	}
	*b = hexbyte(v)
	return nil
}

func (b *hexbyte) String() string { return fmt.Sprintf("0x%02X", *b) }
