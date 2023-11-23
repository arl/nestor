package emu

import "fmt"

var ops = [256]func(cpu *CPU){
	0x00: BRK,
	0x01: ORA(izx),
	0x02: JAM,
	0x03: SLO(izx),
	0x04: NOP(zp),
	0x05: ORA(zp),
	0x06: ASL(zp),
	0x07: SLO(zp),
	0x08: PHP,
	0x09: ORA(imm),
	0x0A: ASLacc,
	0x0B: ANC,
	0x0C: NOP(abs),
	0x0D: ORA(abs),
	0x0E: ASL(abs),
	0x0F: SLO(abs),
	0x10: BPL,
	0x11: ORA(izy_xp),
	0x12: JAM,
	0x13: SLOizy,
	0x14: NOP(zpx),
	0x15: ORA(zpx),
	0x16: ASL(zpx),
	0x17: SLO(zpx),
	0x18: CLC,
	0x19: ORA(aby_xp),
	0x1A: NOPimp,
	0x1B: SLO(aby),
	0x1C: NOP(abx_xp),
	0x1D: ORA(abx_xp),
	0x1E: ASL(abx),
	0x1F: SLO(abx),
	0x20: JSR,
	0x21: AND(izx),
	0x22: JAM,
	0x23: RLA(izx),
	0x24: BIT(zp),
	0x25: AND(zp),
	0x26: ROL(zp),
	0x27: RLA(zp),
	0x28: PLP,
	0x29: AND(imm),
	0x2A: ROLacc,
	0x2B: ANC,
	0x2C: BIT(abs),
	0x2D: AND(abs),
	0x2E: ROL(abs),
	0x2F: RLA(abs),
	0x30: BMI,
	0x31: AND(izy_xp),
	0x32: JAM,
	0x33: RLAizy,
	0x34: NOP(zpx),
	0x35: AND(zpx),
	0x36: ROL(zpx),
	0x37: RLA(zpx),
	0x38: SEC,
	0x39: AND(aby_xp),
	0x3A: NOPimp,
	0x3B: RLA(aby),
	0x3C: NOP(abx_xp),
	0x3D: AND(abx_xp),
	0x3E: ROL(abx),
	0x3F: RLA(abx),
	0x40: RTI,
	0x41: EOR(izx),
	0x42: JAM,
	0x43: SRE(izx),
	0x44: NOP(zp),
	0x45: EOR(zp),
	0x46: LSR(zp),
	0x47: SRE(zp),
	0x48: PHA,
	0x49: EOR(imm),
	0x4A: LSRacc,
	0x4B: ALR,
	0x4C: JMPabs,
	0x4D: EOR(abs),
	0x4E: LSR(abs),
	0x4F: SRE(abs),
	0x50: BVC,
	0x51: EOR(izy_xp),
	0x52: JAM,
	0x53: SREizy,
	0x54: NOP(zpx),
	0x55: EOR(zpx),
	0x56: LSR(zpx),
	0x57: SRE(zpx),
	0x58: CLI,
	0x59: EOR(aby_xp),
	0x5A: NOPimp,
	0x5B: SRE(aby),
	0x5C: NOP(abx_xp),
	0x5D: EOR(abx_xp),
	0x5E: LSR(abx),
	0x5F: SRE(abx),
	0x60: RTS,
	0x61: ADC(izx),
	0x62: JAM,
	0x63: RRA(izx),
	0x64: NOP(zp),
	0x65: ADC(zp),
	0x66: ROR(zp),
	0x67: RRA(zp),
	0x68: PLA,
	0x69: ADC(imm),
	0x6A: RORacc,
	0x6B: ARR,
	0x6C: JMPind,
	0x6D: ADC(abs),
	0x6E: ROR(abs),
	0x6F: RRA(abs),
	0x70: BVS,
	0x71: ADC(izy_xp),
	0x72: JAM,
	0x73: RRAizy,
	0x74: NOP(zpx),
	0x75: ADC(zpx),
	0x76: ROR(zpx),
	0x77: RRA(zpx),
	0x78: SEI,
	0x79: ADC(aby_xp),
	0x7A: NOPimp,
	0x7B: RRA(aby),
	0x7C: NOP(abx_xp),
	0x7D: ADC(abx_xp),
	0x7E: ROR(abx),
	0x7F: RRA(abx),
	0x80: NOP(imm),
	0x81: STA(izx),
	0x82: NOP(imm),
	0x83: SAX(izx),
	0x84: STY(zp),
	0x85: STA(zp),
	0x86: STXzp,
	0x87: SAX(zp),
	0x88: DEY,
	0x89: NOP(imm),
	0x8A: TXA,
	0x8B: unsupported,
	0x8C: STY(abs),
	0x8D: STA(abs),
	0x8E: STXabs,
	0x8F: SAX(abs),
	0x90: BCC,
	0x91: STAizy,
	0x92: JAM,
	0x93: unsupported,
	0x94: STY(zpx),
	0x95: STA(zpx),
	0x96: STXzpy,
	0x97: SAX(zpy),
	0x98: TYA,
	0x99: STA(aby),
	0x9A: TXS,
	0x9B: unsupported,
	0x9C: SHY,
	0x9D: STA(abx),
	0x9E: SHX,
	0x9F: unsupported,
	0xA0: LDY(imm),
	0xA1: LDA(izx),
	0xA2: LDX(imm),
	0xA3: LAX(izx),
	0xA4: LDY(zp),
	0xA5: LDA(zp),
	0xA6: LDX(zp),
	0xA7: LAX(zp),
	0xA8: TAY,
	0xA9: LDA(imm),
	0xAA: TAX,
	0xAB: unsupported,
	0xAC: LDY(abs),
	0xAD: LDA(abs),
	0xAE: LDX(abs),
	0xAF: LAX(abs),
	0xB0: BCS,
	0xB1: LDA(izy_xp),
	0xB2: JAM,
	0xB3: LAX(izy_xp),
	0xB4: LDY(zpx),
	0xB5: LDA(zpx),
	0xB6: LDX(zpy),
	0xB7: LAX(zpy),
	0xB8: CLV,
	0xB9: LDA(aby_xp),
	0xBA: TSX,
	0xBB: LAS,
	0xBC: LDY(abx_xp),
	0xBD: LDA(abx_xp),
	0xBE: LDX(aby_xp),
	0xBF: LAX(aby_xp),
	0xC0: CPYimm,
	0xC1: CMPizx,
	0xC2: NOP(imm),
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
	0xD4: NOP(zpx),
	0xD5: CMPzpx,
	0xD6: DECzpx,
	0xD7: DCPzpx,
	0xD8: CLD,
	0xD9: CMPaby,
	0xDA: NOPimp,
	0xDB: DCPaby,
	0xDC: NOP(abx_xp),
	0xDD: CMPabx,
	0xDE: DECabx,
	0xDF: DCPabx,
	0xE0: CPXimm,
	0xE1: SBCizx,
	0xE2: NOP(imm),
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
	0xF4: NOP(zpx),
	0xF5: SBCzpx,
	0xF6: INCzpx,
	0xF7: ISBzpx,
	0xF8: SED,
	0xF9: SBCaby,
	0xFA: NOPimp,
	0xFB: ISBaby,
	0xFC: NOP(abx_xp),
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

// 08
func PHP(cpu *CPU) {
	cpu.tick()
	p := cpu.P
	p |= (1 << pbitB) | (1 << pbitU)
	push8(cpu, uint8(p))
}

// 0A
func ASLacc(cpu *CPU) {
	asl(cpu, &cpu.A)
}

// 0B
func ANC(cpu *CPU) {
	and(cpu, cpu.Read8(imm(cpu)))
	cpu.P.writeBit(pbitC, cpu.P.N())
}

// 0C
func NOPabs(cpu *CPU) {
	_ = cpu.Read8(abs(cpu))
}

// 10
func BPL(cpu *CPU) {
	branch(cpu, !cpu.P.N())
}

// 13 (extra tick)
func SLOizy(cpu *CPU) {
	oper := izy(cpu)
	cpu.tick()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	cpu.Write8(oper, val)
}

// 18
func CLC(cpu *CPU) {
	cpu.P.clearBit(pbitC)
	cpu.tick()
}

// 1A
func NOPimp(cpu *CPU) {
	_ = cpu.Read8(cpu.PC)
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

// 28
func PLP(cpu *CPU) {
	cpu.tick()
	cpu.tick()
	p := pull8(cpu)
	const mask = 0b11001111 // ignore B and U bits
	cpu.P = P(copybits(uint8(cpu.P), p, mask))
}

// 2A
func ROLacc(cpu *CPU) {
	rol(cpu, &cpu.A)
}

// 2C
func BITabs(cpu *CPU) {
	bit(cpu, cpu.Read8(abs(cpu)))
}

// 30
func BMI(cpu *CPU) {
	branch(cpu, cpu.P.N())
}

// 33
func RLAizy(cpu *CPU) {
	oper := izy(cpu)
	val := cpu.Read8(oper)
	rla(cpu, &val)
	cpu.tick()
	cpu.Write8(oper, val)
}

// 38
func SEC(cpu *CPU) {
	cpu.P.setBit(pbitC)
	cpu.tick()
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

// 48
func PHA(cpu *CPU) {
	cpu.tick()
	push8(cpu, cpu.A)
}

// 4A
func LSRacc(cpu *CPU) {
	lsracc(cpu)
}

// 4B
func ALR(cpu *CPU) {
	alr(cpu, cpu.Read8(imm(cpu)))
}

// 4C
func JMPabs(cpu *CPU) {
	cpu.PC = abs(cpu)
}

// 50
func BVC(cpu *CPU) {
	branch(cpu, !cpu.P.V())
}

// 53 - SRE with an extra tick
func SREizy(cpu *CPU) {
	oper := izy(cpu)
	cpu.tick()
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
}

// 58
func CLI(cpu *CPU) {
	cpu.P.clearBit(pbitI)
	cpu.tick()
}

// 60
func RTS(cpu *CPU) {
	cpu.tick()
	cpu.tick()
	cpu.PC = pull16(cpu)
	cpu.PC++
	cpu.tick()
}

// 68
func PLA(cpu *CPU) {
	cpu.tick()
	cpu.tick()
	cpu.A = pull8(cpu)
	cpu.P.checkNZ(cpu.A)
}

// 6A
func RORacc(cpu *CPU) {
	ror(cpu, &cpu.A)
}

// 6B
func ARR(cpu *CPU) {
	arr(cpu, cpu.Read8(imm(cpu)))
}

// 6C
func JMPind(cpu *CPU) {
	cpu.PC = ind(cpu)
}

// 70
func BVS(cpu *CPU) {
	branch(cpu, cpu.P.V())
}

// 73 - extra tick
func RRAizy(cpu *CPU) {
	oper := izy(cpu)
	val := cpu.Read8(oper)
	rra(cpu, &val)
	cpu.tick()
	cpu.Write8(oper, val)
}

// 78
func SEI(cpu *CPU) {
	cpu.P.setBit(pbitI)
	cpu.tick()
}

// 86
func STXzp(cpu *CPU) {
	cpu.Write8(zp(cpu), cpu.X)
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

// 8E
func STXabs(cpu *CPU) {
	cpu.Write8(abs(cpu), cpu.X)
}

// 90
func BCC(cpu *CPU) {
	branch(cpu, !cpu.P.C())
}

// 91 - extra tick
func STAizy(cpu *CPU) {
	cpu.tick()
	cpu.Write8(izy(cpu), cpu.A)
}

// 93 - unsupported

// 96
func STXzpy(cpu *CPU) {
	cpu.Write8(zpy(cpu), cpu.X)
}

// 98
func TYA(cpu *CPU) {
	cpu.A = cpu.Y
	cpu.P.checkNZ(cpu.A)
	cpu.tick()
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

// 9E
func SHX(cpu *CPU) {
	shx(cpu)
}

// 9F - unsupported

// A8
func TAY(cpu *CPU) {
	cpu.Y = cpu.A
	cpu.P.checkNZ(cpu.Y)
	cpu.tick()
}

// AA
func TAX(cpu *CPU) {
	cpu.X = cpu.A
	cpu.P.checkNZ(cpu.X)
	cpu.tick()
}

// AB - unsupported

// B0
func BCS(cpu *CPU) {
	branch(cpu, cpu.P.C())
}

// B8
func CLV(cpu *CPU) {
	cpu.P.clearBit(pbitV)
	cpu.tick()
}

// BA
func TSX(cpu *CPU) {
	cpu.X = cpu.SP
	cpu.P.checkNZ(cpu.X)
	cpu.tick()
}

// BB
func LAS(cpu *CPU) {
	las(cpu, cpu.Read8(aby_xp(cpu)))
}

// C0
func CPYimm(cpu *CPU) {
	cpy(cpu, cpu.Read8(imm(cpu)))
}

// C1
func CMPizx(cpu *CPU) {
	cmp_(cpu, cpu.Read8(izx(cpu)))
}

// C3
func DCPizx(cpu *CPU) {
	oper := izx(cpu)
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
}

// C4
func CPYzp(cpu *CPU) {
	cpy(cpu, cpu.Read8(zp(cpu)))
}

// C5
func CMPzp(cpu *CPU) {
	cmp_(cpu, cpu.Read8(zp(cpu)))
}

// C6
func DECzp(cpu *CPU) {
	oper := zp(cpu)
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
}

// C7
func DCPzp(cpu *CPU) {
	oper := zp(cpu)
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
	cmp_(cpu, cpu.Read8(imm(cpu)))
}

// CA
func DEX(cpu *CPU) {
	dec(cpu, &cpu.X)
	cpu.P.checkNZ(cpu.X)
}

// CB
func SBX(cpu *CPU) {
	sbx(cpu, cpu.Read8(imm(cpu)))
}

// CC
func CPYabs(cpu *CPU) {
	cpy(cpu, cpu.Read8(abs(cpu)))
}

// CD
func CMPabs(cpu *CPU) {
	cmp_(cpu, cpu.Read8(abs(cpu)))
}

// CE
func DECabs(cpu *CPU) {
	oper := abs(cpu)
	val := cpu.Read8(uint16(oper))
	dec(cpu, &val)
	cpu.Write8(uint16(oper), val)
}

// CF
func DCPabs(cpu *CPU) {
	oper := abs(cpu)
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
	cmp_(cpu, cpu.Read8(izy_xp(cpu)))
}

// D3
func DCPizy(cpu *CPU) {
	oper := izy(cpu)
	cpu.tick()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
}

// D5
func CMPzpx(cpu *CPU) {
	cmp_(cpu, cpu.Read8(zpx(cpu)))
}

// D6
func DECzpx(cpu *CPU) {
	oper := zpx(cpu)
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
}

// D7
func DCPzpx(cpu *CPU) {
	oper := zpx(cpu)
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
	cmp_(cpu, cpu.Read8(aby_xp(cpu)))
}

// DB
func DCPaby(cpu *CPU) {
	oper := aby(cpu)
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
}

// DD
func CMPabx(cpu *CPU) {
	cmp_(cpu, cpu.Read8(abx_xp(cpu)))
}

// DE
func DECabx(cpu *CPU) {
	oper := abx(cpu)
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
}

// DF
func DCPabx(cpu *CPU) {
	oper := abx(cpu)
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
}

// E0
func CPXimm(cpu *CPU) {
	cpx(cpu, cpu.Read8(imm(cpu)))
}

// E1
func SBCizx(cpu *CPU) {
	sbc(cpu, cpu.Read8(izx(cpu)))
}

// E3
func ISBizx(cpu *CPU) {
	oper := izx(cpu)
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(oper, val)
}

// E4
func CPXzp(cpu *CPU) {
	cpx(cpu, cpu.Read8(zp(cpu)))
}

// E5
func SBCzp(cpu *CPU) {
	sbc(cpu, cpu.Read8(zp(cpu)))
}

// E6
func INCzp(cpu *CPU) {
	oper := zp(cpu)
	val := cpu.Read8(oper)
	inc(cpu, &val)
	cpu.Write8(oper, val)
}

// E7
func ISBzp(cpu *CPU) {
	oper := zp(cpu)
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
	sbc(cpu, cpu.Read8(imm(cpu)))
}

// EA - NOP

// EC
func CPXabs(cpu *CPU) {
	cpx(cpu, cpu.Read8(abs(cpu)))
}

// ED
func SBCabs(cpu *CPU) {
	sbc(cpu, cpu.Read8(abs(cpu)))
}

// EE
func INCabs(cpu *CPU) {
	oper := abs(cpu)
	val := cpu.Read8(oper)
	inc(cpu, &val)
	cpu.Write8(oper, val)
}

// EF
func ISBabs(cpu *CPU) {
	oper := abs(cpu)
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
	sbc(cpu, cpu.Read8(izy_xp(cpu)))
}

// F3
func ISBizy(cpu *CPU) {
	oper := izy(cpu)
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.tick()
	cpu.Write8(oper, val)
}

// F5
func SBCzpx(cpu *CPU) {
	sbc(cpu, cpu.Read8(zpx(cpu)))
}

// F6
func INCzpx(cpu *CPU) {
	oper := zpx(cpu)
	val := cpu.Read8(oper)
	inc(cpu, &val)
	cpu.Write8(oper, val)
}

// F7
func ISBzpx(cpu *CPU) {
	oper := zpx(cpu)
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
	sbc(cpu, cpu.Read8(aby_xp(cpu)))
}

// FB
func ISBaby(cpu *CPU) {
	oper := aby(cpu)
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(oper, val)
}

// FD
func SBCabx(cpu *CPU) {
	sbc(cpu, cpu.Read8(abx_xp(cpu)))
}

// FE
func INCabx(cpu *CPU) {
	oper := abx(cpu)
	val := cpu.Read8(oper)
	inc(cpu, &val)
	cpu.Write8(oper, val)
}

// FF
func ISBabx(cpu *CPU) {
	oper := abx(cpu)
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(oper, val)
}

/* common instruction implementation */

func NOP(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		_ = cpu.Read8(m(cpu))
	}
}

// or memory with accumulator.
func ora(cpu *CPU, val uint8) {
	cpu.A |= val
	cpu.P.checkNZ(cpu.A)
}

// or memory with accumulator.
func ORA(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		val := cpu.Read8(m(cpu))
		cpu.A |= val
		cpu.P.checkNZ(cpu.A)
	}
}

func SLO(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		oper := m(cpu)
		val := cpu.Read8(oper)
		slo(cpu, &val)
		cpu.Write8(oper, val)
	}
}

func ASL(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		oper := m(cpu)
		val := cpu.Read8(oper)
		asl(cpu, &val)
		cpu.Write8(oper, val)
	}
}

func AND(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		and(cpu, cpu.Read8(m(cpu)))
	}
}

