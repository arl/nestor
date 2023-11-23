package emu

import "fmt"

var ops = [256]func(cpu *CPU){
	0x00: BRK,
	0x01: ORAizx,
	0x02: JAM,
	0x03: SLOizx,
	0x04: NOPzp,
	0x05: ORAzp,
	0x06: ASLzp,
	0x07: SLOzp,
	0x08: PHP,
	0x09: ORAimm,
	0x0A: ASLacc,
	0x0B: ANC,
	0x0C: NOPabs,
	0x0D: ORAabs,
	0x0E: ASLabs,
	0x0F: SLOabs,
	0x10: BPL,
	0x11: ORAizy,
	0x12: JAM,
	0x13: SLOizy,
	0x14: NOPzpx,
	0x15: ORAzpx,
	0x16: ASLzpx,
	0x17: SLOzpx,
	0x18: CLC,
	0x19: ORAaby,
	0x1A: NOPimp,
	0x1B: SLOaby,
	0x1C: NOPabx,
	0x1D: ORAabx,
	0x1E: ASLabx,
	0x1F: SLOabx,
	0x20: JSR,
	0x21: ANDizx,
	0x22: JAM,
	0x23: RLAizx,
	0x24: BITzp,
	0x25: ANDzp,
	0x26: ROLzp,
	0x27: RLAzp,
	0x28: PLP,
	0x29: ANDimm,
	0x2A: ROLacc,
	0x2B: ANC,
	0x2C: BITabs,
	0x2D: ANDabs,
	0x2E: ROLabs,
	0x2F: RLAabs,
	0x30: BMI,
	0x31: ANDizy,
	0x32: JAM,
	0x33: RLAizy,
	0x34: NOPzpx,
	0x35: ANDzpx,
	0x36: ROLzpx,
	0x37: RLAzpx,
	0x38: SEC,
	0x39: ANDaby,
	0x3A: NOPimp,
	0x3B: RLAaby,
	0x3C: NOPabx,
	0x3D: ANDabx,
	0x3E: ROLabx,
	0x3F: RLAabx,
	0x40: RTI,
	0x41: EORizx,
	0x42: JAM,
	0x43: SREizx,
	0x44: NOPzp,
	0x45: EORzp,
	0x46: LSRzp,
	0x47: SREzp,
	0x48: PHA,
	0x49: EORimm,
	0x4A: LSRacc,
	0x4B: ALR,
	0x4C: JMPabs,
	0x4D: EORabs,
	0x4E: LSRabs,
	0x4F: SREabs,
	0x50: BVC,
	0x51: EORizy,
	0x52: JAM,
	0x53: SREizy,
	0x54: NOPzpx,
	0x55: EORzpx,
	0x56: LSRzpx,
	0x57: SREzpx,
	0x58: CLI,
	0x59: EORaby,
	0x5A: NOPimp,
	0x5B: SREaby,
	0x5C: NOPabx,
	0x5D: EORabx,
	0x5E: LSRabx,
	0x5F: SREabx,
	0x60: RTS,
	0x61: ADCizx,
	0x62: JAM,
	0x63: RRAizx,
	0x64: NOPzp,
	0x65: ADCzp,
	0x66: RORzp,
	0x67: RRAzp,
	0x68: PLA,
	0x69: ADCimm,
	0x6A: RORacc,
	0x6B: ARR,
	0x6C: JMPind,
	0x6D: ADCabs,
	0x6E: RORabs,
	0x6F: RRAabs,
	0x70: BVS,
	0x71: ADCizy,
	0x72: JAM,
	0x73: RRAizy,
	0x74: NOPzpx,
	0x75: ADCzpx,
	0x76: RORzpx,
	0x77: RRAzpx,
	0x78: SEI,
	0x79: ADCaby,
	0x7A: NOPimp,
	0x7B: RRAaby,
	0x7C: NOPabx,
	0x7D: ADCabx,
	0x7E: RORabx,
	0x7F: RRAabx,
	0x80: NOPimm,
	0x81: STAizx,
	0x82: NOPimm,
	0x83: SAXizx,
	0x84: STYzp,
	0x85: STAzp,
	0x86: STXzp,
	0x87: SAXzp,
	0x88: DEY,
	0x89: NOPimm,
	0x8A: TXA,
	0x8B: unsupported,
	0x8C: STYabs,
	0x8D: STAabs,
	0x8E: STXabs,
	0x8F: SAXabs,
	0x90: BCC,
	0x91: STAizy,
	0x92: JAM,
	0x93: unsupported,
	0x94: STYzpx,
	0x95: STAzpx,
	0x96: STXzpy,
	0x97: SAXzpy,
	0x98: TYA,
	0x99: STAaby,
	0x9A: TXS,
	0x9B: unsupported,
	0x9C: SHY,
	0x9D: STAabx,
	0x9E: SHX,
	0x9F: unsupported,
	0xA0: LDYimm,
	0xA1: LDAizx,
	0xA2: LDXimm,
	0xA3: LAXizx,
	0xA4: LDYzp,
	0xA5: LDAzp,
	0xA6: LDXzp,
	0xA7: LAXzp,
	0xA8: TAY,
	0xA9: LDAimm,
	0xAA: TAX,
	0xAB: unsupported,
	0xAC: LDYabs,
	0xAD: LDAabs,
	0xAE: LDXabs,
	0xAF: LAXabs,
	0xB0: BCS,
	0xB1: LDAizy,
	0xB2: JAM,
	0xB3: LAXizy,
	0xB4: LDYzpx,
	0xB5: LDAzpx,
	0xB6: LDXzpy,
	0xB7: LAXzpy,
	0xB8: CLV,
	0xB9: LDAaby,
	0xBA: TSX,
	0xBB: LAS,
	0xBC: LDYabx,
	0xBD: LDAabx,
	0xBE: LDXaby,
	0xBF: LAXaby,
	0xC0: CPYimm,
	0xC1: CMPizx,
	0xC2: NOPimm,
	0xC3: DCPizx,
	0xC4: CPYzp,
	0xC5: CMPzp,
	0xC6: DECzp,
	0xC7: DCPzp,
	0xC8: INY,
	0xC9: CMPimm,
	0xCA: DEX,
	0xCB: SBX,
	0xCC: CPYabs,
	0xCD: CMPabs,
	0xCE: DECabs,
	0xCF: DCPabs,
	0xD0: BNE,
	0xD1: CMPizy,
	0xD2: JAM,
	0xD3: DCPizy,
	0xD4: NOPzpx,
	0xD5: CMPzpx,
	0xD6: DECzpx,
	0xD7: DCPzpx,
	0xD8: CLD,
	0xD9: CMPaby,
	0xDA: NOPimp,
	0xDB: DCPaby,
	0xDC: NOPabx,
	0xDD: CMPabx,
	0xDE: DECabx,
	0xDF: DCPabx,
	0xE0: CPXimm,
	0xE1: SBCizx,
	0xE2: NOPimm,
	0xE3: ISBizx,
	0xE4: CPXzp,
	0xE5: SBCzp,
	0xE6: INCzp,
	0xE7: ISBzp,
	0xE8: INX,
	0xE9: SBCimm,
	0xEA: NOPimp,
	0xEB: SBCimm,
	0xEC: CPXabs,
	0xED: SBCabs,
	0xEE: INCabs,
	0xEF: ISBabs,
	0xF0: BEQ,
	0xF1: SBCizy,
	0xF2: JAM,
	0xF3: ISBizy,
	0xF4: NOPzpx,
	0xF5: SBCzpx,
	0xF6: INCzpx,
	0xF7: ISBzpx,
	0xF8: SED,
	0xF9: SBCaby,
	0xFA: NOPimp,
	0xFB: ISBaby,
	0xFC: NOPabx,
	0xFD: SBCabx,
	0xFE: INCabx,
	0xFF: ISBabx,
}

