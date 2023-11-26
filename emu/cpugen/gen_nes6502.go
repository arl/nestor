package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
)

type opdef struct {
	n  string // name
	m  string // addressing mode
	rw string // " " -> do nothing, "w" -> read/write memory/accumulator
	f  func(g *Generator)
}

var defs = [256]opdef{
	0x00: {n: "BRK", rw: " ", m: "imp", f: BRK},
	0x01: {n: "ORA", rw: " ", m: "izx", f: ORA},
	0x02: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0x03: {n: "SLO", rw: "w", m: "izx", f: SLO},
	0x04: {n: "NOP", rw: " ", m: "zpg", f: NOP},
	0x05: {n: "ORA", rw: " ", m: "zpg", f: ORA},
	0x06: {n: "ASL", rw: " ", m: "zpg"},
	0x07: {n: "SLO", rw: "w", m: "zpg", f: SLO},
	0x08: {n: "PHP", rw: " ", m: "imp"},
	0x09: {n: "ORA", rw: " ", m: "imm", f: ORA},
	0x0A: {n: "ASL", rw: " ", m: "acc"},
	0x0B: {n: "ANC", rw: " ", m: "imm"},
	0x0C: {n: "NOP", rw: " ", m: "abs", f: NOP},
	0x0D: {n: "ORA", rw: " ", m: "abs", f: ORA},
	0x0E: {n: "ASL", rw: " ", m: "abs"},
	0x0F: {n: "SLO", rw: "w", m: "abs", f: SLO},
	0x10: {n: "BPL", rw: " ", m: "rel", f: branch(7, false)},
	0x11: {n: "ORA", rw: " ", m: "izy", f: ORA},
	0x12: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0x13: {n: "SLO", rw: "w", m: "izy", f: SLO},
	0x14: {n: "NOP", rw: " ", m: "zpx", f: NOP},
	0x15: {n: "ORA", rw: " ", m: "zpx", f: ORA},
	0x16: {n: "ASL", rw: " ", m: "zpx"},
	0x17: {n: "SLO", rw: "w", m: "zpx", f: SLO},
	0x18: {n: "CLC", rw: " ", m: "imp"},
	0x19: {n: "ORA", rw: " ", m: "aby", f: ORA},
	0x1A: {n: "NOP", rw: " ", m: "imp", f: NOP},
	0x1B: {n: "SLO", rw: "w", m: "aby", f: SLO},
	0x1C: {n: "NOP", rw: " ", m: "abx", f: NOP},
	0x1D: {n: "ORA", rw: " ", m: "abx", f: ORA},
	0x1E: {n: "ASL", rw: " ", m: "abx"},
	0x1F: {n: "SLO", rw: "w", m: "abx", f: SLO},
	0x20: {n: "JSR", rw: " ", m: "abs"},
	0x21: {n: "AND", rw: " ", m: "izx"},
	0x22: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0x23: {n: "RLA", rw: " ", m: "izx"},
	0x24: {n: "BIT", rw: " ", m: "zpg"},
	0x25: {n: "AND", rw: " ", m: "zpg"},
	0x26: {n: "ROL", rw: " ", m: "zpg"},
	0x27: {n: "RLA", rw: " ", m: "zpg"},
	0x28: {n: "PLP", rw: " ", m: "imp"},
	0x29: {n: "AND", rw: " ", m: "imm"},
	0x2A: {n: "ROL", rw: " ", m: "acc"},
	0x2B: {n: "ANC", rw: " ", m: "imm"},
	0x2C: {n: "BIT", rw: " ", m: "abs"},
	0x2D: {n: "AND", rw: " ", m: "abs"},
	0x2E: {n: "ROL", rw: " ", m: "abs"},
	0x2F: {n: "RLA", rw: " ", m: "abs"},
	0x30: {n: "BMI", rw: " ", m: "rel", f: branch(7, true)},
	0x31: {n: "AND", rw: " ", m: "izy"},
	0x32: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0x33: {n: "RLA", rw: " ", m: "izy"},
	0x34: {n: "NOP", rw: " ", m: "zpx", f: NOP},
	0x35: {n: "AND", rw: " ", m: "zpx"},
	0x36: {n: "ROL", rw: " ", m: "zpx"},
	0x37: {n: "RLA", rw: " ", m: "zpx"},
	0x38: {n: "SEC", rw: " ", m: "imp"},
	0x39: {n: "AND", rw: " ", m: "aby"},
	0x3A: {n: "NOP", rw: " ", m: "imp", f: NOP},
	0x3B: {n: "RLA", rw: " ", m: "aby"},
	0x3C: {n: "NOP", rw: " ", m: "abx", f: NOP},
	0x3D: {n: "AND", rw: " ", m: "abx"},
	0x3E: {n: "ROL", rw: " ", m: "abx"},
	0x3F: {n: "RLA", rw: " ", m: "abx"},
	0x40: {n: "RTI", rw: " ", m: "imp"},
	0x41: {n: "EOR", rw: " ", m: "izx"},
	0x42: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0x43: {n: "SRE", rw: " ", m: "izx"},
	0x44: {n: "NOP", rw: " ", m: "zpg", f: NOP},
	0x45: {n: "EOR", rw: " ", m: "zpg"},
	0x46: {n: "LSR", rw: " ", m: "zpg"},
	0x47: {n: "SRE", rw: " ", m: "zpg"},
	0x48: {n: "PHA", rw: " ", m: "imp"},
	0x49: {n: "EOR", rw: " ", m: "imm"},
	0x4A: {n: "LSR", rw: " ", m: "acc"},
	0x4B: {n: "ALR", rw: " ", m: "imm"},
	0x4C: {n: "JMP", rw: " ", m: "abs"},
	0x4D: {n: "EOR", rw: " ", m: "abs"},
	0x4E: {n: "LSR", rw: " ", m: "abs"},
	0x4F: {n: "SRE", rw: " ", m: "abs"},
	0x50: {n: "BVC", rw: " ", m: "rel", f: branch(6, false)},
	0x51: {n: "EOR", rw: " ", m: "izy"},
	0x52: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0x53: {n: "SRE", rw: " ", m: "izy"},
	0x54: {n: "NOP", rw: " ", m: "zpx", f: NOP},
	0x55: {n: "EOR", rw: " ", m: "zpx"},
	0x56: {n: "LSR", rw: " ", m: "zpx"},
	0x57: {n: "SRE", rw: " ", m: "zpx"},
	0x58: {n: "CLI", rw: " ", m: "imp"},
	0x59: {n: "EOR", rw: " ", m: "aby"},
	0x5A: {n: "NOP", rw: " ", m: "imp", f: NOP},
	0x5B: {n: "SRE", rw: " ", m: "aby"},
	0x5C: {n: "NOP", rw: " ", m: "abx", f: NOP},
	0x5D: {n: "EOR", rw: " ", m: "abx"},
	0x5E: {n: "LSR", rw: " ", m: "abx"},
	0x5F: {n: "SRE", rw: " ", m: "abx"},
	0x60: {n: "RTS", rw: " ", m: "imp"},
	0x61: {n: "ADC", rw: " ", m: "izx"},
	0x62: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0x63: {n: "RRA", rw: " ", m: "izx"},
	0x64: {n: "NOP", rw: " ", m: "zpg", f: NOP},
	0x65: {n: "ADC", rw: " ", m: "zpg"},
	0x66: {n: "ROR", rw: " ", m: "zpg"},
	0x67: {n: "RRA", rw: " ", m: "zpg"},
	0x68: {n: "PLA", rw: " ", m: "imp"},
	0x69: {n: "ADC", rw: " ", m: "imm"},
	0x6A: {n: "ROR", rw: " ", m: "acc"},
	0x6B: {n: "ARR", rw: " ", m: "imm"},
	0x6C: {n: "JMP", rw: " ", m: "ind"},
	0x6D: {n: "ADC", rw: " ", m: "abs"},
	0x6E: {n: "ROR", rw: " ", m: "abs"},
	0x6F: {n: "RRA", rw: " ", m: "abs"},
	0x70: {n: "BVS", rw: " ", m: "rel", f: branch(6, true)},
	0x71: {n: "ADC", rw: " ", m: "izy"},
	0x72: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0x73: {n: "RRA", rw: " ", m: "izy"},
	0x74: {n: "NOP", rw: " ", m: "zpx", f: NOP},
	0x75: {n: "ADC", rw: " ", m: "zpx"},
	0x76: {n: "ROR", rw: " ", m: "zpx"},
	0x77: {n: "RRA", rw: " ", m: "zpx"},
	0x78: {n: "SEI", rw: " ", m: "imp"},
	0x79: {n: "ADC", rw: " ", m: "aby"},
	0x7A: {n: "NOP", rw: " ", m: "imp", f: NOP},
	0x7B: {n: "RRA", rw: " ", m: "aby"},
	0x7C: {n: "NOP", rw: " ", m: "abx", f: NOP},
	0x7D: {n: "ADC", rw: " ", m: "abx"},
	0x7E: {n: "ROR", rw: " ", m: "abx"},
	0x7F: {n: "RRA", rw: " ", m: "abx"},
	0x80: {n: "NOP", rw: " ", m: "imm", f: NOP},
	0x81: {n: "STA", rw: " ", m: "izx"},
	0x82: {n: "NOP", rw: " ", m: "imm", f: NOP},
	0x83: {n: "SAX", rw: " ", m: "izx"},
	0x84: {n: "STY", rw: " ", m: "zpg"},
	0x85: {n: "STA", rw: " ", m: "zpg"},
	0x86: {n: "STX", rw: " ", m: "zpg"},
	0x87: {n: "SAX", rw: " ", m: "zpg"},
	0x88: {n: "DEY", rw: " ", m: "imp"},
	0x89: {n: "NOP", rw: " ", m: "imm", f: NOP},
	0x8A: {n: "TXA", rw: " ", m: "imp"},
	0x8B: {n: "ANE", rw: " ", m: "imm"},
	0x8C: {n: "STY", rw: " ", m: "abs"},
	0x8D: {n: "STA", rw: " ", m: "abs"},
	0x8E: {n: "STX", rw: " ", m: "abs"},
	0x8F: {n: "SAX", rw: " ", m: "abs"},
	0x90: {n: "BCC", rw: " ", m: "rel", f: branch(0, false)},
	0x91: {n: "STA", rw: " ", m: "izy"},
	0x92: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0x93: {n: "SHA", rw: " ", m: "izy"},
	0x94: {n: "STY", rw: " ", m: "zpx"},
	0x95: {n: "STA", rw: " ", m: "zpx"},
	0x96: {n: "STX", rw: " ", m: "zpy"},
	0x97: {n: "SAX", rw: " ", m: "zpy"},
	0x98: {n: "TYA", rw: " ", m: "imp"},
	0x99: {n: "STA", rw: " ", m: "aby"},
	0x9A: {n: "TXS", rw: " ", m: "imp"},
	0x9B: {n: "TAS", rw: " ", m: "aby"},
	0x9C: {n: "SHY", rw: " ", m: "abx"},
	0x9D: {n: "STA", rw: " ", m: "abx"},
	0x9E: {n: "SHX", rw: " ", m: "aby"},
	0x9F: {n: "SHA", rw: " ", m: "aby"},
	0xA0: {n: "LDY", rw: " ", m: "imm"},
	0xA1: {n: "LDA", rw: " ", m: "izx"},
	0xA2: {n: "LDX", rw: " ", m: "imm"},
	0xA3: {n: "LAX", rw: " ", m: "izx"},
	0xA4: {n: "LDY", rw: " ", m: "zpg"},
	0xA5: {n: "LDA", rw: " ", m: "zpg"},
	0xA6: {n: "LDX", rw: " ", m: "zpg"},
	0xA7: {n: "LAX", rw: " ", m: "zpg"},
	0xA8: {n: "TAY", rw: " ", m: "imp"},
	0xA9: {n: "LDA", rw: " ", m: "imm"},
	0xAA: {n: "TAX", rw: " ", m: "imp"},
	0xAB: {n: "LXA", rw: " ", m: "imm"},
	0xAC: {n: "LDY", rw: " ", m: "abs"},
	0xAD: {n: "LDA", rw: " ", m: "abs"},
	0xAE: {n: "LDX", rw: " ", m: "abs"},
	0xAF: {n: "LAX", rw: " ", m: "abs"},
	0xB0: {n: "BCS", rw: " ", m: "rel", f: branch(0, true)},
	0xB1: {n: "LDA", rw: " ", m: "izy"},
	0xB2: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0xB3: {n: "LAX", rw: " ", m: "izy"},
	0xB4: {n: "LDY", rw: " ", m: "zpx"},
	0xB5: {n: "LDA", rw: " ", m: "zpx"},
	0xB6: {n: "LDX", rw: " ", m: "zpy"},
	0xB7: {n: "LAX", rw: " ", m: "zpy"},
	0xB8: {n: "CLV", rw: " ", m: "imp"},
	0xB9: {n: "LDA", rw: " ", m: "aby"},
	0xBA: {n: "TSX", rw: " ", m: "imp"},
	0xBB: {n: "LAS", rw: " ", m: "aby"},
	0xBC: {n: "LDY", rw: " ", m: "abx"},
	0xBD: {n: "LDA", rw: " ", m: "abx"},
	0xBE: {n: "LDX", rw: " ", m: "aby"},
	0xBF: {n: "LAX", rw: " ", m: "aby"},
	0xC0: {n: "CPY", rw: " ", m: "imm"},
	0xC1: {n: "CMP", rw: " ", m: "izx"},
	0xC2: {n: "NOP", rw: " ", m: "imm", f: NOP},
	0xC3: {n: "DCP", rw: " ", m: "izx"},
	0xC4: {n: "CPY", rw: " ", m: "zpg"},
	0xC5: {n: "CMP", rw: " ", m: "zpg"},
	0xC6: {n: "DEC", rw: " ", m: "zpg"},
	0xC7: {n: "DCP", rw: " ", m: "zpg"},
	0xC8: {n: "INY", rw: " ", m: "imp"},
	0xC9: {n: "CMP", rw: " ", m: "imm"},
	0xCA: {n: "DEX", rw: " ", m: "imp"},
	0xCB: {n: "SBX", rw: " ", m: "imm"},
	0xCC: {n: "CPY", rw: " ", m: "abs"},
	0xCD: {n: "CMP", rw: " ", m: "abs"},
	0xCE: {n: "DEC", rw: " ", m: "abs"},
	0xCF: {n: "DCP", rw: " ", m: "abs"},
	0xD0: {n: "BNE", rw: " ", m: "rel", f: branch(1, false)},
	0xD1: {n: "CMP", rw: " ", m: "izy"},
	0xD2: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0xD3: {n: "DCP", rw: " ", m: "izy"},
	0xD4: {n: "NOP", rw: " ", m: "zpx", f: NOP},
	0xD5: {n: "CMP", rw: " ", m: "zpx"},
	0xD6: {n: "DEC", rw: " ", m: "zpx"},
	0xD7: {n: "DCP", rw: " ", m: "zpx"},
	0xD8: {n: "CLD", rw: " ", m: "imp"},
	0xD9: {n: "CMP", rw: " ", m: "aby"},
	0xDA: {n: "NOP", rw: " ", m: "imp", f: NOP},
	0xDB: {n: "DCP", rw: " ", m: "aby"},
	0xDC: {n: "NOP", rw: " ", m: "abx", f: NOP},
	0xDD: {n: "CMP", rw: " ", m: "abx"},
	0xDE: {n: "DEC", rw: " ", m: "abx"},
	0xDF: {n: "DCP", rw: " ", m: "abx"},
	0xE0: {n: "CPX", rw: " ", m: "imm"},
	0xE1: {n: "SBC", rw: " ", m: "izx"},
	0xE2: {n: "NOP", rw: " ", m: "imm", f: NOP},
	0xE3: {n: "ISC", rw: " ", m: "izx"},
	0xE4: {n: "CPX", rw: " ", m: "zpg"},
	0xE5: {n: "SBC", rw: " ", m: "zpg"},
	0xE6: {n: "INC", rw: " ", m: "zpg"},
	0xE7: {n: "ISC", rw: " ", m: "zpg"},
	0xE8: {n: "INX", rw: " ", m: "imp"},
	0xE9: {n: "SBC", rw: " ", m: "imm"},
	0xEA: {n: "NOP", rw: " ", m: "imp", f: NOP},
	0xEB: {n: "SBC", rw: " ", m: "imm"},
	0xEC: {n: "CPX", rw: " ", m: "abs"},
	0xED: {n: "SBC", rw: " ", m: "abs"},
	0xEE: {n: "INC", rw: " ", m: "abs"},
	0xEF: {n: "ISC", rw: " ", m: "abs"},
	0xF0: {n: "BEQ", rw: " ", m: "rel", f: branch(1, true)},
	0xF1: {n: "SBC", rw: " ", m: "izy"},
	0xF2: {n: "JAM", rw: " ", m: "imm", f: JAM},
	0xF3: {n: "ISC", rw: " ", m: "izy"},
	0xF4: {n: "NOP", rw: " ", m: "zpx", f: NOP},
	0xF5: {n: "SBC", rw: " ", m: "zpx"},
	0xF6: {n: "INC", rw: " ", m: "zpx"},
	0xF7: {n: "ISC", rw: " ", m: "zpx"},
	0xF8: {n: "SED", rw: " ", m: "imp"},
	0xF9: {n: "SBC", rw: " ", m: "aby"},
	0xFA: {n: "NOP", rw: " ", m: "imp", f: NOP},
	0xFB: {n: "ISC", rw: " ", m: "aby"},
	0xFC: {n: "NOP", rw: " ", m: "abx", f: NOP},
	0xFD: {n: "SBC", rw: " ", m: "abx"},
	0xFE: {n: "INC", rw: " ", m: "abx"},
	0xFF: {n: "ISC", rw: " ", m: "abx"},
}

