package hw

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nestor/emu/hwio"
	"nestor/emu/log"
	"nestor/tests"
)

func TestAllOpcodesAreImplemented(t *testing.T) {
	for opcode, op := range ops {
		if op == nil {
			t.Errorf("opcode %02x not implemented", opcode)
		}
	}
}

func TestOpcodes(t *testing.T) {
	if !testing.Verbose() {
		log.SetOutput(io.Discard)
	}

	testsDir := tests.TomHarteProcTestsPath(t)

	// Run tests for all implemented opcodes.
	for opcode := range ops {
		opstr := fmt.Sprintf("%02x", opcode)
		switch {
		case unstableOps[uint8(opcode)] == 1:
			t.Run(opstr, func(t *testing.T) { t.Skipf("skipping unsupported opcode") })
		default:
			opfile := filepath.Join(testsDir, opstr+".json")
			t.Run(opstr, testOpcodes(opfile))
		}
	}
}

type testMem struct {
	MEM      hwio.Manual      `hwio:"offset=0x0000,size=0x10000"`
	m        map[uint16]uint8 // actual mapped mem
	accesses []memAccess      // stores all accesses
}

type memAccess struct {
	addr uint16
	val  uint8
	typ  string // "r" or "w"
}

func (tm *testMem) prefill(addr uint16, val uint8) {
	if tm.m == nil {
		tm.m = make(map[uint16]uint8)
	}
	tm.m[addr] = val
}

func (tm *testMem) clear() {
	tm.accesses = nil
	tm.m = nil
}

func (tm *testMem) ReadMEM(addr uint16, _ bool) uint8 {
	tm.accesses = append(tm.accesses, memAccess{addr, 0, "r"})
	if val, ok := tm.m[addr]; ok {
		return val
	}
	return 0
}

func (tm *testMem) WriteMEM(addr uint16, val uint8) {
	tm.accesses = append(tm.accesses, memAccess{addr, val, "w"})
	tm.m[addr] = val
}

