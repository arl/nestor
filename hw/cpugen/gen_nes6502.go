package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const pkgname = "hw"

type opdef struct {
	n string // name
	m string // addressing mode, if prepended with !, adressing mode is manually written in the opcode
	f func() // if nil, opcode is manually written
	d rwType
}

type rwType int

const (
	no rwType = 1 << iota
	rd        // read operand into 'val'
	rw        // read operand into 'val' and write 'val' back into operand after opcode
)

var defs = [256]opdef{
	0x00: {n: "BRK", d: no, m: "imp"},
	0x01: {n: "ORA", d: rd, m: "izx", f: ORA},
	0x02: {n: "STP", d: no, m: "imp", f: STP},
	0x03: {n: "SLO", d: rw, m: "izx", f: SLO},
	0x04: {n: "NOP", d: no, m: "zpg", f: NOP(true)},
	0x05: {n: "ORA", d: rd, m: "zpg", f: ORA},
	0x06: {n: "ASL", d: rw, m: "zpg", f: ASL_mem},
	0x07: {n: "SLO", d: rw, m: "zpg", f: SLO},
	0x08: {n: "PHP", d: no, m: "imp", f: PHP},
	0x09: {n: "ORA", d: rd, m: "imm", f: ORA},
	0x0A: {n: "ASL", d: no, m: "acc", f: ASL},
	0x0B: {n: "ANC", d: rd, m: "imm", f: ANC},
	0x0C: {n: "NOP", d: no, m: "abs", f: NOP(true)},
	0x0D: {n: "ORA", d: rd, m: "abs", f: ORA},
	0x0E: {n: "ASL", d: rw, m: "abs", f: ASL_mem},
	0x0F: {n: "SLO", d: rw, m: "abs", f: SLO},
	0x10: {n: "BPL", d: no, m: "rel", f: branch(Negative, false)},
	0x11: {n: "ORA", d: rd, m: "izy", f: ORA},
	0x12: {n: "STP", d: no, m: "imp", f: STP},
	0x13: {n: "SLO", d: rw, m: "izyd", f: SLO},
	0x14: {n: "NOP", d: no, m: "zpx", f: NOP(true)},
	0x15: {n: "ORA", d: rd, m: "zpx", f: ORA},
	0x16: {n: "ASL", d: rw, m: "zpx", f: ASL_mem},
	0x17: {n: "SLO", d: rw, m: "zpx", f: SLO},
	0x18: {n: "CLC", d: no, m: "imp", f: clear(Carry)},
	0x19: {n: "ORA", d: rd, m: "aby", f: ORA},
	0x1A: {n: "NOP", d: no, m: "imp", f: NOP(false)},
	0x1B: {n: "SLO", d: rw, m: "abyd", f: SLO},
	0x1C: {n: "NOP", d: no, m: "abx", f: NOP(true)},
	0x1D: {n: "ORA", d: rd, m: "abx", f: ORA},
	0x1E: {n: "ASL", d: rw, m: "abxd", f: ASL_mem},
	0x1F: {n: "SLO", d: rw, m: "abxd", f: SLO},
	0x20: {n: "JSR", d: no, m: "abs"},
	0x21: {n: "AND", d: rd, m: "izx", f: AND},
	0x22: {n: "STP", d: no, m: "imp", f: STP},
	0x23: {n: "RLA", d: rw, m: "izx", f: RLA},
	0x24: {n: "BIT", d: rd, m: "zpg", f: BIT},
	0x25: {n: "AND", d: rd, m: "zpg", f: AND},
	0x26: {n: "ROL", d: rw, m: "zpg", f: ROL_mem},
	0x27: {n: "RLA", d: rw, m: "zpg", f: RLA},
	0x28: {n: "PLP", d: no, m: "imp", f: PLP},
	0x29: {n: "AND", d: rd, m: "imm", f: AND},
	0x2A: {n: "ROL", d: no, m: "acc", f: ROL},
	0x2B: {n: "ANC", d: rd, m: "imm", f: ANC},
	0x2C: {n: "BIT", d: rd, m: "abs", f: BIT},
	0x2D: {n: "AND", d: rd, m: "abs", f: AND},
	0x2E: {n: "ROL", d: rw, m: "abs", f: ROL_mem},
	0x2F: {n: "RLA", d: rw, m: "abs", f: RLA},
	0x30: {n: "BMI", d: no, m: "rel", f: branch(Negative, true)},
	0x31: {n: "AND", d: rd, m: "izy", f: AND},
	0x32: {n: "STP", d: no, m: "imp", f: STP},
	0x33: {n: "RLA", d: rw, m: "izyd", f: RLA},
	0x34: {n: "NOP", d: no, m: "zpx", f: NOP(true)},
	0x35: {n: "AND", d: rd, m: "zpx", f: AND},
	0x36: {n: "ROL", d: rw, m: "zpx", f: ROL_mem},
	0x37: {n: "RLA", d: rw, m: "zpx", f: RLA},
	0x38: {n: "SEC", d: no, m: "imp", f: set(Carry)},
	0x39: {n: "AND", d: rd, m: "aby", f: AND},
	0x3A: {n: "NOP", d: no, m: "imp", f: NOP(false)},
	0x3B: {n: "RLA", d: rw, m: "abyd", f: RLA},
	0x3C: {n: "NOP", d: no, m: "abx", f: NOP(true)},
	0x3D: {n: "AND", d: rd, m: "abx", f: AND},
	0x3E: {n: "ROL", d: rw, m: "abxd", f: ROL_mem},
	0x3F: {n: "RLA", d: rw, m: "abxd", f: RLA},
	0x40: {n: "RTI", d: no, m: "imp", f: RTI},
	0x41: {n: "EOR", d: rd, m: "izx", f: EOR},
	0x42: {n: "STP", d: no, m: "imp", f: STP},
	0x43: {n: "SRE", d: rw, m: "izx", f: SRE},
	0x44: {n: "NOP", d: no, m: "zpg", f: NOP(true)},
	0x45: {n: "EOR", d: rd, m: "zpg", f: EOR},
	0x46: {n: "LSR", d: rw, m: "zpg", f: LSR_mem},
	0x47: {n: "SRE", d: rw, m: "zpg", f: SRE},
	0x48: {n: "PHA", d: no, m: "imp", f: PHA},
	0x49: {n: "EOR", d: rd, m: "imm", f: EOR},
	0x4A: {n: "LSR", d: no, m: "acc", f: LSR},
	0x4B: {n: "ALR", d: rd, m: "imm", f: ALR},
	0x4C: {n: "JMP", d: no, m: "abs", f: JMP},
	0x4D: {n: "EOR", d: rd, m: "abs", f: EOR},
	0x4E: {n: "LSR", d: rw, m: "abs", f: LSR_mem},
	0x4F: {n: "SRE", d: rw, m: "abs", f: SRE},
	0x50: {n: "BVC", d: no, m: "rel", f: branch(Overflow, false)},
	0x51: {n: "EOR", d: rd, m: "izy", f: EOR},
	0x52: {n: "STP", d: no, m: "imp", f: STP},
	0x53: {n: "SRE", d: rw, m: "izyd", f: SRE},
	0x54: {n: "NOP", d: no, m: "zpx", f: NOP(true)},
	0x55: {n: "EOR", d: rd, m: "zpx", f: EOR},
	0x56: {n: "LSR", d: rw, m: "zpx", f: LSR_mem},
	0x57: {n: "SRE", d: rw, m: "zpx", f: SRE},
	0x58: {n: "CLI", d: no, m: "imp", f: clear(Interrupt)},
	0x59: {n: "EOR", d: rd, m: "aby", f: EOR},
	0x5A: {n: "NOP", d: no, m: "imp", f: NOP(false)},
	0x5B: {n: "SRE", d: rw, m: "abyd", f: SRE},
	0x5C: {n: "NOP", d: no, m: "abx", f: NOP(true)},
	0x5D: {n: "EOR", d: rd, m: "abx", f: EOR},
	0x5E: {n: "LSR", d: rw, m: "abxd", f: LSR_mem},
	0x5F: {n: "SRE", d: rw, m: "abxd", f: SRE},
	0x60: {n: "RTS", d: no, m: "imp", f: RTS},
	0x61: {n: "ADC", d: rd, m: "izx", f: ADC},
	0x62: {n: "STP", d: no, m: "imp", f: STP},
	0x63: {n: "RRA", d: rw, m: "izx", f: RRA},
	0x64: {n: "NOP", d: no, m: "zpg", f: NOP(true)},
	0x65: {n: "ADC", d: rd, m: "zpg", f: ADC},
	0x66: {n: "ROR", d: rw, m: "zpg", f: ROR_mem},
	0x67: {n: "RRA", d: rw, m: "zpg", f: RRA},
	0x68: {n: "PLA", d: no, m: "imp", f: PLA},
	0x69: {n: "ADC", d: rd, m: "imm", f: ADC},
	0x6A: {n: "ROR", d: no, m: "acc", f: ROR},
	0x6B: {n: "ARR", d: rd, m: "imm", f: ARR},
	0x6C: {n: "JMP", d: no, m: "ind", f: JMP},
	0x6D: {n: "ADC", d: rd, m: "abs", f: ADC},
	0x6E: {n: "ROR", d: rw, m: "abs", f: ROR_mem},
	0x6F: {n: "RRA", d: rw, m: "abs", f: RRA},
	0x70: {n: "BVS", d: no, m: "rel", f: branch(Overflow, true)},
	0x71: {n: "ADC", d: rd, m: "izy", f: ADC},
	0x72: {n: "STP", d: no, m: "imp", f: STP},
	0x73: {n: "RRA", d: rw, m: "izyd", f: RRA},
	0x74: {n: "NOP", d: no, m: "zpx", f: NOP(true)},
	0x75: {n: "ADC", d: rd, m: "zpx", f: ADC},
	0x76: {n: "ROR", d: rw, m: "zpx", f: ROR_mem},
	0x77: {n: "RRA", d: rw, m: "zpx", f: RRA},
	0x78: {n: "SEI", d: no, m: "imp", f: set(Interrupt)},
	0x79: {n: "ADC", d: rd, m: "aby", f: ADC},
	0x7A: {n: "NOP", d: no, m: "imp", f: NOP(false)},
	0x7B: {n: "RRA", d: rw, m: "abyd", f: RRA},
	0x7C: {n: "NOP", d: no, m: "abx", f: NOP(true)},
	0x7D: {n: "ADC", d: rd, m: "abx", f: ADC},
	0x7E: {n: "ROR", d: rw, m: "abxd", f: ROR_mem},
	0x7F: {n: "RRA", d: rw, m: "abxd", f: RRA},
	0x80: {n: "NOP", d: rd, m: "imm", f: NOP(false)},
	0x81: {n: "STA", d: no, m: "izx", f: ST("A")},
	0x82: {n: "NOP", d: rd, m: "imm", f: NOP(false)},
	0x83: {n: "SAX", d: no, m: "izx", f: SAX},
	0x84: {n: "STY", d: no, m: "zpg", f: ST("Y")},
	0x85: {n: "STA", d: no, m: "zpg", f: ST("A")},
	0x86: {n: "STX", d: no, m: "zpg", f: ST("X")},
	0x87: {n: "SAX", d: no, m: "zpg", f: SAX},
	0x88: {n: "DEY", d: no, m: "imp", f: dey},
	0x89: {n: "NOP", d: rd, m: "imm", f: NOP(false)},
	0x8A: {n: "TXA", d: no, m: "imp", f: T("X", "A")},
	0x8B: {n: "ANE", d: rd, m: "imm", f: ANE},
	0x8C: {n: "STY", d: no, m: "abs", f: ST("Y")},
	0x8D: {n: "STA", d: no, m: "abs", f: ST("A")},
	0x8E: {n: "STX", d: no, m: "abs", f: ST("X")},
	0x8F: {n: "SAX", d: no, m: "abs", f: SAX},
	0x90: {n: "BCC", d: no, m: "rel", f: branch(Carry, false)},
	0x91: {n: "STA", d: no, m: "izyd", f: ST("A")},
	0x92: {n: "STP", d: no, m: "imp", f: STP},
	0x93: {n: "SHA", d: no, m: "!izy", f: SHAZ},
	0x94: {n: "STY", d: no, m: "zpx", f: ST("Y")},
	0x95: {n: "STA", d: no, m: "zpx", f: ST("A")},
	0x96: {n: "STX", d: no, m: "zpy", f: ST("X")},
	0x97: {n: "SAX", d: no, m: "zpy", f: SAX},
	0x98: {n: "TYA", d: no, m: "imp", f: T("Y", "A")},
	0x99: {n: "STA", d: no, m: "abyd", f: ST("A")},
	0x9A: {n: "TXS", d: no, m: "imp", f: T("X", "SP")},
	0x9B: {n: "TAS", d: no, m: "!abx", f: TAS},
	0x9C: {n: "SHY", d: no, m: "!aby", f: SHY},
	0x9D: {n: "STA", d: no, m: "abxd", f: ST("A")},
	0x9E: {n: "SHX", d: no, m: "!abx", f: SHX},
	0x9F: {n: "SHA", d: no, m: "!aby", f: SHA},
	0xA0: {n: "LDY", d: rd, m: "imm", f: LD("Y")},
	0xA1: {n: "LDA", d: rd, m: "izx", f: LD("A")},
	0xA2: {n: "LDX", d: rd, m: "imm", f: LD("X")},
	0xA3: {n: "LAX", d: rd, m: "izx", f: LD("A", "X")},
	0xA4: {n: "LDY", d: rd, m: "zpg", f: LD("Y")},
	0xA5: {n: "LDA", d: rd, m: "zpg", f: LD("A")},
	0xA6: {n: "LDX", d: rd, m: "zpg", f: LD("X")},
	0xA7: {n: "LAX", d: rd, m: "zpg", f: LD("A", "X")},
	0xA8: {n: "TAY", d: no, m: "imp", f: T("A", "Y")},
	0xA9: {n: "LDA", d: rd, m: "imm", f: LD("A")},
	0xAA: {n: "TAX", d: no, m: "imp", f: T("A", "X")},
	0xAB: {n: "LXA", d: rd, m: "imm", f: LXA},
	0xAC: {n: "LDY", d: rd, m: "abs", f: LD("Y")},
	0xAD: {n: "LDA", d: rd, m: "abs", f: LD("A")},
	0xAE: {n: "LDX", d: rd, m: "abs", f: LD("X")},
	0xAF: {n: "LAX", d: rd, m: "abs", f: LD("A", "X")},
	0xB0: {n: "BCS", d: no, m: "rel", f: branch(Carry, true)},
	0xB1: {n: "LDA", d: rd, m: "izy", f: LD("A")},
	0xB2: {n: "STP", d: no, m: "imp", f: STP},
	0xB3: {n: "LAX", d: rd, m: "izy", f: LD("A", "X")},
	0xB4: {n: "LDY", d: rd, m: "zpx", f: LD("Y")},
	0xB5: {n: "LDA", d: rd, m: "zpx", f: LD("A")},
	0xB6: {n: "LDX", d: rd, m: "zpy", f: LD("X")},
	0xB7: {n: "LAX", d: rd, m: "zpy", f: LD("A", "X")},
	0xB8: {n: "CLV", d: no, m: "imp", f: clear(Overflow)},
	0xB9: {n: "LDA", d: rd, m: "aby", f: LD("A")},
	0xBA: {n: "TSX", d: no, m: "imp", f: T("SP", "X")},
	0xBB: {n: "LAS", d: rd, m: "aby", f: LAS},
	0xBC: {n: "LDY", d: rd, m: "abx", f: LD("Y")},
	0xBD: {n: "LDA", d: rd, m: "abx", f: LD("A")},
	0xBE: {n: "LDX", d: rd, m: "aby", f: LD("X")},
	0xBF: {n: "LAX", d: rd, m: "aby", f: LD("A", "X")},
	0xC0: {n: "CPY", d: rd, m: "imm", f: cmp("Y")},
	0xC1: {n: "CMP", d: rd, m: "izx", f: cmp("A")},
	0xC2: {n: "NOP", d: rd, m: "imm", f: NOP(false)},
	0xC3: {n: "DCP", d: rd, m: "izx", f: DCP},
	0xC4: {n: "CPY", d: rd, m: "zpg", f: cmp("Y")},
	0xC5: {n: "CMP", d: rd, m: "zpg", f: cmp("A")},
	0xC6: {n: "DEC", d: rd, m: "zpg", f: DEC},
	0xC7: {n: "DCP", d: rd, m: "zpg", f: DCP},
	0xC8: {n: "INY", d: no, m: "imp", f: iny},
	0xC9: {n: "CMP", d: rd, m: "imm", f: cmp("A")},
	0xCA: {n: "DEX", d: no, m: "imp", f: dex},
	0xCB: {n: "SBX", d: rd, m: "imm", f: SBX},
	0xCC: {n: "CPY", d: rd, m: "abs", f: cmp("Y")},
	0xCD: {n: "CMP", d: rd, m: "abs", f: cmp("A")},
	0xCE: {n: "DEC", d: rd, m: "abs", f: DEC},
	0xCF: {n: "DCP", d: rd, m: "abs", f: DCP},
	0xD0: {n: "BNE", d: no, m: "rel", f: branch(Zero, false)},
	0xD1: {n: "CMP", d: rd, m: "izy", f: cmp("A")},
	0xD2: {n: "STP", d: no, m: "imp", f: STP},
	0xD3: {n: "DCP", d: rd, m: "izyd", f: DCP},
	0xD4: {n: "NOP", d: no, m: "zpx", f: NOP(true)},
	0xD5: {n: "CMP", d: rd, m: "zpx", f: cmp("A")},
	0xD6: {n: "DEC", d: rd, m: "zpx", f: DEC},
	0xD7: {n: "DCP", d: rd, m: "zpx", f: DCP},
	0xD8: {n: "CLD", d: no, m: "imp", f: clear(Decimal)},
	0xD9: {n: "CMP", d: rd, m: "aby", f: cmp("A")},
	0xDA: {n: "NOP", d: no, m: "imp", f: NOP(false)},
	0xDB: {n: "DCP", d: rd, m: "abyd", f: DCP},
	0xDC: {n: "NOP", d: no, m: "abx", f: NOP(true)},
	0xDD: {n: "CMP", d: rd, m: "abx", f: cmp("A")},
	0xDE: {n: "DEC", d: rd, m: "abxd", f: DEC},
	0xDF: {n: "DCP", d: rd, m: "abxd", f: DCP},
	0xE0: {n: "CPX", d: rd, m: "imm", f: cmp("X")},
	0xE1: {n: "SBC", d: rd, m: "izx", f: SBC},
	0xE2: {n: "NOP", d: rd, m: "imm", f: NOP(false)},
	0xE3: {n: "ISC", d: rd, m: "izx", f: ISC},
	0xE4: {n: "CPX", d: rd, m: "zpg", f: cmp("X")},
	0xE5: {n: "SBC", d: rd, m: "zpg", f: SBC},
	0xE6: {n: "INC", d: rd, m: "zpg", f: INC},
	0xE7: {n: "ISC", d: rd, m: "zpg", f: ISC},
	0xE8: {n: "INX", d: no, m: "imp", f: inx},
	0xE9: {n: "SBC", d: rd, m: "imm", f: SBC},
	0xEA: {n: "NOP", d: no, m: "imp", f: NOP(false)},
	0xEB: {n: "SBC", d: rd, m: "imm", f: SBC},
	0xEC: {n: "CPX", d: rd, m: "abs", f: cmp("X")},
	0xED: {n: "SBC", d: rd, m: "abs", f: SBC},
	0xEE: {n: "INC", d: rd, m: "abs", f: INC},
	0xEF: {n: "ISC", d: rd, m: "abs", f: ISC},
	0xF0: {n: "BEQ", d: no, m: "rel", f: branch(Zero, true)},
	0xF1: {n: "SBC", d: rd, m: "izy", f: SBC},
	0xF2: {n: "STP", d: no, m: "imp", f: STP},
	0xF3: {n: "ISC", d: rd, m: "izyd", f: ISC},
	0xF4: {n: "NOP", d: no, m: "zpx", f: NOP(true)},
	0xF5: {n: "SBC", d: rd, m: "zpx", f: SBC},
	0xF6: {n: "INC", d: rd, m: "zpx", f: INC},
	0xF7: {n: "ISC", d: rd, m: "zpx", f: ISC},
	0xF8: {n: "SED", d: no, m: "imp", f: set(Decimal)},
	0xF9: {n: "SBC", d: rd, m: "aby", f: SBC},
	0xFA: {n: "NOP", d: no, m: "imp", f: NOP(false)},
	0xFB: {n: "ISC", d: rd, m: "abyd", f: ISC},
	0xFC: {n: "NOP", d: no, m: "abx", f: NOP(true)},
	0xFD: {n: "SBC", d: rd, m: "abx", f: SBC},
	0xFE: {n: "INC", d: rd, m: "abxd", f: INC},
	0xFF: {n: "ISC", d: rd, m: "abxd", f: ISC},
}

