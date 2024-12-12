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
	"sort"
	"strconv"
	"strings"
)

const pkgname = "hw"

type opdef struct {
	i uint8  // opcode value (same as index into 'defs')
	n string // name
	m string // addressing mode
	f func(g *Generator, def opdef)

	dontgen bool // manually written

	// opcode detail string
	// 	- r: declare 'val' and set it to 'oper/accumulator'
	// 	- w: write 'val' back into 'oper/accumulator'
	// 	- x: extra cycle for page crosses
	// 	- a: extra cycle always
	d string
}

var defs = [256]opdef{
	{i: 0x00, n: "BRK", d: "     ", m: "imp", dontgen: true},
	{i: 0x01, n: "ORA", d: "r    ", m: "izx"},
	{i: 0x02, n: "STP", d: "     ", m: "imp"},
	{i: 0x03, n: "SLO", d: "rw   ", m: "izx"},
	{i: 0x04, n: "NOP", d: "     ", m: "zpg"},
	{i: 0x05, n: "ORA", d: "r    ", m: "zpg"},
	{i: 0x06, n: "ASL", d: "rw   ", m: "zpg"},
	{i: 0x07, n: "SLO", d: "rw   ", m: "zpg"},
	{i: 0x08, n: "PHP", d: "     ", m: "imp"},
	{i: 0x09, n: "ORA", d: "r    ", m: "imm"},
	{i: 0x0A, n: "ASL", d: "rw   ", m: "acc"},
	{i: 0x0B, n: "ANC", d: "r    ", m: "imm"},
	{i: 0x0C, n: "NOP", d: "     ", m: "abs"},
	{i: 0x0D, n: "ORA", d: "r    ", m: "abs"},
	{i: 0x0E, n: "ASL", d: "rw   ", m: "abs"},
	{i: 0x0F, n: "SLO", d: "rw   ", m: "abs"},
	{i: 0x10, n: "BPL", d: "     ", m: "rel", f: branch(7, false)},
	{i: 0x11, n: "ORA", d: "r  x ", m: "izy"},
	{i: 0x12, n: "STP", d: "     ", m: "imp"},
	{i: 0x13, n: "SLO", d: "rw  a", m: "izy"},
	{i: 0x14, n: "NOP", d: "     ", m: "zpx"},
	{i: 0x15, n: "ORA", d: "r    ", m: "zpx"},
	{i: 0x16, n: "ASL", d: "rw   ", m: "zpx"},
	{i: 0x17, n: "SLO", d: "rw   ", m: "zpx"},
	{i: 0x18, n: "CLC", d: "     ", m: "imp", f: clear(0)},
	{i: 0x19, n: "ORA", d: "r  x ", m: "aby"},
	{i: 0x1A, n: "NOP", d: "     ", m: "imp"},
	{i: 0x1B, n: "SLO", d: "rw   ", m: "aby"},
	{i: 0x1C, n: "NOP", d: "   x ", m: "abx"},
	{i: 0x1D, n: "ORA", d: "r  x ", m: "abx"},
	{i: 0x1E, n: "ASL", d: "rw   ", m: "abx"},
	{i: 0x1F, n: "SLO", d: "rw   ", m: "abx"},
	{i: 0x20, n: "JSR", d: "     ", m: "abs", dontgen: true},
	{i: 0x21, n: "AND", d: "r    ", m: "izx"},
	{i: 0x22, n: "STP", d: "     ", m: "imp"},
	{i: 0x23, n: "RLA", d: "rw   ", m: "izx"},
	{i: 0x24, n: "BIT", d: "r    ", m: "zpg"},
	{i: 0x25, n: "AND", d: "r    ", m: "zpg"},
	{i: 0x26, n: "ROL", d: "rw   ", m: "zpg"},
	{i: 0x27, n: "RLA", d: "rw   ", m: "zpg"},
	{i: 0x28, n: "PLP", d: "     ", m: "imp"},
	{i: 0x29, n: "AND", d: "r    ", m: "imm"},
	{i: 0x2A, n: "ROL", d: "rw   ", m: "acc"},
	{i: 0x2B, n: "ANC", d: "r    ", m: "imm"},
	{i: 0x2C, n: "BIT", d: "r    ", m: "abs"},
	{i: 0x2D, n: "AND", d: "r    ", m: "abs"},
	{i: 0x2E, n: "ROL", d: "rw   ", m: "abs"},
	{i: 0x2F, n: "RLA", d: "rw   ", m: "abs"},
	{i: 0x30, n: "BMI", d: "     ", m: "rel", f: branch(7, true)},
	{i: 0x31, n: "AND", d: "r  x ", m: "izy"},
	{i: 0x32, n: "STP", d: "     ", m: "imp"},
	{i: 0x33, n: "RLA", d: "rw  a", m: "izy"},
	{i: 0x34, n: "NOP", d: "     ", m: "zpx"},
	{i: 0x35, n: "AND", d: "r    ", m: "zpx"},
	{i: 0x36, n: "ROL", d: "rw   ", m: "zpx"},
	{i: 0x37, n: "RLA", d: "rw   ", m: "zpx"},
	{i: 0x38, n: "SEC", d: "     ", m: "imp", f: set(0)},
	{i: 0x39, n: "AND", d: "r  x ", m: "aby"},
	{i: 0x3A, n: "NOP", d: "     ", m: "imp"},
	{i: 0x3B, n: "RLA", d: "rw   ", m: "aby"},
	{i: 0x3C, n: "NOP", d: "   x ", m: "abx"},
	{i: 0x3D, n: "AND", d: "r  x ", m: "abx"},
	{i: 0x3E, n: "ROL", d: "rw   ", m: "abx"},
	{i: 0x3F, n: "RLA", d: "rw   ", m: "abx"},
	{i: 0x40, n: "RTI", d: "     ", m: "imp"},
	{i: 0x41, n: "EOR", d: "r    ", m: "izx"},
	{i: 0x42, n: "STP", d: "     ", m: "imp"},
	{i: 0x43, n: "SRE", d: "rw   ", m: "izx"},
	{i: 0x44, n: "NOP", d: "     ", m: "zpg"},
	{i: 0x45, n: "EOR", d: "r    ", m: "zpg"},
	{i: 0x46, n: "LSR", d: "rw   ", m: "zpg"},
	{i: 0x47, n: "SRE", d: "rw   ", m: "zpg"},
	{i: 0x48, n: "PHA", d: "     ", m: "imp"},
	{i: 0x49, n: "EOR", d: "r    ", m: "imm"},
	{i: 0x4A, n: "LSR", d: "rw   ", m: "acc"},
	{i: 0x4B, n: "ALR", d: "r    ", m: "imm"},
	{i: 0x4C, n: "JMP", d: "     ", m: "abs"},
	{i: 0x4D, n: "EOR", d: "r    ", m: "abs"},
	{i: 0x4E, n: "LSR", d: "rw   ", m: "abs"},
	{i: 0x4F, n: "SRE", d: "rw   ", m: "abs"},
	{i: 0x50, n: "BVC", d: "     ", m: "rel", f: branch(6, false)},
	{i: 0x51, n: "EOR", d: "r  x ", m: "izy"},
	{i: 0x52, n: "STP", d: "     ", m: "imp"},
	{i: 0x53, n: "SRE", d: "rw  a", m: "izy"},
	{i: 0x54, n: "NOP", d: "     ", m: "zpx"},
	{i: 0x55, n: "EOR", d: "r    ", m: "zpx"},
	{i: 0x56, n: "LSR", d: "rw   ", m: "zpx"},
	{i: 0x57, n: "SRE", d: "rw   ", m: "zpx"},
	{i: 0x58, n: "CLI", d: "     ", m: "imp", f: clear(2)},
	{i: 0x59, n: "EOR", d: "r  x ", m: "aby"},
	{i: 0x5A, n: "NOP", d: "     ", m: "imp"},
	{i: 0x5B, n: "SRE", d: "rw   ", m: "aby"},
	{i: 0x5C, n: "NOP", d: "   x ", m: "abx"},
	{i: 0x5D, n: "EOR", d: "r  x ", m: "abx"},
	{i: 0x5E, n: "LSR", d: "rw   ", m: "abx"},
	{i: 0x5F, n: "SRE", d: "rw   ", m: "abx"},
	{i: 0x60, n: "RTS", d: "     ", m: "imp"},
	{i: 0x61, n: "ADC", d: "r    ", m: "izx"},
	{i: 0x62, n: "STP", d: "     ", m: "imp"},
	{i: 0x63, n: "RRA", d: "rw   ", m: "izx"},
	{i: 0x64, n: "NOP", d: "     ", m: "zpg"},
	{i: 0x65, n: "ADC", d: "r    ", m: "zpg"},
	{i: 0x66, n: "ROR", d: "rw   ", m: "zpg"},
	{i: 0x67, n: "RRA", d: "rw   ", m: "zpg"},
	{i: 0x68, n: "PLA", d: "     ", m: "imp"},
	{i: 0x69, n: "ADC", d: "r    ", m: "imm"},
	{i: 0x6A, n: "ROR", d: "rw   ", m: "acc"},
	{i: 0x6B, n: "ARR", d: "r    ", m: "imm"},
	{i: 0x6C, n: "JMP", d: "     ", m: "ind"},
	{i: 0x6D, n: "ADC", d: "r    ", m: "abs"},
	{i: 0x6E, n: "ROR", d: "rw   ", m: "abs"},
	{i: 0x6F, n: "RRA", d: "rw   ", m: "abs"},
	{i: 0x70, n: "BVS", d: "     ", m: "rel", f: branch(6, true)},
	{i: 0x71, n: "ADC", d: "r  x ", m: "izy"},
	{i: 0x72, n: "STP", d: "     ", m: "imp"},
	{i: 0x73, n: "RRA", d: "rw  a", m: "izy"},
	{i: 0x74, n: "NOP", d: "     ", m: "zpx"},
	{i: 0x75, n: "ADC", d: "r    ", m: "zpx"},
	{i: 0x76, n: "ROR", d: "rw   ", m: "zpx"},
	{i: 0x77, n: "RRA", d: "rw   ", m: "zpx"},
	{i: 0x78, n: "SEI", d: "     ", m: "imp", f: set(2)},
	{i: 0x79, n: "ADC", d: "r  x ", m: "aby"},
	{i: 0x7A, n: "NOP", d: "     ", m: "imp"},
	{i: 0x7B, n: "RRA", d: "rw   ", m: "aby"},
	{i: 0x7C, n: "NOP", d: "   x ", m: "abx"},
	{i: 0x7D, n: "ADC", d: "r  x ", m: "abx"},
	{i: 0x7E, n: "ROR", d: "rw   ", m: "abx"},
	{i: 0x7F, n: "RRA", d: "rw   ", m: "abx"},
	{i: 0x80, n: "NOP", d: "r    ", m: "imm"},
	{i: 0x81, n: "STA", d: "     ", m: "izx", f: ST("A")},
	{i: 0x82, n: "NOP", d: "r    ", m: "imm"},
	{i: 0x83, n: "SAX", d: "     ", m: "izx"},
	{i: 0x84, n: "STY", d: "     ", m: "zpg", f: ST("Y")},
	{i: 0x85, n: "STA", d: "     ", m: "zpg", f: ST("A")},
	{i: 0x86, n: "STX", d: "     ", m: "zpg", f: ST("X")},
	{i: 0x87, n: "SAX", d: "     ", m: "zpg"},
	{i: 0x88, n: "DEY", d: "     ", m: "imp", f: dec("Y")},
	{i: 0x89, n: "NOP", d: "r    ", m: "imm"},
	{i: 0x8A, n: "TXA", d: "     ", m: "imp", f: T("X", "A")},
	{i: 0x8B, n: "ANE", d: "     ", m: "imm", f: unstable},
	{i: 0x8C, n: "STY", d: "     ", m: "abs", f: ST("Y")},
	{i: 0x8D, n: "STA", d: "     ", m: "abs", f: ST("A")},
	{i: 0x8E, n: "STX", d: "     ", m: "abs", f: ST("X")},
	{i: 0x8F, n: "SAX", d: "     ", m: "abs"},
	{i: 0x90, n: "BCC", d: "     ", m: "rel", f: branch(0, false)},
	{i: 0x91, n: "STA", d: "    a", m: "izy", f: ST("A")},
	{i: 0x92, n: "STP", d: "     ", m: "imp"},
	{i: 0x93, n: "SHA", d: "     ", m: "izy", f: unstable},
	{i: 0x94, n: "STY", d: "     ", m: "zpx", f: ST("Y")},
	{i: 0x95, n: "STA", d: "     ", m: "zpx", f: ST("A")},
	{i: 0x96, n: "STX", d: "     ", m: "zpy", f: ST("X")},
	{i: 0x97, n: "SAX", d: "     ", m: "zpy"},
	{i: 0x98, n: "TYA", d: "     ", m: "imp", f: T("Y", "A")},
	{i: 0x99, n: "STA", d: "     ", m: "aby", f: ST("A")},
	{i: 0x9A, n: "TXS", d: "     ", m: "imp", f: T("X", "SP")},
	{i: 0x9B, n: "TAS", d: "     ", m: "aby", f: unstable},
	{i: 0x9C, n: "SHY", d: "     ", m: "abx", f: unstable},
	{i: 0x9D, n: "STA", d: "     ", m: "abx", f: ST("A")},
	{i: 0x9E, n: "SHX", d: "     ", m: "aby", f: unstable},
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
	{i: 0xAB, n: "LXA", d: "r    ", m: "imm"},
	{i: 0xAC, n: "LDY", d: "r    ", m: "abs", f: LD("Y")},
	{i: 0xAD, n: "LDA", d: "r    ", m: "abs", f: LD("A")},
	{i: 0xAE, n: "LDX", d: "r    ", m: "abs", f: LD("X")},
	{i: 0xAF, n: "LAX", d: "r    ", m: "abs", f: LD("A", "X")},
	{i: 0xB0, n: "BCS", d: "     ", m: "rel", f: branch(0, true)},
	{i: 0xB1, n: "LDA", d: "r  x ", m: "izy", f: LD("A")},
	{i: 0xB2, n: "STP", d: "     ", m: "imp"},
	{i: 0xB3, n: "LAX", d: "r  x ", m: "izy", f: LD("A", "X")},
	{i: 0xB4, n: "LDY", d: "r    ", m: "zpx", f: LD("Y")},
	{i: 0xB5, n: "LDA", d: "r    ", m: "zpx", f: LD("A")},
	{i: 0xB6, n: "LDX", d: "r    ", m: "zpy", f: LD("X")},
	{i: 0xB7, n: "LAX", d: "r    ", m: "zpy", f: LD("A", "X")},
	{i: 0xB8, n: "CLV", d: "     ", m: "imp", f: clear(6)},
	{i: 0xB9, n: "LDA", d: "r  x ", m: "aby", f: LD("A")},
	{i: 0xBA, n: "TSX", d: "     ", m: "imp", f: T("SP", "X")},
	{i: 0xBB, n: "LAS", d: "r  x ", m: "aby"},
	{i: 0xBC, n: "LDY", d: "r  x ", m: "abx", f: LD("Y")},
	{i: 0xBD, n: "LDA", d: "r  x ", m: "abx", f: LD("A")},
	{i: 0xBE, n: "LDX", d: "r  x ", m: "aby", f: LD("X")},
	{i: 0xBF, n: "LAX", d: "r  x ", m: "aby", f: LD("A", "X")},
	{i: 0xC0, n: "CPY", d: "r    ", m: "imm", f: cmp("Y")},
	{i: 0xC1, n: "CMP", d: "r    ", m: "izx", f: cmp("A")},
	{i: 0xC2, n: "NOP", d: "r    ", m: "imm"},
	{i: 0xC3, n: "DCP", d: "rw   ", m: "izx"},
	{i: 0xC4, n: "CPY", d: "r    ", m: "zpg", f: cmp("Y")},
	{i: 0xC5, n: "CMP", d: "r    ", m: "zpg", f: cmp("A")},
	{i: 0xC6, n: "DEC", d: "rw   ", m: "zpg", f: dec("")},
	{i: 0xC7, n: "DCP", d: "rw   ", m: "zpg"},
	{i: 0xC8, n: "INY", d: "     ", m: "imp", f: inc("Y")},
	{i: 0xC9, n: "CMP", d: "r    ", m: "imm", f: cmp("A")},
	{i: 0xCA, n: "DEX", d: "     ", m: "imp", f: dec("X")},
	{i: 0xCB, n: "SBX", d: "r    ", m: "imm"},
	{i: 0xCC, n: "CPY", d: "r    ", m: "abs", f: cmp("Y")},
	{i: 0xCD, n: "CMP", d: "r    ", m: "abs", f: cmp("A")},
	{i: 0xCE, n: "DEC", d: "rw   ", m: "abs", f: dec("")},
	{i: 0xCF, n: "DCP", d: "rw   ", m: "abs"},
	{i: 0xD0, n: "BNE", d: "     ", m: "rel", f: branch(1, false)},
	{i: 0xD1, n: "CMP", d: "r  x ", m: "izy", f: cmp("A")},
	{i: 0xD2, n: "STP", d: "     ", m: "imp"},
	{i: 0xD3, n: "DCP", d: "rw  a", m: "izy"},
	{i: 0xD4, n: "NOP", d: "     ", m: "zpx"},
	{i: 0xD5, n: "CMP", d: "r    ", m: "zpx", f: cmp("A")},
	{i: 0xD6, n: "DEC", d: "rw   ", m: "zpx", f: dec("")},
	{i: 0xD7, n: "DCP", d: "rw   ", m: "zpx"},
	{i: 0xD8, n: "CLD", d: "     ", m: "imp", f: clear(3)},
	{i: 0xD9, n: "CMP", d: "r  x ", m: "aby", f: cmp("A")},
	{i: 0xDA, n: "NOP", d: "     ", m: "imp"},
	{i: 0xDB, n: "DCP", d: "rw   ", m: "aby"},
	{i: 0xDC, n: "NOP", d: "   x ", m: "abx"},
	{i: 0xDD, n: "CMP", d: "r  x ", m: "abx", f: cmp("A")},
	{i: 0xDE, n: "DEC", d: "rw   ", m: "abx", f: dec("")},
	{i: 0xDF, n: "DCP", d: "rw   ", m: "abx"},
	{i: 0xE0, n: "CPX", d: "r    ", m: "imm", f: cmp("X")},
	{i: 0xE1, n: "SBC", d: "r    ", m: "izx"},
	{i: 0xE2, n: "NOP", d: "r    ", m: "imm"},
	{i: 0xE3, n: "ISC", d: "rw   ", m: "izx"},
	{i: 0xE4, n: "CPX", d: "r    ", m: "zpg", f: cmp("X")},
	{i: 0xE5, n: "SBC", d: "r    ", m: "zpg"},
	{i: 0xE6, n: "INC", d: "rw   ", m: "zpg", f: inc("")},
	{i: 0xE7, n: "ISC", d: "rw   ", m: "zpg"},
	{i: 0xE8, n: "INX", d: "     ", m: "imp", f: inc("X")},
	{i: 0xE9, n: "SBC", d: "r    ", m: "imm"},
	{i: 0xEA, n: "NOP", d: "     ", m: "imp"},
	{i: 0xEB, n: "SBC", d: "r    ", m: "imm"},
	{i: 0xEC, n: "CPX", d: "r    ", m: "abs", f: cmp("X")},
	{i: 0xED, n: "SBC", d: "r    ", m: "abs"},
	{i: 0xEE, n: "INC", d: "rw   ", m: "abs", f: inc("")},
	{i: 0xEF, n: "ISC", d: "rw   ", m: "abs"},
	{i: 0xF0, n: "BEQ", d: "     ", m: "rel", f: branch(1, true)},
	{i: 0xF1, n: "SBC", d: "r  x ", m: "izy"},
	{i: 0xF2, n: "STP", d: "     ", m: "imp"},
	{i: 0xF3, n: "ISC", d: "rw  a", m: "izy"},
	{i: 0xF4, n: "NOP", d: "     ", m: "zpx"},
	{i: 0xF5, n: "SBC", d: "r    ", m: "zpx"},
	{i: 0xF6, n: "INC", d: "rw   ", m: "zpx", f: inc("")},
	{i: 0xF7, n: "ISC", d: "rw   ", m: "zpx"},
	{i: 0xF8, n: "SED", d: "     ", m: "imp", f: set(3)},
	{i: 0xF9, n: "SBC", d: "r  x ", m: "aby"},
	{i: 0xFA, n: "NOP", d: "     ", m: "imp"},
	{i: 0xFB, n: "ISC", d: "rw   ", m: "aby"},
	{i: 0xFC, n: "NOP", d: "   x ", m: "abx"},
	{i: 0xFD, n: "SBC", d: "r  x ", m: "abx"},
	{i: 0xFE, n: "INC", d: "rw   ", m: "abx", f: inc("")},
	{i: 0xFF, n: "ISC", d: "rw   ", m: "abx"},
}

