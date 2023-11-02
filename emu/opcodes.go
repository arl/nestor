package emu

var ops = [256]func(cpu *CPU){
	0x00: BRK,
	0x01: ORAizx,
	0x04: NOP(2, 3),
	0x05: ORAzp,
	0x06: ASLzp,
	0x08: PHP,
	0x09: ORAimm,
	0x0A: ASLacc,
	0x0C: NOP(3, 4),
	0x0D: ORAabs,
	0x0E: ASLabs,
	0x10: BPL,
	0x11: ORAizy,
	0x14: NOP(2, 4),
	0x15: ORAzpx,
	0x16: ASLzpx,
	0x18: CLC,
	0x19: ORAaby,
	0x1A: NOP(1, 2),
	0x1D: ORAabx,
	0x1E: ASLabx,
	0x20: JSR,
	0x21: ANDizx,
	0x24: BITzp,
	0x25: ANDzp,
	0x26: ROLzp,
	0x28: PLP,
	0x29: ANDimm,
	0x2A: ROLacc,
	0x2C: BITabs,
	0x2D: ANDabs,
	0x2E: ROLabs,
	0x30: BMI,
	0x31: ANDizy,
	0x34: NOP(2, 4),
	0x35: ANDzpx,
	0x36: ROLzpx,
	0x38: SEC,
	0x39: ANDaby,
	0x3A: NOP(1, 2),
	0x3D: ANDabx,
	0x3E: ROLabx,
	0x40: RTI,
	0x41: EORizx,
	0x44: NOP(2, 3),
	0x45: EORzp,
	0x46: LSRzp,
	0x48: PHA,
	0x49: EORimm,
	0x4A: LSRacc,
	0x4C: JMPabs,
	0x4D: EORabs,
	0x4E: LSRabs,
	0x50: BVC,
	0x51: EORizy,
	0x54: NOP(2, 4),
	0x55: EORzpx,
	0x56: LSRzpx,
	0x58: CLI,
	0x59: EORaby,
	0x5A: NOP(1, 2),
	0x5D: EORabx,
	0x5E: LSRabx,
	0x60: RTS,
	0x61: ADCizx,
	0x64: NOP(2, 3),
	0x65: ADCzp,
	0x66: RORzp,
	0x68: PLA,
	0x69: ADCimm,
	0x6A: RORacc,
	0x6C: JMPind,
	0x6D: ADCabs,
	0x6E: RORabs,
	0x70: BVS,
	0x71: ADCizy,
	0x74: NOP(2, 4),
	0x75: ADCzpx,
	0x76: RORzpx,
	0x78: SEI,
	0x79: ADCaby,
	0x7A: NOP(1, 2),
	0x7D: ADCabx,
	0x7E: RORabx,
	0x80: NOP(2, 2),
	0x81: STAizx,
	0x82: NOP(2, 2),
	0x84: STYzp,
	0x85: STAzp,
	0x86: STXzp,
	0x88: DEY,
	0x89: NOP(2, 2),
	0x8A: TXA,
	0x8C: STYabs,
	0x8D: STAabs,
	0x8E: STXabs,
	0x90: BCC,
	0x91: STAizy,
	0x94: STYzpx,
	0x95: STAzpx,
	0x96: STXzpy,
	0x98: TYA,
	0x99: STAaby,
	0x9A: TXS,
	0x9D: STAabx,
	0xA0: LDYimm,
	0xA1: LDAizx,
	0xA2: LDXimm,
	0xA4: LDYzp,
	0xA5: LDAzp,
	0xA6: LDXzp,
	0xA8: TAY,
	0xA9: LDAimm,
	0xAA: TAX,
	0xAC: LDYabs,
	0xAD: LDAabs,
	0xAE: LDXabs,
	0xB0: BCS,
	0xB1: LDAizy,
	0xB4: LDYzpx,
	0xB5: LDAzpx,
	0xB6: LDXzpy,
	0xB8: CLV,
	0xB9: LDAaby,
	0xBA: TSX,
	0xBC: LDYabx,
	0xBD: LDAabx,
	0xBE: LDXaby,
	0xC0: CPYimm,
	0xC1: CMPizx,
	0xC2: NOP(2, 2),
	0xC4: CPYzp,
	0xC5: CMPzp,
	0xC8: INY,
	0xC9: CMPimm,
	0xCA: DEX,
	0xCC: CPYabs,
	0xCD: CMPabs,
	0xD0: BNE,
	0xD1: CMPizy,
	0xD4: NOP(2, 4),
	0xD5: CMPzpx,
	0xD8: CLD,
	0xD9: CMPaby,
	0xDA: NOP(1, 2),
	0xDD: CMPabx,
	0xE0: CPXimm,
	0xE1: SBCizx,
	0xE2: NOP(2, 2),
	0xE4: CPXzp,
	0xE5: SBCzp,
	0xE6: INCzp,
	0xE8: INX,
	0xE9: SBCimm,
	0xEA: NOP(1, 2),
	0xEC: CPXabs,
	0xED: SBCabs,
	0xEE: INCabs,
	0xF0: BEQ,
	0xF1: SBCizy,
	0xF4: NOP(2, 4),
	0xF5: SBCzpx,
	0xF6: INCzpx,
	0xF8: SED,
	0xF9: SBCaby,
	0xFA: NOP(1, 2),
	0xFD: SBCabx,
	0xFE: INCabx,
}

