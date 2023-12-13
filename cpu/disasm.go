package cpu

import (
	"bytes"
	"fmt"
	"io"

	"nestor/emu/hwio"
)

var opsDisasm = [256]disasmFunc{
	0x00: disasm_imp("BRK"),
	0x01: disasm_izx("ORA"),
	0x02: disasm_jam(),
	0x03: disasm_izx("*SLO"),
	0x04: disasm_zp("*NOP"),
	0x05: disasm_zp("ORA"),
	0x06: disasm_zp("ASL"),
	0x07: disasm_zp("*SLO"),
	0x08: disasm_imp("PHP"),
	0x09: disasm_imm("ORA"),
	0x0A: disasm_acc("ASL"),
	0x0B: disasm_imm("*ANC"),
	0x0C: disasm_abs("*NOP"),
	0x0D: disasm_abs("ORA"),
	0x0E: disasm_abs("ASL"),
	0x0F: disasm_abs("*SLO"),
	0x10: disasm_rel("BPL"),
	0x11: disasm_izy("ORA"),
	0x12: disasm_jam(),
	0x13: disasm_izy("*SLO"),
	0x14: disasm_zpx("*NOP"),
	0x15: disasm_zpx("ORA"),
	0x16: disasm_zpx("ASL"),
	0x17: disasm_zpx("*SLO"),
	0x18: disasm_imp("CLC"),
	0x19: disasm_aby("ORA"),
	0x1A: disasm_imp("*NOP"),
	0x1B: disasm_aby("*SLO"),
	0x1C: disasm_abx("*NOP"),
	0x1D: disasm_abx("ORA"),
	0x1E: disasm_abx("ASL"),
	0x1F: disasm_abx("*SLO"),
	0x20: disasm_abs("JSR"),
	0x21: disasm_izx("AND"),
	0x22: disasm_jam(),
	0x23: disasm_izx("*RLA"),
	0x24: disasm_zp("BIT"),
	0x25: disasm_zp("AND"),
	0x26: disasm_zp("ROL"),
	0x27: disasm_zp("*RLA"),
	0x28: disasm_imp("PLP"),
	0x29: disasm_imm("AND"),
	0x2A: disasm_acc("ROL"),
	0x2C: disasm_abs("BIT"),
	0x2D: disasm_abs("AND"),
	0x2E: disasm_abs("ROL"),
	0x2F: disasm_abs("*RLA"),
	0x30: disasm_rel("BMI"),
	0x31: disasm_izy("AND"),
	0x32: disasm_jam(),
	0x33: disasm_izy("*RLA"),
	0x34: disasm_zpx("*NOP"),
	0x35: disasm_zpx("AND"),
	0x36: disasm_zpx("ROL"),
	0x37: disasm_zpx("*RLA"),
	0x38: disasm_imp("SEC"),
	0x39: disasm_aby("AND"),
	0x3A: disasm_imp("*NOP"),
	0x3B: disasm_aby("*RLA"),
	0x3C: disasm_abx("*NOP"),
	0x3D: disasm_abx("AND"),
	0x3E: disasm_abx("ROL"),
	0x3F: disasm_abx("*RLA"),
	0x40: disasm_imp("RTI"),
	0x41: disasm_izx("EOR"),
	0x42: disasm_jam(),
	0x43: disasm_izx("*SRE"),
	0x44: disasm_zp("*NOP"),
	0x45: disasm_zp("EOR"),
	0x46: disasm_zp("LSR"),
	0x47: disasm_zp("*SRE"),
	0x48: disasm_imp("PHA"),
	0x49: disasm_imm("EOR"),
	0x4A: disasm_acc("LSR"),
	0x4B: disasm_imm("*ALR"),
	0x4C: disasm_abs("JMP"),
	0x4D: disasm_abs("EOR"),
	0x4E: disasm_abs("LSR"),
	0x4F: disasm_abs("*SRE"),
	0x50: disasm_rel("BVC"),
	0x51: disasm_izy("EOR"),
	0x52: disasm_jam(),
	0x53: disasm_izy("*SRE"),
	0x54: disasm_zpx("*NOP"),
	0x55: disasm_zpx("EOR"),
	0x56: disasm_zpx("LSR"),
	0x57: disasm_zpx("*SRE"),
	0x58: disasm_imp("CLI"),
	0x59: disasm_aby("EOR"),
	0x5A: disasm_imp("*NOP"),
	0x5B: disasm_aby("*SRE"),
	0x5C: disasm_abx("*NOP"),
	0x5D: disasm_abx("EOR"),
	0x5E: disasm_abx("LSR"),
	0x5F: disasm_abx("*SRE"),
	0x60: disasm_imp("RTS"),
	0x61: disasm_izx("ADC"),
	0x62: disasm_jam(),
	0x63: disasm_izx("*RRA"),
	0x64: disasm_zp("*NOP"),
	0x65: disasm_zp("ADC"),
	0x66: disasm_zp("ROR"),
	0x67: disasm_zp("*RRA"),
	0x68: disasm_imp("PLA"),
	0x69: disasm_imm("ADC"),
	0x6A: disasm_acc("ROR"),
	0x6B: disasm_imm("*ARR"),
	0x6C: disasm_ind("JMP"),
	0x6D: disasm_abs("ADC"),
	0x6E: disasm_abs("ROR"),
	0x6F: disasm_abs("*RRA"),
	0x70: disasm_rel("BVS"),
	0x71: disasm_izy("ADC"),
	0x72: disasm_jam(),
	0x73: disasm_izy("*RRA"),
	0x74: disasm_zpx("*NOP"),
	0x75: disasm_zpx("ADC"),
	0x76: disasm_zpx("ROR"),
	0x77: disasm_zpx("*RRA"),
	0x78: disasm_imp("SEI"),
	0x79: disasm_aby("ADC"),
	0x7A: disasm_imp("*NOP"),
	0x7B: disasm_aby("*RRA"),
	0x7C: disasm_abx("*NOP"),
	0x7D: disasm_abx("ADC"),
	0x7E: disasm_abx("ROR"),
	0x7F: disasm_abx("*RRA"),
	0x80: disasm_imm("*NOP"),
	0x81: disasm_izx("STA"),
	0x82: disasm_imp("*NOP"),
	0x83: disasm_izx("*SAX"),
	0x84: disasm_zp("STY"),
	0x85: disasm_zp("STA"),
	0x86: disasm_zp("STX"),
	0x87: disasm_zp("*SAX"),
	0x88: disasm_imp("DEY"),
	0x89: disasm_imm("*NOP"),
	0x8A: disasm_imp("TXA"),
	0x8B: disasm_imm("*ANE"),
	0x8C: disasm_abs("STY"),
	0x8D: disasm_abs("STA"),
	0x8E: disasm_abs("STX"),
	0x8F: disasm_abs("*SAX"),
	0x90: disasm_rel("BCC"),
	0x91: disasm_izy("STA"),
	0x92: disasm_jam(),
	0x93: disasm_izy("*SHA"),
	0x94: disasm_zpx("STY"),
	0x95: disasm_zpx("STA"),
	0x96: disasm_zpy("STX"),
	0x97: disasm_zpy("*SAX"),
	0x98: disasm_imp("TYA"),
	0x99: disasm_aby("STA"),
	0x9A: disasm_imp("TXS"),
	0x9B: disasm_abs("*TAS"),
	0x9C: disasm_abx("*SHY"),
	0x9D: disasm_abx("STA"),
	0x9E: disasm_aby("*SHX"),
	0x9F: disasm_aby("*SHA"),
	0xA0: disasm_imm("LDY"),
	0xA1: disasm_izx("LDA"),
	0xA2: disasm_imm("LDX"),
	0xA3: disasm_izx("*LAX"),
	0xA4: disasm_zp("LDY"),
	0xA5: disasm_zp("LDA"),
	0xA6: disasm_zp("LDX"),
	0xA7: disasm_zp("*LAX"),
	0xA8: disasm_imp("TAY"),
	0xA9: disasm_imm("LDA"),
	0xAA: disasm_imp("TAX"),
	0xAB: disasm_imm("*LXA"),
	0xAC: disasm_abs("LDY"),
	0xAD: disasm_abs("LDA"),
	0xAE: disasm_abs("LDX"),
	0xAF: disasm_abs("*LAX"),
	0xB0: disasm_rel("BCS"),
	0xB1: disasm_izy("LDA"),
	0xB2: disasm_jam(),
	0xB3: disasm_izy("*LAX"),
	0xB4: disasm_zpx("LDY"),
	0xB5: disasm_zpx("LDA"),
	0xB6: disasm_zpy("LDX"),
	0xB7: disasm_zpy("*LAX"),
	0xB8: disasm_imp("CLV"),
	0xB9: disasm_aby("LDA"),
	0xBA: disasm_imp("TSX"),
	0xBB: disasm_aby("*LAS"),
	0xBC: disasm_abx("LDY"),
	0xBD: disasm_abx("LDA"),
	0xBE: disasm_aby("LDX"),
	0xBF: disasm_aby("*LAX"),
	0xC0: disasm_imm("CPY"),
	0xC1: disasm_izx("CMP"),
	0xC2: disasm_imm("*NOP"),
	0xC3: disasm_izx("*DCP"),
	0xC4: disasm_zp("CPY"),
	0xC5: disasm_zp("CMP"),
	0xC6: disasm_zp("DEC"),
	0xC7: disasm_zp("*DCP"),
	0xC8: disasm_imp("INY"),
	0xC9: disasm_imm("CMP"),
	0xCA: disasm_imp("DEX"),
	0xCB: disasm_imm("*SBX"),
	0xCC: disasm_abs("CPY"),
	0xCD: disasm_abs("CMP"),
	0xCE: disasm_abs("DEC"),
	0xCF: disasm_abs("*DCP"),
	0xD0: disasm_rel("BNE"),
	0xD1: disasm_izy("CMP"),
	0xD2: disasm_jam(),
	0xD3: disasm_izy("*DCP"),
	0xD4: disasm_zpx("*NOP"),
	0xD5: disasm_zpx("CMP"),
	0xD6: disasm_zpx("DEC"),
	0xD7: disasm_zpx("*DCP"),
	0xD8: disasm_imp("CLD"),
	0xD9: disasm_aby("CMP"),
	0xDA: disasm_imp("*NOP"),
	0xDB: disasm_aby("*DCP"),
	0xDC: disasm_abx("*NOP"),
	0xDD: disasm_abx("CMP"),
	0xDE: disasm_abx("DEC"),
	0xDF: disasm_abx("*DCP"),
	0xE0: disasm_imm("CPX"),
	0xE1: disasm_izx("SBC"),
	0xE2: disasm_imm("*NOP"),
	0xE3: disasm_izx("*ISB"),
	0xE4: disasm_zp("CPX"),
	0xE5: disasm_zp("SBC"),
	0xE6: disasm_zp("INC"),
	0xE7: disasm_zp("*ISB"),
	0xE8: disasm_imp("INX"),
	0xE9: disasm_imm("SBC"),
	0xEA: disasm_imp("NOP"),
	0xEB: disasm_imm("*SBC"),
	0xEC: disasm_abs("CPX"),
	0xED: disasm_abs("SBC"),
	0xEE: disasm_abs("INC"),
	0xEF: disasm_abs("*ISB"),
	0xF0: disasm_rel("BEQ"),
	0xF1: disasm_izy("SBC"),
	0xF2: disasm_jam(),
	0xF3: disasm_izy("*ISB"),
	0xF4: disasm_zpx("*NOP"),
	0xF5: disasm_zpx("SBC"),
	0xF6: disasm_zpx("INC"),
	0xF7: disasm_zpx("*ISB"),
	0xF8: disasm_imp("SED"),
	0xF9: disasm_aby("SBC"),
	0xFA: disasm_imp("*NOP"),
	0xFB: disasm_aby("*ISB"),
	0xFC: disasm_abx("*NOP"),
	0xFD: disasm_abx("SBC"),
	0xFE: disasm_abx("INC"),
	0xFF: disasm_abx("*ISB"),
}