// 00
func BRK(cpu *CPU) {
	cpu.tick()
	push16(cpu, cpu.PC+1)
	p := cpu.P
	p.setBit(pbitB)
	push8(cpu, uint8(p))
	cpu.P.writeBit(pbitI, true)
	cpu.PC = cpu.Read16(IRQvector)
}

// 01
func ORAizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	ora(cpu, val)
}

// 03
func SLOizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	cpu.Write8(oper, val)
}

// 05
func ORAzp(cpu *CPU) {
	ora(cpu, cpu.Read8(cpu.zp()))
}

// 06
func ASLzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	asl(cpu, &val)
	cpu.Write8(oper, val)
}

// 07
func SLOzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	cpu.Write8(oper, val)
}

// 08
func PHP(cpu *CPU) {
	cpu.tick()
	p := cpu.P
	p |= (1 << pbitB) | (1 << pbitU)
	push8(cpu, uint8(p))
}

// 09
func ORAimm(cpu *CPU) {
	ora(cpu, cpu.imm())
}

// 0A
func ASLacc(cpu *CPU) {
	asl(cpu, &cpu.A)
}

// 0B
func ANC(cpu *CPU) {
	and(cpu, cpu.imm())
	cpu.P.writeBit(pbitC, cpu.P.N())
}

// 0C
func NOPabs(cpu *CPU) {
	_ = cpu.Read8(cpu.abs())
}

// 0D
func ORAabs(cpu *CPU) {
	ora(cpu, cpu.Read8(cpu.abs()))
}

// 0E
func ASLabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	asl(cpu, &val)
	cpu.Write8(oper, val)
}

// 0F
func SLOabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	cpu.Write8(oper, val)
}

// 10
func BPL(cpu *CPU) {
	branch(cpu, !cpu.P.N())
}

// 11
func ORAizy(cpu *CPU) {
	oper := cpu.izy_xp()
	val := cpu.Read8(oper)
	ora(cpu, val)
}

// 13
func SLOizy(cpu *CPU) {
	oper := cpu.izy()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	cpu.tick()
	cpu.Write8(oper, val)
}

