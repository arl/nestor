package emu

import (
	"bytes"
	"fmt"
	"io"
)

var opsDisasm = [256]func(*disasm) (string, int){
	0x00: disasmOp("BRK", implied),
	0x04: disasmOp("NOP", implied),
	0x05: disasmOp("ORA", zeropage),
	0x08: disasmOp("PHP", implied),
	0x09: disasmOp("ORA", immediate),
	0x0C: disasmOp("NOP", implied),
	0x10: disasmOp("BPL", relative),
	0x14: disasmOp("NOP", implied),
	0x18: disasmOp("CLC", implied),
	0x1A: disasmOp("NOP", implied),
	0x20: disasmOp("JSR", absolute),
	0x24: disasmOp("BIT", zeropage),
	0x28: disasmOp("PLP", implied),
	0x29: disasmOp("AND", immediate),
	0x2C: disasmOp("BIT", absolute),
	0x30: disasmOp("BMI", relative),
	0x34: disasmOp("NOP", implied),
	0x38: disasmOp("SEC", implied),
	0x3A: disasmOp("NOP", implied),
	0x44: disasmOp("NOP", implied),
	0x45: disasmOp("EOR", zeropage),
	0x48: disasmOp("PHA", implied),
	0x49: disasmOp("EOR", immediate),
	0x4C: disasmOp("JMP", absolute),
	0x50: disasmOp("BVC", relative),
	0x54: disasmOp("NOP", implied),
	0x58: disasmOp("CLI", implied),
	0x5A: disasmOp("NOP", implied),
	0x60: disasmOp("RTS", implied),
	0x64: disasmOp("NOP", implied),
	0x66: disasmOp("ROR", zeropage),
	0x68: disasmOp("PLA", implied),
	0x69: disasmOp("ADC", immediate),
	0x6A: disasmOp("ROR", accumulator),
	0x6C: disasmOp("JMP", absindirect),
	0x70: disasmOp("BVS", relative),
	0x74: disasmOp("NOP", implied),
	0x78: disasmOp("SEI", implied),
	0x7A: disasmOp("NOP", implied),
	0x80: disasmOp("NOP", implied),
	0x81: disasmOp("STA", preidxindirect),
	0x82: disasmOp("NOP", implied),
	0x84: disasmOp("STY", zeropage),
	0x85: disasmOp("STA", zeropage),
	0x86: disasmOp("STX", zeropage),
	0x89: disasmOp("NOP", implied),
	0x8A: disasmOp("TXA", implied),
	0x8D: disasmOp("STA", absolute),
	0x8E: disasmOp("STX", absolute),
	0x90: disasmOp("BCC", relative),
	0x91: disasmOp("STA", postidxindirect),
	0x95: disasmOp("STA", zeropagex),
	0x99: disasmOp("STA", absolutey),
	0x9A: disasmOp("TXS", implied),
	0x9D: disasmOp("STA", absolutex),
	0xA0: disasmOp("LDY", immediate),
	0xA2: disasmOp("LDX", immediate),
	0xA9: disasmOp("LDA", immediate),
	0xAA: disasmOp("TAX", implied),
	0xAD: disasmOp("LDA", absolute),
	0xB0: disasmOp("BCS", relative),
	0xB8: disasmOp("CLV", implied),
	0xBA: disasmOp("TSX", implied),
	0xC0: disasmOp("CPY", immediate),
	0xC2: disasmOp("NOP", implied),
	0xC8: disasmOp("INY", implied),
	0xC9: disasmOp("CMP", immediate),
	0xCA: disasmOp("DEX", implied),
	0xD0: disasmOp("BNE", relative),
	0xD4: disasmOp("NOP", implied),
	0xD8: disasmOp("CLD", implied),
	0xDA: disasmOp("NOP", implied),
	0xE0: disasmOp("CPX", immediate),
	0xE2: disasmOp("NOP", implied),
	0xE6: disasmOp("INC", zeropage),
	0xE8: disasmOp("INX", implied),
	0xEA: disasmOp("NOP", implied),
	0xF0: disasmOp("BEQ", relative),
	0xF4: disasmOp("NOP", implied),
	0xF8: disasmOp("SED", implied),
	0xFA: disasmOp("NOP", implied),
}