type addrmode struct {
	human string // human readable name
	n     int    // number of bytes
	f     func()
}

var addrModes = map[string]addrmode{
	"imp":  {f: imp, n: 1, human: `implied addressing.`},
	"acc":  {f: acc, n: 1, human: `adressing accumulator.`},
	"rel":  {f: rel, n: 2, human: `relative addressing.`},
	"abs":  {f: abs, n: 3, human: `absolute addressing.`},
	"abx":  {f: abx(false), n: 3, human: `absolute indexed X.`},
	"abxd": {f: abx(true), n: 3, human: `absolute indexed X.`},
	"aby":  {f: aby(false), n: 3, human: `absolute indexed Y.`},
	"abyd": {f: aby(true), n: 3, human: `absolute indexed Y.`},
	"imm":  {f: imm, n: 2, human: `immediate addressing.`},
	"ind":  {f: ind, n: 3, human: `indirect addressing.`},
	"izx":  {f: izx, n: 2, human: `indexed addressing (abs, X).`},
	"izy":  {f: izy(false), n: 2, human: `indexed addressing (abs),Y.`},
	"izyd": {f: izy(true), n: 2, human: `indexed addressing (abs),Y.`},
	"zpg":  {f: zpg, n: 2, human: `zero page addressing.`},
	"zpx":  {f: zpx, n: 2, human: `indexed addressing: zeropage,X.`},
	"zpy":  {f: zpy, n: 2, human: `indexed addressing: zeropage,Y.`},
}

