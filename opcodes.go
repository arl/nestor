package main

import "fmt"

var opCodes = [255]func(cpu *CPU){
	0x20: JSR,
	0x4C: JMPabs,
	0x6C: JMPind,
	0x78: SEI,
	0x8D: STAabs,
	0x8E: STXabs,
	0x84: STYzer,
	0x86: STXzer,
	0x9A: TXS,
	0xA0: LDYimm,
	0xA2: LDXimm,
	0xA9: LDAimm,
	0xD8: CLD,
	0xE8: INX,
}

var disasmCodes = [255]func(cpu *CPU) string{
	0x20: JSRDisasm,
	0x4C: JMPabsDisasm,
	0x6C: JMPindDisasm,
	0x78: SEIDisasm,
	0x8D: STAabsDisasm,
	0x8E: STXabsDisasm,
	0x84: STYzerDisasm,
	0x86: STXzerDisasm,
	0x9A: TXSDisasm,
	0xA0: LDYimmDisasm,
	0xA2: LDXimmDisasm,
	0xA9: LDAimmDisasm,
	0xD8: CLDDisasm,
	0xE8: INXDisasm,
}

// 20
func JSR(cpu *CPU) {
	// Get jump address
	oper := cpu.Read16(uint32(cpu.PC + 1))

	// Push return address the stack
	ret := cpu.PC + 3
	actualSP := uint32(cpu.SP) + 0x0100
	cpu.Write16(actualSP, ret)

	cpu.PC = oper
	cpu.Clock += 6
}

func JSRDisasm(cpu *CPU) string {
	oper := cpu.Read16(uint32(cpu.PC + 1))
	return fmt.Sprintf("JSR $%04X", oper)
}

// 78
func SEI(cpu *CPU) {
	cpu.P |= 1 << pbitI
	cpu.Clock += 2
	cpu.PC += 1
}

func SEIDisasm(cpu *CPU) string {
	return "SEI"
}

// 4C
func JMPabs(cpu *CPU) {
	oper := cpu.Read16(uint32(cpu.PC + 1))
	cpu.PC = oper
	cpu.Clock += 3
}

func JMPabsDisasm(cpu *CPU) string {
	oper := cpu.Read16(uint32(cpu.PC + 1))
	return fmt.Sprintf("JMP $%04X", oper)
}

// 6C
func JMPind(cpu *CPU) {
	oper := cpu.Read16(uint32(cpu.PC + 1))
	dst := cpu.Read16(uint32(oper))
	cpu.PC = dst
	cpu.Clock += 5
}

func JMPindDisasm(cpu *CPU) string {
	oper := cpu.Read16(uint32(cpu.PC + 1))
	return fmt.Sprintf("JMP ($%04X)", oper)
}

// 8D
func STAabs(cpu *CPU) {
	oper := cpu.Read16(uint32(cpu.PC + 1))
	cpu.bus.Write8(uint32(oper), cpu.A)
	cpu.PC += 3
	cpu.Clock += 5
}

func STAabsDisasm(cpu *CPU) string {
	oper := cpu.Read16(uint32(cpu.PC + 1))
	return fmt.Sprintf("STA $%04X", oper)
}

// 8E
func STXabs(cpu *CPU) {
	oper := cpu.Read16(uint32(cpu.PC + 1))
	cpu.bus.Write8(uint32(oper), cpu.X)
	cpu.PC += 3
	cpu.Clock += 4
}

func STXabsDisasm(cpu *CPU) string {
	oper := cpu.Read16(uint32(cpu.PC + 1))
	return fmt.Sprintf("STX $%04X", oper)
}

// 84
func STYzer(cpu *CPU) {
	oper := cpu.Read8(uint32(cpu.PC + 1))
	cpu.bus.Write8(uint32(oper), cpu.Y)
	cpu.PC += 2
	cpu.Clock += 3
}

func STYzerDisasm(cpu *CPU) string {
	oper := cpu.Read8(uint32(cpu.PC + 1))
	return fmt.Sprintf("STY %02X", oper)
}

// 86
func STXzer(cpu *CPU) {
	oper := cpu.Read8(uint32(cpu.PC + 1))
	cpu.bus.Write8(uint32(oper), cpu.X)
	cpu.PC += 2
	cpu.Clock += 3
}

func STXzerDisasm(cpu *CPU) string {
	oper := cpu.Read8(uint32(cpu.PC + 1))
	return fmt.Sprintf("STX %02X", oper)
}

// 9A
func TXS(cpu *CPU) {
	cpu.SP = cpu.X
	cpu.PC += 1
	cpu.Clock += 2
}

func TXSDisasm(cpu *CPU) string {
	return "TXS"
}

// A0
func LDYimm(cpu *CPU) {
	cpu.Y = cpu.Read8(uint32(cpu.PC + 1))
	cpu.P.maybeSetN(cpu.Y)
	cpu.P.maybeSetZ(cpu.Y)
	cpu.PC += 2
	cpu.Clock += 2
}

func LDYimmDisasm(cpu *CPU) string {
	oper := cpu.Read8(uint32(cpu.PC + 1))
	return fmt.Sprintf("LDY #$%02X", oper)
}

// A2
func LDXimm(cpu *CPU) {
	cpu.X = cpu.Read8(uint32(cpu.PC + 1))
	cpu.P.maybeSetN(cpu.X)
	cpu.P.maybeSetZ(cpu.X)
	cpu.PC += 2
	cpu.Clock += 2
}

func LDXimmDisasm(cpu *CPU) string {
	oper := cpu.Read8(uint32(cpu.PC + 1))
	return fmt.Sprintf("LDX #$%02X", oper)
}

// A9
func LDAimm(cpu *CPU) {
	cpu.A = cpu.Read8(uint32(cpu.PC + 1))
	cpu.P.maybeSetN(cpu.A)
	cpu.P.maybeSetZ(cpu.A)
	cpu.PC += 2
	cpu.Clock += 2
}

func LDAimmDisasm(cpu *CPU) string {
	oper := cpu.Read8(uint32(cpu.PC + 1))
	return fmt.Sprintf("LDA #$%02X", oper)
}

// D8
func CLD(cpu *CPU) {
	cpu.P &= ^P(1 << pbitD)
	cpu.Clock += 2
	cpu.PC += 1
}

func CLDDisasm(cpu *CPU) string {
	return "CLD"
}

// E8
func INX(cpu *CPU) {
	cpu.X++
	cpu.P.maybeSetN(cpu.X)
	cpu.P.maybeSetZ(cpu.X)
	cpu.Clock += 2
	cpu.PC += 1
}

func INXDisasm(cpu *CPU) string {
	return "INX"
}