// 00
func BRK(cpu *CPU) {
	push16(cpu, cpu.PC+2)
	p := cpu.P
	p.setBit(pbitB)
	push8(cpu, uint8(p))
	cpu.P.writeBit(pbitI, true)
	cpu.PC = cpu.Read16(IRQvector)
	cpu.Clock += 7
}

// 01
func ORAizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	ora(cpu, val)
	cpu.PC += 2
	cpu.Clock += 6
}

// 05
func ORAzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	ora(cpu, val)
	cpu.PC += 2
	cpu.Clock += 3
}

// 06
func ASLzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	asl(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
	cpu.Clock += 5
}

// 0E
func ASLabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	asl(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
	cpu.Clock += 6
}

// 08
func PHP(cpu *CPU) {
	p := cpu.P
	p |= (1 << pbitB) | (1 << pbitU)
	push8(cpu, uint8(p))
	cpu.PC += 1
	cpu.Clock += 3
}

// 09
func ORAimm(cpu *CPU) {
	ora(cpu, cpu.imm())
	cpu.PC += 2
	cpu.Clock += 2
}

// 0A
func ASLacc(cpu *CPU) {
	asl(cpu, &cpu.A)
	cpu.PC += 1
	cpu.Clock += 2
}

// 0D
func ORAabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	ora(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4
}

// 10
func BPL(cpu *CPU) {
	if cpu.P.N() {
		cpu.PC += 2
		cpu.Clock += 2
		return
	}

	branch(cpu)
}

// 11
func ORAizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	val := cpu.Read8(oper)
	ora(cpu, val)
	cpu.PC += 2
	cpu.Clock += 5 + int64(crossed)
}

// 15
func ORAzpx(cpu *CPU) {
	addr := cpu.zpx()
	val := cpu.Read8(uint16(addr))
	ora(cpu, val)
	cpu.PC += 2
	cpu.Clock += 4
}

// 16
func ASLzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	asl(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
	cpu.Clock += 6
}

// 18
func CLC(cpu *CPU) {
	cpu.P.clearBit(pbitC)
	cpu.PC += 1
	cpu.Clock += 2
}

// 19
func ORAaby(cpu *CPU) {
	addr, crossed := cpu.aby()
	val := cpu.Read8(uint16(addr))
	ora(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// 1D
func ORAabx(cpu *CPU) {
	addr, crossed := cpu.abx()
	val := cpu.Read8(uint16(addr))
	ora(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// 1E
func ASLabx(cpu *CPU) {
	oper, _ := cpu.abx()
	val := cpu.Read8(oper)
	asl(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
	cpu.Clock += 7
}

// 20
func JSR(cpu *CPU) {
	// Get jump address
	oper := cpu.Read16(cpu.PC + 1)
	// Push return address on the stack
	push16(cpu, cpu.PC+2)
	cpu.PC = oper
	cpu.Clock += 6
}

// 21
func ANDizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	and(cpu, val)
	cpu.PC += 2
	cpu.Clock += 6
}

// 24
func BITzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	bit(cpu, val)
	cpu.PC += 2
	cpu.Clock += 3
}

// 25
func ANDzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	and(cpu, val)
	cpu.PC += 2
	cpu.Clock += 3
}

// 26
func ROLzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	rol(cpu, &val)
	cpu.Write8(uint16(oper), val)

	cpu.PC += 2
	cpu.Clock += 5
}

// 28
func PLP(cpu *CPU) {
	p := pull8(cpu)

	const mask = 0b11001111 // ignore B and U bits
	cpu.P = P(copybits(uint8(cpu.P), p, mask))

	cpu.PC += 1
	cpu.Clock += 4
}

// 29
func ANDimm(cpu *CPU) {
	and(cpu, cpu.imm())
	cpu.PC += 2
	cpu.Clock += 2
}

// 2A
func ROLacc(cpu *CPU) {
	rol(cpu, &cpu.A)
	cpu.PC += 1
	cpu.Clock += 2
}

// 2C
func BITabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	bit(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4
}

// 2D
func ANDabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	and(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4
}

// 2E
func ROLabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	rol(cpu, &val)
	cpu.Write8(oper, val)

	cpu.PC += 3
	cpu.Clock += 6
}