const (
	xcpc       = 1 << iota // extra cycle for page crosses
	xca                    // extra cycle always
	unofficial             // so-called 'illegal' opcodes
)

var details = [256]uint8{
	0x00: 0, 0x01: 0, 0x02: 4, 0x03: 4, 0x04: 4, 0x05: 0, 0x06: 0, 0x07: 4, 0x08: 0, 0x09: 0, 0x0A: 0, 0x0B: 4, 0x0C: 4, 0x0D: 0, 0x0E: 0, 0x0F: 4,
	0x10: 0, 0x11: 1, 0x12: 4, 0x13: 6, 0x14: 4, 0x15: 0, 0x16: 0, 0x17: 4, 0x18: 0, 0x19: 1, 0x1A: 4, 0x1B: 4, 0x1C: 5, 0x1D: 1, 0x1E: 0, 0x1F: 4,
	0x20: 0, 0x21: 0, 0x22: 4, 0x23: 4, 0x24: 0, 0x25: 0, 0x26: 0, 0x27: 4, 0x28: 0, 0x29: 0, 0x2A: 0, 0x2B: 4, 0x2C: 0, 0x2D: 0, 0x2E: 0, 0x2F: 4,
	0x30: 0, 0x31: 1, 0x32: 4, 0x33: 6, 0x34: 4, 0x35: 0, 0x36: 0, 0x37: 4, 0x38: 0, 0x39: 1, 0x3A: 4, 0x3B: 4, 0x3C: 5, 0x3D: 1, 0x3E: 0, 0x3F: 4,
	0x40: 0, 0x41: 0, 0x42: 4, 0x43: 4, 0x44: 4, 0x45: 0, 0x46: 0, 0x47: 4, 0x48: 0, 0x49: 0, 0x4A: 0, 0x4B: 4, 0x4C: 0, 0x4D: 0, 0x4E: 0, 0x4F: 4,
	0x50: 0, 0x51: 1, 0x52: 4, 0x53: 6, 0x54: 4, 0x55: 0, 0x56: 0, 0x57: 4, 0x58: 0, 0x59: 1, 0x5A: 4, 0x5B: 4, 0x5C: 5, 0x5D: 1, 0x5E: 0, 0x5F: 4,
	0x60: 0, 0x61: 0, 0x62: 4, 0x63: 4, 0x64: 4, 0x65: 0, 0x66: 0, 0x67: 4, 0x68: 0, 0x69: 0, 0x6A: 0, 0x6B: 4, 0x6C: 0, 0x6D: 0, 0x6E: 0, 0x6F: 4,
	0x70: 0, 0x71: 1, 0x72: 4, 0x73: 6, 0x74: 4, 0x75: 0, 0x76: 0, 0x77: 4, 0x78: 0, 0x79: 1, 0x7A: 4, 0x7B: 4, 0x7C: 5, 0x7D: 1, 0x7E: 0, 0x7F: 4,
	0x80: 4, 0x81: 0, 0x82: 4, 0x83: 4, 0x84: 0, 0x85: 0, 0x86: 0, 0x87: 4, 0x88: 0, 0x89: 4, 0x8A: 0, 0x8B: 4, 0x8C: 0, 0x8D: 0, 0x8E: 0, 0x8F: 4,
	0x90: 0, 0x91: 2, 0x92: 4, 0x93: 4, 0x94: 0, 0x95: 0, 0x96: 0, 0x97: 4, 0x98: 0, 0x99: 0, 0x9A: 0, 0x9B: 4, 0x9C: 4, 0x9D: 0, 0x9E: 4, 0x9F: 4,
	0xA0: 0, 0xA1: 0, 0xA2: 0, 0xA3: 4, 0xA4: 0, 0xA5: 0, 0xA6: 0, 0xA7: 4, 0xA8: 0, 0xA9: 0, 0xAA: 0, 0xAB: 4, 0xAC: 0, 0xAD: 0, 0xAE: 0, 0xAF: 4,
	0xB0: 0, 0xB1: 1, 0xB2: 4, 0xB3: 5, 0xB4: 0, 0xB5: 0, 0xB6: 0, 0xB7: 4, 0xB8: 0, 0xB9: 1, 0xBA: 0, 0xBB: 4, 0xBC: 1, 0xBD: 1, 0xBE: 1, 0xBF: 5,
	0xC0: 0, 0xC1: 0, 0xC2: 4, 0xC3: 4, 0xC4: 0, 0xC5: 0, 0xC6: 0, 0xC7: 4, 0xC8: 0, 0xC9: 0, 0xCA: 0, 0xCB: 4, 0xCC: 0, 0xCD: 0, 0xCE: 0, 0xCF: 4,
	0xD0: 0, 0xD1: 1, 0xD2: 4, 0xD3: 4, 0xD4: 4, 0xD5: 0, 0xD6: 0, 0xD7: 4, 0xD8: 0, 0xD9: 1, 0xDA: 4, 0xDB: 4, 0xDC: 5, 0xDD: 1, 0xDE: 0, 0xDF: 4,
	0xE0: 0, 0xE1: 0, 0xE2: 4, 0xE3: 4, 0xE4: 0, 0xE5: 0, 0xE6: 0, 0xE7: 4, 0xE8: 0, 0xE9: 0, 0xEA: 0, 0xEB: 4, 0xEC: 0, 0xED: 0, 0xEE: 0, 0xEF: 4,
	0xF0: 0, 0xF1: 1, 0xF2: 4, 0xF3: 6, 0xF4: 4, 0xF5: 0, 0xF6: 0, 0xF7: 4, 0xF8: 0, 0xF9: 1, 0xFA: 4, 0xFB: 4, 0xFC: 5, 0xFD: 1, 0xFE: 0, 0xFF: 4,
}

