package cpu

import (
	"encoding/json"
	"fmt"
	"nestor/emu/hwio"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestAllOpcodesAreImplemented(t *testing.T) {
	for opcode, op := range ops {
		if op == nil {
			t.Errorf("opcode %02x not implemented", opcode)
		}
	}
}

func funcname(temp any) string {
	strs := strings.Split((runtime.FuncForPC(reflect.ValueOf(temp).Pointer()).Name()), ".")
	return strs[len(strs)-1]
}

func TestOpcodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	// Run tests for all implemented opcodes.
	for opcode := range ops {
		opstr := fmt.Sprintf("%02x", opcode)
		switch {
		case unstableOps[uint8(opcode)] == 1:
			t.Run(opstr, func(t *testing.T) { t.Skipf("skipping unsupported opcode") })
		default:
			t.Run(opstr, testOpcodes(opstr))
		}
	}
}

var slicePool = sync.Pool{
	New: func() any {
		s := make([]uint8, 0x10000)
		return &s
	},
}

func newSlice() *[]uint8 {
	return slicePool.Get().(*[]uint8)
}

func putSlice(s *[]uint8) {
	clear(*s)
	slicePool.Put(s)
}

// testOpcodes runs the opcode tests in testdata/<op>.json
// these comes from github.com/TomHarte/ProcessorTests/blob/main/nes6502.
func testOpcodes(op string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		path := filepath.Join("testdata", "tomharte.processor.tests", "v1", op+".json")
		buf, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var tests []struct {
			Name    string `json:"name"`
			Initial struct {
				PC  uint16     `json:"pc"`
				SP  uint8      `json:"s"`
				A   uint8      `json:"a"`
				X   uint8      `json:"x"`
				Y   uint8      `json:"y"`
				P   uint8      `json:"p"`
				RAM [][]uint16 `json:"ram"`
			} `json:"initial"`
			Final struct {
				PC  uint16     `json:"pc"`
				SP  uint8      `json:"s"`
				A   uint8      `json:"a"`
				X   uint8      `json:"x"`
				Y   uint8      `json:"y"`
				P   uint8      `json:"p"`
				RAM [][]uint16 `json:"ram"`
			} `json:"final"`
			Cycles [][]any `json:"cycles"`
		}

		if err := json.Unmarshal(buf, &tests); err != nil {
			t.Fatal(err)
		}

		for _, tt := range tests {
			t.Run(tt.Name, func(t *testing.T) {
				mem := hwio.NewTable("cputest")
				slice := newSlice()
				defer putSlice(slice)

				mem.MapMem(0x0000, &hwio.Mem{
					Data:  *slice,
					Flags: hwio.MemFlag8,
					VSize: int(0x10000),
				})

				cpu := NewCPU(mem, &ticker{})
				cpu.A = tt.Initial.A
				cpu.X = tt.Initial.X
				cpu.Y = tt.Initial.Y
				cpu.P = P(tt.Initial.P)
				cpu.SP = tt.Initial.SP
				cpu.PC = tt.Initial.PC

				// preload RAM
				for _, row := range tt.Initial.RAM {
					mem.Write8(row[0], uint8(row[1]))
				}

				if testing.Verbose() {
					t.Logf("initial {A=0x%02x X=0x%02x Y=0x%02x P=0x%02x(%s) SP=0x%02x PC=0x%04x}\n",
						cpu.A, cpu.X, cpu.Y, uint8(cpu.P), cpu.P.String(), cpu.SP, cpu.PC)
					t.Log("run:")
				}

				if testing.Verbose() {
					t.Logf("expecting cycles:\n%s\n\n", strings.Join(prettyCycles(tt.Cycles), "\n"))
					t.Log("test output:")
				}

				// check cpu state
				runAndCheckState(t, cpu, int64(len(tt.Cycles))-1,
					"PC", tt.Final.PC,
					"SP", tt.Final.SP,
					"A", tt.Final.A,
					"X", tt.Final.X,
					"Y", tt.Final.Y,
					"P", tt.Final.P,
				)

				// check cycles
				if len(tt.Cycles) != int(cpu.Clock) {
					t.Errorf("cycles count mismatch: got %d want %d\ndebug:\n%s", cpu.Clock, len(tt.Cycles), strings.Join(prettyCycles(tt.Cycles), "\n"))
				}

				// check ram
				for _, row := range tt.Final.RAM {
					addr := row[0]
					val := uint8(row[1])
					got := mem.Read8(addr)
					if got != val {
						t.Errorf("ram[0x%x] = 0x%x, want 0x%x", addr, got, val)
					}
				}
			})
		}
	}
}

func prettyCycles(cycles [][]any) []string {
	strs := make([]string, len(cycles))
	for i, row := range cycles {
		addr := int(row[0].(float64))
		val := int(row[1].(float64))
		strs[i] = fmt.Sprintf("%s 0x%04x = 0x%02x", row[2], addr, val)
	}
	return strs
}