//
// Process status flag constants
//

type cpuFlag int

const (
	Carry cpuFlag = 1 << iota
	Zero
	Interrupt
	Decimal
	Break
	Reserved
	Overflow
	Negative
)

func (f cpuFlag) String() string {
	switch f {
	case Carry:
		return "Carry"
	case Zero:
		return "Zero"
	case Interrupt:
		return "Interrupt"
	case Decimal:
		return "Decimal"
	case Break:
		return "Break"
	case Reserved:
		return "Reserved"
	case Overflow:
		return "Overflow"
	case Negative:
		return "Negative"
	}

	panic("unexpected")
}

func setFlags(flags ...cpuFlag) {
	var flagstr []string
	for _, f := range flags {
		flagstr = append(flagstr, f.String())
	}
	printf(`cpu.P.setFlags(%s)`, strings.Join(flagstr, "|"))
}

func clearFlags(flags ...cpuFlag) {
	var flagstr []string
	for _, f := range flags {
		flagstr = append(flagstr, f.String())
	}
	printf(`cpu.P.clearFlags(%s)`, strings.Join(flagstr, "|"))
}

func If(format string, args ...any) block {
	printf(`if %s {`, fmt.Sprintf(format, args...))
	return block{}
}

func IfFlag(f cpuFlag) block {
	printf(`if cpu.P.hasFlag(%s) {`, f)
	return block{}
}