func RLA(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		oper := m(cpu)
		val := cpu.Read8(oper)
		rla(cpu, &val)
		cpu.Write8(oper, val)
	}
}

func BIT(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		bit(cpu, cpu.Read8(m(cpu)))
	}
}

func ROL(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		oper := m(cpu)
		val := cpu.Read8(oper)
		rol(cpu, &val)
		cpu.Write8(oper, val)
	}
}

func EOR(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		eor(cpu, cpu.Read8(m(cpu)))
	}
}

func SRE(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		oper := m(cpu)
		val := cpu.Read8(oper)
		sre(cpu, &val)
		cpu.Write8(oper, val)
	}
}

func LSR(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		oper := m(cpu)
		val := cpu.Read8(oper)
		lsrmem(cpu, &val)
		cpu.Write8(oper, val)
	}
}

func ADC(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		adc(cpu, cpu.Read8(m(cpu)))
	}
}

func RRA(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		oper := m(cpu)
		val := cpu.Read8(oper)
		rra(cpu, &val)
		cpu.Write8(oper, val)
	}
}

func ROR(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		oper := m(cpu)
		val := cpu.Read8(oper)
		ror(cpu, &val)
		cpu.Write8(oper, val)
	}
}

func STA(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		cpu.Write8(m(cpu), cpu.A)
	}
}