type disasm struct {
	cpu       *CPU
	prevPC    uint16
	prevClock int64
	bb        bytes.Buffer

	// use nestest 'golden log' format for automatic diff.
	isNestest bool

	w io.Writer
}

func NewDisasm(cpu *CPU, w io.Writer, nestest bool) *disasm {
	return &disasm{
		cpu:       cpu,
		w:         w,
		isNestest: nestest,
	}
}

func (d *disasm) Run(until int64) {
	for d.cpu.Clock < until {
		d.prevPC = d.cpu.PC
		d.prevClock = d.cpu.Clock

		pc := d.cpu.PC
		opcode := d.cpu.Read8(d.cpu.PC)
		d.cpu.PC++
		d.op(pc)
		ops[opcode](d.cpu)
	}
}

func (d *disasm) op(pc uint16) {
	d.bb.Reset()

	opcode := d.cpu.Bus.Read8(pc)
	opstr, nbytes := opsDisasm[opcode](d)

	var tmp []byte
	for i := uint16(0); i < uint16(nbytes); i++ {
		b := d.cpu.Bus.Read8(pc + i)
		tmp = append(tmp, fmt.Sprintf("%02X ", b)...)
	}

	if d.isNestest {
		fmt.Fprintf(&d.bb, "%04X  %-9s%-33sA:%02X X:%02X Y:%02X P:%02X SP:%02X PPU:%3d,%3d CYC:%d\n",
			pc, tmp, opstr, d.cpu.A, d.cpu.X, d.cpu.Y, byte(d.cpu.P), d.cpu.SP, d.cpu.PPU.Scanline, d.cpu.PPU.Cycle, d.prevClock)
	} else {
		fmt.Fprintf(&d.bb, "%04X  %-9s%-33sA:%02X X:%02X Y:%02X P:%s SP:%02X PPU:%3d,%3d CYC:%d\n",
			pc, tmp, opstr, d.cpu.A, d.cpu.X, d.cpu.Y, d.cpu.P, d.cpu.SP, d.cpu.PPU.Scanline, d.cpu.PPU.Cycle, d.prevClock)
	}
	d.w.Write(d.bb.Bytes())
}