type block struct{}

func (b block) Do(f func()) block {
	f()
	return b
}

func (b block) setFlags(flags ...cpuFlag) block {
	setFlags(flags...)
	return b
}

func (b block) printf(format string, args ...any) block {
	printf("%s", fmt.Sprintf(format, args...))
	return b
}

func (b block) End() { printf(`}`) }

// addressing modes
func imm()                      {}
func acc()                      { printf("cpu.acc()") }
func imp()                      { printf("cpu.imp()") }
func ind()                      { printf("oper := cpu.ind()") }
func rel()                      { printf("oper := cpu.rel()") }
func zpg()                      { printf(`oper := cpu.zpg()`) }
func zpx()                      { printf(`oper := cpu.zpx()`) }
func zpy()                      { printf(`oper := cpu.zpy()`) }
func abs()                      { printf("oper := cpu.abs()") }
func abx(dummyread bool) func() { return func() { printf("oper := cpu.abx(%t)", dummyread) } }
func aby(dummyread bool) func() { return func() { printf("oper := cpu.aby(%t)", dummyread) } }
func izx()                      { printf(`oper := cpu.izx()`) }
func izy(dummyread bool) func() { return func() { printf("oper := cpu.izy(%t)", dummyread) } }

// helpers

func push8(v string) {
	printf(`cpu.push8(%s)`, v)
}