// 15
func ORAzpx(cpu *CPU) {
	addr := cpu.zpx()
	val := cpu.Read8(uint16(addr))
	ora(cpu, val)
}

// 16
func ASLzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	asl(cpu, &val)
	cpu.Write8(uint16(oper), val)
}

// 17
func SLOzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	slo(cpu, &val)
	cpu.Write8(uint16(oper), val)
}

// 18
func CLC(cpu *CPU) {
	cpu.P.clearBit(pbitC)
	cpu.tick()
}

// 19
func ORAaby(cpu *CPU) {
	addr := cpu.abypx()
	val := cpu.Read8(uint16(addr))
	ora(cpu, val)
}

// 1A
func NOPimp(cpu *CPU) {
	_ = cpu.Read8(cpu.PC + 1)
}

// 1B
func SLOaby(cpu *CPU) {
	oper := cpu.aby()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	cpu.Write8(oper, val)
}

// 1D
func ORAabx(cpu *CPU) {
	addr := cpu.abxpx()
	val := cpu.Read8(uint16(addr))
	ora(cpu, val)
}

// 1E
func ASLabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	asl(cpu, &val)
	cpu.Write8(oper, val)
}

// 1F
func SLOabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	cpu.Write8(oper, val)
}

// 20
func JSR(cpu *CPU) {
	// Get jump address
	oper := cpu.Read16(cpu.PC)
	cpu.tick()
	// Push return address on the stack
	push16(cpu, cpu.PC+1)
	cpu.PC = oper
}

// 21
func ANDizx(cpu *CPU) {
	and(cpu, cpu.Read8(cpu.izx()))
}

// 23
func RLAizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	cpu.Write8(oper, val)
}

// 24
func BITzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	bit(cpu, val)
}

// 25
func ANDzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	and(cpu, val)
}

// 26
func ROLzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	rol(cpu, &val)
	cpu.Write8(oper, val)
}

// 27
func RLAzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	cpu.Write8(oper, val)
}

// 28
func PLP(cpu *CPU) {
	cpu.tick()
	cpu.tick()
	p := pull8(cpu)
	const mask = 0b11001111 // ignore B and U bits
	cpu.P = P(copybits(uint8(cpu.P), p, mask))
}

// 29
func ANDimm(cpu *CPU) {
	and(cpu, cpu.imm())
}

// 2A
func ROLacc(cpu *CPU) {
	rol(cpu, &cpu.A)
}

// 2B (see  ANC 0B)

// 2C
func BITabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	bit(cpu, val)
}

// 2D
func ANDabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	and(cpu, val)
}

// 2E
func ROLabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	rol(cpu, &val)
	cpu.Write8(oper, val)
}

// 2F
func RLAabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	cpu.Write8(oper, val)
}

// 30
func BMI(cpu *CPU) {
	branch(cpu, cpu.P.N())
}

// 31
func ANDizy(cpu *CPU) {
	oper := cpu.izy_xp()
	val := cpu.Read8(oper)
	and(cpu, val)
}

// 33
func RLAizy(cpu *CPU) {
	oper := cpu.izy()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	cpu.tick()
	cpu.Write8(oper, val)
}

// 35
func ANDzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	and(cpu, val)
}

// 36
func ROLzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	rol(cpu, &val)
	cpu.Write8(oper, val)
}

// 37
func RLAzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	cpu.Write8(oper, val)
}

// 38
func SEC(cpu *CPU) {
	cpu.P.setBit(pbitC)
	cpu.tick()
}

// 39
func ANDaby(cpu *CPU) {
	oper := cpu.abypx()
	val := cpu.Read8(oper)
	and(cpu, val)
}

// 3B
func RLAaby(cpu *CPU) {
	oper := cpu.aby()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	cpu.Write8(oper, val)
}

// 3D
func ANDabx(cpu *CPU) {
	oper := cpu.abxpx()
	val := cpu.Read8(oper)
	and(cpu, val)
}

// 3E
func ROLabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	rol(cpu, &val)
	cpu.Write8(oper, val)
}

// 3F
func RLAabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	cpu.Write8(oper, val)
}

// 40
func RTI(cpu *CPU) {
	cpu.tick()
	cpu.tick()
	p := pull8(cpu)
	const mask = 0b11001111 // ignore B and U bits
	cpu.P = P(copybits(uint8(cpu.P), p, mask))
	cpu.PC = pull16(cpu)
}

// 41
func EORizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	eor(cpu, val)
}

// 43
func SREizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
}

// 45
func EORzp(cpu *CPU) {
	eor(cpu, cpu.Read8(cpu.zp()))
}

// 46
func LSRzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	lsrmem(cpu, &val)
	cpu.Write8(oper, val)
}

// 47
func SREzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
}

// 48
func PHA(cpu *CPU) {
	cpu.tick()
	push8(cpu, cpu.A)
}

// 49
func EORimm(cpu *CPU) {
	eor(cpu, cpu.imm())
}

// 4A
func LSRacc(cpu *CPU) {
	lsracc(cpu)
}

// 4B
func ALR(cpu *CPU) {
	alr(cpu, cpu.imm())
}

