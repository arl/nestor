package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

const pkgname = "hw"

type opdef struct {
	i uint8       // opcode value (same as index into 'defs')
	n string      // name
	m string      // addressing mode
	f func(opdef) // if nil, opcode is manually written

	// opcode detail string
	// 	- r: declare 'val' and set it to 'oper/accumulator'
	// 	- w: write 'val' back into 'oper/accumulator'
	d string
}

var defs = [256]opdef{
	{i: 0x00, n: "BRK", d: "     ", m: "imp"},
	{i: 0x01, n: "ORA", d: "r    ", m: "izx", f: ORA},
	{i: 0x02, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0x03, n: "SLO", d: "rw   ", m: "izx", f: SLO},
	{i: 0x04, n: "NOP", d: "     ", m: "zpg", f: NOP},
	{i: 0x05, n: "ORA", d: "r    ", m: "zpg", f: ORA},
	{i: 0x06, n: "ASL", d: "rw   ", m: "zpg", f: ASL},
	{i: 0x07, n: "SLO", d: "rw   ", m: "zpg", f: SLO},
	{i: 0x08, n: "PHP", d: "     ", m: "imp", f: PHP},
	{i: 0x09, n: "ORA", d: "r    ", m: "imm", f: ORA},
	{i: 0x0A, n: "ASL", d: "rw   ", m: "acc", f: ASL},
	{i: 0x0B, n: "ANC", d: "r    ", m: "imm", f: ANC},
	{i: 0x0C, n: "NOP", d: "     ", m: "abs", f: NOP},
	{i: 0x0D, n: "ORA", d: "r    ", m: "abs", f: ORA},
	{i: 0x0E, n: "ASL", d: "rw   ", m: "abs", f: ASL},
	{i: 0x0F, n: "SLO", d: "rw   ", m: "abs", f: SLO},
	{i: 0x10, n: "BPL", d: "     ", m: "rel", f: branch(Negative, false)},
	{i: 0x11, n: "ORA", d: "r    ", m: "izy", f: ORA},
	{i: 0x12, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0x13, n: "SLO", d: "rw   ", m: "izyd", f: SLO},
	{i: 0x14, n: "NOP", d: "     ", m: "zpx", f: NOP},
	{i: 0x15, n: "ORA", d: "r    ", m: "zpx", f: ORA},
	{i: 0x16, n: "ASL", d: "rw   ", m: "zpx", f: ASL},
	{i: 0x17, n: "SLO", d: "rw   ", m: "zpx", f: SLO},
	{i: 0x18, n: "CLC", d: "     ", m: "imp", f: clear(Carry)},
	{i: 0x19, n: "ORA", d: "r    ", m: "aby", f: ORA},
	{i: 0x1A, n: "NOP", d: "     ", m: "imp", f: NOP},
	{i: 0x1B, n: "SLO", d: "rw   ", m: "abyd", f: SLO},
	{i: 0x1C, n: "NOP", d: "     ", m: "abx", f: NOP},
	{i: 0x1D, n: "ORA", d: "r    ", m: "abx", f: ORA},
	{i: 0x1E, n: "ASL", d: "rw   ", m: "abxd", f: ASL},
	{i: 0x1F, n: "SLO", d: "rw   ", m: "abxd", f: SLO},
	{i: 0x20, n: "JSR", d: "     ", m: "abs"},
	{i: 0x21, n: "AND", d: "r    ", m: "izx", f: AND},
	{i: 0x22, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0x23, n: "RLA", d: "rw   ", m: "izx", f: RLA},
	{i: 0x24, n: "BIT", d: "r    ", m: "zpg", f: BIT},
	{i: 0x25, n: "AND", d: "r    ", m: "zpg", f: AND},
	{i: 0x26, n: "ROL", d: "rw   ", m: "zpg", f: ROL},
	{i: 0x27, n: "RLA", d: "rw   ", m: "zpg", f: RLA},
	{i: 0x28, n: "PLP", d: "     ", m: "imp", f: PLP},
	{i: 0x29, n: "AND", d: "r    ", m: "imm", f: AND},
	{i: 0x2A, n: "ROL", d: "rw   ", m: "acc", f: ROL},
	{i: 0x2B, n: "ANC", d: "r    ", m: "imm", f: ANC},
	{i: 0x2C, n: "BIT", d: "r    ", m: "abs", f: BIT},
	{i: 0x2D, n: "AND", d: "r    ", m: "abs", f: AND},
	{i: 0x2E, n: "ROL", d: "rw   ", m: "abs", f: ROL},
	{i: 0x2F, n: "RLA", d: "rw   ", m: "abs", f: RLA},
	{i: 0x30, n: "BMI", d: "     ", m: "rel", f: branch(Negative, true)},
	{i: 0x31, n: "AND", d: "r    ", m: "izy", f: AND},
	{i: 0x32, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0x33, n: "RLA", d: "rw   ", m: "izyd", f: RLA},
	{i: 0x34, n: "NOP", d: "     ", m: "zpx", f: NOP},
	{i: 0x35, n: "AND", d: "r    ", m: "zpx", f: AND},
	{i: 0x36, n: "ROL", d: "rw   ", m: "zpx", f: ROL},
	{i: 0x37, n: "RLA", d: "rw   ", m: "zpx", f: RLA},
	{i: 0x38, n: "SEC", d: "     ", m: "imp", f: set(Carry)},
	{i: 0x39, n: "AND", d: "r    ", m: "aby", f: AND},
	{i: 0x3A, n: "NOP", d: "     ", m: "imp", f: NOP},
	{i: 0x3B, n: "RLA", d: "rw   ", m: "abyd", f: RLA},
	{i: 0x3C, n: "NOP", d: "     ", m: "abx", f: NOP},
	{i: 0x3D, n: "AND", d: "r    ", m: "abx", f: AND},
	{i: 0x3E, n: "ROL", d: "rw   ", m: "abxd", f: ROL},
	{i: 0x3F, n: "RLA", d: "rw   ", m: "abxd", f: RLA},
	{i: 0x40, n: "RTI", d: "     ", m: "imp", f: RTI},
	{i: 0x41, n: "EOR", d: "r    ", m: "izx", f: EOR},
	{i: 0x42, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0x43, n: "SRE", d: "rw   ", m: "izx", f: SRE},
	{i: 0x44, n: "NOP", d: "     ", m: "zpg", f: NOP},
	{i: 0x45, n: "EOR", d: "r    ", m: "zpg", f: EOR},
	{i: 0x46, n: "LSR", d: "rw   ", m: "zpg", f: LSR},
	{i: 0x47, n: "SRE", d: "rw   ", m: "zpg", f: SRE},
	{i: 0x48, n: "PHA", d: "     ", m: "imp", f: PHA},
	{i: 0x49, n: "EOR", d: "r    ", m: "imm", f: EOR},
	{i: 0x4A, n: "LSR", d: "rw   ", m: "acc", f: LSR},
	{i: 0x4B, n: "ALR", d: "r    ", m: "imm", f: ALR},
	{i: 0x4C, n: "JMP", d: "     ", m: "abs", f: JMP},
	{i: 0x4D, n: "EOR", d: "r    ", m: "abs", f: EOR},
	{i: 0x4E, n: "LSR", d: "rw   ", m: "abs", f: LSR},
	{i: 0x4F, n: "SRE", d: "rw   ", m: "abs", f: SRE},
	{i: 0x50, n: "BVC", d: "     ", m: "rel", f: branch(Overflow, false)},
	{i: 0x51, n: "EOR", d: "r    ", m: "izy", f: EOR},
	{i: 0x52, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0x53, n: "SRE", d: "rw   ", m: "izyd", f: SRE},
	{i: 0x54, n: "NOP", d: "     ", m: "zpx", f: NOP},
	{i: 0x55, n: "EOR", d: "r    ", m: "zpx", f: EOR},
	{i: 0x56, n: "LSR", d: "rw   ", m: "zpx", f: LSR},
	{i: 0x57, n: "SRE", d: "rw   ", m: "zpx", f: SRE},
	{i: 0x58, n: "CLI", d: "     ", m: "imp", f: clear(Interrupt)},
	{i: 0x59, n: "EOR", d: "r    ", m: "aby", f: EOR},
	{i: 0x5A, n: "NOP", d: "     ", m: "imp", f: NOP},
	{i: 0x5B, n: "SRE", d: "rw   ", m: "abyd", f: SRE},
	{i: 0x5C, n: "NOP", d: "     ", m: "abx", f: NOP},
	{i: 0x5D, n: "EOR", d: "r    ", m: "abx", f: EOR},
	{i: 0x5E, n: "LSR", d: "rw   ", m: "abxd", f: LSR},
	{i: 0x5F, n: "SRE", d: "rw   ", m: "abxd", f: SRE},
	{i: 0x60, n: "RTS", d: "     ", m: "imp", f: RTS},
	{i: 0x61, n: "ADC", d: "r    ", m: "izx", f: ADC},
	{i: 0x62, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0x63, n: "RRA", d: "rw   ", m: "izx", f: RRA},
	{i: 0x64, n: "NOP", d: "     ", m: "zpg", f: NOP},
	{i: 0x65, n: "ADC", d: "r    ", m: "zpg", f: ADC},
	{i: 0x66, n: "ROR", d: "rw   ", m: "zpg", f: ROR},
	{i: 0x67, n: "RRA", d: "rw   ", m: "zpg", f: RRA},
	{i: 0x68, n: "PLA", d: "     ", m: "imp", f: PLA},
	{i: 0x69, n: "ADC", d: "r    ", m: "imm", f: ADC},
	{i: 0x6A, n: "ROR", d: "rw   ", m: "acc", f: ROR},
	{i: 0x6B, n: "ARR", d: "r    ", m: "imm", f: ARR},
	{i: 0x6C, n: "JMP", d: "     ", m: "ind", f: JMP},
	{i: 0x6D, n: "ADC", d: "r    ", m: "abs", f: ADC},
	{i: 0x6E, n: "ROR", d: "rw   ", m: "abs", f: ROR},
	{i: 0x6F, n: "RRA", d: "rw   ", m: "abs", f: RRA},
	{i: 0x70, n: "BVS", d: "     ", m: "rel", f: branch(Overflow, true)},
	{i: 0x71, n: "ADC", d: "r    ", m: "izy", f: ADC},
	{i: 0x72, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0x73, n: "RRA", d: "rw   ", m: "izyd", f: RRA},
	{i: 0x74, n: "NOP", d: "     ", m: "zpx", f: NOP},
	{i: 0x75, n: "ADC", d: "r    ", m: "zpx", f: ADC},
	{i: 0x76, n: "ROR", d: "rw   ", m: "zpx", f: ROR},
	{i: 0x77, n: "RRA", d: "rw   ", m: "zpx", f: RRA},
	{i: 0x78, n: "SEI", d: "     ", m: "imp", f: set(Interrupt)},
	{i: 0x79, n: "ADC", d: "r    ", m: "aby", f: ADC},
	{i: 0x7A, n: "NOP", d: "     ", m: "imp", f: NOP},
	{i: 0x7B, n: "RRA", d: "rw   ", m: "abyd", f: RRA},
	{i: 0x7C, n: "NOP", d: "     ", m: "abx", f: NOP},
	{i: 0x7D, n: "ADC", d: "r    ", m: "abx", f: ADC},
	{i: 0x7E, n: "ROR", d: "rw   ", m: "abxd", f: ROR},
	{i: 0x7F, n: "RRA", d: "rw   ", m: "abxd", f: RRA},
	{i: 0x80, n: "NOP", d: "r    ", m: "imm", f: NOP},
	{i: 0x81, n: "STA", d: "     ", m: "izx", f: ST("A")},
	{i: 0x82, n: "NOP", d: "r    ", m: "imm", f: NOP},
	{i: 0x83, n: "SAX", d: "     ", m: "izx", f: SAX},
	{i: 0x84, n: "STY", d: "     ", m: "zpg", f: ST("Y")},
	{i: 0x85, n: "STA", d: "     ", m: "zpg", f: ST("A")},
	{i: 0x86, n: "STX", d: "     ", m: "zpg", f: ST("X")},
	{i: 0x87, n: "SAX", d: "     ", m: "zpg", f: SAX},
	{i: 0x88, n: "DEY", d: "     ", m: "imp", f: dey},
	{i: 0x89, n: "NOP", d: "r    ", m: "imm", f: NOP},
	{i: 0x8A, n: "TXA", d: "     ", m: "imp", f: T("X", "A")},
	{i: 0x8B, n: "ANE", d: "     ", m: "imm", f: unstable},
	{i: 0x8C, n: "STY", d: "     ", m: "abs", f: ST("Y")},
	{i: 0x8D, n: "STA", d: "     ", m: "abs", f: ST("A")},
	{i: 0x8E, n: "STX", d: "     ", m: "abs", f: ST("X")},
	{i: 0x8F, n: "SAX", d: "     ", m: "abs", f: SAX},
	{i: 0x90, n: "BCC", d: "     ", m: "rel", f: branch(Carry, false)},
	{i: 0x91, n: "STA", d: "     ", m: "izyd", f: ST("A")},
	{i: 0x92, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0x93, n: "SHA", d: "     ", m: "izy", f: unstable},
	{i: 0x94, n: "STY", d: "     ", m: "zpx", f: ST("Y")},
	{i: 0x95, n: "STA", d: "     ", m: "zpx", f: ST("A")},
	{i: 0x96, n: "STX", d: "     ", m: "zpy", f: ST("X")},
	{i: 0x97, n: "SAX", d: "     ", m: "zpy", f: SAX},
	{i: 0x98, n: "TYA", d: "     ", m: "imp", f: T("Y", "A")},
	{i: 0x99, n: "STA", d: "     ", m: "abyd", f: ST("A")},
	{i: 0x9A, n: "TXS", d: "     ", m: "imp", f: T("X", "SP")},
	{i: 0x9B, n: "TAS", d: "     ", m: "abx", f: unstable},
	{i: 0x9C, n: "SHY", d: "     ", m: "aby", f: unstable},
	{i: 0x9D, n: "STA", d: "     ", m: "abxd", f: ST("A")},
	{i: 0x9E, n: "SHX", d: "     ", m: "abx", f: unstable},
	{i: 0x9F, n: "SHA", d: "     ", m: "aby", f: unstable},
	{i: 0xA0, n: "LDY", d: "r    ", m: "imm", f: LD("Y")},
	{i: 0xA1, n: "LDA", d: "r    ", m: "izx", f: LD("A")},
	{i: 0xA2, n: "LDX", d: "r    ", m: "imm", f: LD("X")},
	{i: 0xA3, n: "LAX", d: "r    ", m: "izx", f: LD("A", "X")},
	{i: 0xA4, n: "LDY", d: "r    ", m: "zpg", f: LD("Y")},
	{i: 0xA5, n: "LDA", d: "r    ", m: "zpg", f: LD("A")},
	{i: 0xA6, n: "LDX", d: "r    ", m: "zpg", f: LD("X")},
	{i: 0xA7, n: "LAX", d: "r    ", m: "zpg", f: LD("A", "X")},
	{i: 0xA8, n: "TAY", d: "     ", m: "imp", f: T("A", "Y")},
	{i: 0xA9, n: "LDA", d: "r    ", m: "imm", f: LD("A")},
	{i: 0xAA, n: "TAX", d: "     ", m: "imp", f: T("A", "X")},
	{i: 0xAB, n: "LXA", d: "r    ", m: "imm", f: LXA},
	{i: 0xAC, n: "LDY", d: "r    ", m: "abs", f: LD("Y")},
	{i: 0xAD, n: "LDA", d: "r    ", m: "abs", f: LD("A")},
	{i: 0xAE, n: "LDX", d: "r    ", m: "abs", f: LD("X")},
	{i: 0xAF, n: "LAX", d: "r    ", m: "abs", f: LD("A", "X")},
	{i: 0xB0, n: "BCS", d: "     ", m: "rel", f: branch(Carry, true)},
	{i: 0xB1, n: "LDA", d: "r    ", m: "izy", f: LD("A")},
	{i: 0xB2, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0xB3, n: "LAX", d: "r    ", m: "izy", f: LD("A", "X")},
	{i: 0xB4, n: "LDY", d: "r    ", m: "zpx", f: LD("Y")},
	{i: 0xB5, n: "LDA", d: "r    ", m: "zpx", f: LD("A")},
	{i: 0xB6, n: "LDX", d: "r    ", m: "zpy", f: LD("X")},
	{i: 0xB7, n: "LAX", d: "r    ", m: "zpy", f: LD("A", "X")},
	{i: 0xB8, n: "CLV", d: "     ", m: "imp", f: clear(Overflow)},
	{i: 0xB9, n: "LDA", d: "r    ", m: "aby", f: LD("A")},
	{i: 0xBA, n: "TSX", d: "     ", m: "imp", f: T("SP", "X")},
	{i: 0xBB, n: "LAS", d: "r    ", m: "aby", f: LAS},
	{i: 0xBC, n: "LDY", d: "r    ", m: "abx", f: LD("Y")},
	{i: 0xBD, n: "LDA", d: "r    ", m: "abx", f: LD("A")},
	{i: 0xBE, n: "LDX", d: "r    ", m: "aby", f: LD("X")},
	{i: 0xBF, n: "LAX", d: "r    ", m: "aby", f: LD("A", "X")},
	{i: 0xC0, n: "CPY", d: "r    ", m: "imm", f: cmp("Y")},
	{i: 0xC1, n: "CMP", d: "r    ", m: "izx", f: cmp("A")},
	{i: 0xC2, n: "NOP", d: "r    ", m: "imm", f: NOP},
	{i: 0xC3, n: "DCP", d: "rw   ", m: "izx", f: DCP},
	{i: 0xC4, n: "CPY", d: "r    ", m: "zpg", f: cmp("Y")},
	{i: 0xC5, n: "CMP", d: "r    ", m: "zpg", f: cmp("A")},
	{i: 0xC6, n: "DEC", d: "rw   ", m: "zpg", f: dec("")},
	{i: 0xC7, n: "DCP", d: "rw   ", m: "zpg", f: DCP},
	{i: 0xC8, n: "INY", d: "     ", m: "imp", f: iny},
	{i: 0xC9, n: "CMP", d: "r    ", m: "imm", f: cmp("A")},
	{i: 0xCA, n: "DEX", d: "     ", m: "imp", f: dex},
	{i: 0xCB, n: "SBX", d: "r    ", m: "imm", f: SBX},
	{i: 0xCC, n: "CPY", d: "r    ", m: "abs", f: cmp("Y")},
	{i: 0xCD, n: "CMP", d: "r    ", m: "abs", f: cmp("A")},
	{i: 0xCE, n: "DEC", d: "rw   ", m: "abs", f: dec("")},
	{i: 0xCF, n: "DCP", d: "rw   ", m: "abs", f: DCP},
	{i: 0xD0, n: "BNE", d: "     ", m: "rel", f: branch(Zero, false)},
	{i: 0xD1, n: "CMP", d: "r    ", m: "izy", f: cmp("A")},
	{i: 0xD2, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0xD3, n: "DCP", d: "rw   ", m: "izyd", f: DCP},
	{i: 0xD4, n: "NOP", d: "     ", m: "zpx", f: NOP},
	{i: 0xD5, n: "CMP", d: "r    ", m: "zpx", f: cmp("A")},
	{i: 0xD6, n: "DEC", d: "rw   ", m: "zpx", f: dec("")},
	{i: 0xD7, n: "DCP", d: "rw   ", m: "zpx", f: DCP},
	{i: 0xD8, n: "CLD", d: "     ", m: "imp", f: clear(Decimal)},
	{i: 0xD9, n: "CMP", d: "r    ", m: "aby", f: cmp("A")},
	{i: 0xDA, n: "NOP", d: "     ", m: "imp", f: NOP},
	{i: 0xDB, n: "DCP", d: "rw   ", m: "abyd", f: DCP},
	{i: 0xDC, n: "NOP", d: "     ", m: "abx", f: NOP},
	{i: 0xDD, n: "CMP", d: "r    ", m: "abx", f: cmp("A")},
	{i: 0xDE, n: "DEC", d: "rw   ", m: "abxd", f: dec("")},
	{i: 0xDF, n: "DCP", d: "rw   ", m: "abxd", f: DCP},
	{i: 0xE0, n: "CPX", d: "r    ", m: "imm", f: cmp("X")},
	{i: 0xE1, n: "SBC", d: "r    ", m: "izx", f: SBC},
	{i: 0xE2, n: "NOP", d: "r    ", m: "imm", f: NOP},
	{i: 0xE3, n: "ISC", d: "r    ", m: "izx", f: ISC},
	{i: 0xE4, n: "CPX", d: "r    ", m: "zpg", f: cmp("X")},
	{i: 0xE5, n: "SBC", d: "r    ", m: "zpg", f: SBC},
	{i: 0xE6, n: "INC", d: "r    ", m: "zpg", f: INC},
	{i: 0xE7, n: "ISC", d: "r    ", m: "zpg", f: ISC},
	{i: 0xE8, n: "INX", d: "     ", m: "imp", f: inx},
	{i: 0xE9, n: "SBC", d: "r    ", m: "imm", f: SBC},
	{i: 0xEA, n: "NOP", d: "     ", m: "imp", f: NOP},
	{i: 0xEB, n: "SBC", d: "r    ", m: "imm", f: SBC},
	{i: 0xEC, n: "CPX", d: "r    ", m: "abs", f: cmp("X")},
	{i: 0xED, n: "SBC", d: "r    ", m: "abs", f: SBC},
	{i: 0xEE, n: "INC", d: "r    ", m: "abs", f: INC},
	{i: 0xEF, n: "ISC", d: "r    ", m: "abs", f: ISC},
	{i: 0xF0, n: "BEQ", d: "     ", m: "rel", f: branch(Zero, true)},
	{i: 0xF1, n: "SBC", d: "r    ", m: "izy", f: SBC},
	{i: 0xF2, n: "STP", d: "     ", m: "imp", f: STP},
	{i: 0xF3, n: "ISC", d: "r    ", m: "izyd", f: ISC},
	{i: 0xF4, n: "NOP", d: "     ", m: "zpx", f: NOP},
	{i: 0xF5, n: "SBC", d: "r    ", m: "zpx", f: SBC},
	{i: 0xF6, n: "INC", d: "r    ", m: "zpx", f: INC},
	{i: 0xF7, n: "ISC", d: "r    ", m: "zpx", f: ISC},
	{i: 0xF8, n: "SED", d: "     ", m: "imp", f: set(Decimal)},
	{i: 0xF9, n: "SBC", d: "r    ", m: "aby", f: SBC},
	{i: 0xFA, n: "NOP", d: "     ", m: "imp", f: NOP},
	{i: 0xFB, n: "ISC", d: "r    ", m: "abyd", f: ISC},
	{i: 0xFC, n: "NOP", d: "     ", m: "abx", f: NOP},
	{i: 0xFD, n: "SBC", d: "r    ", m: "abx", f: SBC},
	{i: 0xFE, n: "INC", d: "r    ", m: "abxd", f: INC},
	{i: 0xFF, n: "ISC", d: "r    ", m: "abxd", f: ISC},
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
	printf(`cpu.P |= %s`, strings.Join(flagstr, "|"))
}

