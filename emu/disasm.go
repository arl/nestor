package emu

import (
	"bytes"
	"fmt"
	"io"
)

var opsDisasm = [256]disasmFunc{
	0x00: imp("BRK"),
	0x01: izx("ORA"),
	0x02: jam(),
	0x03: izx("*SLO"),
	0x04: zp("*NOP"),
	0x05: zp("ORA"),
	0x06: zp("ASL"),
	0x07: zp("*SLO"),
	0x08: imp("PHP"),
	0x09: imm("ORA"),
	0x0A: acc("ASL"),
	0x0B: imm("*ANC"),
	0x0C: abs("*NOP"),
	0x0D: abs("ORA"),
	0x0E: abs("ASL"),
	0x0F: abs("*SLO"),
	0x10: rel("BPL"),
	0x11: izy("ORA"),
	0x12: jam(),
	0x13: izy("*SLO"),
	0x14: zpx("*NOP"),
	0x15: zpx("ORA"),
	0x16: zpx("ASL"),
	0x17: zpx("*SLO"),
	0x18: imp("CLC"),
	0x19: aby("ORA"),
	0x1A: imp("*NOP"),
	0x1B: aby("*SLO"),
	0x1C: abx("*NOP"),
	0x1D: abx("ORA"),
	0x1E: abx("ASL"),
	0x1F: abx("*SLO"),
	0x20: abs("JSR"),
	0x21: izx("AND"),
	0x22: jam(),
	0x23: izx("*RLA"),
	0x24: zp("BIT"),
	0x25: zp("AND"),
	0x26: zp("ROL"),
	0x27: zp("*RLA"),
	0x28: imp("PLP"),
	0x29: imm("AND"),
	0x2A: acc("ROL"),
	0x2C: abs("BIT"),
	0x2D: abs("AND"),
	0x2E: abs("ROL"),
	0x2F: abs("*RLA"),
	0x30: rel("BMI"),
	0x31: izy("AND"),
	0x32: jam(),
	0x33: izy("*RLA"),
	0x34: zpx("*NOP"),
	0x35: zpx("AND"),
	0x36: zpx("ROL"),
	0x37: zpx("*RLA"),
	0x38: imp("SEC"),
	0x39: aby("AND"),
	0x3A: imp("*NOP"),
	0x3B: aby("*RLA"),
	0x3C: abx("*NOP"),
	0x3D: abx("AND"),
	0x3E: abx("ROL"),
	0x3F: abx("*RLA"),
	0x40: imp("RTI"),
	0x41: izx("EOR"),
	0x42: jam(),
	0x43: izx("*SRE"),
	0x44: zp("*NOP"),
	0x45: zp("EOR"),
	0x46: zp("LSR"),
	0x47: zp("*SRE"),
	0x48: imp("PHA"),
	0x49: imm("EOR"),
	0x4A: acc("LSR"),
	0x4B: imm("*ALR"),
	0x4C: abs("JMP"),
	0x4D: abs("EOR"),
	0x4E: abs("LSR"),
	0x4F: abs("*SRE"),
	0x50: rel("BVC"),
	0x51: izy("EOR"),
	0x52: jam(),
	0x53: izy("*SRE"),
	0x54: zpx("*NOP"),
	0x55: zpx("EOR"),
	0x56: zpx("LSR"),
	0x57: zpx("*SRE"),
	0x58: imp("CLI"),
	0x59: aby("EOR"),
	0x5A: imp("*NOP"),
	0x5B: aby("*SRE"),
	0x5C: abx("*NOP"),
	0x5D: abx("EOR"),
	0x5E: abx("LSR"),
	0x5F: abx("*SRE"),
	0x60: imp("RTS"),
	0x61: izx("ADC"),
	0x62: jam(),
	0x63: izx("*RRA"),
	0x64: zp("*NOP"),
	0x65: zp("ADC"),
	0x66: zp("ROR"),
	0x67: zp("*RRA"),
	0x68: imp("PLA"),
	0x69: imm("ADC"),
	0x6A: acc("ROR"),
	0x6B: imm("*ARR"),
	0x6C: ind("JMP"),
	0x6D: abs("ADC"),
	0x6E: abs("ROR"),
	0x6F: abs("*RRA"),
	0x70: rel("BVS"),
	0x71: izy("ADC"),
	0x72: jam(),
	0x73: izy("*RRA"),
	0x74: zpx("*NOP"),
	0x75: zpx("ADC"),
	0x76: zpx("ROR"),
	0x77: zpx("*RRA"),
	0x78: imp("SEI"),
	0x79: aby("ADC"),
	0x7A: imp("*NOP"),
	0x7B: aby("*RRA"),
	0x7C: abx("*NOP"),
	0x7D: abx("ADC"),
	0x7E: abx("ROR"),
	0x7F: abx("*RRA"),
	0x80: imm("*NOP"),
	0x81: izx("STA"),
	0x82: imp("*NOP"),
	0x83: izx("*SAX"),
	0x84: zp("STY"),
	0x85: zp("STA"),
	0x86: zp("STX"),
	0x87: zp("*SAX"),
	0x88: imp("DEY"),
	0x89: imm("*NOP"),
	0x8A: imp("TXA"),
	0x8B: imm("*ANE"),
	0x8C: abs("STY"),
	0x8D: abs("STA"),
	0x8E: abs("STX"),
	0x8F: abs("*SAX"),
	0x90: rel("BCC"),
	0x91: izy("STA"),
	0x92: jam(),
	0x93: izy("*SHA"),
	0x94: zpx("STY"),
	0x95: zpx("STA"),
	0x96: zpy("STX"),
	0x97: zpy("*SAX"),
	0x98: imp("TYA"),
	0x99: aby("STA"),
	0x9A: imp("TXS"),
	0x9B: abs("*TAS"),
	0x9C: abx("*SHY"),
	0x9D: abx("STA"),
	0x9E: aby("*SHX"),
	0x9F: aby("*SHA"),
	0xA0: imm("LDY"),
	0xA1: izx("LDA"),
	0xA2: imm("LDX"),
	0xA3: izx("*LAX"),
	0xA4: zp("LDY"),
	0xA5: zp("LDA"),
	0xA6: zp("LDX"),
	0xA7: zp("*LAX"),
	0xA8: imp("TAY"),
	0xA9: imm("LDA"),
	0xAA: imp("TAX"),
	0xAC: abs("LDY"),
	0xAD: abs("LDA"),
	0xAE: abs("LDX"),
	0xAF: abs("*LAX"),
	0xB0: rel("BCS"),
	0xB1: izy("LDA"),
	0xB2: jam(),
	0xB3: izy("*LAX"),
	0xB4: zpx("LDY"),
	0xB5: zpx("LDA"),
	0xB6: zpy("LDX"),
	0xB7: zpy("*LAX"),
	0xB8: imp("CLV"),
	0xB9: aby("LDA"),
	0xBA: imp("TSX"),
	0xBB: aby("*LAS"),
	0xBC: abx("LDY"),
	0xBD: abx("LDA"),
	0xBE: aby("LDX"),
	0xBF: aby("*LAX"),
	0xC0: imm("CPY"),
	0xC1: izx("CMP"),
	0xC2: imm("*NOP"),
	0xC3: izx("*DCP"),
	0xC4: zp("CPY"),
	0xC5: zp("CMP"),
	0xC6: zp("DEC"),
	0xC7: zp("*DCP"),
	0xC8: imp("INY"),
	0xC9: imm("CMP"),
	0xCA: imp("DEX"),
	0xCC: abs("CPY"),
	0xCD: abs("CMP"),
	0xCE: abs("DEC"),
	0xCF: abs("*DCP"),
	0xD0: rel("BNE"),
	0xD1: izy("CMP"),
	0xD2: jam(),
	0xD3: izy("*DCP"),
	0xD4: zpx("*NOP"),
	0xD5: zpx("CMP"),
	0xD6: zpx("DEC"),
	0xD7: zpx("*DCP"),
	0xD8: imp("CLD"),
	0xD9: aby("CMP"),
	0xDA: imp("*NOP"),
	0xDB: aby("*DCP"),
	0xDC: abx("*NOP"),
	0xDD: abx("CMP"),
	0xDE: abx("DEC"),
	0xDF: abx("*DCP"),
	0xE0: imm("CPX"),
	0xE1: izx("SBC"),
	0xE2: imm("*NOP"),
	0xE3: izx("*ISB"),
	0xE4: zp("CPX"),
	0xE5: zp("SBC"),
	0xE6: zp("INC"),
	0xE7: zp("*ISB"),
	0xE8: imp("INX"),
	0xE9: imm("SBC"),
	0xEA: imp("NOP"),
	0xEB: imm("*SBC"),
	0xEC: abs("CPX"),
	0xED: abs("SBC"),
	0xEE: abs("INC"),
	0xEF: abs("*ISB"),
	0xF0: rel("BEQ"),
	0xF1: izy("SBC"),
	0xF2: jam(),
	0xF3: izy("*ISB"),
	0xF4: zpx("*NOP"),
	0xF5: zpx("SBC"),
	0xF6: zpx("INC"),
	0xF7: zpx("*ISB"),
	0xF8: imp("SED"),
	0xF9: aby("SBC"),
	0xFA: imp("*NOP"),
	0xFB: aby("*ISB"),
	0xFC: abx("*NOP"),
	0xFD: abx("SBC"),
	0xFE: abx("INC"),
	0xFF: abx("*ISB"),
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

	opcode := d.cpu.bus.Read8(d.cpu.PC)
	opstr, nbytes := opsDisasm[opcode](d)

	var tmp []byte
	for i := uint16(0); i < uint16(nbytes); i++ {
		b := d.cpu.bus.Read8(d.cpu.PC + i)
		tmp = append(tmp, fmt.Sprintf("%02X ", b)...)
	}

	if d.isNestest {
		// TODO: re-add PPU when we'll have anything else than 0,0
		// fmt.Fprintf(&d.bb, "A:%02X X:%02X Y:%02X P:%02X SP:%02X PPU:%3X,%3X CYC:%d",
		// 	d.cpu.A, d.cpu.X, d.cpu.Y, byte(d.cpu.P), d.cpu.SP, 0, 0, d.cpu.Clock)
		// fmt.Fprintf(&d.bb, "A:%02X X:%02X Y:%02X P:%02X SP:%02X CYC:%d",
		// 	d.cpu.A, d.cpu.X, d.cpu.Y, byte(d.cpu.P), d.cpu.SP, d.cpu.Clock)
		fmt.Fprintf(&d.bb, "%04X  %-9s%-33sA:%02X X:%02X Y:%02X P:%02X SP:%02X CYC:%d\n", d.cpu.PC, tmp, opstr, d.cpu.A, d.cpu.X, d.cpu.Y, byte(d.cpu.P), d.cpu.SP, d.cpu.Clock)
	} else {
		// TODO: re-add PPU when we'll have anything else than 0,0
		// fmt.Fprintf(&d.bb, "A:%02X X:%02X Y:%02X P:%s SP:%02X PPU:%3X,%3X CYC:%d",
		// 	d.cpu.A, d.cpu.X, d.cpu.Y, d.cpu.P, d.cpu.SP, 0, 0, d.cpu.Clock)
		fmt.Fprintf(&d.bb, "%04X  %-9s%-33sA:%02X X:%02X Y:%02X P:%s SP:%02X CYC:%d\n", d.cpu.PC, tmp, opstr, d.cpu.A, d.cpu.X, d.cpu.Y, d.cpu.P, d.cpu.SP, d.cpu.Clock)
	}
	d.w.Write(d.bb.Bytes())
}

