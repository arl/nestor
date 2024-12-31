package hw

import (
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-faster/jx"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"nestor/emu/log"
	"nestor/hw/hwio"
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

	var dontTest = [256]uint8{
		0x02: 1, // STP
		0x12: 1, // STP
		0x22: 1, // STP
		0x32: 1, // STP
		0x42: 1, // STP
		0x52: 1, // STP
		0x62: 1, // STP
		0x72: 1, // STP
		0x8B: 1, // ANE
		0x92: 1, // STP
		0x93: 1, // SHA
		0x9B: 1, // TAS
		0x9C: 1, // SHY
		0x9E: 1, // SHX
		0x9F: 1, // SHA
		0xAB: 1, // LXA
		0xB2: 1, // STP
		0xD2: 1, // STP
		0xF2: 1, // STP
	}

	testsDir := tests.TomHarteProcTestsPath(t)

	// Run tests for all implemented opcodes.
	for opcode := range ops {
		opstr := fmt.Sprintf("%02x", opcode)

		if dontTest[opcode] == 1 {
			t.Run(opstr, func(t *testing.T) { t.Skipf("skipping unsupported opcode") })
			continue
		}

		t.Run(opstr, testOpcodes(filepath.Join(testsDir, opstr+".json")))
	}
}

type testMem struct {
	MEM      hwio.Manual      `hwio:"offset=0x0000,size=0x10000"`
	m        map[uint16]uint8 // actual mapped mem
	accesses []memAccess      // stores all accesses
	verbose  bool
}

type memAccess struct {
	addr uint16
	val  uint8
	typ  string // "read" or "wwrite"
}

func (tm *testMem) prefill(addr uint16, val uint8) {
	if tm.m == nil {
		tm.m = make(map[uint16]uint8)
		tm.accesses = make([]memAccess, 0, 7)
		tm.verbose = testing.Verbose()
	}
	tm.m[addr] = val
}

func (tm *testMem) clear() {
	tm.accesses = nil
	tm.m = nil
}

func (tm *testMem) ReadMEM(addr uint16, _ bool) uint8 {
	val := tm.m[addr]
	if tm.verbose {
		fmt.Printf("[%d] read 0x%04x = 0x%02x\n", len(tm.accesses), addr, val)
	}
	tm.accesses = append(tm.accesses, memAccess{addr, val, "read"})
	return val
}