func push16(v string) {
	printf(`cpu.push16(%s)`, v)
}

func pull8(v string) {
	printf(`%s = cpu.pull8()`, v)
}

func pull16(v string) {
	printf(`%s = cpu.pull16()`, v)
}

func branch(f cpuFlag, val bool) func() {
	return func() {
		if val {
			printf(`cpu.branch(oper, %s, 0)`, f)
		} else {
			printf(`cpu.branch(oper, %s, %s)`, f, f)
		}
	}
}

func copybits(dst, src, mask string) string {
	return fmt.Sprintf(`((%s) & (^%s)) | ((%s) & (%s))`, dst, mask, src, mask)
}

func dummyread(addr string) {
	printf(`_ = cpu.Read8(%s) // dummy read`, addr)
}

func dummywrite(addr, value string) {
	printf(`cpu.Write8(%s, %s) // dummy write`, addr, value)
}

//
// opcode generators
//

func ADC() {
	printf(`cpu.add(val)`)
}

func ALR() {
	printf(`// like and + lsr but saves one tick`)
	printf(`cpu.A &= val`)
	printf(`carry := cpu.A & 0x01 // carry is bit 0`)
	printf(`cpu.A = (cpu.A >> 1) & 0x7f`)
	clearFlags(Zero, Negative, Carry)
	printf(`cpu.P.setNZ(cpu.A)`)
	If(`carry != 0`).
		setFlags(Carry).
		End()
}