// 30
func BMI(cpu *CPU) {
	if !cpu.P.N() {
		cpu.PC += 2
		cpu.Clock += 2
		return
	}

	branch(cpu)
}

// 31
func ANDizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	val := cpu.Read8(oper)
	and(cpu, val)
	cpu.PC += 2
	cpu.Clock += 5 + int64(crossed)
}

// 35
func ANDzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	and(cpu, val)
	cpu.PC += 2
	cpu.Clock += 4
}

// 36
func ROLzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	rol(cpu, &val)
	cpu.Write8(uint16(oper), val)

	cpu.PC += 2
	cpu.Clock += 6
}

// 38
func SEC(cpu *CPU) {
	cpu.P.setBit(pbitC)
	cpu.PC += 1
	cpu.Clock += 2
}

// 39
func ANDaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	val := cpu.Read8(oper)
	and(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// 3D
func ANDabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	and(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// 3E
func ROLabx(cpu *CPU) {
	oper, _ := cpu.abx()
	val := cpu.Read8(oper)
	rol(cpu, &val)
	cpu.Write8(oper, val)

	cpu.PC += 3
	cpu.Clock += 7
}

// 40
func RTI(cpu *CPU) {
	p := pull8(cpu)

	const mask = 0b11001111 // ignore B and U bits
	cpu.P = P(copybits(uint8(cpu.P), p, mask))

	cpu.PC = pull16(cpu)
	cpu.Clock += 6
}

// 41
func EORizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	eor(cpu, val)
	cpu.PC += 2
	cpu.Clock += 6
}

// 45
func EORzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	eor(cpu, val)
	cpu.PC += 2
	cpu.Clock += 3
}

// 46
func LSRzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	lsr(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
	cpu.Clock += 5
}

// 48
func PHA(cpu *CPU) {
	push8(cpu, cpu.A)
	cpu.PC += 1
	cpu.Clock += 3
}

// 49
func EORimm(cpu *CPU) {
	eor(cpu, cpu.imm())
	cpu.PC += 2
	cpu.Clock += 2
}

// 4A
func LSRacc(cpu *CPU) {
	lsr(cpu, &cpu.A)
	cpu.PC += 1
	cpu.Clock += 2
}

// 4C
func JMPabs(cpu *CPU) {
	cpu.PC = cpu.abs()
	cpu.Clock += 3
}

// 4D
func EORabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	eor(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4
}

// 4E
func LSRabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	lsr(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
	cpu.Clock += 6
}

// 50
func BVC(cpu *CPU) {
	if cpu.P.V() {
		cpu.PC += 2
		cpu.Clock += 2
		return
	}

	branch(cpu)
}

// 51
func EORizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	val := cpu.Read8(oper)
	eor(cpu, val)
	cpu.PC += 2
	cpu.Clock += 5 + int64(crossed)
}

// 55
func EORzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	eor(cpu, val)
	cpu.PC += 2
	cpu.Clock += 4
}

// 56
func LSRzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	lsr(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
	cpu.Clock += 6
}

// 58
func CLI(cpu *CPU) {
	cpu.P.clearBit(pbitI)
	cpu.PC += 1
	cpu.Clock += 2
}

