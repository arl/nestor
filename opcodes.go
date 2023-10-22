package main

// TODO(arl): we can factor branch subfunctions

var ops = [256]func(cpu *CPU){
	0x08: PHP,
	0x10: BPL,
	0x18: CLC,
	0x20: JSR,
	0x24: BITzer,
	0x30: BMI,
	0x38: SEC,
	0x45: EORzer,
	0x4C: JMPabs,
	0x48: PHA,
	0x50: BVC,
	0x58: CLI,
	0x66: RORzer,
	0x6A: RORacc,
	0x6C: JMPind,
	0x70: BVS,
	0x78: SEI,
	0x8D: STAabs,
	0x8E: STXabs,
	0x84: STYzer,
	0x85: STAzer,
	0x86: STXzer,
	0x90: BCC,
	0x91: STAindy,
	0x9A: TXS,
	0xA0: LDYimm,
	0xA2: LDXimm,
	0xA9: LDAimm,
	0xAD: LDAabs,
	0xB0: BCS,
	0xB8: CLV,
	0xC8: INY,
	0xCA: DEX,
	0xC9: CMPimm,
	0xD0: BNE,
	0xD8: CLD,
	0xE6: INCzer,
	0xE8: INX,
	0xEA: NOP,
	0xF0: BEQ,
	0xF8: SED,
}

// 08
func PHP(cpu *CPU) {
	p := cpu.P
	p |= (1 << pbitB) | (1 << pbitU)
	push8(cpu, uint8(p))
	cpu.Clock += 3
	cpu.PC += 1
}

