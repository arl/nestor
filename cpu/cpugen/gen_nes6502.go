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
	"strings"
)

type opdef struct {
	i uint8  // opcode value (same as index into 'defs')
	n string // name
	m string // addressing mode
	f func(g *Generator, def opdef)

	// opcode detail string
	// 	- r: declare 'val' and set it to 'oper/accumulator'
	// 	- w: write 'val' back into 'oper/accumulator'
	// 	- x: extra cycle for page crosses
	// 	- a: extra cycle always
	// 	- i: unofficial (illegal) opcode
	d string
}

var defs = [256]opdef{
	{i: 0x00, n: "BRK", d: "      ", m: "imp"},
	{i: 0x01, n: "ORA", d: "r     ", m: "izx"},
	{i: 0x02, n: "JAM", d: "     i", m: "imm"},
	{i: 0x03, n: "SLO", d: "rw   i", m: "izx"},
	{i: 0x04, n: "NOP", d: "     i", m: "zpg"},
	{i: 0x05, n: "ORA", d: "r     ", m: "zpg"},
	{i: 0x06, n: "ASL", d: "rw    ", m: "zpg"},
	{i: 0x07, n: "SLO", d: "rw   i", m: "zpg"},
	{i: 0x08, n: "PHP", d: "      ", m: "imp"},
	{i: 0x09, n: "ORA", d: "r     ", m: "imm"},
	{i: 0x0A, n: "ASL", d: "rw    ", m: "acc"},
	{i: 0x0B, n: "ANC", d: "r    i", m: "imm"},
	{i: 0x0C, n: "NOP", d: "     i", m: "abs"},
	{i: 0x0D, n: "ORA", d: "r     ", m: "abs"},
	{i: 0x0E, n: "ASL", d: "rw    ", m: "abs"},
	{i: 0x0F, n: "SLO", d: "rw   i", m: "abs"},
	{i: 0x10, n: "BPL", d: "      ", m: "rel", f: branch(7, false)},
	{i: 0x11, n: "ORA", d: "r  x  ", m: "izy"},
	{i: 0x12, n: "JAM", d: "     i", m: "imm"},
	{i: 0x13, n: "SLO", d: "rw  ai", m: "izy"},
	{i: 0x14, n: "NOP", d: "     i", m: "zpx"},
	{i: 0x15, n: "ORA", d: "r     ", m: "zpx"},
	{i: 0x16, n: "ASL", d: "rw    ", m: "zpx"},
	{i: 0x17, n: "SLO", d: "rw   i", m: "zpx"},
	{i: 0x18, n: "CLC", d: "      ", m: "imp", f: clear(0)},
	{i: 0x19, n: "ORA", d: "r  x  ", m: "aby"},
	{i: 0x1A, n: "NOP", d: "     i", m: "imp"},
	{i: 0x1B, n: "SLO", d: "rw   i", m: "aby"},
	{i: 0x1C, n: "NOP", d: "   x i", m: "abx"},
	{i: 0x1D, n: "ORA", d: "r  x  ", m: "abx"},
	{i: 0x1E, n: "ASL", d: "rw    ", m: "abx"},
	{i: 0x1F, n: "SLO", d: "rw   i", m: "abx"},
	{i: 0x20, n: "JSR", d: "      ", m: "imp"}, // special case. should be 'abs' but handle it as 'implied'
	{i: 0x21, n: "AND", d: "r     ", m: "izx"},
	{i: 0x22, n: "JAM", d: "     i", m: "imm"},
	{i: 0x23, n: "RLA", d: "rw   i", m: "izx"},
	{i: 0x24, n: "BIT", d: "r     ", m: "zpg"},
	{i: 0x25, n: "AND", d: "r     ", m: "zpg"},
	{i: 0x26, n: "ROL", d: "rw    ", m: "zpg"},
	{i: 0x27, n: "RLA", d: "rw   i", m: "zpg"},
	{i: 0x28, n: "PLP", d: "      ", m: "imp"},
	{i: 0x29, n: "AND", d: "r     ", m: "imm"},
	{i: 0x2A, n: "ROL", d: "rw    ", m: "acc"},
	{i: 0x2B, n: "ANC", d: "r    i", m: "imm"},
	{i: 0x2C, n: "BIT", d: "r     ", m: "abs"},
	{i: 0x2D, n: "AND", d: "r     ", m: "abs"},
	{i: 0x2E, n: "ROL", d: "rw    ", m: "abs"},
	{i: 0x2F, n: "RLA", d: "rw   i", m: "abs"},
	{i: 0x30, n: "BMI", d: "      ", m: "rel", f: branch(7, true)},
	{i: 0x31, n: "AND", d: "r  x  ", m: "izy"},
	{i: 0x32, n: "JAM", d: "     i", m: "imm"},
	{i: 0x33, n: "RLA", d: "rw  ai", m: "izy"},
	{i: 0x34, n: "NOP", d: "     i", m: "zpx"},
	{i: 0x35, n: "AND", d: "r     ", m: "zpx"},
	{i: 0x36, n: "ROL", d: "rw    ", m: "zpx"},
	{i: 0x37, n: "RLA", d: "rw   i", m: "zpx"},
	{i: 0x38, n: "SEC", d: "      ", m: "imp", f: set(0)},
	{i: 0x39, n: "AND", d: "r  x  ", m: "aby"},
	{i: 0x3A, n: "NOP", d: "     i", m: "imp"},
	{i: 0x3B, n: "RLA", d: "rw   i", m: "aby"},
	{i: 0x3C, n: "NOP", d: "   x i", m: "abx"},
	{i: 0x3D, n: "AND", d: "r  x  ", m: "abx"},
	{i: 0x3E, n: "ROL", d: "rw    ", m: "abx"},
	{i: 0x3F, n: "RLA", d: "rw   i", m: "abx"},
	{i: 0x40, n: "RTI", d: "      ", m: "imp"},
	{i: 0x41, n: "EOR", d: "r     ", m: "izx"},
	{i: 0x42, n: "JAM", d: "     i", m: "imm"},
	{i: 0x43, n: "SRE", d: "rw   i", m: "izx"},
	{i: 0x44, n: "NOP", d: "     i", m: "zpg"},
	{i: 0x45, n: "EOR", d: "r     ", m: "zpg"},
	{i: 0x46, n: "LSR", d: "rw    ", m: "zpg"},
	{i: 0x47, n: "SRE", d: "rw   i", m: "zpg"},
	{i: 0x48, n: "PHA", d: "      ", m: "imp"},
	{i: 0x49, n: "EOR", d: "r     ", m: "imm"},
	{i: 0x4A, n: "LSR", d: "rw    ", m: "acc"},
	{i: 0x4B, n: "ALR", d: "r    i", m: "imm"},
	{i: 0x4C, n: "JMP", d: "      ", m: "abs"},
	{i: 0x4D, n: "EOR", d: "r     ", m: "abs"},
	{i: 0x4E, n: "LSR", d: "rw    ", m: "abs"},
	{i: 0x4F, n: "SRE", d: "rw   i", m: "abs"},
	{i: 0x50, n: "BVC", d: "      ", m: "rel", f: branch(6, false)},
	{i: 0x51, n: "EOR", d: "r  x  ", m: "izy"},
	{i: 0x52, n: "JAM", d: "     i", m: "imm"},
	{i: 0x53, n: "SRE", d: "rw  ai", m: "izy"},
	{i: 0x54, n: "NOP", d: "     i", m: "zpx"},
	{i: 0x55, n: "EOR", d: "r     ", m: "zpx"},
	{i: 0x56, n: "LSR", d: "rw    ", m: "zpx"},
	{i: 0x57, n: "SRE", d: "rw   i", m: "zpx"},
	{i: 0x58, n: "CLI", d: "      ", m: "imp", f: clear(2)},
	{i: 0x59, n: "EOR", d: "r  x  ", m: "aby"},
	{i: 0x5A, n: "NOP", d: "     i", m: "imp"},
	{i: 0x5B, n: "SRE", d: "rw   i", m: "aby"},
	{i: 0x5C, n: "NOP", d: "   x i", m: "abx"},
	{i: 0x5D, n: "EOR", d: "r  x  ", m: "abx"},
	{i: 0x5E, n: "LSR", d: "rw    ", m: "abx"},
	{i: 0x5F, n: "SRE", d: "rw   i", m: "abx"},
	{i: 0x60, n: "RTS", d: "      ", m: "imp"},
	{i: 0x61, n: "ADC", d: "r     ", m: "izx"},
	{i: 0x62, n: "JAM", d: "     i", m: "imm"},
	{i: 0x63, n: "RRA", d: "rw   i", m: "izx"},
	{i: 0x64, n: "NOP", d: "     i", m: "zpg"},
	{i: 0x65, n: "ADC", d: "r     ", m: "zpg"},
	{i: 0x66, n: "ROR", d: "rw    ", m: "zpg"},
	{i: 0x67, n: "RRA", d: "rw   i", m: "zpg"},
	{i: 0x68, n: "PLA", d: "      ", m: "imp"},
	{i: 0x69, n: "ADC", d: "r     ", m: "imm"},
	{i: 0x6A, n: "ROR", d: "rw    ", m: "acc"},
	{i: 0x6B, n: "ARR", d: "r    i", m: "imm"},
	{i: 0x6C, n: "JMP", d: "      ", m: "ind"},
	{i: 0x6D, n: "ADC", d: "r     ", m: "abs"},
	{i: 0x6E, n: "ROR", d: "rw    ", m: "abs"},
	{i: 0x6F, n: "RRA", d: "rw   i", m: "abs"},
	{i: 0x70, n: "BVS", d: "      ", m: "rel", f: branch(6, true)},
	{i: 0x71, n: "ADC", d: "r  x  ", m: "izy"},
	{i: 0x72, n: "JAM", d: "     i", m: "imm"},
	{i: 0x73, n: "RRA", d: "rw  ai", m: "izy"},
	{i: 0x74, n: "NOP", d: "     i", m: "zpx"},
	{i: 0x75, n: "ADC", d: "r     ", m: "zpx"},
	{i: 0x76, n: "ROR", d: "rw    ", m: "zpx"},
	{i: 0x77, n: "RRA", d: "rw   i", m: "zpx"},
	{i: 0x78, n: "SEI", d: "      ", m: "imp", f: set(2)},
	{i: 0x79, n: "ADC", d: "r  x  ", m: "aby"},
	{i: 0x7A, n: "NOP", d: "     i", m: "imp"},
	{i: 0x7B, n: "RRA", d: "rw   i", m: "aby"},
	{i: 0x7C, n: "NOP", d: "   x i", m: "abx"},
	{i: 0x7D, n: "ADC", d: "r  x  ", m: "abx"},
	{i: 0x7E, n: "ROR", d: "rw    ", m: "abx"},
	{i: 0x7F, n: "RRA", d: "rw   i", m: "abx"},
	{i: 0x80, n: "NOP", d: "     i", m: "imm"},
	{i: 0x81, n: "STA", d: "      ", m: "izx", f: ST("A")},
	{i: 0x82, n: "NOP", d: "     i", m: "imm"},
	{i: 0x83, n: "SAX", d: "     i", m: "izx"},
	{i: 0x84, n: "STY", d: "      ", m: "zpg", f: ST("Y")},
	{i: 0x85, n: "STA", d: "      ", m: "zpg", f: ST("A")},
	{i: 0x86, n: "STX", d: "      ", m: "zpg", f: ST("X")},
	{i: 0x87, n: "SAX", d: "     i", m: "zpg"},
	{i: 0x88, n: "DEY", d: "      ", m: "imp", f: dec("Y")},
	{i: 0x89, n: "NOP", d: "     i", m: "imm"},
	{i: 0x8A, n: "TXA", d: "      ", m: "imp", f: T("X", "A")},
	{i: 0x8B, n: "ANE", d: "     i", m: "imm", f: unstable},
	{i: 0x8C, n: "STY", d: "      ", m: "abs", f: ST("Y")},
	{i: 0x8D, n: "STA", d: "      ", m: "abs", f: ST("A")},
	{i: 0x8E, n: "STX", d: "      ", m: "abs", f: ST("X")},
	{i: 0x8F, n: "SAX", d: "     i", m: "abs"},
	{i: 0x90, n: "BCC", d: "      ", m: "rel", f: branch(0, false)},
	{i: 0x91, n: "STA", d: "   _a_", m: "izy", f: ST("A")},
	{i: 0x92, n: "JAM", d: "     i", m: "imm"},
	{i: 0x93, n: "SHA", d: "     i", m: "izy", f: unstable},
	{i: 0x94, n: "STY", d: "      ", m: "zpx", f: ST("Y")},
	{i: 0x95, n: "STA", d: "      ", m: "zpx", f: ST("A")},
	{i: 0x96, n: "STX", d: "      ", m: "zpy", f: ST("X")},
	{i: 0x97, n: "SAX", d: "     i", m: "zpy"},
	{i: 0x98, n: "TYA", d: "      ", m: "imp", f: T("Y", "A")},
	{i: 0x99, n: "STA", d: "      ", m: "aby", f: ST("A")},
	{i: 0x9A, n: "TXS", d: "      ", m: "imp", f: T("X", "SP")},
	{i: 0x9B, n: "TAS", d: "     i", m: "aby", f: unstable},
	{i: 0x9C, n: "SHY", d: "     i", m: "abx", f: unstable},
	{i: 0x9D, n: "STA", d: "      ", m: "abx", f: ST("A")},
	{i: 0x9E, n: "SHX", d: "     i", m: "aby", f: unstable},
	{i: 0x9F, n: "SHA", d: "     i", m: "aby", f: unstable},
	{i: 0xA0, n: "LDY", d: "r     ", m: "imm", f: LD("Y")},
	{i: 0xA1, n: "LDA", d: "r     ", m: "izx", f: LD("A")},
	{i: 0xA2, n: "LDX", d: "r     ", m: "imm", f: LD("X")},
	{i: 0xA3, n: "LAX", d: "r    i", m: "izx", f: LD("A", "X")},
	{i: 0xA4, n: "LDY", d: "r     ", m: "zpg", f: LD("Y")},
	{i: 0xA5, n: "LDA", d: "r     ", m: "zpg", f: LD("A")},
	{i: 0xA6, n: "LDX", d: "r     ", m: "zpg", f: LD("X")},
	{i: 0xA7, n: "LAX", d: "r    i", m: "zpg", f: LD("A", "X")},
	{i: 0xA8, n: "TAY", d: "      ", m: "imp", f: T("A", "Y")},
	{i: 0xA9, n: "LDA", d: "r     ", m: "imm", f: LD("A")},
	{i: 0xAA, n: "TAX", d: "      ", m: "imp", f: T("A", "X")},
	{i: 0xAB, n: "LXA", d: "     i", m: "imm", f: unstable},
	{i: 0xAC, n: "LDY", d: "r     ", m: "abs", f: LD("Y")},
	{i: 0xAD, n: "LDA", d: "r     ", m: "abs", f: LD("A")},
	{i: 0xAE, n: "LDX", d: "r     ", m: "abs", f: LD("X")},
	{i: 0xAF, n: "LAX", d: "r    i", m: "abs", f: LD("A", "X")},
	{i: 0xB0, n: "BCS", d: "      ", m: "rel", f: branch(0, true)},
	{i: 0xB1, n: "LDA", d: "r  x  ", m: "izy", f: LD("A")},
	{i: 0xB2, n: "JAM", d: "     i", m: "imm"},
	{i: 0xB3, n: "LAX", d: "r  x i", m: "izy", f: LD("A", "X")},
	{i: 0xB4, n: "LDY", d: "r     ", m: "zpx", f: LD("Y")},
	{i: 0xB5, n: "LDA", d: "r     ", m: "zpx", f: LD("A")},
	{i: 0xB6, n: "LDX", d: "r     ", m: "zpy", f: LD("X")},
	{i: 0xB7, n: "LAX", d: "r    i", m: "zpy", f: LD("A", "X")},
	{i: 0xB8, n: "CLV", d: "      ", m: "imp", f: clear(6)},
	{i: 0xB9, n: "LDA", d: "r  x  ", m: "aby", f: LD("A")},
	{i: 0xBA, n: "TSX", d: "      ", m: "imp", f: T("SP", "X")},
	{i: 0xBB, n: "LAS", d: "r  x i", m: "aby"},
	{i: 0xBC, n: "LDY", d: "r  x  ", m: "abx", f: LD("Y")},
	{i: 0xBD, n: "LDA", d: "r  x  ", m: "abx", f: LD("A")},
	{i: 0xBE, n: "LDX", d: "r  x  ", m: "aby", f: LD("X")},
	{i: 0xBF, n: "LAX", d: "r  x i", m: "aby", f: LD("A", "X")},
	{i: 0xC0, n: "CPY", d: "r     ", m: "imm", f: cmp("Y")},
	{i: 0xC1, n: "CMP", d: "r     ", m: "izx", f: cmp("A")},
	{i: 0xC2, n: "NOP", d: "     i", m: "imm"},
	{i: 0xC3, n: "DCP", d: "rw   i", m: "izx"},
	{i: 0xC4, n: "CPY", d: "r     ", m: "zpg", f: cmp("Y")},
	{i: 0xC5, n: "CMP", d: "r     ", m: "zpg", f: cmp("A")},
	{i: 0xC6, n: "DEC", d: "rw    ", m: "zpg", f: dec("")},
	{i: 0xC7, n: "DCP", d: "rw   i", m: "zpg"},
	{i: 0xC8, n: "INY", d: "      ", m: "imp", f: inc("Y")},
	{i: 0xC9, n: "CMP", d: "r     ", m: "imm", f: cmp("A")},
	{i: 0xCA, n: "DEX", d: "      ", m: "imp", f: dec("X")},
	{i: 0xCB, n: "SBX", d: "r    i", m: "imm"},
	{i: 0xCC, n: "CPY", d: "r     ", m: "abs", f: cmp("Y")},
	{i: 0xCD, n: "CMP", d: "r     ", m: "abs", f: cmp("A")},
	{i: 0xCE, n: "DEC", d: "rw    ", m: "abs", f: dec("")},
	{i: 0xCF, n: "DCP", d: "rw   i", m: "abs"},
	{i: 0xD0, n: "BNE", d: "      ", m: "rel", f: branch(1, false)},
	{i: 0xD1, n: "CMP", d: "r  x  ", m: "izy", f: cmp("A")},
	{i: 0xD2, n: "JAM", d: "     i", m: "imm"},
	{i: 0xD3, n: "DCP", d: "rw  ai", m: "izy"},
	{i: 0xD4, n: "NOP", d: "     i", m: "zpx"},
	{i: 0xD5, n: "CMP", d: "r     ", m: "zpx", f: cmp("A")},
	{i: 0xD6, n: "DEC", d: "rw    ", m: "zpx", f: dec("")},
	{i: 0xD7, n: "DCP", d: "rw   i", m: "zpx"},
	{i: 0xD8, n: "CLD", d: "      ", m: "imp", f: clear(3)},
	{i: 0xD9, n: "CMP", d: "r  x  ", m: "aby", f: cmp("A")},
	{i: 0xDA, n: "NOP", d: "     i", m: "imp"},
	{i: 0xDB, n: "DCP", d: "rw   i", m: "aby"},
	{i: 0xDC, n: "NOP", d: "   x i", m: "abx"},
	{i: 0xDD, n: "CMP", d: "r  x  ", m: "abx", f: cmp("A")},
	{i: 0xDE, n: "DEC", d: "rw    ", m: "abx", f: dec("")},
	{i: 0xDF, n: "DCP", d: "rw   i", m: "abx"},
	{i: 0xE0, n: "CPX", d: "r     ", m: "imm", f: cmp("X")},
	{i: 0xE1, n: "SBC", d: "r     ", m: "izx"},
	{i: 0xE2, n: "NOP", d: "     i", m: "imm"},
	{i: 0xE3, n: "ISB", d: "rw   i", m: "izx"},
	{i: 0xE4, n: "CPX", d: "r     ", m: "zpg", f: cmp("X")},
	{i: 0xE5, n: "SBC", d: "r     ", m: "zpg"},
	{i: 0xE6, n: "INC", d: "rw    ", m: "zpg", f: inc("")},
	{i: 0xE7, n: "ISB", d: "rw   i", m: "zpg"},
	{i: 0xE8, n: "INX", d: "      ", m: "imp", f: inc("X")},
	{i: 0xE9, n: "SBC", d: "r     ", m: "imm"},
	{i: 0xEA, n: "NOP", d: "      ", m: "imp"},
	{i: 0xEB, n: "SBC", d: "r    i", m: "imm"},
	{i: 0xEC, n: "CPX", d: "r     ", m: "abs", f: cmp("X")},
	{i: 0xED, n: "SBC", d: "r     ", m: "abs"},
	{i: 0xEE, n: "INC", d: "rw    ", m: "abs", f: inc("")},
	{i: 0xEF, n: "ISB", d: "rw   i", m: "abs"},
	{i: 0xF0, n: "BEQ", d: "      ", m: "rel", f: branch(1, true)},
	{i: 0xF1, n: "SBC", d: "r  x  ", m: "izy"},
	{i: 0xF2, n: "JAM", d: "     i", m: "imm"},
	{i: 0xF3, n: "ISB", d: "rw  ai", m: "izy"},
	{i: 0xF4, n: "NOP", d: "     i", m: "zpx"},
	{i: 0xF5, n: "SBC", d: "r     ", m: "zpx"},
	{i: 0xF6, n: "INC", d: "rw    ", m: "zpx", f: inc("")},
	{i: 0xF7, n: "ISB", d: "rw   i", m: "zpx"},
	{i: 0xF8, n: "SED", d: "      ", m: "imp", f: set(3)},
	{i: 0xF9, n: "SBC", d: "r  x  ", m: "aby"},
	{i: 0xFA, n: "NOP", d: "     i", m: "imp"},
	{i: 0xFB, n: "ISB", d: "rw   i", m: "aby"},
	{i: 0xFC, n: "NOP", d: "   x i", m: "abx"},
	{i: 0xFD, n: "SBC", d: "r  x  ", m: "abx"},
	{i: 0xFE, n: "INC", d: "rw    ", m: "abx", f: inc("")},
	{i: 0xFF, n: "ISB", d: "rw   i", m: "abx"},
}

