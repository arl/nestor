package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

/* general testing helpers */

func tcheck(tb testing.TB, err error) {
	if err == nil {
		return
	}

	tb.Helper()
	tb.Fatalf("fatal error:\n\n%s\n", err)
}

func tcheckf(tb testing.TB, err error, format string, args ...any) {
	if err == nil {
		return
	}

	tb.Helper()
	tb.Fatalf("fatal error:\n\n%s: %s\n", fmt.Sprintf(format, args...), err)
}

/* cpu specific testing helpers */

func wantMem8(t *testing.T, cpu *CPU, addr uint16, want uint8) {
	t.Helper()

	if got := cpu.Read8(addr); got != want {
		t.Errorf("$%04X = %02X want %02X", addr, got, want)
	}
}

func wantCPUState(t *testing.T, cpu *CPU, states ...any) {
	t.Helper()

	if len(states)%2 != 0 {
		panic("odd number of states")
	}

	checkbool := func(name string, got, want uint8) {
		if got != want {
			t.Errorf("got %s=%d, want %d", name, got, want)
		}
	}
	checkuint8 := func(name string, got, want uint8) {
		if got != want {
			t.Errorf("got %s=$%02X, want $%02X", name, got, want)
		}
	}
	checkuint16 := func(name string, got, want uint16) {
		if got != want {
			t.Errorf("got %s=$%04X, want $%04X", name, got, want)
		}
	}

	for i := 0; i < len(states); i += 2 {
		s := states[i].(string)
		v := states[i+1].(int)
		switch {
		case s == "A":
			checkuint8("A", cpu.A, uint8(v))
		case s == "X":
			checkuint8("X", cpu.X, uint8(v))
		case s == "Y":
			checkuint8("Y", cpu.Y, uint8(v))
		case s == "PC":
			checkuint16("PC", cpu.PC, uint16(v))
		case s == "SP":
			checkuint8("SP", uint8(cpu.SP), uint8(v))
		case len(s) > 1 && s[0] == 'P':
			for i := 1; i < len(s); i++ {
				switch s[i] {
				case 'n':
					checkbool("Pn", b2i(cpu.P.N()), uint8(v))
				case 'v':
					checkbool("Pv", b2i(cpu.P.V()), uint8(v))
				case 'b':
					checkbool("Pb", b2i(cpu.P.B()), uint8(v))
				case 'd':
					checkbool("Pd", b2i(cpu.P.D()), uint8(v))
				case 'i':
					checkbool("Pi", b2i(cpu.P.I()), uint8(v))
				case 'z':
					checkbool("Pz", b2i(cpu.P.Z()), uint8(v))
				case 'c':
					checkbool("Pc", b2i(cpu.P.C()), uint8(v))
				default:
					panic("unknown P bit: " + string(s[i]))
				}
			}
		default:
			panic("unknown state: " + s)
		}
	}

	if t.Failed() {
		t.FailNow()
	}
}

// loadCPUWith loads a CPU with a memory dump.
func loadCPUWith(t *testing.T, dump string) *CPU {
	bus := newCpuBus("cpu")
	scan := bufio.NewScanner(strings.NewReader(dump))
	for scan.Scan() {
		line := scan.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		off, mem, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("malformed line: %s", line)
		}

		ioff, err := strconv.ParseUint(off, 16, 16)
		if err != nil {
			t.Fatalf("offset %s: %s", off, err)
		}
		buf := make([]byte, 0, 32)
		for _, c := range mem {
			if c != ' ' {
				buf = append(buf, byte(c))
			}
		}
		n, err := hex.Decode(buf, buf)
		if err != nil {
			t.Fatalf("hex decode: %s", err)
		}
		// clear the rest of the buffer
		for i := n; i < len(buf); i++ {
			buf[i] = 0
		}
		bus.MapSlice(uint16(ioff), uint16(ioff+15), buf[:16])
		t.Logf("mapping $%04X-$%04X with %s", ioff, ioff+15, hex.Dump(buf[:16]))
	}

	cpu := NewCPU(bus)
	cpu.reset()
	if scan.Err() != nil {
		t.Fatalf("scan error: %s", scan.Err())
	}

	return cpu
}