type Generator struct {
	io.Writer
	outbuf bytes.Buffer
	out    io.Writer

	mode addrmode
}

func (g *Generator) opcodeHeader(code int) {
	modestr := ""
	g.mode = nil
	switch defs[code].m {
	case "imp":
		modestr = `implied addressing.`
	case "acc":
		modestr = `adressing accumulator.`
	case "rel":
		modestr = `relative addressing.`
		g.mode = rel
	case "abs":
		modestr = `absolute addressing.`
		g.mode = abs
	case "abx":
		modestr = `absolute indexed X.`
		g.mode = abx
	case "aby":
		modestr = `absolute indexed Y.`
		g.mode = aby
	case "imm":
		modestr = `immediate addressing.`
		g.mode = imm
	case "ind":
		modestr = `indirect addressing.`
		g.mode = ind
	case "izx":
		modestr = `indexed addressing (abs, X).`
		g.mode = izx
	case "izy":
		modestr = `indexed addressing (abs),Y.`
		g.mode = izy
	case "zpg":
		modestr = `zero page addressing.`
		g.mode = zpg
	case "zpx":
		modestr = `indexed addressing: zeropage,X.`
		g.mode = zpx
	case "zpy":
		modestr = `indexed addressing: zeropage,Y.`
		g.mode = zpy
	}

	g.printf(`// %s   %02X`, defs[code].n, code)
	g.printf(`// %s`, modestr)
	g.printf(`func opcode_%02X(cpu*CPU){`, code)
	if g.mode != nil {
		g.mode(g, details[code])
		g.printf(`_ = oper`)
	}

	switch defs[code].rw {
	case "w":
		switch defs[code].m {
		case "acc":
			panic("not implemented")
		default:
			// Read the value
			g.printf(`val := cpu.Read8(oper)`)
		}
	}
}