type addrmode struct {
	human string // human readable name
	n     int    // number of bytes
	f     func(g *Generator, details string)
}

var addrModes = map[string]addrmode{
	"imp": {f: imp, n: 1, human: `implied addressing.`},
	"acc": {f: acc, n: 1, human: `adressing accumulator.`},
	"rel": {f: rel, n: 2, human: `relative addressing.`},
	"abs": {f: abs, n: 3, human: `absolute addressing.`},
	"abx": {f: abx, n: 3, human: `absolute indexed X.`},
	"aby": {f: aby, n: 3, human: `absolute indexed Y.`},
	"imm": {f: imm, n: 2, human: `immediate addressing.`},
	"ind": {f: ind, n: 3, human: `indirect addressing.`},
	"izx": {f: izx, n: 2, human: `indexed addressing (abs, X).`},
	"izy": {f: izy, n: 2, human: `indexed addressing (abs),Y.`},
	"zpg": {f: zpg, n: 2, human: `zero page addressing.`},
	"zpx": {f: zpx, n: 2, human: `indexed addressing: zeropage,X.`},
	"zpy": {f: zpy, n: 2, human: `indexed addressing: zeropage,Y.`},
}

//
// Process status flag constants
//

const (
	Carry = iota
	Zero
	IntDisable
	Decimal
	Break
	Unused
	Overflow
	Negative
)