// 4C
func JMPabs(cpu *CPU) {
	cpu.PC = cpu.abs()
}

// 4D
func EORabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	eor(cpu, val)
}

// 4E
func LSRabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	lsrmem(cpu, &val)
	cpu.Write8(oper, val)
}

// 4F
func SREabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
}

// 50
func BVC(cpu *CPU) {
	branch(cpu, !cpu.P.V())
}

// 51
func EORizy(cpu *CPU) {
	oper := cpu.izy_xp()
	val := cpu.Read8(oper)
	eor(cpu, val)
}

// 53
func SREizy(cpu *CPU) {
	oper := cpu.izy()
	cpu.tick()
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
}

// 55
func EORzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	eor(cpu, val)
}

// 56
func LSRzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	lsrmem(cpu, &val)
	cpu.Write8(oper, val)
}

// 57
func SREzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
}

// 58
func CLI(cpu *CPU) {
	cpu.P.clearBit(pbitI)
	cpu.tick()
}

// 59
func EORaby(cpu *CPU) {
	oper := cpu.abypx()
	val := cpu.Read8(oper)
	eor(cpu, val)
}

// 5B
func SREaby(cpu *CPU) {
	oper := cpu.aby()
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
}

// 5D
func EORabx(cpu *CPU) {
	oper := cpu.abxpx()
	val := cpu.Read8(oper)
	eor(cpu, val)
}

// 5E
func LSRabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	lsrmem(cpu, &val)
	cpu.Write8(oper, val)
}

// 5F
func SREabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
}

// 60
func RTS(cpu *CPU) {
	cpu.tick()
	cpu.tick()
	cpu.PC = pull16(cpu)
	cpu.PC++
	cpu.tick()
}

// 61
func ADCizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	adc(cpu, val)
}

// 63
func RRAizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	cpu.Write8(oper, val)
}

// 65
func ADCzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	adc(cpu, val)
}

// 66
func RORzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	ror(cpu, &val)
	cpu.Write8(oper, val)
}

// 67
func RRAzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	cpu.Write8(oper, val)
}

// 68
func PLA(cpu *CPU) {
	cpu.tick()
	cpu.tick()
	cpu.A = pull8(cpu)
	cpu.P.checkNZ(cpu.A)
}

// 69
func ADCimm(cpu *CPU) {
	adc(cpu, cpu.imm())
}

// 6A
func RORacc(cpu *CPU) {
	ror(cpu, &cpu.A)
}

// 6B
func ARR(cpu *CPU) {
	arr(cpu, cpu.imm())
}

// 6C
func JMPind(cpu *CPU) {
	cpu.PC = cpu.ind()
}

// 6D
func ADCabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	adc(cpu, val)
}

// 6E
func RORabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	ror(cpu, &val)
	cpu.Write8(oper, val)
}

// 6F
func RRAabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	cpu.Write8(oper, val)
}

// 70
func BVS(cpu *CPU) {
	branch(cpu, cpu.P.V())
}

// 71
func ADCizy(cpu *CPU) {
	oper := cpu.izy_xp()
	val := cpu.Read8(oper)
	adc(cpu, val)
}

// 73
func RRAizy(cpu *CPU) {
	oper := cpu.izy()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	cpu.tick()
	cpu.Write8(oper, val)
}

// 75
func ADCzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	adc(cpu, val)
}

// 76
func RORzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	ror(cpu, &val)
	cpu.Write8(oper, val)
}

// 77
func RRAzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	cpu.Write8(oper, val)
}

// 78
func SEI(cpu *CPU) {
	cpu.P.setBit(pbitI)
	cpu.tick()
}

// 79
func ADCaby(cpu *CPU) {
	oper := cpu.abypx()
	val := cpu.Read8(oper)
	adc(cpu, val)
}

// 7B
func RRAaby(cpu *CPU) {
	oper := cpu.aby()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	cpu.Write8(oper, val)
}

// 7D
func ADCabx(cpu *CPU) {
	oper := cpu.abxpx()
	val := cpu.Read8(oper)
	adc(cpu, val)
}

// 7E
func RORabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	ror(cpu, &val)
	cpu.Write8(oper, val)
}

// 7F
func RRAabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	cpu.Write8(oper, val)
}

// 80
func NOPimm(cpu *CPU) {
	cpu.imm()
}

// 81
func STAizx(cpu *CPU) {
	addr := cpu.izx()
	cpu.Write8(addr, cpu.A)
}

// 83
func SAXizx(cpu *CPU) {
	addr := cpu.izx()
	cpu.Write8(addr, cpu.A&cpu.X)
}

// 84
func STYzp(cpu *CPU) {
	oper := cpu.zp()
	cpu.Write8(oper, cpu.Y)
}

// 85
func STAzp(cpu *CPU) {
	oper := cpu.zp()
	cpu.Write8(oper, cpu.A)
}

// 86
func STXzp(cpu *CPU) {
	oper := cpu.zp()
	cpu.Write8(oper, cpu.X)
}

