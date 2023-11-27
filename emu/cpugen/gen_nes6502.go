package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"strings"
)

type opdef struct {
	i uint8  // opcode value (same as index into 'defs')
	n string // name
	m string // addressing mode
	f func(g *Generator, def opdef)

	// " " -> do nothing
	// "r" -> declare 'val' and set it to 'oper/accumulator'
	// "w" -> write 'val' back into 'oper/accumulator'
	rw string
}

var defs = [256]opdef{
	{i: 0x00, n: "BRK", rw: "  ", m: "imp", f: BRK},
	{i: 0x01, n: "ORA", rw: "r ", m: "izx", f: ORA},
	{i: 0x02, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0x03, n: "SLO", rw: "rw", m: "izx", f: SLO},
	{i: 0x04, n: "NOP", rw: "  ", m: "zpg", f: NOP},
	{i: 0x05, n: "ORA", rw: "r ", m: "zpg", f: ORA},
	{i: 0x06, n: "ASL", rw: "rw", m: "zpg", f: ASL},
	{i: 0x07, n: "SLO", rw: "rw", m: "zpg", f: SLO},
	{i: 0x08, n: "PHP", rw: "  ", m: "imp", f: PHP},
	{i: 0x09, n: "ORA", rw: "r ", m: "imm", f: ORA},
	{i: 0x0A, n: "ASL", rw: "rw", m: "acc", f: ASL},
	{i: 0x0B, n: "ANC", rw: "r ", m: "imm", f: ANC},
	{i: 0x0C, n: "NOP", rw: "  ", m: "abs", f: NOP},
	{i: 0x0D, n: "ORA", rw: "r ", m: "abs", f: ORA},
	{i: 0x0E, n: "ASL", rw: "rw", m: "abs", f: ASL},
	{i: 0x0F, n: "SLO", rw: "rw", m: "abs", f: SLO},
	{i: 0x10, n: "BPL", rw: "  ", m: "rel", f: branch(7, false)},
	{i: 0x11, n: "ORA", rw: "r ", m: "izy", f: ORA},
	{i: 0x12, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0x13, n: "SLO", rw: "rw", m: "izy", f: SLO},
	{i: 0x14, n: "NOP", rw: "  ", m: "zpx", f: NOP},
	{i: 0x15, n: "ORA", rw: "r ", m: "zpx", f: ORA},
	{i: 0x16, n: "ASL", rw: "rw", m: "zpx", f: ASL},
	{i: 0x17, n: "SLO", rw: "rw", m: "zpx", f: SLO},
	{i: 0x18, n: "CLC", rw: "  ", m: "imp", f: clearFlag(0)},
	{i: 0x19, n: "ORA", rw: "r ", m: "aby", f: ORA},
	{i: 0x1A, n: "NOP", rw: "  ", m: "imp", f: NOP},
	{i: 0x1B, n: "SLO", rw: "rw", m: "aby", f: SLO},
	{i: 0x1C, n: "NOP", rw: "  ", m: "abx", f: NOP},
	{i: 0x1D, n: "ORA", rw: "r ", m: "abx", f: ORA},
	{i: 0x1E, n: "ASL", rw: "rw", m: "abx", f: ASL},
	{i: 0x1F, n: "SLO", rw: "rw", m: "abx", f: SLO},
	{i: 0x20, n: "JSR", rw: "  ", m: "imp", f: JSR}, // special case. should be 'abs' but handle it as 'implied'
	{i: 0x21, n: "AND", rw: "r ", m: "izx", f: AND},
	{i: 0x22, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0x23, n: "RLA", rw: "rw", m: "izx", f: RLA},
	{i: 0x24, n: "BIT", rw: "r ", m: "zpg", f: BIT},
	{i: 0x25, n: "AND", rw: "r ", m: "zpg", f: AND},
	{i: 0x26, n: "ROL", rw: "rw", m: "zpg", f: ROL},
	{i: 0x27, n: "RLA", rw: "rw", m: "zpg", f: RLA},
	{i: 0x28, n: "PLP", rw: "  ", m: "imp", f: PLP},
	{i: 0x29, n: "AND", rw: "r ", m: "imm", f: AND},
	{i: 0x2A, n: "ROL", rw: "rw", m: "acc", f: ROL},
	{i: 0x2B, n: "ANC", rw: "r ", m: "imm", f: ANC},
	{i: 0x2C, n: "BIT", rw: "r ", m: "abs", f: BIT},
	{i: 0x2D, n: "AND", rw: "r ", m: "abs", f: AND},
	{i: 0x2E, n: "ROL", rw: "rw", m: "abs", f: ROL},
	{i: 0x2F, n: "RLA", rw: "rw", m: "abs", f: RLA},
	{i: 0x30, n: "BMI", rw: "  ", m: "rel", f: branch(7, true)},
	{i: 0x31, n: "AND", rw: "r ", m: "izy", f: AND},
	{i: 0x32, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0x33, n: "RLA", rw: "rw", m: "izy", f: RLA},
	{i: 0x34, n: "NOP", rw: "  ", m: "zpx", f: NOP},
	{i: 0x35, n: "AND", rw: "r ", m: "zpx", f: AND},
	{i: 0x36, n: "ROL", rw: "rw", m: "zpx", f: ROL},
	{i: 0x37, n: "RLA", rw: "rw", m: "zpx", f: RLA},
	{i: 0x38, n: "SEC", rw: "  ", m: "imp", f: setFlag(0)},
	{i: 0x39, n: "AND", rw: "r ", m: "aby", f: AND},
	{i: 0x3A, n: "NOP", rw: "  ", m: "imp", f: NOP},
	{i: 0x3B, n: "RLA", rw: "rw", m: "aby", f: RLA},
	{i: 0x3C, n: "NOP", rw: "  ", m: "abx", f: NOP},
	{i: 0x3D, n: "AND", rw: "r ", m: "abx", f: AND},
	{i: 0x3E, n: "ROL", rw: "rw", m: "abx", f: ROL},
	{i: 0x3F, n: "RLA", rw: "rw", m: "abx", f: RLA},
	{i: 0x40, n: "RTI", rw: "  ", m: "imp", f: RTI},
	{i: 0x41, n: "EOR", rw: "r ", m: "izx", f: EOR},
	{i: 0x42, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0x43, n: "SRE", rw: "rw", m: "izx", f: SRE},
	{i: 0x44, n: "NOP", rw: "  ", m: "zpg", f: NOP},
	{i: 0x45, n: "EOR", rw: "r ", m: "zpg", f: EOR},
	{i: 0x46, n: "LSR", rw: "rw", m: "zpg", f: LSR},
	{i: 0x47, n: "SRE", rw: "rw", m: "zpg", f: SRE},
	{i: 0x48, n: "PHA", rw: "  ", m: "imp", f: PHA},
	{i: 0x49, n: "EOR", rw: "r ", m: "imm", f: EOR},
	{i: 0x4A, n: "LSR", rw: "rw", m: "acc", f: LSR},
	{i: 0x4B, n: "ALR", rw: "r ", m: "imm", f: ALR},
	{i: 0x4C, n: "JMP", rw: "  ", m: "abs", f: JMP},
	{i: 0x4D, n: "EOR", rw: "r ", m: "abs", f: EOR},
	{i: 0x4E, n: "LSR", rw: "rw", m: "abs", f: LSR},
	{i: 0x4F, n: "SRE", rw: "rw", m: "abs", f: SRE},
	{i: 0x50, n: "BVC", rw: "  ", m: "rel", f: branch(6, false)},
	{i: 0x51, n: "EOR", rw: "r ", m: "izy", f: EOR},
	{i: 0x52, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0x53, n: "SRE", rw: "rw", m: "izy", f: SRE},
	{i: 0x54, n: "NOP", rw: "  ", m: "zpx", f: NOP},
	{i: 0x55, n: "EOR", rw: "r ", m: "zpx", f: EOR},
	{i: 0x56, n: "LSR", rw: "rw", m: "zpx", f: LSR},
	{i: 0x57, n: "SRE", rw: "rw", m: "zpx", f: SRE},
	{i: 0x58, n: "CLI", rw: "  ", m: "imp", f: clearFlag(2)},
	{i: 0x59, n: "EOR", rw: "r ", m: "aby", f: EOR},
	{i: 0x5A, n: "NOP", rw: "  ", m: "imp", f: NOP},
	{i: 0x5B, n: "SRE", rw: "rw", m: "aby", f: SRE},
	{i: 0x5C, n: "NOP", rw: "  ", m: "abx", f: NOP},
	{i: 0x5D, n: "EOR", rw: "r ", m: "abx", f: EOR},
	{i: 0x5E, n: "LSR", rw: "rw", m: "abx", f: LSR},
	{i: 0x5F, n: "SRE", rw: "rw", m: "abx", f: SRE},
	{i: 0x60, n: "RTS", rw: "  ", m: "imp", f: RTS},
	{i: 0x61, n: "ADC", rw: "r ", m: "izx", f: ADC},
	{i: 0x62, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0x63, n: "RRA", rw: "rw", m: "izx", f: RRA},
	{i: 0x64, n: "NOP", rw: "  ", m: "zpg", f: NOP},
	{i: 0x65, n: "ADC", rw: "r ", m: "zpg", f: ADC},
	{i: 0x66, n: "ROR", rw: "rw", m: "zpg", f: ROR},
	{i: 0x67, n: "RRA", rw: "rw", m: "zpg", f: RRA},
	{i: 0x68, n: "PLA", rw: "  ", m: "imp", f: PLA},
	{i: 0x69, n: "ADC", rw: "r ", m: "imm", f: ADC},
	{i: 0x6A, n: "ROR", rw: "rw", m: "acc", f: ROR},
	{i: 0x6B, n: "ARR", rw: "r ", m: "imm", f: ARR},
	{i: 0x6C, n: "JMP", rw: "  ", m: "ind", f: JMP},
	{i: 0x6D, n: "ADC", rw: "r ", m: "abs", f: ADC},
	{i: 0x6E, n: "ROR", rw: "rw", m: "abs", f: ROR},
	{i: 0x6F, n: "RRA", rw: "rw", m: "abs", f: RRA},
	{i: 0x70, n: "BVS", rw: "  ", m: "rel", f: branch(6, true)},
	{i: 0x71, n: "ADC", rw: "r ", m: "izy", f: ADC},
	{i: 0x72, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0x73, n: "RRA", rw: "rw", m: "izy", f: RRA},
	{i: 0x74, n: "NOP", rw: "  ", m: "zpx", f: NOP},
	{i: 0x75, n: "ADC", rw: "r ", m: "zpx", f: ADC},
	{i: 0x76, n: "ROR", rw: "rw", m: "zpx", f: ROR},
	{i: 0x77, n: "RRA", rw: "rw", m: "zpx", f: RRA},
	{i: 0x78, n: "SEI", rw: "  ", m: "imp", f: setFlag(2)},
	{i: 0x79, n: "ADC", rw: "r ", m: "aby", f: ADC},
	{i: 0x7A, n: "NOP", rw: "  ", m: "imp", f: NOP},
	{i: 0x7B, n: "RRA", rw: "rw", m: "aby", f: RRA},
	{i: 0x7C, n: "NOP", rw: "  ", m: "abx", f: NOP},
	{i: 0x7D, n: "ADC", rw: "r ", m: "abx", f: ADC},
	{i: 0x7E, n: "ROR", rw: "rw", m: "abx", f: ROR},
	{i: 0x7F, n: "RRA", rw: "rw", m: "abx", f: RRA},
	{i: 0x80, n: "NOP", rw: "  ", m: "imm", f: NOP},
	{i: 0x81, n: "STA", rw: "  ", m: "izx", f: store("A")},
	{i: 0x82, n: "NOP", rw: "  ", m: "imm", f: NOP},
	{i: 0x83, n: "SAX", rw: "  ", m: "izx", f: SAX},
	{i: 0x84, n: "STY", rw: "  ", m: "zpg", f: store("Y")},
	{i: 0x85, n: "STA", rw: "  ", m: "zpg", f: store("A")},
	{i: 0x86, n: "STX", rw: "  ", m: "zpg", f: store("X")},
	{i: 0x87, n: "SAX", rw: "  ", m: "zpg", f: SAX},
	{i: 0x88, n: "DEY", rw: "  ", m: "imp", f: decrement("Y")},
	{i: 0x89, n: "NOP", rw: "  ", m: "imm", f: NOP},
	{i: 0x8A, n: "TXA", rw: "  ", m: "imp", f: transfer("X", "A")},
	{i: 0x8B, n: "ANE", rw: "  ", m: "imm", f: unstable},
	{i: 0x8C, n: "STY", rw: "  ", m: "abs", f: store("Y")},
	{i: 0x8D, n: "STA", rw: "  ", m: "abs", f: store("A")},
	{i: 0x8E, n: "STX", rw: "  ", m: "abs", f: store("X")},
	{i: 0x8F, n: "SAX", rw: "  ", m: "abs", f: SAX},
	{i: 0x90, n: "BCC", rw: "  ", m: "rel", f: branch(0, false)},
	{i: 0x91, n: "STA", rw: "  ", m: "izy", f: store("A")},
	{i: 0x92, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0x93, n: "SHA", rw: "  ", m: "izy", f: unstable},
	{i: 0x94, n: "STY", rw: "  ", m: "zpx", f: store("Y")},
	{i: 0x95, n: "STA", rw: "  ", m: "zpx", f: store("A")},
	{i: 0x96, n: "STX", rw: "  ", m: "zpy", f: store("X")},
	{i: 0x97, n: "SAX", rw: "  ", m: "zpy", f: SAX},
	{i: 0x98, n: "TYA", rw: "  ", m: "imp", f: transfer("Y", "A")},
	{i: 0x99, n: "STA", rw: "  ", m: "aby", f: store("A")},
	{i: 0x9A, n: "TXS", rw: "  ", m: "imp", f: transfer("X", "SP")},
	{i: 0x9B, n: "TAS", rw: "  ", m: "aby", f: unstable},
	{i: 0x9C, n: "SHY", rw: "  ", m: "abx", f: unstable},
	{i: 0x9D, n: "STA", rw: "  ", m: "abx", f: store("A")},
	{i: 0x9E, n: "SHX", rw: "  ", m: "aby", f: unstable},
	{i: 0x9F, n: "SHA", rw: "  ", m: "aby", f: unstable},
	{i: 0xA0, n: "LDY", rw: "r ", m: "imm", f: load("Y")},
	{i: 0xA1, n: "LDA", rw: "r ", m: "izx", f: load("A")},
	{i: 0xA2, n: "LDX", rw: "r ", m: "imm", f: load("X")},
	{i: 0xA3, n: "LAX", rw: "r ", m: "izx", f: load("A", "X")},
	{i: 0xA4, n: "LDY", rw: "r ", m: "zpg", f: load("Y")},
	{i: 0xA5, n: "LDA", rw: "r ", m: "zpg", f: load("A")},
	{i: 0xA6, n: "LDX", rw: "r ", m: "zpg", f: load("X")},
	{i: 0xA7, n: "LAX", rw: "r ", m: "zpg", f: load("A", "X")},
	{i: 0xA8, n: "TAY", rw: "  ", m: "imp", f: transfer("A", "Y")},
	{i: 0xA9, n: "LDA", rw: "r ", m: "imm", f: load("A")},
	{i: 0xAA, n: "TAX", rw: "  ", m: "imp", f: transfer("A", "X")},
	{i: 0xAB, n: "LXA", rw: "  ", m: "imm", f: unstable},
	{i: 0xAC, n: "LDY", rw: "r ", m: "abs", f: load("Y")},
	{i: 0xAD, n: "LDA", rw: "r ", m: "abs", f: load("A")},
	{i: 0xAE, n: "LDX", rw: "r ", m: "abs", f: load("X")},
	{i: 0xAF, n: "LAX", rw: "r ", m: "abs", f: load("A", "X")},
	{i: 0xB0, n: "BCS", rw: "  ", m: "rel", f: branch(0, true)},
	{i: 0xB1, n: "LDA", rw: "r ", m: "izy", f: load("A")},
	{i: 0xB2, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0xB3, n: "LAX", rw: "r ", m: "izy", f: load("A", "X")},
	{i: 0xB4, n: "LDY", rw: "r ", m: "zpx", f: load("Y")},
	{i: 0xB5, n: "LDA", rw: "r ", m: "zpx", f: load("A")},
	{i: 0xB6, n: "LDX", rw: "r ", m: "zpy", f: load("X")},
	{i: 0xB7, n: "LAX", rw: "r ", m: "zpy", f: load("A", "X")},
	{i: 0xB8, n: "CLV", rw: "  ", m: "imp", f: clearFlag(6)},
	{i: 0xB9, n: "LDA", rw: "r ", m: "aby", f: load("A")},
	{i: 0xBA, n: "TSX", rw: "  ", m: "imp", f: transfer("SP", "X")},
	{i: 0xBB, n: "LAS", rw: "r ", m: "aby", f: LAS},
	{i: 0xBC, n: "LDY", rw: "r ", m: "abx", f: load("Y")},
	{i: 0xBD, n: "LDA", rw: "r ", m: "abx", f: load("A")},
	{i: 0xBE, n: "LDX", rw: "r ", m: "aby", f: load("X")},
	{i: 0xBF, n: "LAX", rw: "r ", m: "aby", f: load("A", "X")},
	{i: 0xC0, n: "CPY", rw: "r ", m: "imm", f: compare("Y")},
	{i: 0xC1, n: "CMP", rw: "r ", m: "izx", f: compare("A")},
	{i: 0xC2, n: "NOP", rw: "  ", m: "imm", f: NOP},
	{i: 0xC3, n: "DCP", rw: "rw", m: "izx", f: DCP},
	{i: 0xC4, n: "CPY", rw: "r ", m: "zpg", f: compare("Y")},
	{i: 0xC5, n: "CMP", rw: "r ", m: "zpg", f: compare("A")},
	{i: 0xC6, n: "DEC", rw: "rw", m: "zpg", f: decrement("mem")},
	{i: 0xC7, n: "DCP", rw: "rw", m: "zpg", f: DCP},
	{i: 0xC8, n: "INY", rw: "  ", m: "imp", f: increment("Y")},
	{i: 0xC9, n: "CMP", rw: "r ", m: "imm", f: compare("A")},
	{i: 0xCA, n: "DEX", rw: "  ", m: "imp", f: decrement("X")},
	{i: 0xCB, n: "SBX", rw: "r ", m: "imm", f: SBX},
	{i: 0xCC, n: "CPY", rw: "r ", m: "abs", f: compare("Y")},
	{i: 0xCD, n: "CMP", rw: "r ", m: "abs", f: compare("A")},
	{i: 0xCE, n: "DEC", rw: "rw", m: "abs", f: decrement("mem")},
	{i: 0xCF, n: "DCP", rw: "rw", m: "abs", f: DCP},
	{i: 0xD0, n: "BNE", rw: "  ", m: "rel", f: branch(1, false)},
	{i: 0xD1, n: "CMP", rw: "r ", m: "izy", f: compare("A")},
	{i: 0xD2, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0xD3, n: "DCP", rw: "rw", m: "izy", f: DCP},
	{i: 0xD4, n: "NOP", rw: "  ", m: "zpx", f: NOP},
	{i: 0xD5, n: "CMP", rw: "r ", m: "zpx", f: compare("A")},
	{i: 0xD6, n: "DEC", rw: "rw", m: "zpx", f: decrement("mem")},
	{i: 0xD7, n: "DCP", rw: "rw", m: "zpx", f: DCP},
	{i: 0xD8, n: "CLD", rw: "  ", m: "imp", f: clearFlag(3)},
	{i: 0xD9, n: "CMP", rw: "r ", m: "aby", f: compare("A")},
	{i: 0xDA, n: "NOP", rw: "  ", m: "imp", f: NOP},
	{i: 0xDB, n: "DCP", rw: "rw", m: "aby", f: DCP},
	{i: 0xDC, n: "NOP", rw: "  ", m: "abx", f: NOP},
	{i: 0xDD, n: "CMP", rw: "r ", m: "abx", f: compare("A")},
	{i: 0xDE, n: "DEC", rw: "rw", m: "abx", f: decrement("mem")},
	{i: 0xDF, n: "DCP", rw: "rw", m: "abx", f: DCP},
	{i: 0xE0, n: "CPX", rw: "r ", m: "imm", f: compare("X")},
	{i: 0xE1, n: "SBC", rw: "r ", m: "izx", f: SBC},
	{i: 0xE2, n: "NOP", rw: "  ", m: "imm", f: NOP},
	{i: 0xE3, n: "ISB", rw: "rw", m: "izx", f: ISB},
	{i: 0xE4, n: "CPX", rw: "r ", m: "zpg", f: compare("X")},
	{i: 0xE5, n: "SBC", rw: "r ", m: "zpg", f: SBC},
	{i: 0xE6, n: "INC", rw: "rw", m: "zpg", f: increment("mem")},
	{i: 0xE7, n: "ISB", rw: "rw", m: "zpg", f: ISB},
	{i: 0xE8, n: "INX", rw: "  ", m: "imp", f: increment("X")},
	{i: 0xE9, n: "SBC", rw: "r ", m: "imm", f: SBC},
	{i: 0xEA, n: "NOP", rw: "  ", m: "imp", f: NOP},
	{i: 0xEB, n: "SBC", rw: "r ", m: "imm", f: SBC},
	{i: 0xEC, n: "CPX", rw: "r ", m: "abs", f: compare("X")},
	{i: 0xED, n: "SBC", rw: "r ", m: "abs", f: SBC},
	{i: 0xEE, n: "INC", rw: "rw", m: "abs", f: increment("mem")},
	{i: 0xEF, n: "ISB", rw: "rw", m: "abs", f: ISB},
	{i: 0xF0, n: "BEQ", rw: "  ", m: "rel", f: branch(1, true)},
	{i: 0xF1, n: "SBC", rw: "r ", m: "izy", f: SBC},
	{i: 0xF2, n: "JAM", rw: "  ", m: "imm", f: JAM},
	{i: 0xF3, n: "ISB", rw: "rw", m: "izy", f: ISB},
	{i: 0xF4, n: "NOP", rw: "  ", m: "zpx", f: NOP},
	{i: 0xF5, n: "SBC", rw: "r ", m: "zpx", f: SBC},
	{i: 0xF6, n: "INC", rw: "rw", m: "zpx", f: increment("mem")},
	{i: 0xF7, n: "ISB", rw: "rw", m: "zpx", f: ISB},
	{i: 0xF8, n: "SED", rw: "  ", m: "imp", f: setFlag(3)},
	{i: 0xF9, n: "SBC", rw: "r ", m: "aby", f: SBC},
	{i: 0xFA, n: "NOP", rw: "  ", m: "imp", f: NOP},
	{i: 0xFB, n: "ISB", rw: "rw", m: "aby", f: ISB},
	{i: 0xFC, n: "NOP", rw: "  ", m: "abx", f: NOP},
	{i: 0xFD, n: "SBC", rw: "r ", m: "abx", f: SBC},
	{i: 0xFE, n: "INC", rw: "rw", m: "abx", f: increment("mem")},
	{i: 0xFF, n: "ISB", rw: "rw", m: "abx", f: ISB},
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
	0xB0: 0, 0xB1: 1, 0xB2: 4, 0xB3: 5, 0xB4: 0, 0xB5: 0, 0xB6: 0, 0xB7: 4, 0xB8: 0, 0xB9: 1, 0xBA: 0, 0xBB: 5, 0xBC: 1, 0xBD: 1, 0xBE: 1, 0xBF: 5,
	0xC0: 0, 0xC1: 0, 0xC2: 4, 0xC3: 4, 0xC4: 0, 0xC5: 0, 0xC6: 0, 0xC7: 4, 0xC8: 0, 0xC9: 0, 0xCA: 0, 0xCB: 4, 0xCC: 0, 0xCD: 0, 0xCE: 0, 0xCF: 4,
	0xD0: 0, 0xD1: 1, 0xD2: 4, 0xD3: 6, 0xD4: 4, 0xD5: 0, 0xD6: 0, 0xD7: 4, 0xD8: 0, 0xD9: 1, 0xDA: 4, 0xDB: 4, 0xDC: 5, 0xDD: 1, 0xDE: 0, 0xDF: 4,
	0xE0: 0, 0xE1: 0, 0xE2: 4, 0xE3: 4, 0xE4: 0, 0xE5: 0, 0xE6: 0, 0xE7: 4, 0xE8: 0, 0xE9: 0, 0xEA: 0, 0xEB: 4, 0xEC: 0, 0xED: 0, 0xEE: 0, 0xEF: 4,
	0xF0: 0, 0xF1: 1, 0xF2: 4, 0xF3: 6, 0xF4: 4, 0xF5: 0, 0xF6: 0, 0xF7: 4, 0xF8: 0, 0xF9: 1, 0xFA: 4, 0xFB: 4, 0xFC: 5, 0xFD: 1, 0xFE: 0, 0xFF: 4,
}