var flagOps = [8][2]string{
	{"carry", "setCarry"},
	{"zero", "setZero"},
	{"intDisable", "setIntDisable"},
	{"decimal", "setDecimal"},
	{"brk", "setBrk"},
	{"unused", "setUnused"},
	{"overflow", "setOverflow"},
	{"negative", "setNegative"},
}

func pget(i uint) string { return flagOps[i][0] }
func pset(i uint) string { return flagOps[i][1] }

func (g *Generator) dummyread(oper string) {
	g.printf(`// dummy read.`)
	g.printf(`_ = cpu.Read8(%s)`, oper)
}

func (g *Generator) dummywrite(addr, value string) {
	g.printf(`// dummy write.`)
	g.printf(`cpu.Write8(%s, %s)`, addr, value)
}

//
// addressing modes
//

func acc(g *Generator, _ string) {
	g.dummyread("cpu.PC")
}

func imp(g *Generator, _ string) {
	g.dummyread("cpu.PC")
}

func ind(g *Generator, _ string) {
	g.printf(`oper := cpu.Read16(cpu.PC)`)
	g.printf(`lo := cpu.Read8(oper)`)
	g.printf(`// 2 bytes address wrap around`)
	g.printf(`hi := cpu.Read8((0xff00 & oper) | (0x00ff & (oper + 1)))`)
	g.printf(`oper = uint16(hi)<<8 | uint16(lo)`)
}

