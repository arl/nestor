package main

import "fmt"

var opCodes = [256]func(cpu *CPU){
	0x08: PHP,
	0x20: JSR,
	0x24: BITzer,
	0x30: BMI,
	0x38: SEC,
	0x45: EORzer,
	0x4C: JMPabs,
	0x48: PHA,
	0x6C: JMPind,
	0x78: SEI,
	0x8D: STAabs,
	0x8E: STXabs,
	0x84: STYzer,
	0x86: STXzer,
	0x91: STAindy,
	0x9A: TXS,
	0xA0: LDYimm,
	0xA2: LDXimm,
	0xA9: LDAimm,
	0xAD: LDAabs,
	0xB0: BCS,
	0xC8: INY,
	0xC9: CMPimm,
	0xD0: BNE,
	0xD8: CLD,
	0xE6: INCzer,
	0xE8: INX,
	0xF8: SED,
}

var disasmCodes = [256]func(cpu *CPU) string{
	0x08: opcodestr("PHP"),
	0x20: JSRDisasm,
	0x24: BITzerDisasm,
	0x30: BMIDisasm,
	0x38: opcodestr("SEC"),
	0x45: EORzerDisasm,
	0x4C: JMPabsDisasm,
	0x48: opcodestr("PHA"),
	0x6C: JMPindDisasm,
	0x78: opcodestr("SEI"),
	0x8D: STAabsDisasm,
	0x8E: STXabsDisasm,
	0x84: STYzerDisasm,
	0x86: STXzerDisasm,
	0x91: STAindyDisasm,
	0x9A: opcodestr("TXS"),
	0xA0: LDYimmDisasm,
	0xA2: LDXimmDisasm,
	0xA9: LDAimmDisasm,
	0xAD: LDAabsDisasm,
	0xB0: BCSDisasm,
	0xC8: opcodestr("INY"),
	0xC9: CMPimmDisasm,
	0xD0: BNEDisasm,
	0xD8: opcodestr("CLD"),
	0xE6: INCzerDisasm,
	0xE8: opcodestr("INX"),
	0xF8: opcodestr("SED"),
}

// 08
func PHP(cpu *CPU) {
	p := cpu.P
	p |= (1 << pbitB) | (1 << pbitU)
	push8(cpu, uint8(p))
	cpu.Clock += 3
	cpu.PC += 1
}

// 20
func JSR(cpu *CPU) {
	// Get jump address
	oper := cpu.Read16(cpu.PC + 1)
	// Push return address on the stack
	ret := cpu.PC + 3
	push16(cpu, ret)
	cpu.PC = oper
	cpu.Clock += 6
}

func JSRDisasm(cpu *CPU) string {
	oper := cpu.Read16(cpu.PC + 1)
	return fmt.Sprintf("JSR $%04X", oper)
}

// 24
func BITzer(cpu *CPU) {
	oper := cpu.Read8(cpu.PC + 1)
	val := cpu.Read8(uint16(oper))

	// Copy bits 7 and 6 (N and V)
	cpu.P &= 0b00111111
	cpu.P |= P(val & 0b11000000)

	cpu.P.checkZ(cpu.A & val)

	cpu.PC += 2
	cpu.Clock += 3
}

func BITzerDisasm(cpu *CPU) string {
	oper := cpu.Read8(cpu.PC + 1)
	return fmt.Sprintf("BIT %02X", oper)
}

