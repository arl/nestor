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
	PPUCycle uint32
	Scanline int
}

type disasmer interface {
	Disasm(pc uint16) DisasmOp
}

type tracer struct {
	d disasmer
	w io.Writer
}

func hexEncode(dst []byte, v byte) {
	const hextable = "0123456789ABCDEF"
	dst[0] = hextable[v>>4]
	dst[1] = hextable[v&0x0f]
}

// write the execution trace for current cycle.
func (t *tracer) write(state cpuState) {
	const totalLen = 88
	buf := make([]byte, totalLen)

	dis := t.d.Disasm(state.PC)
	buf = append(buf[:0], dis.Bytes()...)
	off := min(totalLen, len(buf))
	buf = buf[:max(totalLen, len(buf))]

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
	const totalLen = 48
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

	buf = append(buf[:off], d.Oper...)
	off += len(d.Oper)
	if len(buf) > totalLen {
		buf = append(buf, ' ')
	} else {
		buf = buf[:totalLen]
		for i := off; i < totalLen; i++ {
			buf[i] = ' '
		}
	}

	return buf
}

var addressLabels = map[uint16]string{
	0x2000: "PpuControl_2000",
	0x2001: "PpuMask_2001",
	0x2002: "PpuStatus_2002",
	0x2003: "OamAddr_2003",
	0x2004: "OamData_2004",
	0x2005: "PpuScroll_2005",
	0x2006: "PpuAddr_2006",
	0x2007: "PpuData_2007",
	0x4000: "Sq0Duty_4000",
	0x4001: "Sq0Sweep_4001",
	0x4002: "Sq0Timer_4002",
	0x4003: "Sq0Length_4003",
	0x4004: "Sq1Duty_4004",
	0x4005: "Sq1Sweep_4005",
	0x4006: "Sq1Timer_4006",
	0x4007: "Sq1Length_4007",
	0x4008: "TrgLinear_4008",
	0x400A: "TrgTimer_400A",
	0x400B: "TrgLength_400B",
	0x400C: "NoiseVolume_400C",
	0x400E: "NoisePeriod_400E",
	0x400F: "NoiseLength_400F",
	0x4010: "DmcFreq_4010",
	0x4011: "DmcCounter_4011",
	0x4012: "DmcAddress_4012",
	0x4013: "DmcLength_4013",
	0x4014: "SpriteDma_4014",
	0x4015: "ApuStatus_4015",
	0x4016: "Ctrl1_4016",
	0x4017: "Ctrl2_FrameCtr_4017",
}

func formatAddr(addr uint16) string {
	if label, ok := addressLabels[addr]; ok {
		return label
	}
	return fmt.Sprintf("$%04X", addr)
}
