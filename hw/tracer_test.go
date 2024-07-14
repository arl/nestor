package hw

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func BenchmarkDisasmOpString(b *testing.B) {
	const want = `C000  4C F5 C5  JMP $C5F5         `

	op := DisasmOp{
		Opcode: "JMP",
		Oper:   "$C5F5",
		Buf:    []byte{0x4c, 0xf5, 0xc5},
		PC:     0xC000,
	}

	var opbytes []byte
	for range b.N {
		opbytes = op.Bytes()
	}

	if string(opbytes) != want {
		b.Fatalf("\ngot:  \"%s\"\nwant: \"%s\"\n", string(opbytes), want)
	}
}

type dummyDisasm map[uint16]DisasmOp

func (dd dummyDisasm) Disasm(pc uint16) DisasmOp {
	return dd[pc]
}

func TestTraceFormat(t *testing.T) {
	want := []string{
		`E052  A9 32     LDA #$32                         A:00 X:01 Y:00 P:07 S:F4 PPU:0  ,27  8`,
		`E054  20 EE E0  JSR $E0EE                        A:32 X:01 Y:00 P:05 S:F4 PPU:0  ,33  10`,
	}

	var out bytes.Buffer

	tr := tracer{
		d: dummyDisasm{
			0xE052: DisasmOp{
				PC:     0xE052,
				Buf:    []byte{0xA9, 0x32},
				Opcode: "LDA",
				Oper:   "#$32",
			},
			0xE054: DisasmOp{
				PC:     0xE054,
				Buf:    []byte{0x20, 0xEE, 0xE0},
				Opcode: "JSR",
				Oper:   "$E0EE",
			},
		},
		w: &out,
	}

	tr.write(cpuState{
		PC: 0xE052,
		A:  0x00, X: 0x01, Y: 0x00, P: P(0x07), SP: 0xF4,
		Scanline: 0,
		PPUCycle: 27,
		Clock:    8,
	})
	tr.write(cpuState{
		PC: 0xE054,
		A:  0x32, X: 0x01, Y: 0x00, P: P(0x05), SP: 0xF4,
		Scanline: 0,
		PPUCycle: 33,
		Clock:    10,
	})

	wantstr := strings.Join(want, "\n") + "\n"
	if out.String() != wantstr {
		t.Fatalf("trace differs\ngot:\n%s\nwant:\n%s\n", out.String(), wantstr)
	}
}

func BenchmarkTraceFormat(b *testing.B) {
	tr := tracer{
		d: dummyDisasm{
			0xE052: DisasmOp{
				PC:     0xE052,
				Buf:    []byte{0xA9, 0x32},
				Opcode: "LDA",
				Oper:   "#$32",
			},
			0xE054: DisasmOp{
				PC:     0xE054,
				Buf:    []byte{0x20, 0xEE, 0xE0},
				Opcode: "JSR",
				Oper:   "$E0EE",
			},
		},
		w: io.Discard,
	}
	s1 := cpuState{
		PC: 0xE052,
		A:  0x00, X: 0x01, Y: 0x00, P: P(0x07), SP: 0xF4,
		Scanline: 0,
		PPUCycle: 27,
		Clock:    8,
	}
	s2 := cpuState{
		PC: 0xE054,
		A:  0x32, X: 0x01, Y: 0x00, P: P(0x05), SP: 0xF4,
		Scanline: 0,
		PPUCycle: 33,
		Clock:    10,
	}

	for range b.N {
		tr.write(s1)
		tr.write(s2)
	}
}