// 59
func EORaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	val := cpu.Read8(oper)
	eor(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// 5D
func EORabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	eor(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// 5E
func LSRabx(cpu *CPU) {
	oper, _ := cpu.abx()
	val := cpu.Read8(oper)
	lsr(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
	cpu.Clock += 7
}

// 60
func RTS(cpu *CPU) {
	cpu.PC = pull16(cpu)
	cpu.PC++
	cpu.Clock += 6
}

// 61
func ADCizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)

	adc(cpu, val)

	cpu.PC += 2
	cpu.Clock += 6
}

// 65
func ADCzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	adc(cpu, val)
	cpu.PC += 2
	cpu.Clock += 3
}

// 66
func RORzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	ror(cpu, &val)
	cpu.Write8(uint16(oper), val)

	cpu.PC += 2
	cpu.Clock += 5
}

// 68
func PLA(cpu *CPU) {
	cpu.A = pull8(cpu)
	cpu.P.checkNZ(cpu.A)
	cpu.PC += 1
	cpu.Clock += 4
}

// 69
func ADCimm(cpu *CPU) {
	oper := cpu.imm()

	adc(cpu, oper)

	cpu.PC += 2
	cpu.Clock += 2
}

// 6A
func RORacc(cpu *CPU) {
	ror(cpu, &cpu.A)
	cpu.PC += 1
	cpu.Clock += 2
}

// 6C
func JMPind(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	lo := cpu.Read8(oper)
	// 2 bytes address wrap around
	hi := cpu.Read8((0xff00 & oper) | (0x00ff & (oper + 1)))
	addr := uint16(hi)<<8 | uint16(lo)

	cpu.PC = addr
	cpu.Clock += 5
}

// 6D
func ADCabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)

	adc(cpu, val)

	cpu.PC += 3
	cpu.Clock += 4
}

// 6E
func RORabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	ror(cpu, &val)
	cpu.Write8(oper, val)

	cpu.PC += 3
	cpu.Clock += 6
}

// 70
func BVS(cpu *CPU) {
	if !cpu.P.V() {
		cpu.PC += 2
		cpu.Clock += 2
		return
	}

	branch(cpu)
}

// 71
func ADCizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	val := cpu.Read8(oper)

	adc(cpu, val)

	cpu.PC += 2
	cpu.Clock += 5 + int64(crossed)
}

// 75
func ADCzpx(cpu *CPU) {
	addr := cpu.zpx()
	val := cpu.Read8(uint16(addr))

	adc(cpu, val)

	cpu.PC += 2
	cpu.Clock += 4
}

// 76
func RORzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	ror(cpu, &val)
	cpu.Write8(uint16(oper), val)

	cpu.PC += 2
	cpu.Clock += 6
}

// 78
func SEI(cpu *CPU) {
	cpu.P.setBit(pbitI)
	cpu.PC += 1
	cpu.Clock += 2
}

// 79
func ADCaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	val := cpu.Read8(oper)

	adc(cpu, val)

	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// 7D
func ADCabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)

	adc(cpu, val)

	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// 7E
func RORabx(cpu *CPU) {
	oper, _ := cpu.abx()
	val := cpu.Read8(oper)
	ror(cpu, &val)
	cpu.Write8(oper, val)

	cpu.PC += 3
	cpu.Clock += 7
}

// 81
func STAizx(cpu *CPU) {
	addr := cpu.izx()
	cpu.Write8(addr, cpu.A)
	cpu.PC += 2
	cpu.Clock += 6
}

// 84
func STYzp(cpu *CPU) {
	oper := cpu.zp()
	cpu.Write8(uint16(oper), cpu.Y)
	cpu.PC += 2
	cpu.Clock += 3
}

// 85
func STAzp(cpu *CPU) {
	oper := cpu.zp()
	cpu.Write8(uint16(oper), cpu.A)
	cpu.PC += 2
	cpu.Clock += 3
}

// 86
func STXzp(cpu *CPU) {
	oper := cpu.zp()
	cpu.Write8(uint16(oper), cpu.X)
	cpu.PC += 2
	cpu.Clock += 3
}

// 88
func DEY(cpu *CPU) {
	cpu.Y--
	cpu.P.checkNZ(cpu.Y)
	cpu.PC += 1
	cpu.Clock += 2
}

// 8A
func TXA(cpu *CPU) {
	cpu.A = cpu.X
	cpu.P.checkNZ(cpu.A)
	cpu.PC += 1
	cpu.Clock += 2
}

