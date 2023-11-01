package emu

import (
	"bytes"
	"fmt"
	"io"
)

var opsDisasm = [256]disasmFunc{
	0x00: imp("BRK"),
	0x04: imp("NOP"),
	0x05: zp("ORA"),
	0x08: imp("PHP"),
	0x09: imm("ORA"),
	0x0C: imp("NOP"),
	0x10: rel("BPL"),
	0x14: imp("NOP"),
	0x15: zpx("ORA"),
	0x18: imp("CLC"),
	0x1A: imp("NOP"),
	0x20: abs("JSR"),
	0x24: zp("BIT"),
	0x28: imp("PLP"),
	0x29: imm("AND"),
	0x2C: abs("BIT"),
	0x30: rel("BMI"),
	0x34: imp("NOP"),
	0x38: imp("SEC"),
	0x3A: imp("NOP"),
	0x44: imp("NOP"),
	0x45: zp("EOR"),
	0x48: imp("PHA"),
	0x49: imm("EOR"),
	0x4C: abs("JMP"),
	0x50: rel("BVC"),
	0x54: imp("NOP"),
	0x58: imp("CLI"),
	0x5A: imp("NOP"),
	0x60: imp("RTS"),
	0x61: izx("ADC"),
	0x64: imp("NOP"),
	0x65: zp("ADC"),
	0x66: zp("ROR"),
	0x68: imp("PLA"),
	0x69: imm("ADC"),
	0x6A: acc("ROR"),
	0x6C: abi("JMP"),
	0x6D: abs("ADC"),
	0x70: rel("BVS"),
	0x71: ixy("ADC"),
	0x74: imp("NOP"),
	0x75: zpx("ADC"),
	0x78: imp("SEI"),
	0x79: aby("ADC"),
	0x7A: imp("NOP"),
	0x7D: abx("ADC"),
	0x80: imp("NOP"),
	0x81: izx("STA"),
	0x82: imp("NOP"),
	0x84: zp("STY"),
	0x85: zp("STA"),
	0x86: zp("STX"),
	0x89: imp("NOP"),
	0x8A: imp("TXA"),
	0x8D: abs("STA"),
	0x8E: abs("STX"),
	0x90: rel("BCC"),
	0x91: ixy("STA"),
	0x95: zpx("STA"),
	0x96: zpy("STX"),
	0x99: aby("STA"),
	0x9A: imp("TXS"),
	0x9D: abx("STA"),
	0xA0: imm("LDY"),
	0xA2: imm("LDX"),
	0xA9: imm("LDA"),
	0xAA: imp("TAX"),
	0xAD: abs("LDA"),
	0xB0: rel("BCS"),
	0xB8: imp("CLV"),
	0xBA: imp("TSX"),
	0xC0: imm("CPY"),
	0xC2: imp("NOP"),
	0xC8: imp("INY"),
	0xC9: imm("CMP"),
	0xCA: imp("DEX"),
	0xD0: rel("BNE"),
	0xD4: imp("NOP"),
	0xD8: imp("CLD"),
	0xDA: imp("NOP"),
	0xE0: imm("CPX"),
	0xE2: imp("NOP"),
	0xE6: zp("INC"),
	0xE8: imp("INX"),
	0xEA: imp("NOP"),
	0xF0: rel("BEQ"),
	0xF4: imp("NOP"),
	0xF8: imp("SED"),
	0xFA: imp("NOP"),
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

// A disasmFunc returns the disassembly string and the number of bytes read for
// an opcode in its context.
type disasmFunc func(*disasm) (string, int)

func imp(opname string) disasmFunc {
	return func(*disasm) (string, int) {
		return opname, 1
	}
}

func acc(op string) disasmFunc {
	return func(*disasm) (string, int) {
		return fmt.Sprintf("%s A", op), 1
	}
}

func imm(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("%s #$%02X", op, d.cpu.imm()), 2
	}
}

func abs(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("%s $%04X", op, d.cpu.abs()), 3
	}
}

func abx(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("%s $%04X,X", op, d.cpu.abs()), 3
	}
}

func aby(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("%s $%04X,Y", op, d.cpu.abs()), 3
	}
}

func zp(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.zp()
		value := d.cpu.Read8(uint16(addr))
		return fmt.Sprintf("%s $%02X = %02X", op, addr, value), 2
	}
}

func zpx(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.zp()
		value := d.cpu.Read8(uint16(addr))
		return fmt.Sprintf("%s $%02X,X = %02X", op, addr, value), 2
	}
}

func zpy(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.zp()
		value := d.cpu.Read8(uint16(addr))
		return fmt.Sprintf("%s $%02X,Y = %02X", op, addr, value), 2
	}
}

func rel(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := reladdr(d.cpu)
		return fmt.Sprintf("%s $%04X", op, addr), 2
	}
}

// absolute indirect
func abi(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.abs()
		return fmt.Sprintf("%s ($%04X)", op, addr), 3
	}
}

func izx(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		val := d.cpu.Read8(d.cpu.PC + 1)
		return fmt.Sprintf("%s ($%02X,X)", op, val), 2
	}
}

func ixy(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("%s ($%02X),Y", op, d.cpu.Read8(d.cpu.PC+1)), 2
	}
}
