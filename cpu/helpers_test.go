package cpu

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"nestor/emu/hwio"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func hasPanicked(f func()) (yes bool, msg any) {
	defer func() {
		msg = recover()
		if msg != nil {
			yes = true
		}
	}()
	f()
	return yes, msg
}

/* cpu specific testing helpers */

func wantMem8(t *testing.T, cp *CPU, addr uint16, want uint8) {
	t.Helper()

	if got := cp.Read8(addr); got != want {
		t.Errorf("$%04X = %02X want %02X", addr, got, want)
	}
}

func wantMem(t *testing.T, cpu *CPU, dl dumpline) {
	t.Helper()

	mem := []byte{}
	for i := range dl.bytes {
		mem = append(mem, cpu.Read8(dl.off+uint16(i)))
	}

	if !bytes.Equal(mem, dl.bytes) {
		hd := hex.Dump(mem)
		got := hd[10 : 10+3*len(mem)]
		hd = hex.Dump(dl.bytes)
		want := hd[10 : 10+3*dl.len]
		t.Errorf("mem mismatch at 0x%04x.\ngot: %s\nwant:%s", dl.off, got, want)
	}
}

type runner interface {
	Run(int64)
}

func runAndCheckState(t *testing.T, cpu *CPU, ncycles int64, states ...any) {
	t.Helper()

	if len(states)%2 != 0 {
		panic("odd number of states")
	}

	checkbool := func(name string, got, want uint8) {
		t.Helper()
		if got != want {
			t.Errorf("got %s=%d, want %d", name, got, want)
		}
	}
	checkuint8 := func(name string, got, want uint8) {
		t.Helper()
		if got != want {
			t.Errorf("got %s=$%02X, want $%02X", name, got, want)
		}
	}
	checkuint16 := func(name string, got, want uint16) {
		t.Helper()
		if got != want {
			t.Errorf("got %s=$%04X, want $%04X", name, got, want)
		}
	}

	var r runner = cpu
	if testing.Verbose() {
		r = NewDisasm(cpu, tbwriter{t}, false)
	}

	r.Run(ncycles)

	for i := 0; i < len(states); i += 2 {
		s := states[i].(string)
		switch {
		case s == "A":
			checkuint8("A", cpu.A, states[i+1].(uint8))
		case s == "X":
			checkuint8("X", cpu.X, states[i+1].(uint8))
		case s == "Y":
			checkuint8("Y", cpu.Y, states[i+1].(uint8))
		case s == "PC":
			checkuint16("PC", cpu.PC, states[i+1].(uint16))
		case s == "SP":
			checkuint8("SP", uint8(cpu.SP), states[i+1].(uint8))
		case s == "P":
			if got, want := uint8(cpu.P), states[i+1].(uint8); got != want {
				t.Errorf("got P=$%02X(%s), want $%02X(%s)", got, P(got), want, P(want))
			}
		case len(s) > 1 && s[0] == 'P':
			for j := 1; j < len(s); j++ {
				bit := states[i+1].(uint8)
				switch s[j] {
				case 'n':
					checkbool("Pn", b2i(cpu.P.N()), bit)
				case 'v':
					checkbool("Pv", b2i(cpu.P.V()), bit)
				case 'b':
					checkbool("Pb", b2i(cpu.P.B()), bit)
				case 'd':
					checkbool("Pd", b2i(cpu.P.D()), bit)
				case 'i':
					checkbool("Pi", b2i(cpu.P.I()), bit)
				case 'z':
					checkbool("Pz", b2i(cpu.P.Z()), bit)
				case 'c':
					checkbool("Pc", b2i(cpu.P.C()), bit)
				default:
					panic("unknown P bit: " + string(s[j]))
				}
			}
		case s == "mem":
			lines := loadDump(t, states[i+1].(string))
			for _, line := range lines {
				wantMem(t, cpu, line)
			}

		default:
			panic("unknown state: " + s)
		}
	}

	if t.Failed() {
		t.FailNow()
	}
}

type dumpline struct {
	off   uint16
	len   uint16 // actual length
	bytes []byte // pow2 sized (padded with 0)
}