// 8C
func STYabs(cpu *CPU) {
	oper := cpu.abs()
	cpu.Write8(oper, cpu.Y)
	cpu.PC += 3
	cpu.Clock += 4
}

// 8D
func STAabs(cpu *CPU) {
	oper := cpu.abs()
	cpu.Write8(oper, cpu.A)
	cpu.PC += 3
	cpu.Clock += 4
}

// 8E
func STXabs(cpu *CPU) {
	oper := cpu.abs()
	cpu.Write8(oper, cpu.X)
	cpu.PC += 3
	cpu.Clock += 4
}

// 90
func BCC(cpu *CPU) {
	if cpu.P.C() {
		cpu.PC += 2
		cpu.Clock += 2
		return
	}

	branch(cpu)
}

// 91
func STAizy(cpu *CPU) {
	addr, _ := cpu.izy()
	cpu.Write8(addr, cpu.A)
	cpu.PC += 2
	cpu.Clock += 6
}

// 94
func STYzpx(cpu *CPU) {
	oper := cpu.zpx()
	cpu.Write8(uint16(oper), cpu.Y)
	cpu.PC += 2
	cpu.Clock += 4
}

// 95
func STAzpx(cpu *CPU) {
	addr := cpu.zpx()
	cpu.Write8(uint16(addr), cpu.A)
	cpu.PC += 2
	cpu.Clock += 4
}

// 96
func STXzpy(cpu *CPU) {
	addr := cpu.zpy()
	cpu.Write8(uint16(addr), cpu.X)
	cpu.PC += 2
	cpu.Clock += 4
}

// 98
func TYA(cpu *CPU) {
	cpu.A = cpu.Y
	cpu.P.checkNZ(cpu.A)
	cpu.PC += 1
	cpu.Clock += 2
}

// 99
func STAaby(cpu *CPU) {
	addr, _ := cpu.aby()
	cpu.Write8(addr, cpu.A)
	cpu.PC += 3
	cpu.Clock += 5
}

// 9A
func TXS(cpu *CPU) {
	cpu.SP = cpu.X
	cpu.PC += 1
	cpu.Clock += 2
}

// 9D
func STAabx(cpu *CPU) {
	addr, _ := cpu.abx()
	cpu.Write8(addr, cpu.A)
	cpu.PC += 3
	cpu.Clock += 5
}

// A0
func LDYimm(cpu *CPU) {
	ldy(cpu, cpu.imm())
	cpu.PC += 2
	cpu.Clock += 2
}

// A1
func LDAizx(cpu *CPU) {
	oper := cpu.izx()
	lda(cpu, cpu.Read8(oper))
	cpu.PC += 2
	cpu.Clock += 6
}

// A2
func LDXimm(cpu *CPU) {
	ldx(cpu, cpu.imm())
	cpu.PC += 2
	cpu.Clock += 2
}

// A4
func LDYzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	ldy(cpu, val)
	cpu.PC += 2
	cpu.Clock += 3
}

// A5
func LDAzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	lda(cpu, val)
	cpu.PC += 2
	cpu.Clock += 3
}

// A6
func LDXzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	ldx(cpu, val)
	cpu.PC += 2
	cpu.Clock += 3
}

// A8
func TAY(cpu *CPU) {
	cpu.Y = cpu.A
	cpu.P.checkNZ(cpu.Y)
	cpu.PC += 1
	cpu.Clock += 2
}

// A9
func LDAimm(cpu *CPU) {
	lda(cpu, cpu.imm())
	cpu.PC += 2
	cpu.Clock += 2
}

// AA
func TAX(cpu *CPU) {
	cpu.X = cpu.A
	cpu.P.checkNZ(cpu.X)
	cpu.PC += 1
	cpu.Clock += 2
}

// AC
func LDYabs(cpu *CPU) {
	oper := cpu.abs()
	ldy(cpu, cpu.Read8(oper))
	cpu.PC += 3
	cpu.Clock += 4
}

// AD
func LDAabs(cpu *CPU) {
	oper := cpu.abs()
	lda(cpu, cpu.Read8(oper))
	cpu.PC += 3
	cpu.Clock += 4
}

// AE
func LDXabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	ldx(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4
}

// B0
func BCS(cpu *CPU) {
	if !cpu.P.C() {
		cpu.PC += 2
		cpu.Clock += 2
		return
	}

	branch(cpu)
}