func SAX(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		cpu.Write8(m(cpu), cpu.A&cpu.X)
	}
}

func STY(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		cpu.Write8(m(cpu), cpu.Y)
	}
}

func LDY(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		ldy(cpu, cpu.Read8(m(cpu)))
	}
}

func LDA(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		lda(cpu, cpu.Read8(m(cpu)))
	}
}

func LDX(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		ldx(cpu, cpu.Read8(m(cpu)))
	}
}

func LAX(m addrmode) func(cpu *CPU) {
	return func(cpu *CPU) {
		lax(cpu, cpu.Read8(m(cpu)))
	}
}

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
	addr := abs(cpu)
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
	addr := abs(cpu)
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

type addrmode func(*CPU) uint16

func imm(cpu *CPU) uint16 {
	val := cpu.PC
	cpu.PC++
	return val
}

func abs(cpu *CPU) uint16 {
	val := cpu.Read16(cpu.PC)
	cpu.PC += 2
	return val
}

func zp(cpu *CPU) uint16 {
	val := cpu.Read8(cpu.PC)
	cpu.PC++
	return uint16(val)
}

func zpx(cpu *CPU) uint16 {
	cpu.tick()
	addr := zp(cpu) + uint16(cpu.X)
	return addr & 0xff
}

func zpy(cpu *CPU) uint16 {
	cpu.tick()
	addr := zp(cpu) + uint16(cpu.Y)
	return addr & 0xff
}