func imm(g *Generator, _ string) {}

func rel(g *Generator, _ string) {
	g.printf(`off := int16(int8(cpu.fetch()))`)
	g.printf(`oper := uint16(int16(cpu.PC) + off)`)
}

func abs(g *Generator, _ string) {
	g.printf(`oper := cpu.Read16(cpu.PC)`)
	g.printf(`cpu.PC += 2`)
}

// seems that we don't even need the dummyread bool

func abx(g *Generator, info string) {
	g.printf(`addr := cpu.Read16(cpu.PC)`)
	g.printf(`cpu.PC += 2`)
	g.printf(`oper := addr + uint16(cpu.X)`)

	switch {
	case has(info, 'x'):
		tickIfPageCrossed(g, "addr", "oper")
	default:
		g.dummyread(fmt.Sprintf("%s & 0x00FF | %s & 0xFF00", "oper", "addr"))
	}
}

func aby(g *Generator, info string) {
	g.printf(`addr := cpu.Read16(cpu.PC)`)
	g.printf(`cpu.PC += 2`)
	g.printf(`oper := addr + uint16(cpu.Y)`)

	switch {
	case has(info, 'x'):
		tickIfPageCrossed(g, "addr", "oper")
	default:
		g.dummyread(fmt.Sprintf("%s & 0x00FF | %s & 0xFF00", "oper", "addr"))
	}
}