type addrmode struct {
	human string // human readable name
	f     func(g *Generator, details uint8)
}

var addrModes = map[string]addrmode{
	"imp": {f: nil, human: `implied addressing.`},
	"acc": {f: nil, human: `adressing accumulator.`},
	"rel": {f: rel, human: `relative addressing.`},
	"abs": {f: abs, human: `absolute addressing.`},
	"abx": {f: abx, human: `absolute indexed X.`},
	"aby": {f: aby, human: `absolute indexed Y.`},
	"imm": {f: imm, human: `immediate addressing.`},
	"ind": {f: ind, human: `indirect addressing.`},
	"izx": {f: izx, human: `indexed addressing (abs, X).`},
	"izy": {f: izy, human: `indexed addressing (abs),Y.`},
	"zpg": {f: zpg, human: `zero page addressing.`},
	"zpx": {f: zpx, human: `indexed addressing: zeropage,X.`},
	"zpy": {f: zpy, human: `indexed addressing: zeropage,Y.`},
}

// helpers

func push8(g *Generator, val string) {
	g.printf(`{`)
	g.printf(`top := uint16(cpu.SP) + 0x0100`)
	g.printf(`cpu.Write8(top, (%s))`, val)
	g.printf(`cpu.SP -= 1`)
	g.printf(`}`)
}