type addrmode struct {
	human string // human readable name
	f     func(g *Generator, details string)
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

//
// addressing modes
//

func ind(g *Generator, _ string) {
	g.printf(`oper := cpu.Read16(cpu.PC)`)
	g.printf(`lo := cpu.Read8(oper)`)
	g.printf(`// 2 bytes address wrap around`)
	g.printf(`hi := cpu.Read8((0xff00 & oper) | (0x00ff & (oper + 1)))`)
	g.printf(`oper = uint16(hi)<<8 | uint16(lo)`)
}

func imm(g *Generator, _ string) {
	g.printf(`oper := cpu.PC`)
	g.printf(`cpu.PC++`)
}

func rel(g *Generator, _ string) {
	g.printf(`off := int8(cpu.Read8(cpu.PC))`)
	g.printf(`oper := uint16(int16(cpu.PC+1) + int16(off))`)
}

func abs(g *Generator, _ string) {
	g.printf(`oper := cpu.Read16(cpu.PC)`)
	g.printf(`cpu.PC += 2`)
}

func abx(g *Generator, info string) {
	switch {
	case has(info, 'x'):
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

func aby(g *Generator, info string) {
	switch {
	case has(info, 'x'):
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

func zpg(g *Generator, _ string) {
	g.printf(`oper := uint16(cpu.Read8(cpu.PC))`)
	g.printf(`cpu.PC++`)
}

func zpx(g *Generator, _ string) {
	g.printf(`cpu.tick()`)
	g.printf(`addr := cpu.Read8(cpu.PC)`)
	g.printf(`cpu.PC++`)
	g.printf(`oper := uint16(addr) + uint16(cpu.X)`)
	g.printf(`oper &= 0xff`)
}

func zpy(g *Generator, _ string) {
	g.printf(`cpu.tick()`)
	g.printf(`addr := cpu.Read8(cpu.PC)`)
	g.printf(`cpu.PC++`)
	g.printf(`oper := uint16(addr) + uint16(cpu.Y)`)
	g.printf(`oper &= 0xff`)
}

func izx(g *Generator, info string) {
	g.printf(`cpu.tick()`)
	zpg(g, info)
	g.printf(`oper = uint16(uint8(oper) + cpu.X)`)
	r16zpwrap(g)
}

func izy(g *Generator, info string) {
	switch {
	case has(info, 'x'):
		g.printf(`// extra cycle for page cross`)
		zpg(g, info)
		r16zpwrap(g)
		tickIfPageCrossed(g, "oper", "oper+uint16(cpu.Y)")
		g.printf(`oper += uint16(cpu.Y)`)
	case has(info, 'a'):
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

//
// opcode generators
//

func (g *Generator) ADC(_ opdef) {
	g.printf(`carry := cpu.P.ibit(pbitC)`)
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
	g.printf(`cpu.P.writeBit(pbitC, carry != 0)`)
}

func (g *Generator) ANC(def opdef) {
	g.AND(def)
	g.printf(`cpu.P.writeBit(pbitC, cpu.P.N())`)
}

func (g *Generator) AND(_ opdef) {
	g.printf(`cpu.A &= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func (g *Generator) ARR(_ opdef) {
	g.printf(`cpu.A &= val`)
	g.printf(`cpu.A >>= 1`)
	g.printf(`cpu.P.writeBit(pbitV, (cpu.A>>6)^(cpu.A>>5)&0x01 != 0)`)
	g.printf(`if cpu.P.C() {`)
	g.printf(`	cpu.A |= 1 << 7`)
	g.printf(`}`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
	g.printf(`cpu.P.writeBit(pbitC, cpu.A&(1<<6) != 0)`)
}

func (g *Generator) ASL(_ opdef) {
	g.printf(`carry := val & 0x80`)
	g.printf(`val = (val << 1) & 0xfe`)
	g.printf(`cpu.tick()`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.writeBit(pbitC, carry != 0)`)
}

func (g *Generator) BIT(_ opdef) {
	g.printf(`cpu.P &= 0b00111111`)
	g.printf(`cpu.P |= P(val & 0b11000000)`)
	g.printf(`cpu.P.checkZ(cpu.A & val)`)
}

func (g *Generator) BRK(_ opdef) {
	g.printf(`cpu.tick()`)
	push16(g, `cpu.PC+1`)
	g.printf(`p := cpu.P`)
	g.printf(`p.setBit(pbitB)`)
	push8(g, `uint8(p)`)
	g.printf(`cpu.P.setBit(pbitI)`)
	g.printf(`cpu.PC = cpu.Read16(IRQvector)`)
}

func (g *Generator) DCP(def opdef) {
	dec("")(g, def)
	cmp("A")(g, def)
}

func (g *Generator) EOR(_ opdef) {
	g.printf(`cpu.A ^= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func (g *Generator) ISB(def opdef) {
	inc("")(g, def)
	g.printf(`final := val`)
	g.SBC(def)
	g.printf(`val = final`)
}

func (g *Generator) JAM(def opdef) {
	g.unstable = append(g.unstable, def.i)
	insertPanic(g, `Halt and catch fire!\nJAM called`)
}

func (g *Generator) JMP(_ opdef) {
	g.printf(`cpu.PC = oper`)
}

func (g *Generator) JSR(_ opdef) {
	g.printf(`oper := cpu.Read16(cpu.PC)`)
	g.printf(`cpu.tick()`)
	push16(g, `cpu.PC+1`)
	g.printf(`cpu.PC = oper`)
}

func (g *Generator) LAS(def opdef) {
	g.printf(`cpu.A = cpu.SP & val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
	g.printf(`cpu.X = cpu.A`)
	g.printf(`cpu.SP = cpu.A`)
}

func (g *Generator) LSR(_ opdef) {
	g.printf(`{`)
	g.printf(`carry := val & 0x01 // carry is bit 0`)
	g.printf(`val = (val >> 1)&0x7f`)
	g.printf(`cpu.tick()`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.writeBit(pbitC, carry != 0)`)
	g.printf(`}`)
}

func (g *Generator) NOP(_ opdef) {
	g.printf(`cpu.tick()`)
}

func (g *Generator) ORA(_ opdef) {
	g.printf(`cpu.A |= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func (g *Generator) PHA(_ opdef) {
	g.printf(`cpu.tick()`)
	push8(g, `cpu.A`)
}

func (g *Generator) PHP(_ opdef) {
	g.printf(`cpu.tick()`)
	g.printf(`p := cpu.P`)
	g.printf(`p |= (1 << pbitB) | (1 << pbitU)`)
	push8(g, `uint8(p)`)
}

func (g *Generator) PLA(_ opdef) {
	g.printf(`cpu.tick()`)
	g.printf(`cpu.tick()`)
	pull8(g, `cpu.A`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func (g *Generator) PLP(_ opdef) {
	g.printf(`cpu.tick()`)
	g.printf(`cpu.tick()`)
	g.printf(`var p uint8`)
	pull8(g, `p`)
	g.printf(`const mask uint8 = 0b11001111 // ignore B and U bits`)
	g.printf(`cpu.P = P(%s)`, copybits(`uint8(cpu.P)`, `p`, `mask`))
}

func (g *Generator) RLA(def opdef) {
	g.ROL(def)
	g.AND(def)
}

func (g *Generator) ROL(_ opdef) {
	g.printf(`carry := val & 0x80`)
	g.printf(`val <<= 1`)
	g.printf(`if cpu.P.C() {`)
	g.printf(`	val |= 1 << 0`)
	g.printf(`}`)
	g.printf(`cpu.tick()`)
	g.printf(`cpu.P.checkNZ(val)`)
	g.printf(`cpu.P.writeBit(pbitC, carry != 0)`)
}

func (g *Generator) ROR(_ opdef) {
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

func (g *Generator) RRA(def opdef) {
	g.ROR(def)
	g.ADC(def)
}

func (g *Generator) RTI(_ opdef) {
	g.printf(`cpu.tick()`)
	g.printf(`cpu.tick()`)
	g.printf(`var p uint8`)
	pull8(g, `p`)
	g.printf(`const mask uint8 = 0b11001111 // ignore B and U bits`)
	g.printf(`cpu.P = P(%s)`, copybits(`uint8(cpu.P)`, `p`, `mask`))
	pull16(g, `cpu.PC`)
}

func (g *Generator) RTS(_ opdef) {
	g.printf(`cpu.tick()`)
	g.printf(`cpu.tick()`)
	pull16(g, `cpu.PC`)
	g.printf(`cpu.PC++`)
	g.printf(`cpu.tick()`)
}

func (g *Generator) SAX(_ opdef) {
	g.printf(`cpu.Write8(oper, cpu.A&cpu.X)`)
}

func (g *Generator) SBC(def opdef) {
	g.printf(`val ^= 0xff`)
	g.printf(`carry := cpu.P.ibit(pbitC)`)
	g.printf(`sum := uint16(cpu.A) + uint16(val) + uint16(carry)`)
	g.printf(`cpu.P.checkCV(cpu.A, val, sum)`)
	g.printf(`cpu.A = uint8(sum)`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
}

func (g *Generator) SBX(def opdef) {
	g.printf(`ival := (int16(cpu.A) & int16(cpu.X)) - int16(val)`)
	g.printf(`cpu.X = uint8(ival)`)
	g.printf(`cpu.P.checkNZ(uint8(ival))`)
	g.printf(`cpu.P.writeBit(pbitC, ival >= 0)`)
}

func (g *Generator) SLO(def opdef) {
	g.ASL(def)
	g.printf(`cpu.A |= val`)
	g.printf(`cpu.P.checkNZ(cpu.A)`)
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
		g.printf(`cpu.P.writeBit(pbitC, val <= %s)`, v)
	}
}

func T(src, dst string) func(g *Generator, _ opdef) {
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
	case "mem", "":
		return `val`
	}
	panic("regOrMem " + v)
}

func inc(v string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.tick()`)
		v = regOrMem(v)
		g.printf(`%s++`, v)
		g.printf(`cpu.P.checkNZ(%s)`, v)
	}
}

func dec(v string) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.tick()`)
		v = regOrMem(v)
		g.printf(`%s--`, v)
		g.printf(`cpu.P.checkNZ(%s)`, v)
	}
}

func clear(ibit uint) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.P.clearBit(%d)`, ibit)
		g.printf(`cpu.tick()`)
	}
}

func set(ibit uint) func(g *Generator, _ opdef) {
	return func(g *Generator, _ opdef) {
		g.printf(`cpu.P.setBit(%d)`, ibit)
		g.printf(`cpu.tick()`)
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
	outbuf bytes.Buffer
	out    io.Writer

	unstable []uint8
}

func (g *Generator) header() {
	g.printf(`// Code generated by cpugen/gen_nes6502.go. DO NOT EDIT.`)
	g.printf(`package cpu`)
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
		g.printf(`_ = oper`)
	}

	switch {
	case strings.Contains(defs[code].d, "r"):
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
		g.opcodeHeader(def.i)
		if def.f != nil {
			def.f(g, def)
		} else {
			f := reflect.ValueOf(g).MethodByName(def.n)
			f.Call([]reflect.Value{reflect.ValueOf(def)})
		}

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