// dissasembly helpers

// A disasmFunc returns the disassembly string and the number of bytes read for
// an opcode in its context.
type disasmFunc func(*disasm) (string, int)

func imp(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("% 4s", op), 1
	}
}

func acc(op string) disasmFunc {
	return func(*disasm) (string, int) {
		return fmt.Sprintf("% 4s A", op), 1
	}
}

func imm(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("% 4s #$%02X", op, d.cpu.imm()), 2
	}
}

func jam() disasmFunc {
	return func(d *disasm) (string, int) {
		return "*JAM", 2
	}
}

func abs(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.abs()
		switch op {
		case "JMP", "JSR":
			return fmt.Sprintf("% 4s $%04X", op, addr), 3
		default:
			return fmt.Sprintf("% 4s $%04X = %02X", op, addr, d.cpu.Read8(addr)), 3
		}
	}
}

func abx(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		oper := d.cpu.abs()
		addr, _ := d.cpu.abx()
		return fmt.Sprintf("% 4s $%04X,X @ %04X = %02X", op, oper, addr, d.cpu.Read8(addr)), 3
	}
}

func aby(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		oper := d.cpu.abs()
		addr, _ := d.cpu.aby()
		return fmt.Sprintf("% 4s $%04X,Y @ %04X = %02X", op, oper, addr, d.cpu.Read8(addr)), 3
	}
}