// absolute indexed x (with page cross handling).
func abx_xp(cpu *CPU) uint16 {
	addr := abs(cpu)
	dst := addr + uint16(cpu.X)
	if pagecrossed(addr, dst) {
		cpu.tick()
	}
	return dst
}

func abx(cpu *CPU) uint16 {
	cpu.tick()
	return abs(cpu) + uint16(cpu.X)
}

// absolute indexed y (with page cross handling).
func aby_xp(cpu *CPU) uint16 {
	addr := abs(cpu)
	dst := addr + uint16(cpu.Y)
	if pagecrossed(addr, dst) {
		cpu.tick()
	}
	return dst
}

func aby(cpu *CPU) uint16 {
	cpu.tick()
	return abs(cpu) + uint16(cpu.Y)
}

// zeropage indexed indirect (zp,x)
func izx(cpu *CPU) uint16 {
	cpu.tick()
	oper := uint8(zp(cpu)) + cpu.X
	return cpu.zpr16(uint16(oper))
}

// zeropage indexed indirect (zp),y.
func izy(cpu *CPU) uint16 {
	return cpu.zpr16(zp(cpu)) + uint16(cpu.Y)
}

// zeropage indexed indirect (zp),y.
// (like izy but with an additional cycle if page boundary is crossed)
func izy_xp(cpu *CPU) uint16 {
	oper := zp(cpu)
	addr := cpu.zpr16(oper)
	if pagecrossed(addr, addr+uint16(cpu.Y)) {
		cpu.tick()
	}
	return addr + uint16(cpu.Y)
}

func ind(cpu *CPU) uint16 {
	oper := cpu.Read16(cpu.PC)
	lo := cpu.Read8(oper)
	// 2 bytes address wrap around
	hi := cpu.Read8((0xff00 & oper) | (0x00ff & (oper + 1)))
	return uint16(hi)<<8 | uint16(lo)
}