func TestCPx(t *testing.T) {
	t.Run("40 - 41", func(t *testing.T) {
		// LDX #$40
		// CPX #$41
		cpu := loadCPUWith(t, `0600: a2 40 e0 41`)
		cpu.Clock = 0
		cpu.PC = 0x0600
		cpu.P = 0b00110000
		runAndCheckState(t, cpu, 4,
			"A", uint8(0x00),
			"X", uint8(0x40),
			"Y", uint8(0x00),
			"P", uint8(0b10110000),
		)
	})
	t.Run("40 - 40", func(t *testing.T) {
		// LDX #$40
		// CPX #$40
		cpu := loadCPUWith(t, `0600: a2 40 e0 40`)
		cpu.Clock = 0
		cpu.PC = 0x0600
		cpu.P = 0b00110000
		runAndCheckState(t, cpu, 4,
			"A", uint8(0x00),
			"X", uint8(0x40),
			"Y", uint8(0x00),
			"P", uint8(0b00110011),
		)
	})
	t.Run("40 - 39", func(t *testing.T) {
		// LDX #$40
		// CPX #$39
		cpu := loadCPUWith(t, `0600: a2 40 e0 39`)
		cpu.Clock = 0
		cpu.PC = 0x0600
		cpu.P = 0b00110000
		runAndCheckState(t, cpu, 4,
			"A", uint8(0x00),
			"X", uint8(0x40),
			"Y", uint8(0x00),
			"P", uint8(0b00110001),
		)
	})
}

func TestLDA_STA(t *testing.T) {
	dump := `0600: a9 01 8d 00 02 a9 05 8d 01 02 a9 08 8d 02 02`
	cpu := loadCPUWith(t, dump)
	cpu.Clock = 0
	cpu.PC = 0x0600
	runAndCheckState(t, cpu, 6*3,
		"A", uint8(0x08),
		"Pb", uint8(1),
		"PC", uint16(0x060F),
		"SP", uint8(0xfd),
	)
}

func TestEOR(t *testing.T) {
	t.Run("zeropage", func(t *testing.T) {
		dump := `
0000: 06
0100: 45 00`
		cpu := loadCPUWith(t, dump)
		cpu.Clock = 0
		cpu.PC = 0x0100
		cpu.A = 0x80
		runAndCheckState(t, cpu, 3,
			"A", uint8(0x86),
			"Pn", uint8(1),
			"Pz", uint8(0),
		)
	})
}

func TestROR(t *testing.T) {
	t.Run("zeropage", func(t *testing.T) {
		dump := `
0000: 55
0100: 66 00
# reset vector
FFFC: 00 01`
		cpu := loadCPUWith(t, dump)
		cpu.A = 0x80
		cpu.P.writeBit(pbitC, true)
		runAndCheckState(t, cpu, 5,
			"Pn", uint8(1),
			"Pc", uint8(1),
			"Pz", uint8(0),
		)
		wantMem8(t, cpu, 0x0000, 0xAA)
	})
}

func TestStack(t *testing.T) {
	dump := `
# upper stack
01E0: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
01F0: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
# ram
0200: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
0210: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
# instructions
0600: a2 00 a0 00 8a 99 00 02 48 e8 c8 c0 10 d0 f5 68
0610: 99 00 02 c8 c0 20 d0 f7
# reset vector
FFFC: 00 06
`
	cpu := loadCPUWith(t, dump)
	cpu.Clock = 0
	cpu.P = 0x30
	cpu.SP = 0xFF
	runAndCheckState(t, cpu, 562,
		"PC", uint16(0x0618),
		"A", uint8(0x00),
		"X", uint8(0x10),
		"Y", uint8(0x20),
		"SP", uint8(0xFF),
		"mem", `
01f0: 0f 0e 0d 0c 0b 0a 09 08 07 06 05 04 03 02 01 00
0200: 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f
0210: 0f 0e 0d 0c 0b 0a 09 08 07 06 05 04 03 02 01 00`,
	)
}

func TestStackSmall(t *testing.T) {
	dump := `
# upper stack
01E0: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
01F0: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
# instructions
0600: a9 aa 48 a9 11 68`
	cpu := loadCPUWith(t, dump)
	cpu.Clock = 0
	cpu.PC = 0x0600
	cpu.P = 0x30
	cpu.SP = 0xFF
	runAndCheckState(t, cpu, 8,
		"PC", uint16(0x0606),
		"A", uint8(0xAA),
		"SP", uint8(0xFF),
		"Pn", uint8(1),
	)
}

func TestJSR_RTS(t *testing.T) {
	dump := `
# upper stack
01F0: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
# JSR $0620
# LDA #$FF
0600: 20 20 06 A9 FF
# LDA #$88
# RTS
0620: A9 88 60`
	cpu := loadCPUWith(t, dump)
	cpu.Clock = 0
	cpu.PC = 0x0600
	cpu.P = 0x30
	runAndCheckState(t, cpu, 6, "PC", uint16(0x0620))
	runAndCheckState(t, cpu, 6+2, "A", uint8(0x88))
	runAndCheckState(t, cpu, 6+2+6, "PC", uint16(0x0603))
	runAndCheckState(t, cpu, 6+2+6+2, "A", uint8(0xFF))
}

type mapbus struct {
	t testing.TB
	m map[uint16]uint8
}

func (b *mapbus) Reset() {
	b.m = make(map[uint16]uint8)
}

func (b *mapbus) Read8(addr uint16) uint8 {
	val := b.m[addr]
	if b.t != nil && testing.Verbose() {
		b.t.Logf("read8:  addr[0x%04x] -> 0x%02x\n", addr, val)
	}
	return val
}

func (b *mapbus) Write8(addr uint16, val uint8) {
	if b.t != nil && testing.Verbose() {
		b.t.Logf("write8: addr[0x%04x] <- 0x%02x\n", addr, val)
	}
	b.m[addr] = val
}

func (b *mapbus) MapSlice(addr, end uint16, buf []byte) {
	panic("not implemented")
}