// B1
func LDAizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	lda(cpu, cpu.Read8(oper))
	cpu.PC += 2
	cpu.Clock += 5 + int64(crossed)
}

// B4
func LDYzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	ldy(cpu, val)
	cpu.PC += 2
	cpu.Clock += 4
}

// B5
func LDAzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	lda(cpu, val)
	cpu.PC += 2
	cpu.Clock += 4
}

// B6
func LDXzpy(cpu *CPU) {
	oper := cpu.zpy()
	val := cpu.Read8(uint16(oper))
	ldx(cpu, val)
	cpu.PC += 2
	cpu.Clock += 4
}

// B8
func CLV(cpu *CPU) {
	cpu.P.clearBit(pbitV)
	cpu.PC += 1
	cpu.Clock += 2
}

// B9
func LDAaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	lda(cpu, cpu.Read8(oper))
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// BA
func TSX(cpu *CPU) {
	cpu.X = cpu.SP
	cpu.P.checkNZ(cpu.X)
	cpu.PC += 1
	cpu.Clock += 2
}

// BC
func LDYabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	ldy(cpu, cpu.Read8(oper))
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// BD
func LDAabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	lda(cpu, cpu.Read8(oper))
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// BE
func LDXaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	val := cpu.Read8(oper)
	ldx(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// C0
func CPYimm(cpu *CPU) {
	oper := cpu.imm()
	cpy(cpu, oper)
	cpu.PC += 2
	cpu.Clock += 2
}

// C1
func CMPizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	cmp_(cpu, val)
	cpu.PC += 2
	cpu.Clock += 6
}

// C4
func CPYzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	cpy(cpu, val)
	cpu.PC += 2
	cpu.Clock += 3
}

// C5
func CMPzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	cmp_(cpu, val)
	cpu.PC += 2
	cpu.Clock += 3
}

// C8
func INY(cpu *CPU) {
	cpu.Y++
	cpu.P.checkNZ(cpu.Y)
	cpu.PC += 1
	cpu.Clock += 2
}

// C9
func CMPimm(cpu *CPU) {
	cmp_(cpu, cpu.imm())
	cpu.PC += 2
	cpu.Clock += 2
}

// CA
func DEX(cpu *CPU) {
	cpu.X--
	cpu.P.checkNZ(cpu.X)
	cpu.PC += 1
	cpu.Clock += 2
}

// CC
func CPYabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	cpy(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4
}

// CD
func CMPabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	cmp_(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4
}

// D0
func BNE(cpu *CPU) {
	if cpu.P.Z() {
		cpu.PC += 2
		cpu.Clock += 2
		return
	}

	branch(cpu)
}

// D1
func CMPizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	val := cpu.Read8(oper)
	cmp_(cpu, val)
	cpu.PC += 2
	cpu.Clock += 6 + int64(crossed)
}

// D5
func CMPzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	cmp_(cpu, val)
	cpu.PC += 2
	cpu.Clock += 4
}

// D8
func CLD(cpu *CPU) {
	cpu.P.clearBit(pbitD)
	cpu.PC += 1
	cpu.Clock += 2
}

// D9
func CMPaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	val := cpu.Read8(oper)
	cmp_(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// DD
func CMPabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	cmp_(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// E0
func CPXimm(cpu *CPU) {
	oper := cpu.imm()
	cpx(cpu, oper)
	cpu.PC += 2
	cpu.Clock += 2
}

// E1
func SBCizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	sbc(cpu, val)
	cpu.PC += 2
	cpu.Clock += 6
}

// E4
func CPXzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	cpx(cpu, val)
	cpu.PC += 2
	cpu.Clock += 3
}

// E5
func SBCzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	sbc(cpu, val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
	cpu.Clock += 3
}

// E6
func INCzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
	cpu.Clock += 5
}

// E8
func INX(cpu *CPU) {
	cpu.X++
	cpu.P.checkNZ(cpu.X)
	cpu.PC += 1
	cpu.Clock += 2
}

// E9
func SBCimm(cpu *CPU) {
	oper := cpu.imm()
	sbc(cpu, oper)
	cpu.PC += 2
	cpu.Clock += 2
}

// EA
func NOP(nb uint16, nc int64) func(*CPU) {
	return func(cpu *CPU) {
		cpu.PC += nb
		cpu.Clock += nc
	}
}