func (g *Generator) opcodeFooter(code int) {
	switch defs[code].rw {
	case "w":
		switch defs[code].m {
		case "acc":
			panic("not implemented")
		default:
			// Write the value
			g.printf(`cpu.Write8(oper, val)`)
		}
	}
	g.printf(`}`)
}

func (g *Generator) printf(format string, args ...any) {
	fmt.Fprintf(g, "%s\n", fmt.Sprintf(format, args...))
}

// read 16 bytes from the zero page, handling page wrap.
func r16zpwrap(g *Generator) {
	g.printf(`// read 16 bytes from the zero page, handling page wrap`)
	g.printf(`lo := cpu.Read8(oper)`)
	g.printf(`hi := cpu.Read8(uint16(uint8(oper) + 1))`)
	g.printf(`oper = uint16(hi)<<8 | uint16(lo)`)
}

func branch(ibit int, val bool) func(g *Generator) {
	return func(g *Generator) {
		g.printf(`if cpu.P.bit(%d) == %t {`, ibit, val)
		g.printf(`// branching`)
		pagecrossed(g, "cpu.PC+1", "oper")
		g.printf(`	cpu.tick()`)
		g.printf(`	cpu.PC = oper`)
		g.printf(`	return`)
		g.printf(`}`)
		g.printf(`cpu.PC++`)
	}
}