func push16(g *Generator, val string) {
	push8(g, fmt.Sprintf(`uint8((%s)>>8)`, val))
	push8(g, fmt.Sprintf(`uint8((%s)&0xFF)`, val))
}

func pull8(g *Generator, ret string) {
	g.printf(`{`)
	g.printf(`cpu.SP += 1`)
	g.printf(`top := uint16(cpu.SP) + 0x0100`)
	g.printf(`%s = cpu.Read8(top)`, ret)
	g.printf(`}`)
}

func pull16(g *Generator, ret string) {
	g.printf(` var lo, hi uint8`)
	pull8(g, `lo`)
	pull8(g, `hi`)
	g.printf(`%s = uint16(hi)<<8 | uint16(lo)`, ret)
}

// read 16 bytes from the zero page, handling page wrap.
func r16zpwrap(g *Generator) {
	g.printf(`// read 16 bytes from the zero page, handling page wrap`)
	g.printf(`lo := cpu.Read8(oper)`)
	g.printf(`hi := cpu.Read8(uint16(uint8(oper) + 1))`)
	g.printf(`oper = uint16(hi)<<8 | uint16(lo)`)
}

func clearFlag(ibit uint) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.P.clearBit(%d)`, ibit)
		g.printf(`cpu.tick()`)
	}
}

func setFlag(ibit uint) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.P.setBit(%d)`, ibit)
		g.printf(`cpu.tick()`)
	}
}