func clearFlags(f cpuFlag) {
	val := uint8(f)
	val = ^val
	printf(`cpu.P &= 0x%02x`, val)
}

func checkFlags(f cpuFlag) string {
	return fmt.Sprintf(`(cpu.P&0x%02x == 0x%02x)`, int(f), int(f))
}

func If(format string, args ...any) block {
	printf(`if %s {`, fmt.Sprintf(format, args...))
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
func ind()                      { printf("cpu.ind()") }
func rel()                      { printf("cpu.rel()") }
func abs()                      { printf("cpu.abs()") }
func abx(dummyread bool) func() { return func() { printf("cpu.abx(%t)", dummyread) } }
func aby(dummyread bool) func() { return func() { printf("cpu.aby(%t)", dummyread) } }
func zpg()                      { printf(`cpu.zpg()`) }
func zpx()                      { printf(`cpu.zpx()`) }
func zpy()                      { printf(`cpu.zpy()`) }
func izx()                      { printf(`cpu.izx()`) }
func izy(dummyread bool) func() { return func() { printf("cpu.izy(%t)", dummyread) } }

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

func branch(f cpuFlag, val bool) func(_ opdef) {
	return func(_ opdef) {
		if val {
			printf(`cpu.branch(%s, 0)`, f)
		} else {
			printf(`cpu.branch(%s, %s)`, f, f)
		}
	}
}

func copybits(dst, src, mask string) string {
	return fmt.Sprintf(`((%s) & (^%s)) | ((%s) & (%s))`, dst, mask, src, mask)
}

func checkNZ(val string) {
	printf(`cpu.P.clearFlags(Zero | Negative)`)
	printf(`cpu.P.setNZ(%s)`, val)
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

func STP(def opdef) {
	g.unstable = append(g.unstable, def.i)
	printf(`cpu.halt()`)
}

func ADC(_ opdef) {
	printf(`cpu.add(val)`)
}

func ALR(_ opdef) {
	printf(`// like and + lsr but saves one tick`)
	printf(`cpu.A &= val`)
	printf(`carry := cpu.A & 0x01 // carry is bit 0`)
	printf(`cpu.A = (cpu.A >> 1) & 0x7f`)
	checkNZ(`cpu.A`)
	clearFlags(Carry)
	If(`carry != 0`).setFlags(Carry).End()
}

func ANC(def opdef) {
	AND(def)
	clearFlags(Carry)
	If(checkFlags(Negative)).setFlags(Carry).End()
}

func AND(_ opdef) {
	printf(`cpu.A &= val`)
	checkNZ(`cpu.A`)
}

func ARR(_ opdef) {
	printf(`cpu.A &= val`)
	printf(`cpu.A >>= 1`)
	clearFlags(Overflow)

	If(`(cpu.A>>6)^(cpu.A>>5)&0x01 != 0`).
		setFlags(Overflow).
		End()
	If(checkFlags(Carry)).
		printf(`cpu.A |= 1 << 7`).
		End()

	checkNZ(`cpu.A`)
	clearFlags(Carry)

	If(`(cpu.A&(1<<6) != 0)`).
		setFlags(Carry).
		End()
}

func ASL(def opdef) {
	if def.m != "acc" {
		dummywrite("cpu.operand", "val")
	}
	printf(`carry := val & 0x80`)
	printf(`val = (val << 1) & 0xfe`)
	checkNZ(`val`)
	clearFlags(Carry)

	If(`carry != 0`).
		setFlags(Carry).
		End()
}

func BIT(_ opdef) {
	clearFlags(Zero | Overflow | Negative)
	printf(`cpu.P |= P(val & 0b11000000)`)
	If(`cpu.A&val == 0`).
		setFlags(Zero).
		End()
}

func DCP(def opdef) {
	dec("")(def)
	cmp("A")(def)
}

func EOR(_ opdef) {
	printf(`cpu.A ^= val`)
	checkNZ(`cpu.A`)
}

func INC(_ opdef) {
	dummywrite("cpu.operand", "val")
	printf(`val++`)
	clearFlags(Zero | Negative)
	checkNZ(`val`)
	printf(`cpu.Write8(cpu.operand, val)`)
}

func ISC(def opdef) {
	INC(def)
	printf(`final := val`)
	SBC(def)
	printf(`val = final`)
}

func JMP(_ opdef) {
	printf(`cpu.PC = cpu.operand`)
}

func LAS(def opdef) {
	printf(`cpu.A = cpu.SP & val`)
	checkNZ(`cpu.A`)
	printf(`cpu.X = cpu.A`)
	printf(`cpu.SP = cpu.A`)
}

func LSR(def opdef) {
	if def.m != "acc" {
		dummywrite("cpu.operand", "val")
	}

	printf(`carry := val & 0x01 // carry is bit 0`)
	printf(`val = (val >> 1)&0x7f`)
	checkNZ(`val`)
	clearFlags(Carry)
	If(`carry != 0`).
		setFlags(Carry).
		End()
}

func LXA(def opdef) {
	g.unstable = append(g.unstable, def.i)

	const mask = 0xff
	printf(`val = (cpu.A | 0x%02x) & val`, mask)
	printf(`cpu.A = val`)
	printf(`cpu.X = val`)
	checkNZ(`cpu.A`)
}

func NOP(def opdef) {
	if !slices.Contains([]string{"acc", "imp", "rel", "imm"}, def.m) {
		dummyread("cpu.operand")
	}
	if def.m == "imm" {
		printf(`_ = val`)
	}
}

func ORA(_ opdef) {
	printf(`cpu.setreg(&cpu.A, cpu.A|val)`)
}

func PHA(_ opdef) {
	push8(`cpu.A`)
}

func PHP(_ opdef) {
	printf(`p := cpu.P | 0x%02x`, int(Break|Reserved))
	push8(`uint8(p)`)
}

func PLA(_ opdef) {
	dummyread("uint16(cpu.SP) + 0x0100")
	pull8(`cpu.A`)
	checkNZ(`cpu.A`)
}

func PLP(_ opdef) {
	printf(`var p uint8`)
	dummyread("uint16(cpu.SP) + 0x0100")
	pull8(`p`)
	printf(`const mask uint8 = 0b11001111 // ignore B and U bits`)
	printf(`cpu.P = P(%s)`, copybits(`uint8(cpu.P)`, `p`, `mask`))
}

func RLA(def opdef) {
	ROL(def)
	AND(def)
}

func ROL(def opdef) {
	if def.m != "acc" {
		dummywrite("cpu.operand", "val")
	}
	printf(`carry := val & 0x80`)
	printf(`val <<= 1`)

	If(checkFlags(Carry)).
		printf(`val |= 1 << 0`).
		End()

	checkNZ(`val`)
	clearFlags(Carry)

	If(`carry != 0`).
		setFlags(Carry).
		End()
}

func ROR(def opdef) {
	if def.m != "acc" {
		dummywrite("cpu.operand", "val")
	}
	printf(`carry := val & 0x01`)
	printf(`val >>= 1`)

	If(checkFlags(Carry)).
		printf(`val |= 1 << 7`).
		End()

	checkNZ(`val`)
	clearFlags(Carry)
	If(`carry != 0`).
		setFlags(Carry).
		End()
}

func RRA(def opdef) {
	ROR(def)
	ADC(def)
}

func RTI(_ opdef) {
	printf(`var p uint8`)
	dummyread("uint16(cpu.SP) + 0x0100")
	pull8(`p`)
	printf(`const mask uint8 = 0b11001111 // ignore B and U bits`)
	printf(`cpu.P = P(%s)`, copybits(`uint8(cpu.P)`, `p`, `mask`))
	pull16(`cpu.PC`)
}

func RTS(_ opdef) {
	dummyread("uint16(cpu.SP) + 0x0100")
	pull16(`cpu.PC`)
	printf(`cpu.fetch()`)
}

func SAX(_ opdef) {
	printf(`cpu.Write8(cpu.operand, cpu.A&cpu.X)`)
}

func SBC(def opdef) {
	printf(`val ^= 0xff`)
	printf(`cpu.add(val)`)
}

func SBX(def opdef) {
	printf(`ival := (int16(cpu.A) & int16(cpu.X)) - int16(val)`)
	printf(`cpu.X = uint8(ival)`)
	checkNZ(`cpu.X`)
	clearFlags(Carry)

	If(`ival >= 0`).
		setFlags(Carry).
		End()
}

func SLO(def opdef) {
	ASL(def)
	printf(`cpu.setreg(&cpu.A, cpu.A|val)`)
}

func SRE(def opdef) {
	LSR(def)
	EOR(def)
}

//
// opcode helpers
//

func LD(reg ...string) func(_ opdef) {
	return func(_ opdef) {
		for _, r := range reg {
			printf(`cpu.setreg(&cpu.%s, val)`, r)
		}
	}
}

func ST(reg string) func(_ opdef) {
	return func(_ opdef) {
		printf(`cpu.Write8(cpu.operand, cpu.%s)`, reg)
	}
}

func cmp(v string) func(_ opdef) {
	return func(_ opdef) {
		v = regOrMem(v)
		checkNZ(fmt.Sprintf("%s - val", v))
		clearFlags(Carry)

		If(`val <= %s`, v).
			setFlags(Carry).
			End()
	}
}

func T(src, dst string) func(_ opdef) {
	return func(_ opdef) {
		printf(`cpu.%s = cpu.%s`, dst, src)
		if dst != "SP" {
			checkNZ(fmt.Sprintf(`cpu.%s`, src))
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

func inx(_ opdef) { printf(`cpu.setreg(&cpu.X, cpu.X+1)`) }
func iny(_ opdef) { printf(`cpu.setreg(&cpu.Y, cpu.Y+1)`) }
func dex(_ opdef) { printf(`cpu.setreg(&cpu.X, cpu.X-1)`) }
func dey(_ opdef) { printf(`cpu.setreg(&cpu.Y, cpu.Y-1)`) }

func dec(v string) func(_ opdef) {
	return func(_ opdef) {
		v = regOrMem(v)
		if v == "val" {
			// TODO: works but ugly
			dummywrite("cpu.operand", "val")
		}
		printf(`%s--`, v)
		checkNZ(v)
	}
}

func clear(f cpuFlag) func(_ opdef) {
	return func(_ opdef) {
		clearFlags(f)
	}
}

func set(f cpuFlag) func(_ opdef) {
	return func(_ opdef) {
		setFlags(f)
	}
}

func unstable(def opdef) {
	g.unstable = append(g.unstable, def.i)
	printf(`msg := fmt.Sprintf("unsupported unstable opcode 0x%02X (%s)\nPC:0x%%04X", cpu.PC)`, def.i, def.n)
	printf(`panic(msg)`)
}

func header() {
	printf(`// Code generated by cpugen/gen_nes6502.go. DO NOT EDIT.`)
	printf(`package %s`, pkgname)
	printf(`import (`)
	printf(`"fmt"`)
	printf(`)`)
}

func opcodeHeader(code uint8) {
	mode, ok := addrModes[defs[code].m]
	if !ok {
		panic(fmt.Sprintf("unknown addressing mode (opcode: 0x%02X)", code))
	}

	printf(`// %s - %s`, defs[code].n, mode.human)
	printf(`func opcode%02X(cpu*CPU) {`, code)
	if mode.f != nil {
		mode.f()
	}

	switch {
	case strings.Contains(defs[code].d, "r"):
		switch defs[code].m {
		case "acc":
			printf(`val := cpu.A`)
		case "imm":
			printf(`val := cpu.fetch()`)
		default:
			printf(`val := cpu.Read8(cpu.operand)`)
		}
	}
}

func opcodeFooter(code uint8) {
	switch {
	case strings.Contains(defs[code].d, "w"):
		switch defs[code].m {
		case "acc":
			printf(`cpu.A = val`)
		default:
			printf(`cpu.Write8(cpu.operand, val)`)
		}
	}
	printf(`}`)
}

func opcodes() {
	for _, def := range defs {
		if def.f == nil {
			continue
		}
		opcodeHeader(def.i)
		if def.f != nil {
			def.f(def)
		} else {
			f := reflect.ValueOf(g).MethodByName(def.n)
			f.Call([]reflect.Value{reflect.ValueOf(def)})
		}

		opcodeFooter(def.i)
		printf("\n")
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
	printf(bb.String())
	printf(`}`)
	printf(``)
}

func disasmTable() {
	bb := &strings.Builder{}
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			name := defs[i*16+j].m
			name = strings.ToUpper(name[:1]) + name[1:]
			name = name[:3]
			fmt.Fprintf(bb, "disasm%s, ", name)
		}
		bb.WriteByte('\n')
	}
	printf(`// nes 6502 opcodes disassembly table`)
	printf(`var disasmOps = [256]func(*CPU, uint16) DisasmOp {`)
	printf(bb.String())
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

func unstableOpcodes() {
	printf(`// list of unstable opcodes (unsupported)`)
	printf(`var unstableOps = [256]uint8{`)
	for _, code := range g.unstable {
		printf(`0x%02X: 1, // %s`, code, defs[code].n)
	}
	printf(`}`)
}

func printf(format string, args ...any) {
	fmt.Fprintf(g, "%s\n", fmt.Sprintf(format, args...))
}

type Generator struct {
	io.Writer
	unstable []uint8
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
	unstableOpcodes()
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
