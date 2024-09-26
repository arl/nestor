package hw

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
		`E052  A9 32     LDA #$32                         A:00 X:01 Y:02 P:07 S:F4 PPU:0  ,27  8`,
		`E054  20 EE E0  JSR $E0EE                        A:03 X:02 Y:01 P:05 S:F4 PPU:0  ,33  10`,
		`E060  AD 02 20  LDA PpuStatus_2002 = $00         A:0F X:F0 Y:FF P:04 S:FD PPU:0  ,39  12`,
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
			0xE060: DisasmOp{
				PC:     0xE060,
				Buf:    []byte{0xAD, 0x02, 0x20},
				Opcode: "LDA",
				Oper:   "PpuStatus_2002 = $00",
			},
		},
		w: &out,
	}

	tr.write(cpuState{
		PC: 0xE052,
		A:  0x00, X: 0x01, Y: 0x02, P: P(0x07), SP: 0xF4,
		Scanline: 0,
		PPUCycle: 27,
		Clock:    8,
	})
	tr.write(cpuState{
		PC: 0xE054,
		A:  0x03, X: 0x02, Y: 0x01, P: P(0x05), SP: 0xF4,
		Scanline: 0,
		PPUCycle: 33,
		Clock:    10,
	})
	tr.write(cpuState{
		PC: 0xE060,
		A:  0x0F, X: 0xF0, Y: 0xFF, P: P(0x04), SP: 0xFD,
		Scanline: 0,
		PPUCycle: 39,
		Clock:    12,
	})

	wantstr := strings.Join(want, "\n") + "\n"
	if diff := cmp.Diff(out.String(), wantstr); diff != "" {
		t.Fatalf("trace differs\ngot:\n%s\nwant:\n%s\n\ndifferences:\n%s", out.String(), wantstr, diff)
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