// EC
func CPXabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	cpx(cpu, val)
	cpu.PC += 3
	cpu.Clock += 4
}

// ED
func SBCabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	sbc(cpu, val)
	cpu.Write8(oper, val)
	cpu.PC += 3
	cpu.Clock += 4
}

// EE
func INCabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 3
	cpu.Clock += 6
}

// F0
func BEQ(cpu *CPU) {
	if !cpu.P.Z() {
		cpu.PC += 2
		cpu.Clock += 2
		return
	}

	branch(cpu)
}

// F1
func SBCizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	val := cpu.Read8(oper)
	sbc(cpu, val)
	cpu.PC += 2
	cpu.Clock += 5 + int64(crossed)
}

// F5
func SBCzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	sbc(cpu, val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
	cpu.Clock += 4
}

// F6
func INCzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
	cpu.Clock += 6
}

// F8
func SED(cpu *CPU) {
	cpu.P.setBit(pbitD)
	cpu.PC += 1
	cpu.Clock += 2
}

// F9
func SBCaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	val := cpu.Read8(oper)
	sbc(cpu, val)
	cpu.Write8(oper, val)
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// FD
func SBCabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	sbc(cpu, val)
	cpu.Write8(oper, val)
	cpu.PC += 3
	cpu.Clock += 4 + int64(crossed)
}

// FE
func INCabx(cpu *CPU) {
	oper, _ := cpu.abx()
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 3
	cpu.Clock += 7
}

/* common instruction implementation */

// add memory to accumulator with carry.
func adc(cpu *CPU, val uint8) {
	carry := cpu.P.ibit(pbitC)
	sum := uint16(cpu.A) + uint16(val) + uint16(carry)

	cpu.P.checkCV(cpu.A, val, sum)
	cpu.A = uint8(sum)
	cpu.P.checkNZ(cpu.A)
}

// substract memory from accumulator with borrow.
func sbc(cpu *CPU, val uint8) {
	adc(cpu, val^0xff)
}

// and memory with accumulator.
func and(cpu *CPU, val uint8) {
	cpu.A &= val
	cpu.P.checkNZ(cpu.A)
}

// or memory with accumulator.
func ora(cpu *CPU, val uint8) {
	cpu.A |= val
	cpu.P.checkNZ(cpu.A)
}

// exlusive-or memory with accumulator.
func eor(cpu *CPU, val uint8) {
	cpu.A ^= val
	cpu.P.checkNZ(cpu.A)
}

// rotate one bit left.
func rol(cpu *CPU, val *uint8) {
	carry := *val & 0x80 // next carry is bit 7
	*val <<= 1

	// bit 0 is set to prev carry
	if cpu.P.C() {
		*val |= 1 << 0
	}

	cpu.P.checkNZ(*val)
	cpu.P.writeBit(pbitC, carry != 0)
}

// rotate one bit right.
func ror(cpu *CPU, val *uint8) {
	carry := *val & 0x01 // next carry is bit 0
	*val >>= 1

	// bit 7 is set to prev carry
	if cpu.P.C() {
		*val |= 1 << 7
	}

	cpu.P.checkNZ(*val)
	cpu.P.writeBit(pbitC, carry != 0)
}

// shift one bit left (memory or accumulator).
func asl(cpu *CPU, val *uint8) {
	carry := *val & 0x80 // carry is bit 7
	*val <<= 1
	*val &= 0xfe

	cpu.P.checkNZ(*val)
	cpu.P.writeBit(pbitC, carry != 0)
}

// shift one bit right (memory or accumulator).
func lsr(cpu *CPU, val *uint8) {
	carry := *val & 0x01 // carry is bit 0
	*val >>= 1
	*val &= 0x7f

	cpu.P.checkNZ(*val)
	cpu.P.writeBit(pbitC, carry != 0)
}

// test bits in memory with accumulator.
func bit(cpu *CPU, val uint8) {
	// Copy bits 7 and 6 (N and V)
	cpu.P &= 0b00111111
	cpu.P |= P(val & 0b11000000)
	cpu.P.checkZ(cpu.A & val)
}

// compare memory with accumulator.
func cmp_(cpu *CPU, val uint8) {
	cpu.P.checkNZ(cpu.A - val)
	cpu.P.writeBit(pbitC, val <= cpu.A)
}

