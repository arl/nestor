package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"nestor/cpu"
	"nestor/ines"
)

func main() {
	hexbyte := hexbyte(0x34)
	disasmLog := new(outfile)

	flag.Var(&hexbyte, "P", "P register after first cpu reset (hex)")
	flag.Var(disasmLog, "dbglog", "write execution log to [file|stdout|stderr] (for testing/debugging")
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
	checkf(nes.PowerUp(rom), "error during power up")

	nes.Reset()
	nes.CPU.P = cpu.P(hexbyte)
	if disasmLog.w != nil {
		defer disasmLog.Close()
		nes.CPU.SetDisasm(disasmLog, false)
	}

	nes.Run()
	startScreen(nes)
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "fatal error:")
	fmt.Fprintf(os.Stderr, "\n\t%s\n", err)
	os.Exit(1)
}

func checkf(err error, format string, args ...any) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "fatal error:")
	fmt.Fprintf(os.Stderr, "\n\t%s: %s\n", fmt.Sprintf(format, args...), err)
	os.Exit(1)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fatal error:")
	fmt.Fprintf(os.Stderr, "\n\t%s\n", fmt.Sprintf(format, args...))
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

type outfile struct {
	w    io.Writer
	name string
}

func (f *outfile) Set(s string) error {
	f.name = s
	switch s {
	case "stdout":
		f.w = os.Stdout
	case "stderr":
		f.w = os.Stderr
	default:
		fd, err := os.Create(s)
		if err != nil {
			return err
		}
		f.w = fd
	}
	return nil
}

func (f *outfile) String() string              { return f.name }
func (f *outfile) Write(p []byte) (int, error) { return f.w.Write(p) }
func (f *outfile) Close() error {
	if f.name == "stdout" || f.name == "stderr" {
		return nil
	}
	return f.w.(io.Closer).Close()
}
