package hw

import (
	"bytes"
	"fmt"
	"io"
)

type disasm struct {
	cpu *CPU
	bb  bytes.Buffer

	w io.Writer
}

func NewDisasm(cpu *CPU, w io.Writer) *disasm {
	return &disasm{
		cpu: cpu,
		w:   w,
	}
}

type cpuState struct {
	A, X, Y uint8
	P       P
	SP      uint8
	PC      uint16

	Clock    int64
	PPUCycle int
	Scanline int
}

func (d *disasm) Run(until int64) {
	for d.cpu.Clock < until {
		state := cpuState{
			A:        d.cpu.A,
			X:        d.cpu.X,
			Y:        d.cpu.Y,
			P:        d.cpu.P,
			SP:       d.cpu.SP,
			Clock:    d.cpu.Clock,
			PPUCycle: d.cpu.ppu.Cycle,
			Scanline: d.cpu.ppu.Scanline,
			PC:       d.cpu.PC,
		}

		d.cpu.dbg.Trace(d.cpu.PC)
		opcode := d.cpu.Read8(d.cpu.PC)
		d.cpu.PC++
		d.op(state)
		ops[opcode](d.cpu)

		if d.cpu.prevRunIRQ || d.cpu.prevNeedNmi {
			d.cpu.IRQ()
		}
	}
}

func (d *disasm) read8(addr uint16) uint8 {
	return d.cpu.Bus.Peek8(addr)
}

func (d *disasm) op(state cpuState) {
	d.bb.Reset()

	// Write disassembly.
	opcode := d.read8(state.PC)
	dis := disasmOps[opcode](d.cpu, state.PC)

	fmt.Fprintf(&d.bb, "%-30s A:%02X X:%02X Y:%02X P:%02X SP:%02X PPU:%3d,%3d CYC:%d\n",
		dis.String(), state.A, state.X, state.Y, byte(state.P), state.SP,
		state.Scanline, state.PPUCycle, state.Clock)
	d.w.Write(d.bb.Bytes())
}