func pagecrossed(g *Generator, a, b string) {
	g.printf(`	if pagecrossed(%s, %s) {`, a, b)
	g.printf(`		cpu.tick()`)
	g.printf(`	}`)
}

type addrmode func(g *Generator, details uint8)

func ind(g *Generator, _ uint8) {
	g.printf(`oper := cpu.Read16(cpu.PC)`)
	g.printf(`lo := cpu.Read8(oper)`)
	g.printf(`// 2 bytes address wrap around`)
	g.printf(`hi := cpu.Read8((0xff00 & oper) | (0x00ff & (oper + 1)))`)
	g.printf(`oper = uint16(hi)<<8 | uint16(lo)`)
}

func acc(g *Generator, _ uint8) {
	g.printf(`panic("not implemented")`)
}

func imm(g *Generator, _ uint8) {
	g.printf(`oper := cpu.PC`)
	g.printf(`cpu.PC++`)
}

func rel(g *Generator, _ uint8) {
	g.printf(`off := int8(cpu.Read8(cpu.PC))`)
	g.printf(`oper := uint16(int16(cpu.PC+1) + int16(off))`)
}

func abs(g *Generator, _ uint8) {
	g.printf(`oper := cpu.Read16(cpu.PC)`)
	g.printf(`cpu.PC += 2`)
}

