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
	_ = dst[1]
	dst[0] = hextable[v>>4]
	dst[1] = hextable[v&0x0f]
}

func (t *tracer) append(buf []byte, state cpuState) {
	copy(buf, "A:XX X:XX Y:XX P:XX S:XX ")

	_ = buf[24]
	hexEncode(buf[2:4], state.A)
	hexEncode(buf[7:9], state.X)
	hexEncode(buf[12:14], state.Y)
	hexEncode(buf[17:19], byte(state.P))
	hexEncode(buf[22:24], state.SP)
}

// write the execution trace for current cycle.
func (t *tracer) write(state cpuState) {
	const maxTraceBytes = maxDisasmOpBytes + 41

	dis := t.d.Disasm(state.PC)
	buf := make([]byte, maxTraceBytes)
	off := copy(buf[0:], dis.Bytes())

	for off < maxDisasmOpBytes+1 {
		buf[off] = ' '
		off++
	}
	t.append(buf[off:], state)

	scanline := state.Scanline
	if scanline == 261 {
		scanline = -1
	}

	buf = fmt.Appendf(buf[:off+25], "PPU:%-3d,%-3d %d\n", scanline, state.PPUCycle, state.Clock)
	t.w.Write(buf)
}

type DisasmOp struct {
	Opcode string
	Oper   string
	Buf    []byte
	PC     uint16
}

const maxDisasmOpBytes = 48

// Bytes returns the string representation of a DisasmOp, this is optimized
// version, suitable for the execution tracer.
func (d DisasmOp) Bytes() []byte {
	buf := make([]byte, 128)
	copy(buf, "XXXX  xx        XXX                                                                                                             ")

	_ = buf[127]
	hexEncode(buf[0:], byte(d.PC>>8))
	hexEncode(buf[2:], byte(d.PC))

	hexEncode(buf[6:], d.Buf[0])
	if len(d.Buf) >= 2 {
		hexEncode(buf[9:], d.Buf[1])
	}
	if len(d.Buf) == 3 {
		hexEncode(buf[12:], d.Buf[2])
	}

	copy(buf[16:], d.Opcode)
	n := copy(buf[20:], d.Oper)
	return buf[:20+n+1]
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

func disasmAbs(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper1 := cpu.Bus.Peek8(pc + 1)
	oper2 := cpu.Bus.Peek8(pc + 2)
	operaddr := uint16(oper1) | uint16(oper2)<<8
	oper := ""

	if oper0 == 0x20 || oper0 == 0x4C {
		// JSR / JMP
		oper = fmt.Sprintf("$%04X", operaddr)
	} else {
		pointee := cpu.Bus.Peek8(operaddr)
		oper = fmt.Sprintf("%s = $%02X", formatAddr(operaddr), pointee)
	}

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0, oper1, oper2},
		Oper:   oper,
	}
}

func disasmAbx(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper1 := cpu.Bus.Peek8(pc + 1)
	oper2 := cpu.Bus.Peek8(pc + 2)
	operaddr := uint16(oper1) | uint16(oper2)<<8
	oper := ""

	addr := operaddr + uint16(cpu.X)
	pointee := cpu.Bus.Peek8(addr)
	oper = fmt.Sprintf("%s,X [%s] = $%02X", formatAddr(operaddr), formatAddr(addr), pointee)

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0, oper1, oper2},
		Oper:   oper,
	}
}

func disasmAby(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper1 := cpu.Bus.Peek8(pc + 1)
	oper2 := cpu.Bus.Peek8(pc + 2)
	operaddr := uint16(oper1) | uint16(oper2)<<8
	oper := ""

	addr := operaddr + uint16(cpu.Y)
	pointee := cpu.Bus.Peek8(addr)
	oper = fmt.Sprintf("%s,Y [%s] = $%02X", formatAddr(operaddr), formatAddr(addr), pointee)

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0, oper1, oper2},
		Oper:   oper,
	}
}

func disasmAcc(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper := ""

	oper = "A"

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0},
		Oper:   oper,
	}
}