// 87
func SAXzp(cpu *CPU) {
	oper := cpu.zp()
	cpu.Write8(oper, cpu.A&cpu.X)
}

// 88
func DEY(cpu *CPU) {
	dec(cpu, &cpu.Y)
	cpu.P.checkNZ(cpu.Y)
}

// 8A
func TXA(cpu *CPU) {
	cpu.A = cpu.X
	cpu.P.checkNZ(cpu.A)
	cpu.tick()
}

// 8B - unsupported

// 8C
func STYabs(cpu *CPU) {
	cpu.Write8(cpu.abs(), cpu.Y)
}

// 8D
func STAabs(cpu *CPU) {
	cpu.Write8(cpu.abs(), cpu.A)
}

// 8E
func STXabs(cpu *CPU) {
	cpu.Write8(cpu.abs(), cpu.X)
}

// 8F
func SAXabs(cpu *CPU) {
	cpu.Write8(cpu.abs(), cpu.A&cpu.X)
}

// 90
func BCC(cpu *CPU) {
	branch(cpu, !cpu.P.C())
}

// 91
func STAizy(cpu *CPU) {
	cpu.tick()
	addr := cpu.izy()
	cpu.Write8(addr, cpu.A)
}

// 93 - unsupported

// 94
func STYzpx(cpu *CPU) {
	cpu.Write8(cpu.zpx(), cpu.Y)
}

// 95
func STAzpx(cpu *CPU) {
	cpu.Write8(cpu.zpx(), cpu.A)
}

// 96
func STXzpy(cpu *CPU) {
	cpu.Write8(cpu.zpy(), cpu.X)
}

// 97
func SAXzpy(cpu *CPU) {
	cpu.Write8(cpu.zpy(), cpu.A&cpu.X)
}

// 98
func TYA(cpu *CPU) {
	cpu.A = cpu.Y
	cpu.P.checkNZ(cpu.A)
	cpu.tick()
}

// 99
func STAaby(cpu *CPU) {
	cpu.Write8(cpu.aby(), cpu.A)
}

// 9A
func TXS(cpu *CPU) {
	cpu.SP = cpu.X
	cpu.tick()
}

// 9B - unsupported

// 9C
func SHY(cpu *CPU) {
	shy(cpu)
}

// 9D
func STAabx(cpu *CPU) {
	cpu.Write8(cpu.abx(), cpu.A)
}

// 9E
func SHX(cpu *CPU) {
	shx(cpu)
}

// 9F - unsupported

// A0
func LDYimm(cpu *CPU) {
	ldy(cpu, cpu.imm())
}

// A1
func LDAizx(cpu *CPU) {
	lda(cpu, cpu.Read8(cpu.izx()))
}

// A2
func LDXimm(cpu *CPU) {
	ldx(cpu, cpu.imm())
}

// A3
func LAXizx(cpu *CPU) {
	lax(cpu, cpu.Read8(cpu.izx()))
}

// A4
func LDYzp(cpu *CPU) {
	val := cpu.Read8(cpu.zp())
	ldy(cpu, val)
}

// A5
func LDAzp(cpu *CPU) {
	val := cpu.Read8(cpu.zp())
	lda(cpu, val)
}

// A6
func LDXzp(cpu *CPU) {
	val := cpu.Read8(cpu.zp())
	ldx(cpu, val)
}

// A7
func LAXzp(cpu *CPU) {
	val := cpu.Read8(cpu.zp())
	lax(cpu, val)
}

// A8
func TAY(cpu *CPU) {
	cpu.Y = cpu.A
	cpu.P.checkNZ(cpu.Y)
	cpu.tick()
}

// A9
func LDAimm(cpu *CPU) {
	lda(cpu, cpu.imm())
}

// AA
func TAX(cpu *CPU) {
	cpu.X = cpu.A
	cpu.P.checkNZ(cpu.X)
	cpu.tick()
}

// AB - unsupported

// AC
func LDYabs(cpu *CPU) {
	ldy(cpu, cpu.Read8(cpu.abs()))
}

// AD
func LDAabs(cpu *CPU) {
	lda(cpu, cpu.Read8(cpu.abs()))
}

// AE
func LDXabs(cpu *CPU) {
	val := cpu.Read8(cpu.abs())
	ldx(cpu, val)
}

// AF
func LAXabs(cpu *CPU) {
	val := cpu.Read8(cpu.abs())
	lax(cpu, val)
}

// B0
func BCS(cpu *CPU) {
	branch(cpu, cpu.P.C())
}

// B1
func LDAizy(cpu *CPU) {
	oper := cpu.izy_xp()
	lda(cpu, cpu.Read8(oper))
}

// B3
func LAXizy(cpu *CPU) {
	oper := cpu.izy_xp()
	lax(cpu, cpu.Read8(oper))
}

// B4
func LDYzpx(cpu *CPU) {
	ldy(cpu, cpu.Read8(cpu.zpx()))
}

// B5
func LDAzpx(cpu *CPU) {
	lda(cpu, cpu.Read8(cpu.zpx()))
}