type disasm struct {
	cpu       *CPU
	prevP     P
	prevPC    uint16
	prevClock int64
	bb        bytes.Buffer

	// use nestest 'golden log' format for automatic diff.
	isNestest bool

	w io.Writer
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
	if d == nil {
		return
	}
	d.bb.Reset()
	fmt.Fprintf(&d.bb, "%04X", d.cpu.PC)

	opcode := d.cpu.bus.Read8(uint16(d.cpu.PC))
	opstr, nbytes := opsDisasm[opcode](d)

	var tmp []byte
	for i := uint16(0); i < uint16(nbytes); i++ {
		b := d.cpu.bus.Read8(d.cpu.PC + i)
		tmp = append(tmp, fmt.Sprintf("%02X ", b)...)
	}

	fmt.Fprintf(&d.bb, "  %-9s %-32s", tmp, opstr)
	if d.isNestest {
		// TODO: re-add PPU when we'll have anything else than 0,0
		// fmt.Fprintf(&d.bb, "A:%02X X:%02X Y:%02X P:%02X SP:%02X PPU:%3X,%3X CYC:%d",
		// 	d.cpu.A, d.cpu.X, d.cpu.Y, byte(d.cpu.P), d.cpu.SP, 0, 0, d.cpu.Clock)
		fmt.Fprintf(&d.bb, "A:%02X X:%02X Y:%02X P:%02X SP:%02X CYC:%d",
			d.cpu.A, d.cpu.X, d.cpu.Y, byte(d.cpu.P), d.cpu.SP, d.cpu.Clock)
	} else {
		// TODO: re-add PPU when we'll have anything else than 0,0
		// fmt.Fprintf(&d.bb, "A:%02X X:%02X Y:%02X P:%s SP:%02X PPU:%3X,%3X CYC:%d",
		// 	d.cpu.A, d.cpu.X, d.cpu.Y, d.cpu.P, d.cpu.SP, 0, 0, d.cpu.Clock)
		fmt.Fprintf(&d.bb, "A:%02X X:%02X Y:%02X P:%s SP:%02X CYC:%d",
			d.cpu.A, d.cpu.X, d.cpu.Y, d.cpu.P, d.cpu.SP, d.cpu.Clock)
	}
	d.bb.WriteByte('\n')

	d.w.Write(d.bb.Bytes())
}

// dissasembly helpers

func disasmOp(opname string, mode addressing) func(*disasm) (string, int) {
	return mode(opname)
}

type addressing func(op string) func(*disasm) (string, int)

func implied(opname string) func(*disasm) (string, int) {
	return func(*disasm) (string, int) {
		return opname, 1
	}
}

func accumulator(op string) func(*disasm) (string, int) {
	return func(*disasm) (string, int) {
		return fmt.Sprintf("%s A", op), 1
	}
}

func immediate(op string) func(*disasm) (string, int) {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("%s #$%02X", op, d.cpu.immediate()), 2
	}
}

func absolute(op string) func(*disasm) (string, int) {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("%s $%04X", op, d.cpu.absolute()), 3
	}
}

func absolutex(op string) func(*disasm) (string, int) {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("%s $%04X,X", op, d.cpu.absolute()), 3
	}
}

func absolutey(op string) func(*disasm) (string, int) {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("%s $%04X,Y", op, d.cpu.absolute()), 3
	}
}

func zeropage(op string) func(*disasm) (string, int) {
	return func(d *disasm) (string, int) {
		addr := d.cpu.zeropage()
		value := d.cpu.Read8(addr)
		return fmt.Sprintf("%s $%02X = %02X", op, addr, value), 2
	}
}

func zeropagex(op string) func(*disasm) (string, int) {
	return func(d *disasm) (string, int) {
		addr := d.cpu.zeropage()
		value := d.cpu.Read8(addr)
		return fmt.Sprintf("%s $%02X,X = %02X", op, addr, value), 2
	}
}

func relative(op string) func(*disasm) (string, int) {
	return func(d *disasm) (string, int) {
		addr := reladdr(d.cpu)
		return fmt.Sprintf("%s $%04X", op, addr), 2
	}
}

func absindirect(op string) func(*disasm) (string, int) {
	return func(d *disasm) (string, int) {
		addr := d.cpu.absolute()
		return fmt.Sprintf("%s ($%04X)", op, addr), 3
	}
}

func preidxindirect(op string) func(*disasm) (string, int) {
	return func(d *disasm) (string, int) {
		val := d.cpu.Read8(d.cpu.PC + 1)
		return fmt.Sprintf("%s ($%02X,X)", op, val), 2
	}
}

func postidxindirect(op string) func(*disasm) (string, int) {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("%s ($%02X),Y", op, d.cpu.Read8(d.cpu.PC+1)), 2
	}
}