func ANC() {
	AND()
	clearFlags(Carry)
	IfFlag(Negative).
		setFlags(Carry).
		End()
}

func AND() {
	printf(`cpu.A &= val`)
	clearFlags(Zero, Negative)
	printf(`cpu.P.setNZ(cpu.A)`)
}

func ANE() {
	printf(`const Const = 0xEE`)
	printf(`cpu.A = val & cpu.X & (cpu.A | Const) `)
	clearFlags(Zero, Negative)
	printf(`cpu.P.setNZ(cpu.A)`)
}

func ARR() {
	printf(`cpu.A &= val`)
	printf(`cpu.A >>= 1`)
	clearFlags(Overflow)

	If(`(cpu.A>>6)^(cpu.A>>5)&0x01 != 0`).
		setFlags(Overflow).
		End()
	IfFlag(Carry).
		printf(`cpu.A |= 1 << 7`).
		End()

	clearFlags(Zero, Negative, Carry)
	printf(`cpu.P.setNZ(cpu.A)`)
	If(`(cpu.A&(1<<6) != 0)`).
		setFlags(Carry).
		End()
}

func ASL() {
	printf(`carry := val & 0x80`)
	printf(`val = (val << 1) & 0xfe`)
	clearFlags(Zero, Negative, Carry)
	printf(`cpu.P.setNZ(val)`)
	If(`carry != 0`).
		setFlags(Carry).
		End()
}

func ASL_mem() {
	dummywrite("oper", "val")
	ASL()
}

func BIT() {
	clearFlags(Zero, Overflow, Negative)
	printf(`cpu.P |= P(val & 0b11000000)`)
	If(`cpu.A&val == 0`).
		setFlags(Zero).
		End()
}

func DEC() {
	dummywrite("oper", "val")
	printf(`val--`)
	clearFlags(Zero, Negative)
	printf(`cpu.P.setNZ(val)`)
	printf(`cpu.Write8(oper, val)`)
}

func DCP() {
	DEC()
	cmp("A")()
}

func EOR() {
	printf(`cpu.A ^= val`)
	clearFlags(Zero, Negative)
	printf(`cpu.P.setNZ(cpu.A)`)
}

func INC() {
	dummywrite("oper", "val")
	printf(`val++`)
	clearFlags(Zero, Negative)
	printf(`cpu.P.setNZ(val)`)
	printf(`cpu.Write8(oper, val)`)
}

func ISC() {
	INC()
	printf(`final := val`)
	SBC()
	printf(`val = final`)
}

func JMP() {
	printf(`cpu.PC = oper`)
}

func LAS() {
	printf(`cpu.A = cpu.SP & val`)
	clearFlags(Zero, Negative)
	printf(`cpu.P.setNZ(cpu.A)`)
	printf(`cpu.X = cpu.A`)
	printf(`cpu.SP = cpu.A`)
}

func LSR_mem() {
	dummywrite("oper", "val")
	LSR()
}

func LSR() {
	printf(`carry := val & 0x01 // carry is bit 0`)
	printf(`val = (val >> 1)&0x7f`)
	clearFlags(Zero, Negative, Carry)
	printf(`cpu.P.setNZ(val)`)
	If(`carry != 0`).
		setFlags(Carry).
		End()
}