// B6
func LDXzpy(cpu *CPU) {
	ldx(cpu, cpu.Read8(cpu.zpy()))
}

// B7
func LAXzpy(cpu *CPU) {
	lax(cpu, cpu.Read8(cpu.zpy()))
}

// B8
func CLV(cpu *CPU) {
	cpu.P.clearBit(pbitV)
	cpu.tick()
}

// B9
func LDAaby(cpu *CPU) {
	lda(cpu, cpu.Read8(cpu.abypx()))
}

// BA
func TSX(cpu *CPU) {
	cpu.X = cpu.SP
	cpu.P.checkNZ(cpu.X)
	cpu.tick()
}

// BB
func LAS(cpu *CPU) {
	las(cpu, cpu.Read8(cpu.abypx()))
}

// BC
func LDYabx(cpu *CPU) {
	ldy(cpu, cpu.Read8(cpu.abxpx()))
}

// BD
func LDAabx(cpu *CPU) {
	lda(cpu, cpu.Read8(cpu.abxpx()))
}

// BE
func LDXaby(cpu *CPU) {
	ldx(cpu, cpu.Read8(cpu.abypx()))
}

// BF
func LAXaby(cpu *CPU) {
	lax(cpu, cpu.Read8(cpu.abypx()))
}

// C0
func CPYimm(cpu *CPU) {
	cpy(cpu, cpu.imm())
}

// C1
func CMPizx(cpu *CPU) {
	cmp_(cpu, cpu.Read8(cpu.izx()))
}

// C3
func DCPizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
}

// C4
func CPYzp(cpu *CPU) {
	cpy(cpu, cpu.Read8(cpu.zp()))
}

// C5
func CMPzp(cpu *CPU) {
	cmp_(cpu, cpu.Read8(cpu.zp()))
}

// C6
func DECzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
}

// C7
func DCPzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
}

// C8
func INY(cpu *CPU) {
	inc(cpu, &cpu.Y)
	cpu.P.checkNZ(cpu.Y)
}

// C9
func CMPimm(cpu *CPU) {
	cmp_(cpu, cpu.imm())
}

// CA
func DEX(cpu *CPU) {
	dec(cpu, &cpu.X)
	cpu.P.checkNZ(cpu.X)
}

// CB
func SBX(cpu *CPU) {
	sbx(cpu, cpu.imm())
}

// CC
func CPYabs(cpu *CPU) {
	cpy(cpu, cpu.Read8(cpu.abs()))
}

// CD
func CMPabs(cpu *CPU) {
	cmp_(cpu, cpu.Read8(cpu.abs()))
}

// CE
func DECabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(uint16(oper))
	dec(cpu, &val)
	cpu.Write8(uint16(oper), val)
}

// CF
func DCPabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(uint16(oper))
	dec(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cmp_(cpu, val)
}

// D0
func BNE(cpu *CPU) {
	branch(cpu, !cpu.P.Z())
}

// D1
func CMPizy(cpu *CPU) {
	oper := cpu.izy_xp()
	val := cpu.Read8(oper)
	cmp_(cpu, val)
}

// D3
func DCPizy(cpu *CPU) {
	oper := cpu.izy()
	cpu.tick()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
}

// D5
func CMPzpx(cpu *CPU) {
	val := cpu.Read8(cpu.zpx())
	cmp_(cpu, val)
}

// D6
func DECzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
}

// D7
func DCPzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
}

// D8
func CLD(cpu *CPU) {
	cpu.P.clearBit(pbitD)
	cpu.tick()
}

// D9
func CMPaby(cpu *CPU) {
	cmp_(cpu, cpu.Read8(cpu.abypx()))
}

// DB
func DCPaby(cpu *CPU) {
	oper := cpu.aby()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
}

// DD
func CMPabx(cpu *CPU) {
	cmp_(cpu, cpu.Read8(cpu.abxpx()))
}

// DE
func DECabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
}

// DF
func DCPabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
}

// E0
func CPXimm(cpu *CPU) {
	cpx(cpu, cpu.imm())
}

// E1
func SBCizx(cpu *CPU) {
	sbc(cpu, cpu.Read8(cpu.izx()))
}

// E3
func ISBizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(oper, val)
}

// E4
func CPXzp(cpu *CPU) {
	cpx(cpu, cpu.Read8(cpu.zp()))
}

// E5
func SBCzp(cpu *CPU) {
	sbc(cpu, cpu.Read8(cpu.zp()))
}

// E6
func INCzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	cpu.Write8(oper, val)
}

// E7
func ISBzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(oper, val)
}

// E8
func INX(cpu *CPU) {
	inc(cpu, &cpu.X)
	cpu.P.checkNZ(cpu.X)
}

// E9
func SBCimm(cpu *CPU) {
	sbc(cpu, cpu.imm())
}

// EA - NOP

// EC
func CPXabs(cpu *CPU) {
	cpx(cpu, cpu.Read8(cpu.abs()))
}

// ED
func SBCabs(cpu *CPU) {
	sbc(cpu, cpu.Read8(cpu.abs()))
}