// 10
func BPL(cpu *CPU) {
	if cpu.P.N() {
		cpu.Clock += 2
		cpu.PC += 2
		return
	}

	// Branch
	off := cpu.Read8(cpu.PC + 1)
	reladdr := int32(cpu.PC+2) + int32(off)
	addr := uint16(reladdr)
	if pagecrossed(cpu.PC, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
}

// 18
func CLC(cpu *CPU) {
	cpu.P.clearBit(pbitC)
	cpu.Clock += 2
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

// 30
func BMI(cpu *CPU) {
	if !cpu.P.N() {
		cpu.Clock += 2
		cpu.PC += 2
		return
	}

	// Branch
	off := cpu.Read8(cpu.PC + 1)
	reladdr := int32(cpu.PC+2) + int32(off)
	addr := uint16(reladdr)
	if pagecrossed(cpu.PC, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
}

// 38
func SEC(cpu *CPU) {
	cpu.P.setBit(pbitC)
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

// 4C
func JMPabs(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	cpu.PC = oper
	cpu.Clock += 3
}

// 48
func PHA(cpu *CPU) {
	push8(cpu, cpu.A)

	cpu.Clock += 3
	cpu.PC += 1
}

// 50
func BVC(cpu *CPU) {
	if cpu.P.V() {
		cpu.Clock += 2
		cpu.PC += 2
		return
	}

	// Branch
	off := cpu.Read8(cpu.PC + 1)
	reladdr := int32(cpu.PC+2) + int32(off)
	addr := uint16(reladdr)
	if pagecrossed(cpu.PC, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
}

// 58
func CLI(cpu *CPU) {
	cpu.P.clearBit(pbitI)
	cpu.Clock += 2
	cpu.PC += 1
}

// 66
func RORzer(cpu *CPU) {
	oper := cpu.Read8(cpu.PC + 1)
	val := cpu.Read8(uint16(oper))
	carry := val & 1 // carry will be set to bit 0
	val >>= 1
	// bit 7 is set to the carry
	if cpu.P.C() {
		val |= 1 << 7
	}

	cpu.Write8(uint16(oper), val)

	cpu.P.checkN(val)
	cpu.P.checkZ(val)
	cpu.P.writeBit(pbitC, carry != 0)

	cpu.PC += 2
	cpu.Clock += 5
}

// 6A
func RORacc(cpu *CPU) {
	val := cpu.A
	carry := val & 1 // carry will be set to bit 0
	val >>= 1
	// bit 7 is set to the carry
	if cpu.P.C() {
		val |= 1 << 7
	}

	cpu.A = val
	cpu.P.checkN(val)
	cpu.P.checkZ(val)
	cpu.P.writeBit(pbitC, carry != 0)

	cpu.PC += 1
	cpu.Clock += 2
}

// 6C
func JMPind(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	dst := cpu.Read16(oper)
	cpu.PC = dst
	cpu.Clock += 5
}

// 70
func BVS(cpu *CPU) {
	if !cpu.P.V() {
		cpu.Clock += 2
		cpu.PC += 2
		return
	}

	// Branch
	off := cpu.Read8(cpu.PC + 1)
	reladdr := int32(cpu.PC+2) + int32(off)
	addr := uint16(reladdr)
	if pagecrossed(cpu.PC, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
}

// 78
func SEI(cpu *CPU) {
	cpu.P.setBit(pbitI)
	cpu.Clock += 2
	cpu.PC += 1
}

// 8D
func STAabs(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	cpu.bus.Write8(oper, cpu.A)
	cpu.PC += 3
	cpu.Clock += 5
}

// 8E
func STXabs(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	cpu.bus.Write8(oper, cpu.X)
	cpu.PC += 3
	cpu.Clock += 4
}

// 84
func STYzer(cpu *CPU) {
	oper := cpu.Read8(cpu.PC + 1)
	cpu.bus.Write8(uint16(oper), cpu.Y)
	cpu.PC += 2
	cpu.Clock += 3
}

// 85
func STAzer(cpu *CPU) {
	oper := cpu.Read8(cpu.PC + 1)
	cpu.bus.Write8(uint16(oper), cpu.A)
	cpu.PC += 2
	cpu.Clock += 3
}

// 86
func STXzer(cpu *CPU) {
	oper := cpu.Read8(cpu.PC + 1)
	cpu.bus.Write8(uint16(oper), cpu.X)
	cpu.PC += 2
	cpu.Clock += 3
}

// 90
func BCC(cpu *CPU) {
	if cpu.P.C() {
		cpu.Clock += 2
		cpu.PC += 2
		return
	}

	// Branch
	off := cpu.Read8(cpu.PC + 1)
	reladdr := int32(cpu.PC+2) + int32(off)
	addr := uint16(reladdr)
	if pagecrossed(cpu.PC, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
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

// A2
func LDXimm(cpu *CPU) {
	cpu.X = cpu.Read8(cpu.PC + 1)
	cpu.P.checkN(cpu.X)
	cpu.P.checkZ(cpu.X)
	cpu.PC += 2
	cpu.Clock += 2
}

// A9
func LDAimm(cpu *CPU) {
	cpu.A = cpu.Read8(cpu.PC + 1)
	cpu.P.checkN(cpu.A)
	cpu.P.checkZ(cpu.A)
	cpu.PC += 2
	cpu.Clock += 2
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

// B0
func BCS(cpu *CPU) {
	if !cpu.P.C() {
		cpu.Clock += 2
		cpu.PC += 2
		return
	}

	// Branch
	off := cpu.Read8(cpu.PC + 1)
	reladdr := int32(cpu.PC+2) + int32(off)
	addr := uint16(reladdr)
	if pagecrossed(cpu.PC, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
}

// B8
func CLV(cpu *CPU) {
	cpu.P.clearBit(pbitV)
	cpu.Clock += 2
	cpu.PC += 1
}

// C8
func INY(cpu *CPU) {
	cpu.Y++
	cpu.P.checkN(cpu.Y)
	cpu.P.checkZ(cpu.Y)
	cpu.Clock += 2
	cpu.PC += 1
}

// CA
func DEX(cpu *CPU) {
	cpu.X--
	cpu.P.checkN(cpu.X)
	cpu.P.checkZ(cpu.X)
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

// D0
func BNE(cpu *CPU) {
	if cpu.P.Z() {
		cpu.Clock += 2
		cpu.PC += 2
		return
	}

	// Branch
	off := cpu.Read8(cpu.PC + 1)
	reladdr := int32(cpu.PC+2) + int32(off)
	addr := uint16(reladdr)
	if pagecrossed(cpu.PC, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
}

// D8
func CLD(cpu *CPU) {
	cpu.P.clearBit(pbitD)
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

// E8
func INX(cpu *CPU) {
	cpu.X++
	cpu.P.checkN(cpu.X)
	cpu.P.checkZ(cpu.X)
	cpu.Clock += 2
	cpu.PC += 1
}

// EA
func NOP(cpu *CPU) {
	cpu.Clock += 2
	cpu.PC += 1
}

// F0
func BEQ(cpu *CPU) {
	if !cpu.P.Z() {
		cpu.Clock += 2
		cpu.PC += 2
		return
	}

	// Branch
	off := cpu.Read8(cpu.PC + 1)
	reladdr := int32(cpu.PC+2) + int32(off)
	addr := uint16(reladdr)
	if pagecrossed(cpu.PC, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
}

// F8
func SED(cpu *CPU) {
	cpu.P.setBit(pbitD)
	cpu.Clock += 2
	cpu.PC += 1
}

/* helpers */

func pagecrossed(a, b uint16) bool {
	return 0xFF00&a != 0xFF00&b
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