func LXA() {
	const mask = 0xFF
	printf(`val = (cpu.A | 0x%02x) & val`, mask)
	printf(`cpu.A = val`)
	printf(`cpu.X = val`)
	clearFlags(Zero, Negative)
	printf(`cpu.P.setNZ(cpu.A)`)
}

func NOP(dummy bool) func() {
	return func() {
		if dummy {
			dummyread("oper")
		}
	}
}

func ORA() {
	printf(`cpu.setreg(&cpu.A, cpu.A|val)`)
}

func PHA() {
	push8(`cpu.A`)
}

func PHP() {
	printf(`p := cpu.P | Break |Reserved`)
	push8(`uint8(p)`)
}

func PLA() {
	dummyread("uint16(cpu.SP) + 0x0100")
	pull8(`cpu.A`)
	clearFlags(Zero, Negative)
	printf(`cpu.P.setNZ(cpu.A)`)
}

func PLP() {
	printf(`var p uint8`)
	dummyread("uint16(cpu.SP) + 0x0100")
	pull8(`p`)
	printf(`const mask uint8 = 0b11001111 // ignore B and U bits`)
	printf(`cpu.P = P(%s)`, copybits(`uint8(cpu.P)`, `p`, `mask`))
}

func RLA() {
	ROL_mem()
	AND()
}

func ROL_mem() {
	dummywrite("oper", "val")
	ROL()
}

func ROL() {
	printf(`carry := val & 0x80`)
	printf(`val <<= 1`)

	IfFlag(Carry).
		printf(`val |= 1 << 0`).
		End()

	clearFlags(Zero, Negative, Carry)
	printf(`cpu.P.setNZ(val)`)

	If(`carry != 0`).
		setFlags(Carry).
		End()
}

func ROR_mem() {
	dummywrite("oper", "val")
	ROR()
}

func ROR() {
	printf(`carry := val & 0x01`)
	printf(`val >>= 1`)

	IfFlag(Carry).
		printf(`val |= 1 << 7`).
		End()

	clearFlags(Zero, Negative, Carry)
	printf(`cpu.P.setNZ(val)`)

	If(`carry != 0`).
		setFlags(Carry).
		End()
}

func RRA() {
	ROR_mem()
	ADC()
}

func RTI() {
	printf(`var p uint8`)
	dummyread("uint16(cpu.SP) + 0x0100")
	pull8(`p`)
	printf(`const mask uint8 = 0b11001111 // ignore B and U bits`)
	printf(`cpu.P = P(%s)`, copybits(`uint8(cpu.P)`, `p`, `mask`))
	pull16(`cpu.PC`)
}

func RTS() {
	dummyread("uint16(cpu.SP) + 0x0100")
	pull16(`cpu.PC`)
	printf(`cpu.fetch8()`)
}

func SAX() {
	printf(`cpu.Write8(oper, cpu.A&cpu.X)`)
}

func SBC() {
	printf(`val ^= 0xff`)
	printf(`cpu.add(val)`)
}

func SBX() {
	printf(`ival := (int16(cpu.A) & int16(cpu.X)) - int16(val)`)
	printf(`cpu.X = uint8(ival)`)
	clearFlags(Zero, Negative, Carry)
	printf(`cpu.P.setNZ(cpu.X)`)
	If(`ival >= 0`).
		setFlags(Carry).
		End()
}

func SHA() {
	printf(`cpu.sh(cpu.fetch16(), cpu.Y, cpu.X & cpu.A)`)
}

func SHAZ() {
	printf(`zero := cpu.fetch8()`)
	printf(`var baseaddr uint16`)
	printf(`if(zero == 0xFF) {`)
	printf(`	lo := cpu.Read8(0xFF)`)
	printf(`	hi := cpu.Read8(0x00)`)
	printf(`	baseaddr = uint16(lo) | uint16(hi) << 8`)
	printf(`} else {`)
	printf(`	baseaddr = cpu.Read16(uint16(zero))`)
	printf(`}`)
	printf(`cpu.sh(baseaddr, cpu.Y, cpu.X & cpu.A)`)
}

func SHX() {
	printf(`cpu.sh(cpu.fetch16(), cpu.Y, cpu.X)`)
}

func SHY() {
	printf(`cpu.sh(cpu.fetch16(), cpu.X, cpu.Y)`)
}

func SLO() {
	ASL_mem()
	printf(`cpu.setreg(&cpu.A, cpu.A|val)`)
}

func SRE() {
	LSR_mem()
	EOR()
}

func STP() {
	printf(`cpu.halt()`)
}

func TAS() {
	// Same as "SHA abs, y", but also sets SP = A & X
	SHA()
	printf(`cpu.SP = cpu.X & cpu.A`)
}

//
// opcode helpers
//

func LD(reg ...string) func() {
	return func() {
		for _, r := range reg {
			printf(`cpu.setreg(&cpu.%s, val)`, r)
		}
	}
}