// compare memory and index x.
func cpx(cpu *CPU, val uint8) {
	cpu.P.checkNZ(cpu.X - val)
	cpu.P.writeBit(pbitC, val <= cpu.X)
}

// compare memory and index y.
func cpy(cpu *CPU, val uint8) {
	cpu.P.checkNZ(cpu.Y - val)
	cpu.P.writeBit(pbitC, val <= cpu.Y)
}

// increment memory by one.
func inc(cpu *CPU, val *uint8) {
	*val++
	cpu.P.checkNZ(*val)
}

// load accumulator with memory.
func lda(cpu *CPU, val uint8) {
	cpu.A = val
	cpu.P.checkNZ(cpu.A)
}

// load index x with memory.
func ldx(cpu *CPU, val uint8) {
	cpu.X = val
	cpu.P.checkNZ(cpu.X)
}

// load index y with memory.
func ldy(cpu *CPU, val uint8) {
	cpu.Y = val
	cpu.P.checkNZ(cpu.Y)
}

/* helpers */

func pagecrossed(a, b uint16) bool {
	// fmt.Printf("pagecrossed: a=%04X, b=%04X\n", a, b)
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
	push8(cpu, uint8(val>>8))
	push8(cpu, uint8(val&0xFF))
}

// pull a 8-bit value from the stack
func pull8(cpu *CPU) uint8 {
	cpu.SP += 1
	top := uint16(cpu.SP) + 0x0100
	val := cpu.Read8(top)
	return val
}

// pull a 16-bit value from the stack
func pull16(cpu *CPU) uint16 {
	lo := pull8(cpu)
	hi := pull8(cpu)
	return uint16(hi)<<8 | uint16(lo)
}

// reladdr returns the destination address for a jump.
// that is the address at PC+1 + an offset (PC+2)
func reladdr(cpu *CPU) uint16 {
	off := int8(cpu.Read8(cpu.PC + 1))
	reladdr := int16(cpu.PC+2) + int16(off)
	return uint16(reladdr)
}

func branch(cpu *CPU) {
	addr := reladdr(cpu)
	if pagecrossed(cpu.PC+2, addr) {
		cpu.Clock += 4
	} else {
		cpu.Clock += 3
	}
	cpu.PC = addr
}

// Copy bits from src to dst, using mask to select which bits to copy.
func copybits(dst uint8, src uint8, mask uint8) uint8 {
	return (dst & ^mask) | (src & mask)
}

// read 16 bytes from the zero page, handling page wrap.
func (cpu *CPU) zpr16(addr uint16) uint16 {
	lo := cpu.bus.Read8(addr)
	hi := cpu.bus.Read8(uint16(uint8(addr) + 1))
	return uint16(hi)<<8 | uint16(lo)
}

// addressing modes

func (cpu *CPU) imm() uint8  { return cpu.Read8(cpu.PC + 1) }
func (cpu *CPU) abs() uint16 { return cpu.Read16(cpu.PC + 1) }
func (cpu *CPU) zp() uint8   { return cpu.Read8(cpu.PC + 1) }
func (cpu *CPU) zpx() uint8  { return cpu.zp() + cpu.X }
func (cpu *CPU) zpy() uint8  { return cpu.zp() + cpu.Y }

// abolute indexed x. returns the destination address and a integer set to 1 if
// a page boundary was crossed.
func (cpu *CPU) abx() (uint16, uint8) {
	addr := cpu.abs()
	dst := addr + uint16(cpu.X)
	return dst, b2i(pagecrossed(addr, dst))
}

// abolute indexed y. returns the destination address and a integer set to 1 if
// a page boundary was crossed.
func (cpu *CPU) aby() (uint16, uint8) {
	addr := cpu.abs()
	dst := addr + uint16(cpu.Y)
	return dst, b2i(pagecrossed(addr, dst))
}

// zeropage indexed indirect (zp,x)
func (cpu *CPU) izx() uint16 {
	oper := uint8(cpu.zp())
	oper += cpu.X
	return cpu.zpr16(uint16(oper))
}

// zeropage indexed indirect (zp),y. returns the destination address and a integer set to 1 if
// a page boundary was crossed.
func (cpu *CPU) izy() (uint16, uint8) {
	oper := cpu.zp()
	addr := cpu.zpr16(uint16(oper))
	dst := addr + uint16(cpu.Y)
	return dst, b2i(pagecrossed(addr, dst))
}