// addressing modes
//
// For the disassembler, addressing modes use cpu.bus.Read rather than cpu.Read,
// because we don't want to tick the clock.

func read16(b hwio.BankIO8, addr uint16) uint16 {
	lo := b.Read8(addr)
	hi := b.Read8(addr + 1)
	return uint16(hi)<<8 | uint16(lo)
}

func (d *disasm) imm() uint8  { return d.cpu.Bus.Read8(d.prevPC + 1) }
func (d *disasm) abs() uint16 { return read16(d.cpu.Bus, d.prevPC+1) }
func (d *disasm) zp() uint8   { return d.cpu.Bus.Read8(d.prevPC + 1) }
func (d *disasm) zpx() uint8  { return d.zp() + d.cpu.X }
func (d *disasm) zpy() uint8  { return d.zp() + d.cpu.Y }

func (d *disasm) rel() uint16 {
	off := int16(int8(d.cpu.Bus.Read8(d.prevPC + 1)))
	return uint16(int16(d.prevPC+2) + off)
}

func (d *disasm) izx() uint16 {
	oper := uint8(d.zp())
	oper += d.cpu.X
	return d.zpr16(uint16(oper))
}

func (d *disasm) zpr16(addr uint16) uint16 {
	lo := d.cpu.Bus.Read8(addr)
	hi := d.cpu.Bus.Read8(uint16(uint8(addr) + 1))
	return uint16(hi)<<8 | uint16(lo)
}