func abx(g *Generator, info uint8) {
	switch {
	case info&xcpc != 0:
		g.printf(`addr := cpu.Read16(cpu.PC)`)
		g.printf(`cpu.PC += 2`)
		g.printf(`oper := addr + uint16(cpu.X)`)
		pagecrossed(g, "oper", "addr")
	default:
		g.printf(`cpu.tick()`)
		g.printf(`oper := cpu.Read16(cpu.PC)`)
		g.printf(`cpu.PC += 2`)
		g.printf(`oper += uint16(cpu.X)`)
	}
}

func aby(g *Generator, info uint8) {
	switch {
	case info&xcpc != 0:
		g.printf(`// extra cycle for page cross`)
		g.printf(`addr := cpu.Read16(cpu.PC)`)
		g.printf(`cpu.PC += 2`)
		g.printf(`oper := addr + uint16(cpu.Y)`)
		pagecrossed(g, "oper", "addr")
	default:
		g.printf(`// default`)
		g.printf(`cpu.tick()`)
		g.printf(`oper := cpu.Read16(cpu.PC)`)
		g.printf(`cpu.PC += 2`)
		g.printf(`oper += uint16(cpu.Y)`)
	}
}

func zpg(g *Generator, _ uint8) {
	g.printf(`oper := uint16(cpu.Read8(cpu.PC))`)
	g.printf(`cpu.PC++`)
}