func (tm *testMem) WriteMEM(addr uint16, val uint8) {
	if tm.verbose {
		fmt.Printf("[%d] write 0x%04x = 0x%02x\n", len(tm.accesses), addr, val)
	}

	tm.accesses = append(tm.accesses, memAccess{addr, val, "write"})
	tm.m[addr] = val
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustT[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

type (
	Cycle struct {
		Addr uint16 `json:"addr"`
		Val  uint8  `json:"val"`
		RW   string
	}
	CPUState struct {
		PC  uint16  `json:"pc"`
		SP  uint8   `json:"s"`
		A   uint8   `json:"a"`
		X   uint8   `json:"x"`
		Y   uint8   `json:"y"`
		P   uint8   `json:"p"`
		RAM [][]int `json:"ram"`
	}
	TestCase struct {
		Name    string   `json:"name"`
		Initial CPUState `json:"initial"`
		Final   CPUState `json:"final"`
		Cycles  []Cycle  `json:"cycles"`
	}
)

func fastDecode(t *testing.T, path string) []TestCase {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	dec := jx.GetDecoder()
	dec.Reset(f)

	tests := make([]TestCase, 10000)

	i := 0
	must(dec.Arr(func(d *jx.Decoder) error {
		must(d.ObjBytes(func(d *jx.Decoder, key []byte) error {
			switch string(key) {
			case "name":
				tests[i].Name = mustT(d.Str())
			case "initial":
				must(d.ObjBytes(decodeCPUState(&tests[i].Initial)))
			case "final":
				must(d.ObjBytes(decodeCPUState(&tests[i].Final)))
			case "cycles":
				must(d.Arr(decodeCycles(&tests[i].Cycles)))
			default:
				panic("unexpected key: " + string(key))
			}
			return nil
		}))
		i++
		return nil
	}))

	return tests[:i]
}

func decodeCPUState(s *CPUState) func(d *jx.Decoder, key []byte) error {
	return func(d *jx.Decoder, key []byte) error {
		switch string(key) {
		case "pc":
			s.PC = mustT(d.UInt16())
		case "s":
			s.SP = mustT(d.UInt8())
		case "a":
			s.A = mustT(d.UInt8())
		case "x":
			s.X = mustT(d.UInt8())
		case "y":
			s.Y = mustT(d.UInt8())
		case "p":
			s.P = mustT(d.UInt8())
		case "ram":
			must(d.Arr(func(d *jx.Decoder) error {
				row := make([]int, 0, 2)

				it := mustT(d.ArrIter())
				for it.Next() {
					row = append(row, mustT(d.Int()))
				}

				s.RAM = append(s.RAM, row)
				return nil
			}))
		}
		return nil
	}
}

func decodeCycles(cycles *[]Cycle) func(d *jx.Decoder) error {
	*cycles = make([]Cycle, 0, 10)
	return func(d *jx.Decoder) error {
		var c Cycle
		it := mustT(d.ArrIter())
		it.Next()
		c.Addr = mustT(d.UInt16())
		it.Next()
		c.Val = mustT(d.UInt8())
		it.Next()
		c.RW = mustT(d.Str())
		it.Next()
		*cycles = append(*cycles, c)
		return nil
	}
}

var noPPU *PPU = nil

// testOpcodes runs the opcodes tests in the given json file path (should be of
// the form tests/tomharte.processor.tests/<op>.json). These comes from
// github.com/TomHarte/ProcessorTests/blob/main/nes6502.
func testOpcodes(opfile string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		tests := fastDecode(t, opfile)

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
				cpu.A = tt.Initial.A
				cpu.X = tt.Initial.X
				cpu.Y = tt.Initial.Y
				cpu.P = P(tt.Initial.P)
				cpu.SP = tt.Initial.SP
				cpu.PC = tt.Initial.PC

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

				cpu.Run(cpu.Cycles + int64(len(tt.Cycles)-1))

				got := CPUState{
					A:  cpu.A,
					X:  cpu.X,
					Y:  cpu.Y,
					P:  uint8(cpu.P),
					SP: cpu.SP,
					PC: cpu.PC,
				}

				if diff := cmp.Diff(got, tt.Final, cmpopts.IgnoreFields(CPUState{}, "RAM")); diff != "" {
					t.Errorf("cpu state mismatch (-got +want):\n%s", diff)
				}

				// check cycles
				if len(tt.Cycles) != int(cpu.Cycles) {
					cyclesStr := strings.Join(prettyCycles(tt.Cycles), "\n")
					t.Errorf("cycles count mismatch: got %d want %d\ndebug:\n%s",
						cpu.Cycles, len(tt.Cycles), cyclesStr)
				}

				// check ram accesses
				if len(tt.Cycles) != len(tmem.accesses) {
					t.Errorf("ram accesses count mismatch: got %d want %d", len(tmem.accesses), len(tt.Cycles))
				}

				for i, cycle := range tt.Cycles {
					if tmem.accesses[i].addr != cycle.Addr {
						t.Errorf("ram access %d: addr mismatch: got 0x%04x want 0x%04x", i, tmem.accesses[i].addr, cycle.Addr)
					}
					if tmem.accesses[i].val != cycle.Val {
						t.Errorf("ram access %d: val mismatch: got 0x%02x want 0x%02x", i, tmem.accesses[i].val, cycle.Val)
					}
					if tmem.accesses[i].typ != cycle.RW {
						t.Errorf("ram access %d: type mismatch: got %s want %s", i, tmem.accesses[i].typ, cycle.RW)
					}
				}

				// check ram content
				for _, row := range tt.Final.RAM {
					addr := row[0]
					val := uint8(row[1])
					if got := tmem.m[uint16(addr)]; got != val {
						t.Errorf("ram[0x%x] = 0x%x, want 0x%x", addr, got, val)
					}
				}
			})
		}
	}
}

func prettyCycles(cycles []Cycle) []string {
	strs := make([]string, len(cycles))
	for i, c := range cycles {
		strs[i] = fmt.Sprintf("%s 0x%04x = 0x%02x", c.RW, c.Addr, c.Val)
	}
	return strs
}