// 30
func BMI(cpu *CPU) {
	if !cpu.P.N() {
		cpu.Clock += 2
		cpu.PC += 2
		return
	}

	// Branch
	off := int32(cpu.Read8(cpu.PC + 1))
	addr := uint16(int32(cpu.PC+2) + off)
	if pagecrossed(cpu.PC, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
}

func BMIDisasm(cpu *CPU) string {
	off := int32(cpu.Read8(cpu.PC + 1))
	addr := uint16(int32(cpu.PC+2) + off)
	return fmt.Sprintf("BMI $%04X", addr)
}

// 38
func SEC(cpu *CPU) {
	cpu.P |= 1 << pbitC
	cpu.Clock += 2
	cpu.PC += 1
}

// 45
func EORzer(cpu *CPU) {
	oper := cpu.Read8(cpu.PC + 1)
	val := cpu.Read8(uint16(oper))
	cpu.A ^= val

	cpu.P.checkN(cpu.A)
	cpu.P.checkZ(cpu.A)

	cpu.Clock += 3
	cpu.PC += 2
}

func EORzerDisasm(cpu *CPU) string {
	oper := cpu.Read8(cpu.PC + 1)
	return fmt.Sprintf("EOR $%02X", oper)
}

// 78
func SEI(cpu *CPU) {
	cpu.P |= 1 << pbitI
	cpu.Clock += 2
	cpu.PC += 1
}

// 4C
func JMPabs(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	cpu.PC = oper
	cpu.Clock += 3
}

func JMPabsDisasm(cpu *CPU) string {
	oper := cpu.Read16(cpu.PC + 1)
	return fmt.Sprintf("JMP $%04X", oper)
}

// 48
func PHA(cpu *CPU) {
	push8(cpu, cpu.A)

	cpu.Clock += 3
	cpu.PC += 1
}

// 6C
func JMPind(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	dst := cpu.Read16(oper)
	cpu.PC = dst
	cpu.Clock += 5
}

func JMPindDisasm(cpu *CPU) string {
	oper := cpu.Read16(cpu.PC + 1)
	return fmt.Sprintf("JMP ($%04X)", oper)
}

// 8D
func STAabs(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	cpu.bus.Write8(oper, cpu.A)
	cpu.PC += 3
	cpu.Clock += 5
}

func STAabsDisasm(cpu *CPU) string {
	oper := cpu.Read16(cpu.PC + 1)
	return fmt.Sprintf("STA $%04X", oper)
}

// 8E
func STXabs(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	cpu.bus.Write8(oper, cpu.X)
	cpu.PC += 3
	cpu.Clock += 4
}

func STXabsDisasm(cpu *CPU) string {
	oper := cpu.Read16(cpu.PC + 1)
	return fmt.Sprintf("STX $%04X", oper)
}

// 84
func STYzer(cpu *CPU) {
	oper := cpu.Read8(cpu.PC + 1)
	cpu.bus.Write8(uint16(oper), cpu.Y)
	cpu.PC += 2
	cpu.Clock += 3
}

func STYzerDisasm(cpu *CPU) string {
	oper := cpu.Read8(cpu.PC + 1)
	return fmt.Sprintf("STY %02X", oper)
}

// 86
func STXzer(cpu *CPU) {
	oper := cpu.Read8(cpu.PC + 1)
	cpu.bus.Write8(uint16(oper), cpu.X)
	cpu.PC += 2
	cpu.Clock += 3
}

func STXzerDisasm(cpu *CPU) string {
	oper := cpu.Read8(cpu.PC + 1)
	return fmt.Sprintf("STX %02X", oper)
}

// 91
func STAindy(cpu *CPU) {
	// Read from the zero page
	oper := cpu.Read8(cpu.PC + 1)
	addr := uint16(oper)
	addr += uint16(cpu.Y)
	cpu.bus.Write8(addr, cpu.A)
	cpu.PC += 2
	cpu.Clock += 6
}

func STAindyDisasm(cpu *CPU) string {
	oper := cpu.Read8(cpu.PC + 1)
	return fmt.Sprintf("STA ($%02X),Y", oper)
}

// 9A
func TXS(cpu *CPU) {
	cpu.SP = cpu.X
	cpu.PC += 1
	cpu.Clock += 2
}

// A0
func LDYimm(cpu *CPU) {
	cpu.Y = cpu.Read8(cpu.PC + 1)
	cpu.P.checkN(cpu.Y)
	cpu.P.checkZ(cpu.Y)
	cpu.PC += 2
	cpu.Clock += 2
}

func LDYimmDisasm(cpu *CPU) string {
	oper := cpu.Read8(cpu.PC + 1)
	return fmt.Sprintf("LDY #$%02X", oper)
}

// A2
func LDXimm(cpu *CPU) {
	cpu.X = cpu.Read8(cpu.PC + 1)
	cpu.P.checkN(cpu.X)
	cpu.P.checkZ(cpu.X)
	cpu.PC += 2
	cpu.Clock += 2
}

func LDXimmDisasm(cpu *CPU) string {
	oper := cpu.Read8(cpu.PC + 1)
	return fmt.Sprintf("LDX #$%02X", oper)
}

// A9
func LDAimm(cpu *CPU) {
	cpu.A = cpu.Read8(cpu.PC + 1)
	cpu.P.checkN(cpu.A)
	cpu.P.checkZ(cpu.A)
	cpu.PC += 2
	cpu.Clock += 2
}

func LDAimmDisasm(cpu *CPU) string {
	oper := cpu.Read8(cpu.PC + 1)
	return fmt.Sprintf("LDA #$%02X", oper)
}

// AD
func LDAabs(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	cpu.A = cpu.Read8(oper)
	cpu.P.checkN(cpu.A)
	cpu.P.checkZ(cpu.A)
	cpu.PC += 3
	cpu.Clock += 4
}

func LDAabsDisasm(cpu *CPU) string {
	oper := cpu.Read16(cpu.PC + 1)
	return fmt.Sprintf("LDA $%04X", oper)
}

// B0
func BCS(cpu *CPU) {
	if !cpu.P.C() {
		cpu.Clock += 2
		cpu.PC += 2
		return
	}

	// Branch
	off := int32(cpu.Read8(cpu.PC + 1))
	addr := uint16(int32(cpu.PC+2) + off)
	if pagecrossed(cpu.PC, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
}

func BCSDisasm(cpu *CPU) string {
	off := int32(cpu.Read8(cpu.PC + 1))
	addr := uint16(int32(cpu.PC+2) + off)
	return fmt.Sprintf("BCS $%04X", addr)
}

// C8
func INY(cpu *CPU) {
	cpu.Y++
	cpu.P.checkN(cpu.Y)
	cpu.P.checkZ(cpu.Y)
	cpu.Clock += 2
	cpu.PC += 1
}

// C9
func CMPimm(cpu *CPU) {
	oper := cpu.Read8(cpu.PC + 1)
	res := cpu.A - oper
	cpu.P.checkN(res)
	cpu.P.checkZ(res)
	cpu.P.writeBit(pbitC, oper <= cpu.A)
	cpu.PC += 2
	cpu.Clock += 2
}

func CMPimmDisasm(cpu *CPU) string {
	oper := cpu.Read8(cpu.PC + 1)
	return fmt.Sprintf("CMP #$%02X", oper)
}

// D0
func BNE(cpu *CPU) {
	if cpu.P.Z() {
		cpu.Clock += 2
		cpu.PC += 2
		return
	}

	// Branch
	off := int32(cpu.Read8(cpu.PC + 1))
	addr := uint16(int32(cpu.PC+2) + off)
	if pagecrossed(cpu.PC, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
}

func BNEDisasm(cpu *CPU) string {
	off := int32(cpu.Read8(cpu.PC + 1))
	addr := uint16(int32(cpu.PC+2) + off)
	return fmt.Sprintf("BNE $%04X", addr)
}

// D8
func CLD(cpu *CPU) {
	cpu.P &= ^P(1 << pbitD)
	cpu.Clock += 2
	cpu.PC += 1
}

// E6
func INCzer(cpu *CPU) {
	oper := cpu.Read8(cpu.PC + 1)
	val := cpu.Read8(uint16(oper))
	val++
	cpu.P.checkN(val)
	cpu.P.checkZ(val)
	cpu.Clock += 5
	cpu.PC += 2
}

func INCzerDisasm(cpu *CPU) string {
	oper := cpu.Read8(cpu.PC + 1)
	return fmt.Sprintf("INC $%02X", oper)
}

// E8
func INX(cpu *CPU) {
	cpu.X++
	cpu.P.checkN(cpu.X)
	cpu.P.checkZ(cpu.X)
	cpu.Clock += 2
	cpu.PC += 1
}

// F8
func SED(cpu *CPU) {
	cpu.P |= 1 << pbitD
	cpu.Clock += 2
	cpu.PC += 1
}

/* helpers */

func pagecrossed(a, b uint16) bool {
	return 0xFF00&a != 0xFF00&b
}

func opcodestr(opname string) func(*CPU) string {
	return func(_ *CPU) string { return opname }
}

// push 8-bit onto the stack
func push8(cpu *CPU, val uint8) {
	top := uint16(cpu.SP) + 0x0100
	cpu.Write8(top, val)
	cpu.SP -= 1
}

// push a 16-bit value onto the stack
func push16(cpu *CPU, val uint16) {
	top := uint16(cpu.SP) + 0x0100
	cpu.Write16(top, val)
	cpu.SP -= 2
}