func zpx(g *Generator, _ uint8) {
	g.printf(`cpu.tick()`)
	g.printf(`addr := cpu.Read8(cpu.PC)`)
	g.printf(`cpu.PC++`)
	g.printf(`oper := uint16(addr) + uint16(cpu.X)`)
	g.printf(`oper &= 0xff`)
}

func zpy(g *Generator, _ uint8) {
	g.printf(`cpu.tick()`)
	g.printf(`addr := cpu.Read8(cpu.PC)`)
	g.printf(`cpu.PC++`)
	g.printf(`oper := uint16(addr) + uint16(cpu.Y)`)
	g.printf(`oper &= 0xff`)
}

func izx(g *Generator, info uint8) {
	g.printf(`cpu.tick()`)
	zpg(g, info)
	g.printf(`oper = uint16(uint8(oper) + cpu.X)`)
	r16zpwrap(g)
}

func izy(g *Generator, info uint8) {
	switch {
	case info&xcpc != 0:
		g.printf(`// extra cycle for page cross`)
		zpg(g, info)
		r16zpwrap(g)
		pagecrossed(g, "oper", "oper+uint16(cpu.Y)")
		g.printf(`oper += uint16(cpu.Y)`)
	case info&xca != 0:
		g.printf(`// extra cycle always`)
		zpg(g, info)
		r16zpwrap(g)
		g.printf(`cpu.tick()`)
		g.printf(`oper += uint16(cpu.Y)`)
	default:
		g.printf(`// default`)
		zpg(g, info)
		r16zpwrap(g)
		g.printf(`oper += uint16(cpu.Y)`)
	}
}