func disasmImm(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper1 := cpu.Bus.Peek8(pc + 1)
	oper := ""

	oper = fmt.Sprintf("#$%02X", oper1)

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0, oper1},
		Oper:   oper,
	}
}

func disasmImp(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper := ""

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0},
		Oper:   oper,
	}
}

func disasmInd(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper1 := cpu.Bus.Peek8(pc + 1)
	oper2 := cpu.Bus.Peek8(pc + 2)
	operaddr := uint16(oper1) | uint16(oper2)<<8
	oper := ""

	lo := cpu.Bus.Peek8(operaddr)
	// 2 bytes address wrap around
	hi := cpu.Bus.Peek8((0xff00 & operaddr) | (0x00ff & (operaddr + 1)))
	dest := uint16(hi)<<8 | uint16(lo)
	pointee := cpu.Bus.Peek8(dest)
	oper = fmt.Sprintf("(%s) [%s] = $%02X", formatAddr(operaddr), formatAddr(dest), pointee)

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0, oper1, oper2},
		Oper:   oper,
	}
}

func disasmIzx(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper1 := cpu.Bus.Peek8(pc + 1)
	oper := ""

	addr := uint16(uint8(oper1) + cpu.X)
	// read 16 bytes from the zero page, handling page wrap
	lo := cpu.Bus.Peek8(addr)
	hi := cpu.Bus.Peek8(uint16(uint8(addr) + 1))
	addr = uint16(hi)<<8 | uint16(lo)
	pointee := cpu.Bus.Peek8(addr)
	oper = fmt.Sprintf("($%02X,X) [%s] = $%02X", oper1, formatAddr(addr), pointee)

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0, oper1},
		Oper:   oper,
	}
}

func disasmIzy(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper1 := cpu.Bus.Peek8(pc + 1)
	oper := ""

	// read 16 bytes from the zero page, handling page wrap
	lo := cpu.Bus.Peek8(uint16(oper1))
	hi := cpu.Bus.Peek8(uint16(uint8(oper1) + 1))
	addr := uint16(hi)<<8 | uint16(lo)
	addr += uint16(cpu.Y)
	pointee := cpu.Bus.Peek8(addr)
	oper = fmt.Sprintf("($%02X),Y [%s] = $%02X", oper1, formatAddr(addr), pointee)

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0, oper1},
		Oper:   oper,
	}
}

func disasmRel(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper1 := cpu.Bus.Peek8(pc + 1)
	oper := ""

	oper = fmt.Sprintf("$%04X", uint16(int16(pc+2)+int16(int8(oper1))))

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0, oper1},
		Oper:   oper,
	}
}

func disasmZpg(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper1 := cpu.Bus.Peek8(pc + 1)
	oper := ""

	pointee := cpu.Bus.Peek8(uint16(oper1))
	oper = fmt.Sprintf("$%02X = $%02X", oper1, pointee)

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0, oper1},
		Oper:   oper,
	}
}

func disasmZpx(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper1 := cpu.Bus.Peek8(pc + 1)
	oper := ""

	addr := uint16(oper1) + uint16(cpu.X)
	addr &= 0xff
	pointee := cpu.Bus.Peek8(addr)
	oper = fmt.Sprintf("$%02X,X [%s] = $%02X", oper1, formatAddr(addr), pointee)

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0, oper1},
		Oper:   oper,
	}
}

func disasmZpy(cpu *CPU, pc uint16) DisasmOp {
	oper0 := cpu.Bus.Peek8(pc + 0)
	oper1 := cpu.Bus.Peek8(pc + 1)
	oper := ""

	addr := uint16(oper1) + uint16(cpu.Y)
	addr &= 0xff
	pointee := cpu.Bus.Peek8(addr)
	oper = fmt.Sprintf("$%02X,Y [%s] = $%02X", oper1, formatAddr(addr), pointee)

	return DisasmOp{
		PC:     pc,
		Opcode: opcodeNames[oper0],
		Buf:    []byte{oper0, oper1},
		Oper:   oper,
	}
}