func zpg(g *Generator, _ string) {
	g.printf(`oper := uint16(cpu.fetch())`)
}

func zpx(g *Generator, _ string) {
	g.printf(`addr := cpu.fetch()`)
	g.dummyread("uint16(addr)")
	g.printf(`oper := uint16(addr) + uint16(cpu.X)`)
	g.printf(`oper &= 0xff`)
}

func zpy(g *Generator, _ string) {
	g.printf(`addr := cpu.fetch()`)
	g.dummyread("uint16(addr)")
	g.printf(`oper := uint16(addr) + uint16(cpu.Y)`)
	g.printf(`oper &= 0xff`)
}

func izx(g *Generator, info string) {
	g.printf(`oper := uint16(cpu.fetch())`)
	g.dummyread("uint16(oper)")
	g.printf(`oper = uint16(uint8(oper) + cpu.X)`)
	r16zpwrap(g)
}

func izy(g *Generator, info string) {
	g.printf(`oper := uint16(cpu.fetch())`)
	r16zpwrap(g)

	switch {
	case has(info, 'x'):
		g.printf(`if 0xFF00&(oper) != 0xFF00&(oper+uint16(cpu.Y)) {`)
		g.printf(`// extra cycle for page cross`)
		g.dummyread(`oper + uint16(cpu.Y) - 0x100`)
		g.printf(`}`)
	case has(info, 'a'):
		g.printf(`// page crossed?`)
		g.printf(`if 0xFF00&(oper) != 0xFF00&(oper+uint16(cpu.Y)) {`)
		g.dummyread(`oper + uint16(cpu.Y) - 0x100`)
		g.printf(`} else {`)
		g.dummyread(`oper + uint16(cpu.Y)`)
		g.printf(`}`)
	default:
	}

	g.printf(`oper += uint16(cpu.Y)`)
}

// helpers

func push8(g *Generator, v string) {
	g.printf(`cpu.push8(%s)`, v)
}

func push16(g *Generator, v string) {
	g.printf(`cpu.push16(%s)`, v)
}

func pull8(g *Generator, v string) {
	g.printf(`%s = cpu.pull8()`, v)
}

func pull16(g *Generator, v string) {
	g.printf(`%s = cpu.pull16()`, v)
}

func carrybit(g *Generator) {
	g.printf(`var carry uint16`)
	g.printf(`if cpu.P.%s() {`, pget(Carry))
	g.printf(`	carry = 1`)
	g.printf(`}`)
}

// read 16 bytes from the zero page, handling page wrap.
func r16zpwrap(g *Generator) {
	g.printf(`// read 16 bytes from the zero page, handling page wrap`)
	g.printf(`lo := cpu.Read8(oper)`)
	g.printf(`hi := cpu.Read8(uint16(uint8(oper) + 1))`)
	g.printf(`oper = uint16(hi)<<8 | uint16(lo)`)
}

func branch(ibit int, val bool) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		neg := "!"
		if !val {
			neg = ""
		}
		g.printf(`if %scpu.P.%s() {`, neg, pget(uint(ibit)))
		g.printf(`  return // no branch`)
		g.printf(`}`)
		g.printf(`// A taken non-page-crossing branch ignores IRQ/NMI during its last`)
		g.printf(`// clock, so that next instruction executes before the IRQ.`)
		g.printf(`// Fixes 'branch_delays_irq' test.`)
		g.printf(`if cpu.runIRQ && !cpu.prevRunIRQ {`)
		g.printf(`	cpu.runIRQ = false`)
		g.printf(`}`)
		g.dummyread("cpu.PC")
		tickIfPageCrossed(g, "cpu.PC", "oper")
		g.printf(`	cpu.PC = oper`)
		g.printf(`	return`)
	}
}

func tick(g *Generator) {
	g.printf(`cpu.tick()`)
}

func tickIfPageCrossed(g *Generator, a, b string) {
	g.printf(`// extra cycle for page cross`)
	g.printf(`if 0xFF00&(%s) != 0xFF00&(%s) {`, a, b)
	g.dummyread(fmt.Sprintf("(%s) & 0x00FF | (%s) & 0xFF00", b, a))
	g.printf(`}`)
}

func copybits(dst, src, mask string) string {
	return fmt.Sprintf(`((%s) & (^%s)) | ((%s) & (%s))`, dst, mask, src, mask)
}

//
// opcode generators
//

func (g *Generator) STP(def opdef) {
	g.unstable = append(g.unstable, def.i)
	g.printf(`cpu.halt()`)
}