func zp(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.zp()
		value := d.cpu.Read8(uint16(addr))
		return fmt.Sprintf("% 4s $%02X = %02X", op, addr, value), 2
	}
}

func zpx(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.zp()
		addr2 := d.cpu.zpx()
		return fmt.Sprintf("% 4s $%02X,X @ %02X = %02X", op, addr, addr2, d.cpu.Read8(uint16(addr2))), 2
	}
}

func zpy(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.zp()
		addr2 := d.cpu.zpy()
		return fmt.Sprintf("% 4s $%02X,Y @ %02X = %02X", op, addr, addr2, d.cpu.Read8(uint16(addr2))), 2
	}
}

func rel(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := reladdr(d.cpu)
		return fmt.Sprintf("% 4s $%04X", op, addr), 2
	}
}

// indirect (JMP-only)
func ind(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		oper := d.cpu.Read16(d.cpu.PC + 1)
		dst := d.cpu.ind()
		return fmt.Sprintf("% 4s ($%04X) = %04X", op, oper, dst), 3
	}
}

func izx(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.Read8(d.cpu.PC + 1)
		zp := d.cpu.zp() + d.cpu.X
		addr2 := d.cpu.izx()
		return fmt.Sprintf("% 4s ($%02X,X) @ %02X = %04X = %02X", op, addr, zp, addr2, d.cpu.Read8(addr2)), 2
	}
}

func izy(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.Read8(d.cpu.PC + 1)
		oper := d.cpu.zp()
		addr2 := d.cpu.zpr16(uint16(oper))
		dst := addr2 + uint16(d.cpu.Y)
		return fmt.Sprintf("% 4s ($%02X),Y = %04X @ %04X = %02X", op, addr, addr2, dst, d.cpu.Read8(dst)), 2
	}
}
