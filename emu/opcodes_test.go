package emu

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type opcodeAutoTest struct {
	Name    string `json:"name"`
	Initial struct {
		PC  int     `json:"pc"`
		SP  int     `json:"s"`
		A   int     `json:"a"`
		X   int     `json:"x"`
		Y   int     `json:"y"`
		P   int     `json:"p"`
		RAM [][]int `json:"ram"`
	} `json:"initial"`
	Final struct {
		PC  int     `json:"pc"`
		SP  int     `json:"s"`
		A   int     `json:"a"`
		X   int     `json:"x"`
		Y   int     `json:"y"`
		P   int     `json:"p"`
		RAM [][]int `json:"ram"`
	} `json:"final"`
	Cycles [][]any `json:"cycles"`
}

func TestOpcodes(t *testing.T) {
	skip := make([]uint8, 256) // skip tests for those opcodes for now

	// Run tests for all implemented opcodes
	for op, f := range ops {
		if f == nil || skip[op] == 1 {
			continue
		}
		opstr := fmt.Sprintf("%02x", op)
		t.Run(opstr, testOpcodes(opstr))
	}
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
		var tests []opcodeAutoTest
		if err := json.Unmarshal(buf, &tests); err != nil {
			t.Fatal(err)
		}

		for _, tt := range tests {
			t.Run(tt.Name, func(t *testing.T) {
				mem := &mapbus{m: make(map[uint16]uint8), t: t}
				cpu := NewCPU(mem)
				cpu.A = uint8(tt.Initial.A)
				cpu.X = uint8(tt.Initial.X)
				cpu.Y = uint8(tt.Initial.Y)
				cpu.P = P(tt.Initial.P)
				cpu.SP = uint8(tt.Initial.SP)
				cpu.PC = uint16(tt.Initial.PC)

				for _, row := range tt.Initial.RAM {
					mem.Write8(uint16(row[0]), uint8(row[1]))
				}

				if testing.Verbose() {
					t.Log("run")
				}

				cpu.Run(int64(len(tt.Cycles)) - 1)

				if testing.Verbose() {
					for i, row := range tt.Cycles {
						addr := int(row[0].(float64))
						val := int(row[1].(float64))
						t.Logf("cycle %d: %s 0x%4x = 0x%2x\n", i, row[2], addr, val)
					}
				}

				// check ram
				for _, row := range tt.Final.RAM {
					addr := uint16(row[0])
					val := uint8(row[1])
					got := mem.Read8(addr)
					if got != val {
						t.Errorf("ram[0x%x] = 0x%x, want 0x%x", addr, got, val)
					}
				}

				// check cpu state
				wantCPUState(t, cpu,
					"PC", tt.Final.PC,
					"SP", tt.Final.SP,
					"A", tt.Final.A,
					"X", tt.Final.X,
					"Y", tt.Final.Y,
					"P", tt.Final.P,
				)
			})
		}
	}
}

func TestCPx(t *testing.T) {
	t.Run("40 - 41", func(t *testing.T) {
		// LDX #$40
		// CPX #$41
		cpu := loadCPUWith(t, `0600: a2 40 e0 41`)
		cpu.PC = 0x0600
		cpu.P = 0b00110000
		cpu.Run(4)

		wantCPUState(t, cpu, "A", 0x00, "X", 0x40, "Y", 0x00, "P", 0b10110000)
	})
	t.Run("40 - 40", func(t *testing.T) {
		// LDX #$40
		// CPX #$40
		cpu := loadCPUWith(t, `0600: a2 40 e0 40`)
		cpu.PC = 0x0600
		cpu.P = 0b00110000
		cpu.Run(4)

		wantCPUState(t, cpu, "A", 0x00, "X", 0x40, "Y", 0x00, "P", 0b00110011)
	})
	t.Run("40 - 39", func(t *testing.T) {
		// LDX #$40
		// CPX #$39
		cpu := loadCPUWith(t, `0600: a2 40 e0 39`)
		cpu.PC = 0x0600
		cpu.P = 0b00110000
		cpu.Run(4)

		wantCPUState(t, cpu, "A", 0x00, "X", 0x40, "Y", 0x00, "P", 0b00110001)
	})
}

func TestLDA_STA(t *testing.T) {
	dump := `0600: a9 01 8d 00 02 a9 05 8d 01 02 a9 08 8d 02 02`
	cpu := loadCPUWith(t, dump)
	cpu.PC = 0x0600
	cpu.Run(6 * 3)

	wantCPUState(t, cpu, "A", 0x08, "Pb", 1, "PC", 0x060F, "SP", 0xfd)
}

func TestEOR(t *testing.T) {
	t.Run("zeropage", func(t *testing.T) {
		dump := `
0000: 06
0100: 45 00`
		cpu := loadCPUWith(t, dump)
		cpu.PC = 0x0100
		cpu.A = 0x80
		cpu.Run(3)

		wantCPUState(t, cpu, "A", 0x86, "Pn", 1, "Pz", 0)
	})
}

func TestROR(t *testing.T) {
	t.Run("zeropage", func(t *testing.T) {
		dump := `
0000: 55
0100: 66 00`
		cpu := loadCPUWith(t, dump)
		cpu.PC = 0x0100
		cpu.A = 0x80
		cpu.P.writeBit(pbitC, true)

		cpu.Run(5)

		wantMem8(t, cpu, 0x0000, 0xAA)
		wantCPUState(t, cpu, "Pn", 1, "Pc", 1, "Pz", 0)
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
0610: 99 00 02 c8 c0 20 d0 f7`
	cpu := loadCPUWith(t, dump)
	cpu.PC = 0x0600
	cpu.P = 0x30
	cpu.SP = 0xFF
	cpu.SetDisasm(os.Stdout, false)

	cpu.Run(562)

	wantCPUState(t, cpu,
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
	cpu.PC = 0x0600
	cpu.P = 0x30
	cpu.SP = 0xFF
	cpu.SetDisasm(os.Stdout, false)

	cpu.Run(8)

	wantCPUState(t, cpu, "PC", 0x0606, "A", 0xAA, "SP", 0xFF, "Pn", 1)
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
	cpu.PC = 0x0600
	cpu.P = 0x30
	cpu.SetDisasm(os.Stdout, false)

	cpu.Run(6)
	wantCPUState(t, cpu, "PC", 0x0620)
	cpu.Run(6 + 2)
	wantCPUState(t, cpu, "A", 0x88)
	cpu.Run(6 + 2 + 6)
	wantCPUState(t, cpu, "PC", 0x0603)
	cpu.Run(6 + 2 + 6 + 2)
	wantCPUState(t, cpu, "A", 0xFF)
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