func (g *Generator) ADC(_ opdef) {
	carrybit(g)
	g.printf(`sum := uint16(cpu.A) + uint16(val) + uint16(carry)`)
	g.printf(`cpu.P.checkCV(cpu.A, val, sum)`)
	g.printf(`cpu.A = uint8(sum)`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func (g *Generator) ALR(_ opdef) {
	g.printf(`// like and + lsr but saves one tick`)
	g.printf(`cpu.A &= val`)
	g.printf(`carry := cpu.A & 0x01 // carry is bit 0`)
	g.printf(`cpu.A = (cpu.A >> 1) & 0x7f`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
	g.printf(`cpu.P.%s(carry != 0)`, pset(Carry))
}

func (g *Generator) ANC(def opdef) {
	g.AND(def)
	g.printf(`cpu.P.%s(cpu.P.%s())`, pset(Carry), pget(Negative))
}

func (g *Generator) AND(_ opdef) {
	g.printf(`cpu.A &= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func (g *Generator) ARR(_ opdef) {
	g.printf(`cpu.A &= val`)
	g.printf(`cpu.A >>= 1`)
	g.printf(`cpu.P.%s((cpu.A>>6)^(cpu.A>>5)&0x01 != 0)`, pset(Overflow))
	g.printf(`if cpu.P.%s() {`, pget(Carry))
	g.printf(`	cpu.A |= 1 << 7`)
	g.printf(`}`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
	g.printf(`cpu.P.%s(cpu.A&(1<<6) != 0)`, pset(Carry))
}

func (g *Generator) ASL(def opdef) {
	if def.m != "acc" {
		g.dummywrite("oper", "val")
	}
	g.printf(`carry := val & 0x80`)
	g.printf(`val = (val << 1) & 0xfe`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.%s(carry != 0)`, pset(Carry))
}

func (g *Generator) BIT(_ opdef) {
	g.printf(`cpu.P &= 0b00111111`)
	g.printf(`cpu.P |= P(val & 0b11000000)`)
	g.printf(`cpu.P.checkZ(cpu.A & val)`)
}

func (g *Generator) BRK(_ opdef) {
	tick(g)
	push16(g, `cpu.PC+1`)
	g.printf(`p := cpu.P`)
	g.printf(`p.%s(true)`, pset(Break))
	push8(g, `uint8(p)`)
	g.printf(`cpu.P.%s(true)`, pset(IntDisable))
	g.printf(`cpu.PC = cpu.Read16(CpuIRQvector)`)
}

func (g *Generator) DCP(def opdef) {
	dec("")(g, def)
	cmp("A")(g, def)
}

func (g *Generator) EOR(_ opdef) {
	g.printf(`cpu.A ^= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func (g *Generator) ISC(def opdef) {
	inc("")(g, def)
	g.printf(`final := val`)
	g.SBC(def)
	g.printf(`val = final`)
}

func (g *Generator) JMP(_ opdef) {
	g.printf(`cpu.PC = oper`)
}

func (g *Generator) LAS(def opdef) {
	g.printf(`cpu.A = cpu.SP & val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
	g.printf(`cpu.X = cpu.A`)
	g.printf(`cpu.SP = cpu.A`)
}

func (g *Generator) LSR(def opdef) {
	if def.m != "acc" {
		g.dummywrite("oper", "val")
	}

	g.printf(`{`)
	g.printf(`carry := val & 0x01 // carry is bit 0`)
	g.printf(`val = (val >> 1)&0x7f`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.%s(carry != 0)`, pset(Carry))
	g.printf(`}`)
}

func (g *Generator) LXA(def opdef) {
	g.unstable = append(g.unstable, def.i)

	const mask = 0xff
	g.printf(`val = (cpu.A | 0x%02x) & val`, mask)
	g.printf(`cpu.A = val`)
	g.printf(`cpu.X = val`)
	g.printf(`cpu.P.checkNZ(val)`)
}

func (g *Generator) NOP(def opdef) {
	if !slices.Contains([]string{"acc", "imp", "rel", "imm"}, def.m) {
		g.dummyread("oper")
	}
	if def.m == "imm" {
		g.printf(`_ = val`)
	}
}

func (g *Generator) ORA(_ opdef) {
	g.printf(`cpu.A |= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func (g *Generator) PHA(_ opdef) {
	push8(g, `cpu.A`)
}

func (g *Generator) PHP(_ opdef) {
	g.printf(`p := cpu.P`)
	g.printf(`p.%s(true)`, pset(Break))
	g.printf(`p.%s(true)`, pset(Unused))
	push8(g, `uint8(p)`)
}

func (g *Generator) PLA(_ opdef) {
	g.dummyread("uint16(cpu.SP) + 0x0100")
	pull8(g, `cpu.A`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func (g *Generator) PLP(_ opdef) {
	g.printf(`var p uint8`)
	g.dummyread("uint16(cpu.SP) + 0x0100")
	pull8(g, `p`)
	g.printf(`const mask uint8 = 0b11001111 // ignore B and U bits`)
	g.printf(`cpu.P = P(%s)`, copybits(`uint8(cpu.P)`, `p`, `mask`))
}

func (g *Generator) RLA(def opdef) {
	g.ROL(def)
	g.AND(def)
}

func (g *Generator) ROL(def opdef) {
	if def.m != "acc" {
		g.dummywrite("oper", "val")
	}
	g.printf(`carry := val & 0x80`)
	g.printf(`val <<= 1`)
	g.printf(`if cpu.P.%s() {`, pget(Carry))
	g.printf(`	val |= 1 << 0`)
	g.printf(`}`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.%s(carry != 0)`, pset(Carry))
}

func (g *Generator) ROR(def opdef) {
	if def.m != "acc" {
		g.dummywrite("oper", "val")
	}
	g.printf(`{`)
	g.printf(`carry := val & 0x01`)
	g.printf(`val >>= 1`)
	g.printf(`if cpu.P.%s() {`, pget(Carry))
	g.printf(`	val |= 1 << 7`)
	g.printf(`}`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.%s(carry != 0)`, pset(Carry))
	g.printf(`}`)
}

func (g *Generator) RRA(def opdef) {
	g.ROR(def)
	g.ADC(def)
}

func (g *Generator) RTI(_ opdef) {
	g.printf(`var p uint8`)
	g.dummyread("uint16(cpu.SP) + 0x0100")
	pull8(g, `p`)
	g.printf(`const mask uint8 = 0b11001111 // ignore B and U bits`)
	g.printf(`cpu.P = P(%s)`, copybits(`uint8(cpu.P)`, `p`, `mask`))
	pull16(g, `cpu.PC`)
}

func (g *Generator) RTS(_ opdef) {
	g.dummyread("uint16(cpu.SP) + 0x0100")
	pull16(g, `cpu.PC`)
	g.printf(`cpu.fetch()`)
}

func (g *Generator) SAX(_ opdef) {
	g.printf(`cpu.Write8(oper, cpu.A&cpu.X)`)
}

func (g *Generator) SBC(def opdef) {
	g.printf(`val ^= 0xff`)
	carrybit(g)
	g.printf(`sum := uint16(cpu.A) + uint16(val) + uint16(carry)`)
	g.printf(`cpu.P.checkCV(cpu.A, val, sum)`)
	g.printf(`cpu.A = uint8(sum)`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func (g *Generator) SBX(def opdef) {
	g.printf(`ival := (int16(cpu.A) & int16(cpu.X)) - int16(val)`)
	g.printf(`cpu.X = uint8(ival)`)
	g.printf(`cpu.P.checkNZ(uint8(ival))`)
	g.printf(`cpu.P.%s(ival >= 0)`, pset(Carry))
}

func (g *Generator) SLO(def opdef) {
	g.printf(`// SLO start`)
	g.ASL(def)
	g.printf(`cpu.A |= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
	g.printf(`// SLO end`)
}

func (g *Generator) SRE(def opdef) {
	g.LSR(def)
	g.EOR(def)
}

//
// opcode helpers
//

func LD(reg ...string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		for _, r := range reg {
			g.printf(`cpu.%s = val`, r)
		}
		g.printf(`cpu.P.checkNZ(val)`)
	}
}

func ST(reg string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.Write8(oper, cpu.%s)`, reg)
	}
}

func cmp(v string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		v = regOrMem(v)
		g.printf(`cpu.P.checkNZ(%s - val)`, v)
		g.printf(`cpu.P.%s(val <= %s)`, pset(Carry), v)
	}
}

func T(src, dst string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.%s = cpu.%s`, dst, src)
		if dst != "SP" {
			g.printf(`cpu.P.checkNZ(cpu.%s)`, src)
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

func inc(v string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		v = regOrMem(v)
		if v == "val" {
			// TODO: works but ugly
			g.dummywrite("oper", "val")
		}
		g.printf(`%s++`, v)
		g.printf(`cpu.P.checkNZ(%s)`, v)
	}
}

func dec(v string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		v = regOrMem(v)
		if v == "val" {
			// TODO: works but ugly
			g.dummywrite("oper", "val")
		}
		g.printf(`%s--`, v)
		g.printf(`cpu.P.checkNZ(%s)`, v)
	}
}

func clear(ibit uint) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.P.%s(false)`, pset(ibit))
	}
}

func set(ibit uint) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.P.%s(true)`, pset(ibit))
	}
}

func unstable(g *Generator, def opdef) {
	g.unstable = append(g.unstable, def.i)
	insertPanic(g, fmt.Sprintf("unsupported unstable opcode 0x%02X (%s)", def.i, def.n))
}

func insertPanic(g *Generator, msg string) {
	g.printf(`msg := fmt.Sprintf("%s\nPC:0x%%04X", cpu.PC)`, msg)
	g.printf(`panic(msg)`)
}

type Generator struct {
	io.Writer
	unstable []uint8
}

func (g *Generator) header() {
	g.printf(`// Code generated by cpugen/gen_nes6502.go. DO NOT EDIT.`)
	g.printf(`package %s`, pkgname)
	g.printf(`import (`)
	g.printf(`"fmt"`)
	g.printf(`)`)
}

func has(details string, c byte) bool {
	for _, i := range details {
		if byte(i) == c {
			return true
		}
	}
	return false
}

func (g *Generator) opcodeHeader(code uint8) {
	mode, ok := addrModes[defs[code].m]
	if !ok {
		panic(fmt.Sprintf("unknown addressing mode (opcode: 0x%02X)", code))
	}

	g.printf(`// %s - %s`, defs[code].n, mode.human)
	g.printf(`func opcode%02X(cpu*CPU){`, code)
	if mode.f != nil {
		mode.f(g, defs[code].d)
	}

	switch {
	case strings.Contains(defs[code].d, "r"):
		switch defs[code].m {
		case "acc":
			g.printf(`val := cpu.A`)
		case "imm":
			g.printf(`val := cpu.fetch()`)
		default:
			g.printf(`val := cpu.Read8(oper)`)
		}
	}
}

func (g *Generator) opcodeFooter(code uint8) {
	switch {
	case strings.Contains(defs[code].d, "w"):
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
		if def.dontgen {
			continue
		}
		g.opcodeHeader(def.i)
		if def.f != nil {
			def.f(g, def)
		} else {
			f := reflect.ValueOf(g).MethodByName(def.n)
			f.Call([]reflect.Value{reflect.ValueOf(def)})
		}

		g.opcodeFooter(def.i)
		g.printf("\n")
	}
}

func (g *Generator) disasmAddrModes() {
	// order alphabetically to get deterministic output.
	var modes []string
	for k := range addrModes {
		modes = append(modes, k)
	}

	sort.Strings(modes)
	for _, name := range modes {
		am := addrModes[name]
		fname := strings.ToUpper(name[:1]) + name[1:]
		g.printf(`func disasm%s(cpu*CPU, pc uint16) DisasmOp {`, fname)
		for n := 0; n < am.n; n++ {
			g.printf(`oper%d := cpu.Bus.Peek8(pc+%d)`, n, n)
		}
		var bytes []string
		for n := 0; n < am.n; n++ {
			bytes = append(bytes, "oper"+strconv.Itoa(n))
		}

		if am.n == 3 {
			g.printf(`operaddr := uint16(oper1)|uint16(oper2)<<8`)
		}
		g.printf(`oper := ""`)
		g.printf(``)

		abxy := func(xy byte) {
			g.printf(`addr := operaddr + uint16(cpu.%c)`, xy)
			g.printf(`pointee := cpu.Bus.Peek8(addr)`)
			g.printf(`oper = fmt.Sprintf("%%s,%c [%%s] = $%%02X", formatAddr(operaddr), formatAddr(addr), pointee)`, xy)
		}

		switch name {
		case "imp":
		case "acc":
			g.printf(`oper = "A"`)
		case "rel":
			g.printf(`oper = fmt.Sprintf("$%%04X", uint16(int16(pc+2) + int16(int8(oper1))))`)
		case "ind":
			g.printf(`lo := cpu.Bus.Peek8(operaddr)`)
			g.printf(`// 2 bytes address wrap around`)
			g.printf(`hi := cpu.Bus.Peek8((0xff00 & operaddr) | (0x00ff & (operaddr + 1)))`)
			g.printf(`dest := uint16(hi)<<8 | uint16(lo)`)
			g.printf(`pointee := cpu.Bus.Peek8(dest)`)
			g.printf(`oper = fmt.Sprintf("(%%s) [%%s] = $%%02X", formatAddr(operaddr), formatAddr(dest), pointee)`)
		case "abs":
			g.printf(`if oper0 == 0x20 || oper0 == 0x4C {`)
			g.printf("        // JSR / JMP")
			g.printf(`        oper = fmt.Sprintf("$%%04X", operaddr)`)
			g.printf(`} else {`)
			g.printf(`        pointee := cpu.Bus.Peek8(operaddr)`)
			g.printf(`        oper = fmt.Sprintf("%%s = $%%02X", formatAddr(operaddr), pointee)`)
			g.printf(`}`)
		case "abx":
			abxy('X')
		case "aby":
			abxy('Y')
		case "imm":
			g.printf(`oper = fmt.Sprintf("#$%%02X", oper1)`)
		case "zpg":
			g.printf(`pointee := cpu.Bus.Peek8(uint16(oper1))`)
			g.printf(`oper = fmt.Sprintf("$%%02X = $%%02X", oper1, pointee)`)
		case "zpx":
			g.printf(`addr := uint16(oper1) + uint16(cpu.X)`)
			g.printf(`addr &= 0xff`)
			g.printf(`pointee := cpu.Bus.Peek8(addr)`)
			g.printf(`oper = fmt.Sprintf("$%%02X,X [%%s] = $%%02X", oper1, formatAddr(addr), pointee)`)
		case "zpy":
			g.printf(`addr := uint16(oper1) + uint16(cpu.Y)`)
			g.printf(`addr &= 0xff`)
			g.printf(`pointee := cpu.Bus.Peek8(addr)`)
			g.printf(`oper = fmt.Sprintf("$%%02X,Y [%%s] = $%%02X", oper1, formatAddr(addr), pointee)`)
		case "zp":
			g.printf(`oper = fmt.Sprintf("$%%02X", oper1)`)
		case "izx":
			g.printf(`addr := uint16(uint8(oper1) + cpu.X)`)
			g.printf(`// read 16 bytes from the zero page, handling page wrap`)
			g.printf(`lo := cpu.Bus.Peek8(addr)`)
			g.printf(`hi := cpu.Bus.Peek8(uint16(uint8(addr) + 1))`)
			g.printf(`addr = uint16(hi)<<8 | uint16(lo)`)
			g.printf(`pointee := cpu.Bus.Peek8(addr)`)
			g.printf(`oper = fmt.Sprintf("($%%02X,X) [%%s] = $%%02X", oper1, formatAddr(addr), pointee)`)
		case "izy":
			g.printf(`// read 16 bytes from the zero page, handling page wrap`)
			g.printf(`lo := cpu.Bus.Peek8(uint16(oper1))`)
			g.printf(`hi := cpu.Bus.Peek8(uint16(uint8(oper1) + 1))`)
			g.printf(`addr := uint16(hi)<<8 | uint16(lo)`)
			g.printf(`addr += uint16(cpu.Y)`)
			g.printf(`pointee := cpu.Bus.Peek8(addr)`)
			g.printf(`oper = fmt.Sprintf("($%%02X),Y [%%s] = $%%02X", oper1, formatAddr(addr), pointee)`)
		}

		g.printf(``)
		g.printf(`return DisasmOp{`)
		g.printf(`	PC: pc,`)
		g.printf(`	Opcode: opcodeNames[oper0],`)
		g.printf(`	Buf: []byte{%s},`, strings.Join(bytes, ","))
		g.printf(`	Oper: oper,`)
		g.printf(`}`)
		g.printf(`}`)
		g.printf(``)
	}
}

func (g *Generator) opcodesTable() {
	bb := &strings.Builder{}
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			opcode := i*16 + j
			if defs[opcode].dontgen {
				fmt.Fprintf(bb, "%s,", defs[opcode].n)
			} else {
				fmt.Fprintf(bb, "opcode%02X, ", opcode)
			}
		}
		bb.WriteByte('\n')
	}
	g.printf(`// nes 6502 opcodes table`)
	g.printf(`var ops = [256]func(*CPU){`)
	g.printf(bb.String())
	g.printf(`}`)
	g.printf(``)
}

func (g *Generator) disasmTable() {
	bb := &strings.Builder{}
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			name := defs[i*16+j].m
			fmt.Fprintf(bb, "disasm%s, ", strings.ToUpper(name[:1])+name[1:])
		}
		bb.WriteByte('\n')
	}
	g.printf(`// nes 6502 opcodes disassembly table`)
	g.printf(`var disasmOps = [256]func(*CPU, uint16) DisasmOp {`)
	g.printf(bb.String())
	g.printf(`}`)
	g.printf(``)
}

func (g *Generator) opcodeNamesTable() {
	var names [256]string
	for i, def := range defs {
		names[i] = strconv.Quote(def.n)
	}
	g.printf(`var opcodeNames = [256]string{`)
	for i := 0; i < 16; i++ {
		g.printf("%s,", strings.Join(names[i*16:i*16+16], ", "))
	}
	g.printf(`}`)
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
	outf := flag.String("out", "opcodes.go", "output file")
	flag.Parse()

	var w io.Writer = os.Stdout

	bb := &bytes.Buffer{}
	if *outf != "stdout" {
		w = bb
	}

	g := &Generator{Writer: w}

	g.header()
	g.opcodes()
	g.unstableOpcodes()
	g.opcodesTable()
	g.disasmAddrModes()
	g.disasmTable()
	g.opcodeNamesTable()

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