// EE
func INCabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	cpu.Write8(oper, val)
}

// EF
func ISBabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(oper, val)
}

// F0
func BEQ(cpu *CPU) {
	branch(cpu, cpu.P.Z())
}

// F1
func SBCizy(cpu *CPU) {
	oper := cpu.izy_xp()
	val := cpu.Read8(oper)
	sbc(cpu, val)
}

// F3
func ISBizy(cpu *CPU) {
	oper := cpu.izy()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.tick()
	cpu.Write8(oper, val)
}

// F5
func SBCzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	sbc(cpu, val)
}

// F6
func INCzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	cpu.Write8(oper, val)
}

// F7
func ISBzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(oper, val)
}

// F8
func SED(cpu *CPU) {
	cpu.P.setBit(pbitD)
	cpu.tick()
}

// F9
func SBCaby(cpu *CPU) {
	sbc(cpu, cpu.Read8(cpu.abypx()))
}

// FB
func ISBaby(cpu *CPU) {
	oper := cpu.aby()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(oper, val)
}

// FD
func SBCabx(cpu *CPU) {
	sbc(cpu, cpu.Read8(cpu.abxpx()))
}

// FE
func INCabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	cpu.Write8(oper, val)
}

// FF
func ISBabx(cpu *CPU) {
	oper := cpu.abx()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(oper, val)
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
	val ^= 0xff
	carry := cpu.P.ibit(pbitC)
	sum := uint16(cpu.A) + uint16(val) + uint16(carry)

	cpu.P.checkCV(cpu.A, val, sum)
	cpu.A = uint8(sum)
	cpu.P.checkNZ(cpu.A)
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

	cpu.tick()
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

	cpu.tick()
	cpu.P.checkNZ(*val)
	cpu.P.writeBit(pbitC, carry != 0)
}

// shift one bit left (memory or accumulator).
func asl(cpu *CPU, val *uint8) {
	carry := *val & 0x80 // carry is bit 7
	*val <<= 1
	*val &= 0xfe
	cpu.tick()
	cpu.P.checkNZ(*val)
	cpu.P.writeBit(pbitC, carry != 0)
}

// shift one bit right (memory)
func lsrmem(cpu *CPU, val *uint8) {
	carry := *val & 0x01 // carry is bit 0
	*val >>= 1
	*val &= 0x7f
	cpu.tick()
	cpu.P.checkNZ(*val)
	cpu.P.writeBit(pbitC, carry != 0)
}

// shift one bit right (accumulator).
func lsracc(cpu *CPU) {
	carry := cpu.A & 0x01 // carry is bit 0
	cpu.A >>= 1
	cpu.A &= 0x7f
	cpu.P.checkNZ(cpu.A)
	cpu.P.writeBit(pbitC, carry != 0)
	cpu.tick()
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
	cpu.tick()
	*val++
	cpu.P.checkNZ(*val)
}