func loadDump(tb testing.TB, dump string) []dumpline {
	tb.Helper()

	var lines []dumpline
	scan := bufio.NewScanner(strings.NewReader(dump))
	for scan.Scan() {
		line := scan.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		off, octets, ok := strings.Cut(line, ":")
		if !ok {
			tb.Fatalf("malformed line: %s", line)
		}

		ioff, err := strconv.ParseUint(off, 16, 16)
		if err != nil {
			tb.Fatalf("malformed offset %s: %s", off, err)
		}
		var buf []byte
		for _, c := range octets {
			if c != ' ' {
				buf = append(buf, byte(c))
			}
		}
		n, err := hex.Decode(buf, buf)
		if err != nil {
			tb.Fatalf("hex decode: %s", err)
		}
		// clear the rest of the buffer
		nbytes := nextpow2(uint64(n))
		for i := uint64(n); i < nbytes; i++ {
			buf[i] = 0
		}
		dl := dumpline{off: uint16(ioff), len: uint16(nbytes), bytes: buf[:nbytes]}
		lines = append(lines, dl)

	}
	if scan.Err() != nil {
		tb.Fatalf("scan error: %s", scan.Err())
	}

	return lines
}

func nextpow2(v uint64) uint64 {
	v--
	v |= v>>1 | v>>2 | v>>4 | v>>8 | v>>16 | v>>32
	return v + 1
}

type ticker struct{}

func (tt *ticker) Tick() {}

// loadCPUWith loads a CPU with a memory dump.
func loadCPUWith(tb testing.TB, dump string) *CPU {
	mem := hwio.NewTable("cpu")
	lines := loadDump(tb, dump)
	for _, line := range lines {
		hd := hex.Dump(line.bytes)
		tb.Logf("mapping $%04X: %s", line.off, hd[10:10+3*line.len])
		mem.MapMemorySlice(line.off, line.off+uint16(len(line.bytes))-1, line.bytes, false)
	}

	cpu := NewCPU(mem, &ticker{})
	cpu.Reset()
	return cpu
}

type tbwriter struct {
	testing.TB
}

func (t tbwriter) Write(p []byte) (int, error) {
	t.TB.Helper()
	t.TB.Log(string(bytes.TrimSpace((p))))
	return len(p), nil
}

func TestLoadDump(t *testing.T) {
	tests := []struct {
		dump string
		want []dumpline
	}{
		{
			dump: `01f0: 0f 0e 0d`,
			want: []dumpline{
				{0x01f0, 3, []byte{0x0f, 0x0e, 0x0d, 0x00}},
			},
		},
		{
			dump: `01f0: 0f 0e 0d 0c 0b 0a 09 08 07 06 05 04 03 02 01 00`,
			want: []dumpline{
				{0x01f0, 16, []byte{0x0f, 0x0e, 0x0d, 0x0c, 0x0b, 0x0a, 0x09, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00}},
			},
		},
		{
			dump: `
01f0: 0f 0e 0d 0c 0b 0a 09 08 07 06 05 04 03 02 01 00
0210: 0f 0e 0d 0c 0b 0a 09 08 07 06 05 04 03 02 01 00
`,
			want: []dumpline{
				{0x01f0, 16, []byte{0x0f, 0x0e, 0x0d, 0x0c, 0x0b, 0x0a, 0x09, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00}},
				{0x0210, 16, []byte{0x0f, 0x0e, 0x0d, 0x0c, 0x0b, 0x0a, 0x09, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00}},
			},
		},
		{
			dump: `01f0: 0f 0e 0d 0c 0b 0a 09 08 07 06 05 04 03 02 01`,
			want: []dumpline{
				{0x01f0, 15, []byte{0x0f, 0x0e, 0x0d, 0x0c, 0x0b, 0x0a, 0x09, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00}},
			},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := loadDump(t, tt.dump)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d lines, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].off != tt.want[i].off {
					t.Errorf("got offset %04X, want %04X", got[i].off, tt.want[i].off)
				}
				if !bytes.Equal(got[i].bytes, tt.want[i].bytes) {
					t.Fatal(cmp.Diff(got[i].bytes, tt.want[i].bytes))
				}
			}
		})
	}
}