func BRK(g *Generator) {
	g.printf(`cpu.tick()`)
	g.printf(`push16(cpu, cpu.PC+1)`)
	g.printf(`p := cpu.P`)
	g.printf(`p.setBit(pbitB)`)
	g.printf(`push8(cpu, uint8(p))`)
	g.printf(`cpu.P.setBit(pbitI)`)
	g.printf(`cpu.PC = cpu.Read16(IRQvector)`)
}

func ORA(g *Generator) {
	g.printf(`val := cpu.Read8(oper)`)
	g.printf(`cpu.A |= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func SLO(g *Generator) {
	ASL(g)
	g.printf(`cpu.A |= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func ASL(g *Generator) {
	g.printf(`carry := val & 0x80 // carry is bit 7`)
	g.printf(`val <<= 1`)
	g.printf(`val &= 0xfe`)
	g.printf(`cpu.tick()`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.writeBit(pbitC, carry != 0)`)
}

func NOP(g *Generator) {
	g.printf(`cpu.tick()`)
}

func JAM(g *Generator) {
	g.printf(`panic("Halt and catch fire!")`)
}

func (g *Generator) generate() {
	// TODO(arl) temporary code
	var generated [256]bool
	// TODO(arl) end

	g.printf(`// Code generated by cpugen/gen_nes6502.go. DO NOT EDIT.`)
	g.printf(`package emu`)
	for code, def := range defs {
		if def.f == nil {
			log.Printf("skipping 0x%02X opcode", code)
			continue
		}

		g.opcodeHeader(code)
		def.f(g)
		g.opcodeFooter(code)
		generated[code] = true
	}

	// TODO(arl) temporary code
	g.printf(`var gdefs = [256]func(*CPU){`)
	for code := range generated {
		if generated[code] {
			g.printf(`0x%02X: opcode_%02X,`, code, code)
		}
	}
	g.printf(`}`)
	// TODO(arl) end

}

func main() {
	log.SetFlags(0)
	outf := flag.String("out", "cpu_ops.go", "output file")
	flag.Parse()

	bb := &bytes.Buffer{}
	g := &Generator{Writer: bb}

	g.generate()

	buf, err := format.Source(bb.Bytes())
	if err != nil {
		if err := os.WriteFile(*outf, bb.Bytes(), 0644); err != nil {
			fatalf("can't write to %s: %s", *outf, err)
		}
		fatalf("'gofmt' failed\n%s", err)
	}

	if err := os.WriteFile(*outf, buf, 0644); err != nil {
		fatalf("can't write to %s: %s", *outf, err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fatal error:")
	fmt.Fprintf(os.Stderr, "\n\t%s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}
