package hw

import (
	"fmt"
	"io"
)

// cpuState stores the CPU state for the execution trace.
type cpuState struct {
	A, X, Y uint8
	P       P
	SP      uint8
	PC      uint16

	Clock    int64
	PPUCycle int
	Scanline int
}

type disasmer interface {
	Disasm(pc uint16) DisasmOp
}

type tracer struct {
	d disasmer
	w io.Writer
}

// write the execution trace for current cycle.
func (t *tracer) write(state cpuState) {
	const totallen = 88
	buf := make([]byte, totallen)

	dis := t.d.Disasm(state.PC)
	off := copy(buf, dis.Bytes())

	for off < 49 {
		buf[off] = ' '
		off++
	}

	buf[off] = 'A'
	off++
	buf[off] = ':'
	off++
	hexEncode(buf[off:], state.A)
	off += 2
	buf[off] = ' '
	off++

	buf[off] = 'X'
	off++
	buf[off] = ':'
	off++
	hexEncode(buf[off:], state.X)
	off += 2
	buf[off] = ' '
	off++

	buf[off] = 'Y'
	off++
	buf[off] = ':'
	off++
	hexEncode(buf[off:], state.Y)
	off += 2
	buf[off] = ' '
	off++

	buf[off] = 'P'
	off++
	buf[off] = ':'
	off++
	hexEncode(buf[off:], byte(state.P))
	off += 2
	buf[off] = ' '
	off++

	buf[off] = 'S'
	off++
	buf[off] = ':'
	off++
	hexEncode(buf[off:], state.SP)
	off += 2
	buf[off] = ' '
	off++

	scanline := state.Scanline
	if scanline == 261 {
		scanline = -1
	}

	buf = fmt.Appendf(buf[:off], "PPU:%-3d,%-3d %d\n", scanline, state.PPUCycle, state.Clock)
	t.w.Write(buf)
}

type DisasmOp struct {
	Opcode string
	Oper   string
	Buf    []byte
	PC     uint16
}

// Bytes returns the string representation of a DisasmOp, this is optimized
// version, suitable for the execution tracer.
func (d DisasmOp) Bytes() []byte {
	const totalLen = 34
	buf := make([]byte, totalLen)

	hexEncode(buf[0:], byte(d.PC>>8))
	hexEncode(buf[2:], byte(d.PC))
	buf[4] = ' '
	buf[5] = ' '

	off := 6
	for i := range d.Buf {
		hexEncode(buf[off:], d.Buf[i])
		buf[off+2] = ' '
		off += 3
	}

	for ; off < 16; off++ {
		buf[off] = ' '
	}

	off += copy(buf[off:], []byte(d.Opcode))
	buf[off] = ' '
	off++

	off += copy(buf[off:], []byte(d.Oper))

	for ; off < totalLen; off++ {
		buf[off] = ' '
	}

	return buf
}

func hexEncode(dst []byte, v byte) {
	const hextable = "0123456789ABCDEF"
	dst[0] = hextable[v>>4]
	dst[1] = hextable[v&0x0f]
}