// decrement memory by one.
func dec(cpu *CPU, val *uint8) {
	cpu.tick()
	*val--
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

// NOP variants

func NOPabx(cpu *CPU) {
	addr := cpu.abxpx()
	_ = cpu.Read8(addr)
}

func NOPzpx(cpu *CPU) {
	_ = cpu.Read8(uint16(cpu.zpx()))
}

func NOPzp(cpu *CPU) {
	_ = cpu.Read8(uint16(cpu.zp()))
}

/* unofficial instructions */

func lax(cpu *CPU, val uint8) {
	lda(cpu, val)
	ldx(cpu, val)
}

func slo(cpu *CPU, val *uint8) {
	asl(cpu, val)
	ora(cpu, *val)
}

func rla(cpu *CPU, val *uint8) {
	rol(cpu, val)
	and(cpu, *val)
}

func sre(cpu *CPU, val *uint8) {
	lsrmem(cpu, val)
	eor(cpu, *val)
}

func rra(cpu *CPU, val *uint8) {
	ror(cpu, val)
	adc(cpu, *val)
}

func alr(cpu *CPU, val uint8) {
	// like and + lsr but saves one tick
	cpu.A &= val
	carry := cpu.A & 0x01 // carry is bit 0
	cpu.A >>= 1
	cpu.A &= 0x7f
	cpu.P.checkNZ(cpu.A)
	cpu.P.writeBit(pbitC, carry != 0)
}

func las(cpu *CPU, val uint8) {
	cpu.A = cpu.SP & val
	cpu.X = cpu.A
	cpu.SP = cpu.A
	cpu.P.checkNZ(cpu.A)
}

func arr(cpu *CPU, val uint8) {
	cpu.A &= val
	cpu.A >>= 1
	cpu.P.writeBit(pbitV, (cpu.A>>6)^(cpu.A>>5)&0x01 != 0)

	// bit 7 is set to prev carry
	if cpu.P.C() {
		cpu.A |= 1 << 7
	}

	cpu.P.checkNZ(cpu.A)
	cpu.P.writeBit(pbitC, cpu.A&(1<<6) != 0)
}

func shx(cpu *CPU) {
	addr := cpu.abs()
	dst := addr + uint16(cpu.Y)

	var waddr uint16
	val := cpu.X & (uint8(addr>>8) + 1)
	if pagecrossed(addr, dst) {
		waddr = (uint16(val) << 8) | dst&0xff
	} else {
		waddr = (addr & 0xff00) | dst&0xff
	}
	cpu.tick()
	cpu.Write8(waddr, val)
}

func shy(cpu *CPU) {
	addr := cpu.abs()
	dst := addr + uint16(cpu.X)
	val := cpu.Y & (uint8(addr>>8) + 1)

	var waddr uint16
	if pagecrossed(addr, dst) {
		waddr = (uint16(val) << 8) | dst&0xff
	} else {
		waddr = (addr & 0xff00) | dst&0xff
	}
	cpu.tick()
	cpu.Write8(waddr, val)
}

func sbx(cpu *CPU, oper uint8) {
	val := (int16(cpu.A) & int16(cpu.X)) - int16(oper)
	cpu.X = uint8(val)
	cpu.P.checkNZ(uint8(val))
	cpu.P.writeBit(pbitC, val >= 0)
}

func unsupported(cpu *CPU) {
	op := cpu.Read8(cpu.PC)
	msg := fmt.Sprintf("unsupported instruction (0x%02X) at %04X", op, cpu.PC)
	panic(msg)
}

func JAM(cpu *CPU) {
	panic("Halt and catch fire!")
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
	off := int8(cpu.Read8(cpu.PC))
	reladdr := int16(cpu.PC+1) + int16(off)
	return uint16(reladdr)
}

func branch(cpu *CPU, cond bool) {
	addr := reladdr(cpu)
	if cond {
		if pagecrossed(cpu.PC+1, addr) {
			cpu.tick()
		}
		cpu.tick()
		cpu.PC = addr
		return
	}
	cpu.PC++
}

// Copy bits from src to dst, using mask to select which bits to copy.
func copybits(dst uint8, src uint8, mask uint8) uint8 {
	return (dst & ^mask) | (src & mask)
}

// read 16 bytes from the zero page, handling page wrap.
func (cpu *CPU) zpr16(addr uint16) uint16 {
	lo := cpu.Read8(addr)
	hi := cpu.Read8(uint16(uint8(addr) + 1))
	return uint16(hi)<<8 | uint16(lo)
}

// addressing modes

func (cpu *CPU) imm() uint8 {
	val := cpu.Read8(cpu.PC)
	cpu.PC++
	return val
}

func (cpu *CPU) abs() uint16 {
	val := cpu.Read16(cpu.PC)
	cpu.PC += 2
	return val
}

func (cpu *CPU) zp() uint16 {
	val := cpu.Read8(cpu.PC)
	cpu.PC++
	return uint16(val)
}

func (cpu *CPU) zpx() uint16 {
	cpu.tick()
	addr := cpu.zp() + uint16(cpu.X)
	return addr & 0xff
}

func (cpu *CPU) zpy() uint16 {
	cpu.tick()
	addr := cpu.zp() + uint16(cpu.Y)
	return addr & 0xff
}

// absolute indexed x (with page cross handling).
func (cpu *CPU) abxpx() uint16 {
	addr := cpu.abs()
	dst := addr + uint16(cpu.X)
	if pagecrossed(addr, dst) {
		cpu.tick()
	}
	return dst
}

func (cpu *CPU) abx() uint16 {
	cpu.tick()
	return cpu.abs() + uint16(cpu.X)
}

// absolute indexed y (with page cross handling).
func (cpu *CPU) abypx() uint16 {
	addr := cpu.abs()
	dst := addr + uint16(cpu.Y)
	if pagecrossed(addr, dst) {
		cpu.tick()
	}
	return dst
}

func (cpu *CPU) aby() uint16 {
	cpu.tick()
	return cpu.abs() + uint16(cpu.Y)
}

// zeropage indexed indirect (zp,x)
func (cpu *CPU) izx() uint16 {
	cpu.tick()
	oper := uint8(cpu.zp()) + cpu.X
	return cpu.zpr16(uint16(oper))
}

// zeropage indexed indirect (zp),y.
func (cpu *CPU) izy() uint16 {
	return cpu.zpr16(cpu.zp()) + uint16(cpu.Y)
}

// zeropage indexed indirect (zp),y.
// (like izy but with an additional cycle if page boundary is crossed)
func (cpu *CPU) izy_xp() uint16 {
	oper := cpu.zp()
	addr := cpu.zpr16(oper)
	if pagecrossed(addr, addr+uint16(cpu.Y)) {
		cpu.tick()
	}
	return addr + uint16(cpu.Y)
}

func (cpu *CPU) ind() uint16 {
	oper := cpu.Read16(cpu.PC)
	lo := cpu.Read8(oper)
	// 2 bytes address wrap around
	hi := cpu.Read8((0xff00 & oper) | (0x00ff & (oper + 1)))
	return uint16(hi)<<8 | uint16(lo)
}
