package hw

import (
	"bytes"
	"fmt"
	"io"
)

type disasm struct {
	cpu             *CPU
	prevPC          uint16
	prevClock       int64
	prevPPUCycle    int
	prevPPUScanline int
	bb              bytes.Buffer

	w io.Writer
}

func NewDisasm(cpu *CPU, w io.Writer) *disasm {
	return &disasm{
		cpu: cpu,
		w:   w,
	}
}

func (d *disasm) Run(until int64) {
	for d.cpu.Clock < until {
		d.prevPC = d.cpu.PC
		d.prevClock = d.cpu.Clock
		d.prevPPUCycle = d.cpu.ppu.Cycle
		d.prevPPUScanline = d.cpu.ppu.Scanline

		pc := d.cpu.PC
		opcode := d.cpu.Read8(d.cpu.PC)
		d.cpu.PC++
		d.op(pc)
		ops[opcode](d.cpu)

		if d.cpu.prevRunIRQ || d.cpu.prevNeedNmi {
			d.cpu.IRQ()
		}
	}
}

func (d *disasm) read8(addr uint16) uint8 {
	return d.cpu.Bus.Peek8(addr)
}

func (d *disasm) read16(addr uint16) uint16 {
	lo := d.read8(addr)
	hi := d.read8(addr + 1)
	return uint16(lo) | uint16(hi)<<8
}

func (d *disasm) op(pc uint16) {
	d.bb.Reset()

	// Write disassembly.
	opcode := d.read8(pc)
	dis := disasmOps[opcode](d.cpu, pc)
	d.bb.Write(dis)

	// Write CPU state.
	fmt.Fprintf(&d.bb, "%-*s A:%02X X:%02X Y:%02X P:%02X SP:%02X", 40-len(dis), "",
		d.cpu.A, d.cpu.X, d.cpu.Y, byte(d.cpu.P), d.cpu.SP)

	// Write PPU state.
	fmt.Fprintf(&d.bb, " PPU:%3d,%3d CYC:%d",
		d.prevPPUScanline, d.prevPPUCycle, d.prevClock)

	d.bb.WriteByte('\n')
	d.w.Write(d.bb.Bytes())
}