func ST(reg string) func() {
	return func() {
		printf(`cpu.Write8(oper, cpu.%s)`, reg)
	}
}

func cmp(v string) func() {
	return func() {
		v = regOrMem(v)

		clearFlags(Zero, Negative, Carry)
		printf(`cpu.P.setNZ(%s - val)`, v)

		If(`val <= %s`, v).
			setFlags(Carry).
			End()
	}
}

func T(src, dst string) func() {
	return func() {
		printf(`cpu.%s = cpu.%s`, dst, src)
		if dst != "SP" {
			clearFlags(Zero, Negative)
			printf(`cpu.P.setNZ(cpu.%s)`, src)
		}
	}
}

func regOrMem(v string) string {
	switch v {
	case "A", "X", "Y", "SP":
		return `cpu.` + v
	case "mem", "":
		return `val`
	}
	panic("regOrMem " + v)
}

func inx() { printf(`cpu.setreg(&cpu.X, cpu.X+1)`) }
func iny() { printf(`cpu.setreg(&cpu.Y, cpu.Y+1)`) }
func dex() { printf(`cpu.setreg(&cpu.X, cpu.X-1)`) }
func dey() { printf(`cpu.setreg(&cpu.Y, cpu.Y-1)`) }

func clear(f cpuFlag) func() {
	return func() {
		clearFlags(f)
	}
}

func set(f cpuFlag) func() {
	return func() {
		setFlags(f)
	}
}

func header() {
	printf(`// Code generated by cpugen/gen_nes6502.go. DO NOT EDIT.`)
	printf(`package %s`, pkgname)
}

func opcodes() {
	for code, def := range defs {
		if def.f == nil {
			continue
		}

		// header
		m := def.m
		genmode := true
		if m[0] == '!' {
			m = m[1:]
			genmode = false
		}
		mode := addrModes[m]
		printf(`// %s - %s`, def.n, mode.human)
		printf(`func opcode%02X(cpu*CPU) {`, code)
		if genmode {
			mode.f()
		}

		switch {
		case def.m == "acc":
			printf(`val := cpu.A`)
		case def.m == "imm":
			printf(`val := cpu.fetch8()`)
		case def.d == rd, def.d == rw:
			printf(`val := cpu.Read8(oper)`)
		}

		// body
		def.f()

		// footer
		switch {
		case def.m == "imm":
			printf(`_ = val`)
		case def.m == "acc":
			printf(`cpu.A = val`)
		case def.d == rw:
			printf(`cpu.Write8(oper, val)`)
		}

		printf(`}`)
	}
}

func opcodesTable() {
	bb := &strings.Builder{}
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			opcode := i*16 + j
			if defs[opcode].f == nil {
				fmt.Fprintf(bb, "%s,", defs[opcode].n)
			} else {
				fmt.Fprintf(bb, "opcode%02X, ", opcode)
			}
		}
		bb.WriteByte('\n')
	}
	printf(`// nes 6502 opcodes table`)
	printf(`var ops = [256]func(*CPU){`)
	printf("%s", bb.String())
	printf(`}`)
	printf(``)
}

func disasmTable() {
	bb := &strings.Builder{}
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			name := defs[i*16+j].m
			if name[0] == '!' {
				name = name[1:]
			}
			name = strings.ToUpper(name[:1]) + name[1:]
			name = name[:3]
			fmt.Fprintf(bb, "disasm%s, ", name)
		}
		bb.WriteByte('\n')
	}
	printf(`// nes 6502 opcodes disassembly table`)
	printf(`var disasmOps = [256]func(*CPU, uint16) DisasmOp {`)
	printf("%s", bb.String())
	printf(`}`)
	printf(``)
}

func opcodeNamesTable() {
	var names [256]string
	for i, def := range defs {
		names[i] = strconv.Quote(def.n)
	}
	printf(`var opcodeNames = [256]string{`)
	for i := 0; i < 16; i++ {
		printf("%s,", strings.Join(names[i*16:i*16+16], ", "))
	}
	printf(`}`)
}

func printf(format string, args ...any) {
	fmt.Fprintf(g, "%s\n", fmt.Sprintf(format, args...))
}

type Generator struct {
	io.Writer
}

var g Generator

func main() {
	log.SetFlags(0)
	outf := flag.String("out", "opcodes.go", "output file")
	flag.Parse()

	var w io.Writer = os.Stdout

	bb := &bytes.Buffer{}
	if *outf != "stdout" {
		w = bb
	}

	g = Generator{Writer: w}

	header()
	opcodes()
	opcodesTable()
	disasmTable()
	opcodeNamesTable()

	if *outf == "stdout" {
		return
	}
	buf, err := format.Source(bb.Bytes())
	if err != nil {
		if err := os.WriteFile(*outf, bb.Bytes(), 0644); err != nil {
			log.Fatalf("can't write to %s: %s", *outf, err)
		}
		log.Fatalf("'gofmt' failed\n%s", err)
	}

	if err := os.WriteFile(*outf, buf, 0644); err != nil {
		log.Fatalf("can't write to %s: %s", *outf, err)
	}
}
