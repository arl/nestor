package hw

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func BenchmarkDisasmOpString(b *testing.B) {
	const want = `C000  4C F5 C5  JMP $C5F5                       `

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
	tests := map[uint16]struct {
		disasmOp DisasmOp
		state    cpuState
		want     string
	}{
		0xE052: {
			disasmOp: DisasmOp{PC: 0xE052, Buf: []byte{0xA9, 0x32}, Opcode: "LDA", Oper: "#$32"},
			state:    cpuState{PC: 0xE052, A: 0x00, X: 0x01, Y: 0x02, P: P(0x07), SP: 0xF4, Scanline: 0, PPUCycle: 27, Clock: 8},
			want:     `E052  A9 32     LDA #$32                         A:00 X:01 Y:02 P:07 S:F4 PPU:0  ,27  8`,
		},
		0xE054: {
			disasmOp: DisasmOp{PC: 0xE054, Buf: []byte{0x20, 0xEE, 0xE0}, Opcode: "JSR", Oper: "$E0EE"},
			state:    cpuState{PC: 0xE054, A: 0x03, X: 0x02, Y: 0x01, P: P(0x05), SP: 0xF4, Scanline: 0, PPUCycle: 33, Clock: 10},
			want:     `E054  20 EE E0  JSR $E0EE                        A:03 X:02 Y:01 P:05 S:F4 PPU:0  ,33  10`,
		},
		0xE060: {
			disasmOp: DisasmOp{PC: 0xE060, Buf: []byte{0xAD, 0x02, 0x20}, Opcode: "LDA", Oper: "PpuStatus_2002 = $00"},
			state:    cpuState{PC: 0xE060, A: 0x0F, X: 0xF0, Y: 0xFF, P: P(0x04), SP: 0xFD, Scanline: 0, PPUCycle: 39, Clock: 12},
			want:     `E060  AD 02 20  LDA PpuStatus_2002 = $00         A:0F X:F0 Y:FF P:04 S:FD PPU:0  ,39  12`,
		},
		0xE26E: {
			disasmOp: DisasmOp{PC: 0xE26E, Buf: []byte{0x9D, 0x00, 0x40}, Opcode: "STA", Oper: "Sq0Duty_4000,X [Sq0Duty_4000] = $40"},
			state:    cpuState{PC: 0xE26E, A: 0x11, X: 0x00, Y: 0x00, P: P(0x04), SP: 0xFD, Scanline: 241, PPUCycle: 151, Clock: 116785},
			want:     `E26E  9D 00 40  STA Sq0Duty_4000,X [Sq0Duty_4000] = $40 A:11 X:00 Y:00 P:04 S:FD PPU:241,151 116785`,
		},
		0xCA25: {
			disasmOp: DisasmOp{PC: 0xCA25, Buf: []byte{0x99, 0x0C, 0x40}, Opcode: "STA", Oper: "NoiseVolume_400C,Y [NoiseVolume_400C] = $00"},
			state:    cpuState{PC: 0xCA25, A: 0x00, X: 0x04, Y: 0x00, P: P(0x06), SP: 0x31, Scanline: 257, PPUCycle: 232, Clock: 178192},
			want:     `CA25  99 0C 40  STA NoiseVolume_400C,Y [NoiseVolume_400C] = $00 A:00 X:04 Y:00 P:06 S:31 PPU:257,232 178192`,
		},
	}

	for pc, tt := range tests {
		t.Run(fmt.Sprintf("%04X", pc), func(t *testing.T) {
			var out bytes.Buffer
			tr := tracer{
				d: dummyDisasm{pc: tt.disasmOp},
				w: &out,
			}

			tr.write(tt.state)

			if got := strings.TrimRight(out.String(), "\n"); got != tt.want {
				t.Fatalf("trace differs\ngot:\n\t'%s'\nwant:\n\t'%s'\n", got, tt.want)
			}
		})
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