func branch(ibit int, val bool) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`if cpu.P.bit(%d) == %t {`, ibit, val)
		g.printf(`// branching`)
		tickIfPageCrossed(g, "cpu.PC+1", "oper")
		g.printf(`	cpu.tick()`)
		g.printf(`	cpu.PC = oper`)
		g.printf(`	return`)
		g.printf(`}`)
		g.printf(`cpu.PC++`)
	}
}

func tickIfPageCrossed(g *Generator, a, b string) {
	g.printf(`if 0xFF00&(%s) != 0xFF00&(%s) {`, a, b)
	g.printf(`	cpu.tick()`)
	g.printf(`}`)
}

func copybits(dst, src, mask string) string {
	return fmt.Sprintf(`((%s) & (^%s)) | ((%s) & (%s))`, dst, mask, src, mask)
}

func ind(g *Generator, _ uint8) {
	g.printf(`oper := cpu.Read16(cpu.PC)`)
	g.printf(`lo := cpu.Read8(oper)`)
	g.printf(`// 2 bytes address wrap around`)
	g.printf(`hi := cpu.Read8((0xff00 & oper) | (0x00ff & (oper + 1)))`)
	g.printf(`oper = uint16(hi)<<8 | uint16(lo)`)
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
		tickIfPageCrossed(g, "oper", "addr")
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
		tickIfPageCrossed(g, "oper", "addr")
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
		tickIfPageCrossed(g, "oper", "oper+uint16(cpu.Y)")
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

func BRK(g *Generator, _ opdef) {
	g.printf(`cpu.tick()`)
	push16(g, `cpu.PC+1`)
	g.printf(`p := cpu.P`)
	g.printf(`p.setBit(pbitB)`)
	push8(g, `uint8(p)`)
	g.printf(`cpu.P.setBit(pbitI)`)
	g.printf(`cpu.PC = cpu.Read16(IRQvector)`)
}

func PHP(g *Generator, _ opdef) {
	g.printf(`cpu.tick()`)
	g.printf(`p := cpu.P`)
	g.printf(`p |= (1 << pbitB) | (1 << pbitU)`)
	push8(g, `uint8(p)`)
}

func RTI(g *Generator, _ opdef) {
	g.printf(`cpu.tick()`)
	g.printf(`cpu.tick()`)
	g.printf(`var p uint8`)
	pull8(g, `p`)
	g.printf(`const mask uint8 = 0b11001111 // ignore B and U bits`)
	g.printf(`cpu.P = P(%s)`, copybits(`uint8(cpu.P)`, `p`, `mask`))
	pull16(g, `cpu.PC`)
}

func RTS(g *Generator, _ opdef) {
	g.printf(`cpu.tick()`)
	g.printf(`cpu.tick()`)
	pull16(g, `cpu.PC`)
	g.printf(`cpu.PC++`)
	g.printf(`cpu.tick()`)
}

func PHA(g *Generator, _ opdef) {
	g.printf(`cpu.tick()`)
	push8(g, `cpu.A`)
}

func PLA(g *Generator, _ opdef) {
	g.printf(`cpu.tick()`)
	g.printf(`cpu.tick()`)
	pull8(g, `cpu.A`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func PLP(g *Generator, _ opdef) {
	g.printf(`cpu.tick()`)
	g.printf(`cpu.tick()`)
	g.printf(`var p uint8`)
	pull8(g, `p`)
	g.printf(`const mask uint8 = 0b11001111 // ignore B and U bits`)
	g.printf(`cpu.P = P(%s)`, copybits(`uint8(cpu.P)`, `p`, `mask`))
}

func JSR(g *Generator, _ opdef) {
	g.printf(`oper := cpu.Read16(cpu.PC)`)
	g.printf(`cpu.tick()`)
	push16(g, `cpu.PC+1`)
	g.printf(`cpu.PC = oper`)
}

func ORA(g *Generator, _ opdef) {
	g.printf(`cpu.A |= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func SLO(g *Generator, def opdef) {
	ASL(g, def)
	g.printf(`cpu.A |= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func ASL(g *Generator, _ opdef) {
	g.printf(`carry := val & 0x80`)
	g.printf(`val = (val << 1) & 0xfe`)
	g.printf(`cpu.tick()`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.writeBit(pbitC, carry != 0)`)
}

func ANC(g *Generator, def opdef) {
	AND(g, def)
	g.printf(`cpu.P.writeBit(pbitC, cpu.P.N())`)
}

func ROL(g *Generator, _ opdef) {
	g.printf(`carry := val & 0x80`)
	g.printf(`val <<= 1`)
	g.printf(`if cpu.P.C() {`)
	g.printf(`	val |= 1 << 0`)
	g.printf(`}`)
	g.printf(`cpu.tick()`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.writeBit(pbitC, carry != 0)`)
}

func BIT(g *Generator, _ opdef) {
	g.printf(`cpu.P &= 0b00111111`)
	g.printf(`cpu.P |= P(val & 0b11000000)`)
	g.printf(`cpu.P.checkZ(cpu.A & val)`)
}

func RLA(g *Generator, def opdef) {
	ROL(g, def)
	AND(g, def)
}

func AND(g *Generator, _ opdef) {
	g.printf(`cpu.A &= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func load(reg ...string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		for _, r := range reg {
			g.printf(`cpu.%s = val`, r)
		}
		g.printf(`cpu.P.checkNZ(val)`)
	}
}

func store(reg string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.Write8(oper, cpu.%s)`, reg)
	}
}

func compare(v string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		v = regOrMem(v)
		g.printf(`cpu.P.checkNZ(%s - val)`, v)
		g.printf(`cpu.P.writeBit(pbitC, val <= %s)`, v)
	}
}

func transfer(src, dst string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.%s = cpu.%s`, dst, src)
		if dst != "SP" {
			g.printf(`cpu.P.checkNZ(cpu.%s)`, src)
		}
		g.printf(`cpu.tick()`)
	}
}

func regOrMem(v string) string {
	switch v {
	case "A", "X", "Y", "SP":
		return `cpu.` + v
	case "mem":
		return `val`
	}
	panic("regOrMem " + v)
}

func increment(v string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.tick()`)
		v = regOrMem(v)
		g.printf(`%s++`, v)
		g.printf(`cpu.P.checkNZ(%s)`, v)
	}
}

func decrement(v string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.tick()`)
		v = regOrMem(v)
		g.printf(`%s--`, v)
		g.printf(`cpu.P.checkNZ(%s)`, v)
	}
}

func LAS(g *Generator, def opdef) {
	g.printf(`cpu.A = cpu.SP & val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
	g.printf(`cpu.X = cpu.A`)
	g.printf(`cpu.SP = cpu.A`)
}

func SAX(g *Generator, _ opdef) {
	g.printf(`cpu.Write8(oper, cpu.A&cpu.X)`)
}

func EOR(g *Generator, _ opdef) {
	g.printf(`cpu.A ^= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func RRA(g *Generator, def opdef) {
	ROR(g, def)
	ADC(g, def)
}

func DCP(g *Generator, def opdef) {
	decrement("mem")(g, def)
	compare("A")(g, def)
}

func SBX(g *Generator, def opdef) {
	g.printf(`ival := (int16(cpu.A) & int16(cpu.X)) - int16(val)`)
	g.printf(`cpu.X = uint8(ival)`)
	g.printf(`cpu.P.checkNZ(uint8(ival))`)
	g.printf(`cpu.P.writeBit(pbitC, ival >= 0)`)
}

func SBC(g *Generator, def opdef) {
	g.printf(`val ^= 0xff`)
	g.printf(`carry := cpu.P.ibit(pbitC)`)
	g.printf(`sum := uint16(cpu.A) + uint16(val) + uint16(carry)`)
	g.printf(`cpu.P.checkCV(cpu.A, val, sum)`)
	g.printf(`cpu.A = uint8(sum)`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func ISB(g *Generator, def opdef) {
	increment("mem")(g, def)
	g.printf(`final := val`)
	SBC(g, def)
	g.printf(`val = final`)
}

func ROR(g *Generator, _ opdef) {
	g.printf(`{`)
	g.printf(`carry := val & 0x01`)
	g.printf(`val >>= 1`)
	g.printf(`if cpu.P.C() {`)
	g.printf(`	val |= 1 << 7`)
	g.printf(`}`)
	g.printf(`cpu.tick()`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.writeBit(pbitC, carry != 0)`)
	g.printf(`}`)
}

func ARR(g *Generator, _ opdef) {
	g.printf(`cpu.A &= val`)
	g.printf(`cpu.A >>= 1`)
	g.printf(`cpu.P.writeBit(pbitV, (cpu.A>>6)^(cpu.A>>5)&0x01 != 0)`)
	g.printf(`if cpu.P.C() {`)
	g.printf(`	cpu.A |= 1 << 7`)
	g.printf(`}`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
	g.printf(`cpu.P.writeBit(pbitC, cpu.A&(1<<6) != 0)`)
}

func LSR(g *Generator, _ opdef) {
	g.printf(`{`)
	g.printf(`carry := val & 0x01 // carry is bit 0`)
	g.printf(`val = (val >> 1)&0x7f`)
	g.printf(`cpu.tick()`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.writeBit(pbitC, carry != 0)`)
	g.printf(`}`)
}

func ADC(g *Generator, _ opdef) {
	g.printf(`carry := cpu.P.ibit(pbitC)`)
	g.printf(`sum := uint16(cpu.A) + uint16(val) + uint16(carry)`)
	g.printf(`cpu.P.checkCV(cpu.A, val, sum)`)
	g.printf(`cpu.A = uint8(sum)`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func ALR(g *Generator, _ opdef) {
	g.printf(`// like and + lsr but saves one tick`)
	g.printf(`cpu.A &= val`)
	g.printf(`carry := cpu.A & 0x01 // carry is bit 0`)
	g.printf(`cpu.A = (cpu.A >> 1) & 0x7f`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
	g.printf(`cpu.P.writeBit(pbitC, carry != 0)`)
}

func SRE(g *Generator, def opdef) {
	LSR(g, def)
	EOR(g, def)
}

func NOP(g *Generator, _ opdef) {
	g.printf(`cpu.tick()`)
}

func JMP(g *Generator, _ opdef) {
	g.printf(`cpu.PC = oper`)
}

func JAM(g *Generator, def opdef) {
	g.unstable = append(g.unstable, def.i)
	insertPanic(g, `Halt and catch fire!\nJAM called`)
}

func unstable(g *Generator, def opdef) {
	g.unstable = append(g.unstable, def.i)
	insertPanic(g, fmt.Sprintf("unsupported unstable opcode 0x%02X (%s)", def.i, def.n))
}

func opname(code int) string {
	return defs[code].n
}

func insertPanic(g *Generator, msg string) {
	g.printf(`msg := fmt.Sprintf("%s\nPC:0x%%04X", cpu.PC)`, msg)
	g.printf(`panic(msg)`)
}

type Generator struct {
	io.Writer
	outbuf bytes.Buffer
	out    io.Writer

	unstable []uint8
}

func (g *Generator) header() {
	g.printf(`// Code generated by cpugen/gen_nes6502.go. DO NOT EDIT.`)
	g.printf(`package emu`)
	g.printf(`import (`)
	g.printf(`"fmt"`)
	g.printf(`)`)
}

func (g *Generator) opcodeHeader(code uint8) {
	mode, ok := addrModes[defs[code].m]
	if !ok {
		panic(fmt.Sprintf("unknown addressing mode (opcode: 0x%02X)", code))
	}

	g.printf(`// %s - %s`, defs[code].n, mode.human)
	g.printf(`func opcode%02X(cpu*CPU){`, code)
	if mode.f != nil {
		mode.f(g, details[code])
		g.printf(`_ = oper`)
	}

	switch {
	case strings.Contains(defs[code].rw, "r"):
		switch defs[code].m {
		case "acc":
			g.printf(`val := cpu.A`)
		default:
			g.printf(`val := cpu.Read8(oper)`)
		}
	}
}

func (g *Generator) opcodeFooter(code uint8) {
	switch {
	case strings.Contains(defs[code].rw, "w"):
		switch defs[code].m {
		case "acc":
			g.printf(`cpu.A = val`)
		default:
			g.printf(`cpu.Write8(oper, val)`)
		}
	}
	g.printf(`}`)
}

func (g *Generator) opcodes() {
	for _, def := range defs {
		g.opcodeHeader(def.i)
		def.f(g, def)
		g.opcodeFooter(def.i)
	}
}

func (g *Generator) opcodesTable() {
	bb := &strings.Builder{}
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			fmt.Fprintf(bb, "opcode%02X, ", i*16+j)
		}
		bb.WriteByte('\n')
	}
	g.printf(`// nes 6502 opcodes table`)
	g.printf(`var ops = [256]func(*CPU){`)
	g.printf(bb.String())
	g.printf(`}`)
	g.printf(``)
}

func (g *Generator) unstableOpcodes() {
	g.printf(`// list of unstable opcodes (unsupported)`)
	g.printf(`var unstableOps = [256]uint8{`)
	for _, code := range g.unstable {
		g.printf(`0x%02X: 1, // %s`, code, defs[code].n)
	}
	g.printf(`}`)
}

func (g *Generator) printf(format string, args ...any) {
	fmt.Fprintf(g, "%s\n", fmt.Sprintf(format, args...))
}

func main() {
	log.SetFlags(0)
	outf := flag.String("out", "cpu_ops.go", "output file")
	flag.Parse()

	bb := &bytes.Buffer{}
	g := &Generator{Writer: bb}

	g.header()
	g.opcodes()
	g.opcodesTable()
	g.unstableOpcodes()

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