// testOpcodes runs the opcodes tests in the given json file path (should be of
// the form tests/tomharte.processor.tests/<op>.json). These comes from
// github.com/TomHarte/ProcessorTests/blob/main/nes6502.
func testOpcodes(opfile string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		buf, err := os.ReadFile(opfile)
		if err != nil {
			t.Fatal(err)
		}

		type (
			CPUState struct {
				PC  int     `json:"pc"`
				SP  int     `json:"s"`
				A   int     `json:"a"`
				X   int     `json:"x"`
				Y   int     `json:"y"`
				P   int     `json:"p"`
				RAM [][]int `json:"ram"`
			}
			TestCase struct {
				Name    string   `json:"name"`
				Initial CPUState `json:"initial"`
				Final   CPUState `json:"final"`
				Cycles  [][]any  `json:"cycles"`
			}
		)
		var tests []TestCase
		if err := json.Unmarshal(buf, &tests); err != nil {
			t.Fatal(err)
		}

		if testing.Short() {
			t.Log("with --short, just test a single case per opcode, random")
			idx := rand.IntN(len(tests))
			tests = []TestCase{tests[idx]}
		}

		bus := hwio.NewTable("cputest")

		tmem := testMem{}
		hwio.MustInitRegs(&tmem)

		for _, tt := range tests {
			t.Run(tt.Name, func(t *testing.T) {
				cpu := NewCPU(noPPU)
				cpu.A = uint8(tt.Initial.A)
				cpu.X = uint8(tt.Initial.X)
				cpu.Y = uint8(tt.Initial.Y)
				cpu.P = P(tt.Initial.P)
				cpu.SP = uint8(tt.Initial.SP)
				cpu.PC = uint16(tt.Initial.PC)

				// Preload RAM with test values.
				cpu.Bus = bus
				cpu.Bus.MapBank(0, &tmem, 0)

				tmem.clear()
				for _, row := range tt.Initial.RAM {
					tmem.prefill(uint16(row[0]), uint8(row[1]))
				}

				if testing.Verbose() {
					t.Logf("initial {A=0x%02x X=0x%02x Y=0x%02x P=0x%02x(%s) SP=0x%02x PC=0x%04x}\n",
						cpu.A, cpu.X, cpu.Y, uint8(cpu.P), cpu.P.String(), cpu.SP, cpu.PC)
					t.Logf("run:\nexpecting cycles:\n%s\n\n", strings.Join(prettyCycles(tt.Cycles), "\n"))
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
				if len(tt.Cycles) != int(cpu.Cycles) {
					cyclesStr := strings.Join(prettyCycles(tt.Cycles), "\n")
					t.Errorf("cycles count mismatch: got %d want %d\ndebug:\n%s",
						cpu.Cycles, len(tt.Cycles), cyclesStr)
				}

				// check ram
				for _, row := range tt.Final.RAM {
					addr := row[0]
					val := uint8(row[1])
					got := cpu.Bus.Read8(uint16(addr), false)
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
		cpu.Cycles = 0
		cpu.PC = 0x0600
		cpu.P = 0b00110000
		runAndCheckState(t, cpu, 4,
			"A", 0x00,
			"X", 0x40,
			"Y", 0x00,
			"P", 0b10110000,
		)
	})
	t.Run("40 - 40", func(t *testing.T) {
		// LDX #$40
		// CPX #$40
		cpu := loadCPUWith(t, `0600: a2 40 e0 40`)
		cpu.Cycles = 0
		cpu.PC = 0x0600
		cpu.P = 0b00110000
		runAndCheckState(t, cpu, 4,
			"A", 0x00,
			"X", 0x40,
			"Y", 0x00,
			"P", 0b00110011,
		)
	})
	t.Run("40 - 39", func(t *testing.T) {
		// LDX #$40
		// CPX #$39
		cpu := loadCPUWith(t, `0600: a2 40 e0 39`)
		cpu.Cycles = 0
		cpu.PC = 0x0600
		cpu.P = 0b00110000
		runAndCheckState(t, cpu, 4,
			"A", 0x00,
			"X", 0x40,
			"Y", 0x00,
			"P", 0b00110001,
		)
	})
}

func TestLDA_STA(t *testing.T) {
	dump := `0600: a9 01 8d 00 02 a9 05 8d 01 02 a9 08 8d 02 02`
	cpu := loadCPUWith(t, dump)
	cpu.Cycles = 0
	cpu.PC = 0x0600
	runAndCheckState(t, cpu, 6*3,
		"A", 0x08,
		"PC", 0x060F,
		"SP", 0xfd,
	)
}

func TestEOR(t *testing.T) {
	t.Run("zeropage", func(t *testing.T) {
		dump := `
0000: 06
0100: 45 00`
		cpu := loadCPUWith(t, dump)
		cpu.Cycles = 0
		cpu.PC = 0x0100
		cpu.A = 0x80
		runAndCheckState(t, cpu, 3,
			"A", 0x86,
			"Pn", 1,
			"Pz", 0,
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
		cpu.P.setCarry(true)
		runAndCheckState(t, cpu, 5,
			"Pn", 1,
			"Pc", 1,
			"Pz", 0,
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
	cpu.Cycles = 0
	cpu.P = 0x30
	cpu.SP = 0xFF
	runAndCheckState(t, cpu, 562,
		"PC", 0x0618,
		"A", 0x00,
		"X", 0x10,
		"Y", 0x20,
		"SP", 0xFF,
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
	cpu.Cycles = 0
	cpu.PC = 0x0600
	cpu.P = 0x30
	cpu.SP = 0xFF
	runAndCheckState(t, cpu, 8,
		"PC", 0x0606,
		"A", 0xAA,
		"SP", 0xFF,
		"Pn", 1,
	)
}
