package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

var opsDisasm = [256]func(cpu *CPU) (string, int){
	0x08: disasmOp("PHP", implied),
	0x20: disasmOp("JSR", absolute),
	0x24: disasmOp("BIT", zeropage),
	0x30: disasmOp("BMI", relative),
	0x38: disasmOp("SEC", implied),
	0x45: disasmOp("EOR", zeropage),
	0x4C: disasmOp("JMP", absolute),
	0x48: disasmOp("PHA", implied),
	0x66: disasmOp("ROR", zeropage),
	0x6A: disasmOp("ROR", accumulator),
	0x6C: disasmOp("JMP", absindirect),
	0x78: disasmOp("SEI", implied),
	0x8D: disasmOp("STA", absolute),
	0x8E: disasmOp("STX", absolute),
	0x84: disasmOp("STY", zeropage),
	0x86: disasmOp("STX", zeropage),
	0x91: disasmOp("STA", zeroindirectY),
	0x9A: disasmOp("TXS", implied),
	0xA0: disasmOp("LDY", immediate),
	0xA2: disasmOp("LDX", immediate),
	0xA9: disasmOp("LDA", immediate),
	0xAD: disasmOp("LDA", absolute),
	0xB0: disasmOp("BCS", relative),
	0xC8: disasmOp("INY", implied),
	0xCA: disasmOp("DEX", implied),
	0xC9: disasmOp("CMP", immediate),
	0xD0: disasmOp("BNE", relative),
	0xD8: disasmOp("CLD", implied),
	0xE6: disasmOp("INC", zeropage),
	0xE8: disasmOp("INX", implied),
	0xEA: disasmOp("NOP", implied),
	0xF8: disasmOp("SED", implied),
}

// when true, disasm is enabled everywhere (tests, emulator, etc.)
const defDisasm = true

type disasm struct {
	cpu       *CPU
	prevP     P
	prevPC    uint16
	prevClock int64
	bb        bytes.Buffer

	w io.Writer
}

func newDisasm(cpu *CPU) *disasm {
	return &disasm{cpu: cpu, w: os.Stderr}
}

func (d *disasm) loopinit() {
	if d == nil {
		return
	}
	d.prevP = d.cpu.P
	d.prevPC = d.cpu.PC
	d.prevClock = d.cpu.Clock
}

func (d *disasm) op() {
	d.bb.Reset()
	fmt.Fprintf(&d.bb, "%04X", d.cpu.PC)

	opcode := d.cpu.bus.Read8(uint16(d.cpu.PC))
	opstr, nb := opsDisasm[opcode](d.cpu)

	var tmp []byte
	for i := uint16(0); i < uint16(nb); i++ {
		b := d.cpu.bus.Read8(d.cpu.PC + i)
		tmp = append(tmp, fmt.Sprintf("%02X ", b)...)
	}

	fmt.Fprintf(&d.bb, "  %-9s %-32s", tmp, opstr)
	fmt.Fprintf(&d.bb, "A:%02X X:%02X Y:%02X P:%02X SP:%02X PPU:%3X,%3X CYC:%d",
		d.cpu.A, d.cpu.X, d.cpu.Y, byte(d.cpu.P), d.cpu.SP, 0, 0, d.cpu.Clock)
	d.bb.WriteByte('\n')

	d.w.Write(d.bb.Bytes())
}

// dissasembly functions

func disasmOp(opname string, mode addressing) func(*CPU) (string, int) {
	return mode(opname)
}

type addressing func(op string) func(*CPU) (string, int)

func implied(opname string) func(*CPU) (string, int) {
	return func(cpu *CPU) (string, int) {
		return opname, 1
	}
}

func accumulator(opname string) func(*CPU) (string, int) {
	return func(cpu *CPU) (string, int) {
		return fmt.Sprintf("%s A", opname), 1
	}
}

func immediate(op string) func(*CPU) (string, int) {
	return func(cpu *CPU) (string, int) {
		return fmt.Sprintf("%s #$%02X", op, cpu.Read8(cpu.PC+1)), 2
	}
}

func absolute(op string) func(*CPU) (string, int) {
	return func(cpu *CPU) (string, int) {
		return fmt.Sprintf("%s $%04X", op, cpu.Read16(cpu.PC+1)), 3
	}
}

func zeropage(op string) func(*CPU) (string, int) {
	return func(cpu *CPU) (string, int) {
		opcode := cpu.Read8(cpu.PC)
		addr := cpu.Read8(cpu.PC + 1)
		value := "" // for certain opcodes, we also print the value
		switch opcode {
		case 0x86: // STX
			value = fmt.Sprintf(" = %02X", cpu.Read8(uint16(addr)))
		}
		return fmt.Sprintf("%s $%02X%s", op, addr, value), 2
	}
}

func relative(op string) func(*CPU) (string, int) {
	return func(cpu *CPU) (string, int) {
		off := int32(cpu.Read8(cpu.PC + 1))
		return fmt.Sprintf("%s $%04X", op, uint16(int32(cpu.PC+2)+off)), 2
	}
}

func absindirect(op string) func(*CPU) (string, int) {
	return func(cpu *CPU) (string, int) {
		return fmt.Sprintf("%s ($%04X)", op, cpu.Read16(cpu.PC+1)), 3
	}
}

func zeroindirectY(op string) func(*CPU) (string, int) {
	return func(cpu *CPU) (string, int) {
		return fmt.Sprintf("%s ($%02X),Y", op, cpu.Read8(cpu.PC+1)), 2
	}
}