func (d *disasm) ind() uint16 {
	oper := read16(d.cpu.Bus, d.prevPC+1)
	lo := d.cpu.Bus.Read8(oper)
	hi := d.cpu.Bus.Read8((0xff00 & oper) | (0x00ff & (oper + 1)))
	return uint16(hi)<<8 | uint16(lo)
}

func (d *disasm) aby() uint16 {
	return d.abs() + uint16(d.cpu.Y)
}

func (d *disasm) abx() uint16 {
	addr := d.abs()
	return addr + uint16(d.cpu.X)
}

// dissasembly helpers

// A disasmFunc returns the disassembly string and the number of bytes read for
// an opcode in its context.
type disasmFunc func(*disasm) (string, int)

func disasm_imp(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("% 4s", op), 1
	}
}

func disasm_acc(op string) disasmFunc {
	return func(*disasm) (string, int) {
		return fmt.Sprintf("% 4s A", op), 1
	}
}

func disasm_imm(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("% 4s #$%02X", op, d.imm()), 2
	}
}

func disasm_jam() disasmFunc {
	return func(d *disasm) (string, int) {
		return "*JAM", 2
	}
}

func disasm_abs(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.abs()
		switch op {
		case "JMP", "JSR":
			return fmt.Sprintf("% 4s $%04X", op, addr), 3
		default:
			return fmt.Sprintf("% 4s $%04X = %02X", op, addr, d.cpu.Bus.Read8(addr)), 3
		}
	}
}

func disasm_abx(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		oper := d.abs()
		addr := d.abx()
		return fmt.Sprintf("% 4s $%04X,X @ %04X = %02X", op, oper, addr, d.cpu.Bus.Read8(addr)), 3
	}
}

func disasm_aby(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		oper := d.abs()
		addr := d.aby()
		return fmt.Sprintf("% 4s $%04X,Y @ %04X = %02X", op, oper, addr, d.cpu.Bus.Read8(addr)), 3
	}
}

func disasm_zp(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.zp()
		value := d.cpu.Bus.Read8(uint16(addr))
		return fmt.Sprintf("% 4s $%02X = %02X", op, addr, value), 2
	}
}

func disasm_zpx(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.zp()
		addr2 := d.zpx()
		return fmt.Sprintf("% 4s $%02X,X @ %02X = %02X", op, addr, addr2, d.cpu.Bus.Read8(uint16(addr2))), 2
	}
}

func disasm_zpy(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.zp()
		addr2 := d.zpy()
		return fmt.Sprintf("% 4s $%02X,Y @ %02X = %02X", op, addr, addr2, d.cpu.Bus.Read8(uint16(addr2))), 2
	}
}

func disasm_rel(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		return fmt.Sprintf("% 4s $%04X", op, d.rel()), 2
	}
}

// indirect (JMP-only)
func disasm_ind(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		oper := read16(d.cpu.Bus, d.prevPC+1)
		dst := d.ind()
		return fmt.Sprintf("% 4s ($%04X) = %04X", op, oper, dst), 3
	}
}

func disasm_izx(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.Bus.Read8(d.prevPC + 1)
		zp := d.zp() + d.cpu.X
		addr2 := d.izx()
		return fmt.Sprintf("% 4s ($%02X,X) @ %02X = %04X = %02X", op, addr, zp, addr2, d.cpu.Bus.Read8(addr2)), 2
	}
}

func disasm_izy(op string) disasmFunc {
	return func(d *disasm) (string, int) {
		addr := d.cpu.Bus.Read8(d.prevPC + 1)
		oper := d.zp()
		addr2 := d.zpr16(uint16(oper))
		dst := addr2 + uint16(d.cpu.Y)
		return fmt.Sprintf("% 4s ($%02X),Y = %04X @ %04X = %02X", op, addr, addr2, dst, d.cpu.Bus.Read8(dst)), 2
	}
}
