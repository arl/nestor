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
	push16(cpu, cpu.PC+2)
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
	cpu.PC += 2
}

// 03
func SLOizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 2
}

// 05
func ORAzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	ora(cpu, val)
	cpu.PC += 2
}

// 06
func ASLzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	asl(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 07
func SLOzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	slo(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 08
func PHP(cpu *CPU) {
	cpu.tick()
	p := cpu.P
	p |= (1 << pbitB) | (1 << pbitU)
	push8(cpu, uint8(p))
	cpu.PC += 1
}

// 09
func ORAimm(cpu *CPU) {
	ora(cpu, cpu.imm())
	cpu.PC += 2
}

// 0A
func ASLacc(cpu *CPU) {
	asl(cpu, &cpu.A)
	cpu.PC += 1
}

// 0B
func ANC(cpu *CPU) {
	and(cpu, cpu.imm())
	cpu.P.writeBit(pbitC, cpu.P.N())
	cpu.PC += 2
}

// 0C
func NOPabs(cpu *CPU) {
	_ = cpu.Read8(cpu.abs())
	cpu.PC += 3
}

// 0D
func ORAabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	ora(cpu, val)
	cpu.PC += 3
}

// 0E
func ASLabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	asl(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 0F
func SLOabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 10
func BPL(cpu *CPU) {
	branch(cpu, !cpu.P.N())
}

// 11
func ORAizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	val := cpu.Read8(oper)
	ora(cpu, val)
	if crossed == 1 {
		cpu.tick()
	}
	cpu.PC += 2
}

// 13
func SLOizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	cpu.tick()
	_ = crossed
	cpu.Write8(oper, val)
	cpu.PC += 2
}

// 15
func ORAzpx(cpu *CPU) {
	addr := cpu.zpx()
	val := cpu.Read8(uint16(addr))
	ora(cpu, val)
	cpu.PC += 2
}

// 16
func ASLzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	asl(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 17
func SLOzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	slo(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 18
func CLC(cpu *CPU) {
	cpu.P.clearBit(pbitC)
	cpu.tick()
	cpu.PC += 1
}

// 19
func ORAaby(cpu *CPU) {
	addr, _ := cpu.aby()
	val := cpu.Read8(uint16(addr))
	ora(cpu, val)
	cpu.PC += 3
}

// 1A
func NOPimp(cpu *CPU) {
	_ = cpu.Read8(cpu.PC + 1)
	cpu.PC += 1
}

// 1B
func SLOaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	if crossed == 0 {
		cpu.tick()
	}
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 1D
func ORAabx(cpu *CPU) {
	addr, _ := cpu.abx()
	val := cpu.Read8(uint16(addr))
	ora(cpu, val)
	cpu.PC += 3

}

// 1E
func ASLabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	asl(cpu, &val)
	if crossed == 0 {
		cpu.tick()
	}
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 1F
func SLOabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	slo(cpu, &val)
	if crossed == 0 {
		cpu.tick()
	}
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 20
func JSR(cpu *CPU) {
	// Get jump address
	oper := cpu.Read16(cpu.PC + 1)
	cpu.tick()
	// Push return address on the stack
	push16(cpu, cpu.PC+2)
	cpu.PC = oper
}

// 21
func ANDizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	and(cpu, val)
	cpu.PC += 2
}

// 23
func RLAizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 2
}

// 24
func BITzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	bit(cpu, val)
	cpu.PC += 2
}

// 25
func ANDzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	and(cpu, val)
	cpu.PC += 2
}

// 26
func ROLzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	rol(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 27
func RLAzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	rla(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 28
func PLP(cpu *CPU) {
	cpu.tick()
	cpu.tick()
	p := pull8(cpu)
	const mask = 0b11001111 // ignore B and U bits
	cpu.P = P(copybits(uint8(cpu.P), p, mask))
	cpu.PC += 1
}

// 29
func ANDimm(cpu *CPU) {
	and(cpu, cpu.imm())
	cpu.PC += 2
}

// 2A
func ROLacc(cpu *CPU) {
	rol(cpu, &cpu.A)
	cpu.PC += 1
}

// 2B (see  ANC 0B)

// 2C
func BITabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	bit(cpu, val)
	cpu.PC += 3
}

// 2D
func ANDabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	and(cpu, val)
	cpu.PC += 3
}

// 2E
func ROLabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	rol(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 2F
func RLAabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 30
func BMI(cpu *CPU) {
	branch(cpu, cpu.P.N())
}

// 31
func ANDizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	val := cpu.Read8(oper)
	and(cpu, val)
	if crossed == 1 {
		cpu.tick()
	}
	cpu.PC += 2
}

// 33
func RLAizy(cpu *CPU) {
	oper, _ := cpu.izy()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	cpu.tick()
	cpu.Write8(oper, val)
	cpu.PC += 2
}

// 35
func ANDzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	and(cpu, val)
	cpu.PC += 2
}

// 36
func ROLzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	rol(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 37
func RLAzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	rla(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 38
func SEC(cpu *CPU) {
	cpu.P.setBit(pbitC)
	cpu.tick()
	cpu.PC += 1
}

// 39
func ANDaby(cpu *CPU) {
	oper, _ := cpu.aby()
	val := cpu.Read8(oper)
	and(cpu, val)
	cpu.PC += 3
}

// 3B
func RLAaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	if crossed == 0 {
		cpu.tick()
	}
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 3D
func ANDabx(cpu *CPU) {
	oper, _ := cpu.abx()
	val := cpu.Read8(oper)
	and(cpu, val)
	cpu.PC += 3
}

// 3E
func ROLabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	rol(cpu, &val)
	if crossed == 0 {
		cpu.tick()
	}
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 3F
func RLAabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	rla(cpu, &val)
	if crossed == 0 {
		cpu.tick()
	}
	cpu.Write8(oper, val)
	cpu.PC += 3
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
	cpu.PC += 2
}

// 43
func SREizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 2
}

// 45
func EORzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	eor(cpu, val)
	cpu.PC += 2
}

// 46
func LSRzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	lsrmem(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 47
func SREzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	sre(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 48
func PHA(cpu *CPU) {
	cpu.tick()
	push8(cpu, cpu.A)
	cpu.PC += 1
}

// 49
func EORimm(cpu *CPU) {
	eor(cpu, cpu.imm())
	cpu.PC += 2
}

// 4A
func LSRacc(cpu *CPU) {
	lsracc(cpu)
	cpu.PC += 1
}

// 4B
func ALR(cpu *CPU) {
	alr(cpu, cpu.imm())
	cpu.PC += 2
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
	cpu.PC += 3
}

// 4E
func LSRabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	lsrmem(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 4F
func SREabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 50
func BVC(cpu *CPU) {
	branch(cpu, !cpu.P.V())
}

// 51
func EORizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	if crossed == 1 {
		cpu.tick()
	}
	val := cpu.Read8(oper)
	eor(cpu, val)
	cpu.PC += 2
}

// 53
func SREizy(cpu *CPU) {
	oper, _ := cpu.izy()
	cpu.tick()
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 2
}

// 55
func EORzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	eor(cpu, val)
	cpu.PC += 2
}

// 56
func LSRzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	lsrmem(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 57
func SREzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	sre(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 58
func CLI(cpu *CPU) {
	cpu.P.clearBit(pbitI)
	cpu.tick()
	cpu.PC += 1
}

// 59
func EORaby(cpu *CPU) {
	oper, _ := cpu.aby()
	val := cpu.Read8(oper)
	eor(cpu, val)
	cpu.PC += 3
}

// 5B
func SREaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	if crossed == 0 {
		cpu.tick()
	}
	val := cpu.Read8(oper)
	sre(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 5D
func EORabx(cpu *CPU) {
	oper, _ := cpu.abx()
	val := cpu.Read8(oper)
	eor(cpu, val)
	cpu.PC += 3
}

// 5E
func LSRabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	if crossed == 0 {
		cpu.tick()
	}
	val := cpu.Read8(oper)
	lsrmem(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 5F
func SREabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	if crossed == 0 {
		cpu.tick()
	}
	sre(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
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
	cpu.PC += 2
}

// 63
func RRAizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 2
}

// 65
func ADCzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	adc(cpu, val)
	cpu.PC += 2
}

// 66
func RORzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	ror(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 67
func RRAzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	rra(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 68
func PLA(cpu *CPU) {
	cpu.tick()
	cpu.tick()
	cpu.A = pull8(cpu)
	cpu.P.checkNZ(cpu.A)
	cpu.PC += 1
}

// 69
func ADCimm(cpu *CPU) {
	oper := cpu.imm()
	adc(cpu, oper)
	cpu.PC += 2
}

// 6A
func RORacc(cpu *CPU) {
	ror(cpu, &cpu.A)
	cpu.PC += 1
}

// 6B
func ARR(cpu *CPU) {
	arr(cpu, cpu.imm())
	cpu.PC += 2
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
	cpu.PC += 3
}

// 6E
func RORabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	ror(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 6F
func RRAabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 70
func BVS(cpu *CPU) {
	branch(cpu, cpu.P.V())
}

// 71
func ADCizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	val := cpu.Read8(oper)
	adc(cpu, val)
	cpu.PC += 2
	if crossed == 1 {
		cpu.tick()
	}
}

// 73
func RRAizy(cpu *CPU) {
	oper, _ := cpu.izy()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	cpu.tick()
	cpu.Write8(oper, val)
	cpu.PC += 2
}

// 75
func ADCzpx(cpu *CPU) {
	addr := cpu.zpx()
	val := cpu.Read8(uint16(addr))
	adc(cpu, val)
	cpu.PC += 2
}

// 76
func RORzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	ror(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 77
func RRAzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	rra(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// 78
func SEI(cpu *CPU) {
	cpu.P.setBit(pbitI)
	cpu.tick()
	cpu.PC += 1
}

// 79
func ADCaby(cpu *CPU) {
	oper, _ := cpu.aby()
	val := cpu.Read8(oper)
	adc(cpu, val)
	cpu.PC += 3
}

// 7B
func RRAaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	if crossed == 0 {
		cpu.tick()
	}
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 7D
func ADCabx(cpu *CPU) {
	oper, _ := cpu.abx()
	val := cpu.Read8(oper)
	adc(cpu, val)
	cpu.PC += 3
}

// 7E
func RORabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	ror(cpu, &val)
	if crossed == 0 {
		cpu.tick()
	}
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 7F
func RRAabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	rra(cpu, &val)
	if crossed == 0 {
		cpu.tick()
	}
	cpu.Write8(oper, val)
	cpu.PC += 3
}

// 80
func NOPimm(cpu *CPU) {
	cpu.imm()
	cpu.PC += 2
}

// 81
func STAizx(cpu *CPU) {
	addr := cpu.izx()
	cpu.Write8(addr, cpu.A)
	cpu.PC += 2
}

// 83
func SAXizx(cpu *CPU) {
	addr := cpu.izx()
	cpu.Write8(addr, cpu.A&cpu.X)
	cpu.PC += 2
}

// 84
func STYzp(cpu *CPU) {
	oper := cpu.zp()
	cpu.Write8(uint16(oper), cpu.Y)
	cpu.PC += 2
}

// 85
func STAzp(cpu *CPU) {
	oper := cpu.zp()
	cpu.Write8(uint16(oper), cpu.A)
	cpu.PC += 2
}

// 86
func STXzp(cpu *CPU) {
	oper := cpu.zp()
	cpu.Write8(uint16(oper), cpu.X)
	cpu.PC += 2
}

// 87
func SAXzp(cpu *CPU) {
	oper := cpu.zp()
	cpu.Write8(uint16(oper), cpu.A&cpu.X)
	cpu.PC += 2
}

// 88
func DEY(cpu *CPU) {
	dec(cpu, &cpu.Y)
	cpu.P.checkNZ(cpu.Y)
	cpu.PC += 1
}

// 8A
func TXA(cpu *CPU) {
	cpu.A = cpu.X
	cpu.P.checkNZ(cpu.A)
	cpu.tick()
	cpu.PC += 1
}

// 8B - unsupported

// 8C
func STYabs(cpu *CPU) {
	oper := cpu.abs()
	cpu.Write8(oper, cpu.Y)
	cpu.PC += 3
}

// 8D
func STAabs(cpu *CPU) {
	oper := cpu.abs()
	cpu.Write8(oper, cpu.A)
	cpu.PC += 3
}

// 8E
func STXabs(cpu *CPU) {
	oper := cpu.abs()
	cpu.Write8(oper, cpu.X)
	cpu.PC += 3
}

// 8F
func SAXabs(cpu *CPU) {
	oper := cpu.abs()
	cpu.Write8(oper, cpu.A&cpu.X)
	cpu.PC += 3
}

// 90
func BCC(cpu *CPU) {
	branch(cpu, !cpu.P.C())
}

// 91
func STAizy(cpu *CPU) {
	cpu.tick()
	addr, _ := cpu.izy()
	cpu.Write8(addr, cpu.A)
	cpu.PC += 2
}

// 93 - unsupported

// 94
func STYzpx(cpu *CPU) {
	oper := cpu.zpx()
	cpu.Write8(uint16(oper), cpu.Y)
	cpu.PC += 2
}

// 95
func STAzpx(cpu *CPU) {
	addr := cpu.zpx()
	cpu.Write8(uint16(addr), cpu.A)
	cpu.PC += 2
}

// 96
func STXzpy(cpu *CPU) {
	addr := cpu.zpy()
	cpu.Write8(uint16(addr), cpu.X)
	cpu.PC += 2
}

// 97
func SAXzpy(cpu *CPU) {
	addr := cpu.zpy()
	cpu.Write8(uint16(addr), cpu.A&cpu.X)
	cpu.PC += 2
}

// 98
func TYA(cpu *CPU) {
	cpu.A = cpu.Y
	cpu.P.checkNZ(cpu.A)
	cpu.tick()
	cpu.PC += 1
}

// 99
func STAaby(cpu *CPU) {
	addr, crossed := cpu.aby()
	cpu.Write8(addr, cpu.A)
	if crossed == 0 {
		cpu.tick()
	}
	cpu.PC += 3
}

// 9A
func TXS(cpu *CPU) {
	cpu.SP = cpu.X
	cpu.tick()
	cpu.PC += 1
}

// 9B - unsupported

// 9C
func SHY(cpu *CPU) {
	shy(cpu)
	cpu.PC += 3
}

// 9D
func STAabx(cpu *CPU) {
	addr, crossed := cpu.abx()
	if crossed == 0 {
		cpu.tick()
	}
	cpu.Write8(addr, cpu.A)
	cpu.PC += 3
}

// 9E
func SHX(cpu *CPU) {
	shx(cpu)
	cpu.PC += 3
}

// 9F - unsupported

// A0
func LDYimm(cpu *CPU) {
	ldy(cpu, cpu.imm())
	cpu.PC += 2
}

// A1
func LDAizx(cpu *CPU) {
	oper := cpu.izx()
	lda(cpu, cpu.Read8(oper))
	cpu.PC += 2
}

// A2
func LDXimm(cpu *CPU) {
	ldx(cpu, cpu.imm())
	cpu.PC += 2
}

// A3
func LAXizx(cpu *CPU) {
	oper := cpu.izx()
	lax(cpu, cpu.Read8(oper))
	cpu.PC += 2
}

// A4
func LDYzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	ldy(cpu, val)
	cpu.PC += 2
}

// A5
func LDAzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	lda(cpu, val)
	cpu.PC += 2
}

// A6
func LDXzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	ldx(cpu, val)
	cpu.PC += 2
}

// A7
func LAXzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	lax(cpu, val)
	cpu.PC += 2
}

// A8
func TAY(cpu *CPU) {
	cpu.Y = cpu.A
	cpu.P.checkNZ(cpu.Y)
	cpu.tick()
	cpu.PC += 1
}

// A9
func LDAimm(cpu *CPU) {
	lda(cpu, cpu.imm())
	cpu.PC += 2
}

// AA
func TAX(cpu *CPU) {
	cpu.X = cpu.A
	cpu.P.checkNZ(cpu.X)
	cpu.tick()
	cpu.PC += 1
}

// AB - unsupported

// AC
func LDYabs(cpu *CPU) {
	oper := cpu.abs()
	ldy(cpu, cpu.Read8(oper))
	cpu.PC += 3
}

// AD
func LDAabs(cpu *CPU) {
	oper := cpu.abs()
	lda(cpu, cpu.Read8(oper))
	cpu.PC += 3
}

// AE
func LDXabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	ldx(cpu, val)
	cpu.PC += 3
}

// AF
func LAXabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	lax(cpu, val)
	cpu.PC += 3
}

// B0
func BCS(cpu *CPU) {
	branch(cpu, cpu.P.C())
}

// B1
func LDAizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	if crossed == 1 {
		cpu.tick()
	}
	lda(cpu, cpu.Read8(oper))
	cpu.PC += 2
}

// B3
func LAXizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	if crossed == 1 {
		cpu.tick()
	}
	lax(cpu, cpu.Read8(oper))
	cpu.PC += 2
}

// B4
func LDYzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	ldy(cpu, val)
	cpu.PC += 2
}

// B5
func LDAzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	lda(cpu, val)
	cpu.PC += 2
}

// B6
func LDXzpy(cpu *CPU) {
	oper := cpu.zpy()
	val := cpu.Read8(uint16(oper))
	ldx(cpu, val)
	cpu.PC += 2
}

// B7
func LAXzpy(cpu *CPU) {
	oper := cpu.zpy()
	val := cpu.Read8(uint16(oper))
	lax(cpu, val)
	cpu.PC += 2
}

// B8
func CLV(cpu *CPU) {
	cpu.P.clearBit(pbitV)
	cpu.tick()
	cpu.PC += 1
}

// B9
func LDAaby(cpu *CPU) {
	oper, _ := cpu.aby()
	lda(cpu, cpu.Read8(oper))
	cpu.PC += 3
}

// BA
func TSX(cpu *CPU) {
	cpu.X = cpu.SP
	cpu.P.checkNZ(cpu.X)
	cpu.tick()
	cpu.PC += 1
}

// BB
func LAS(cpu *CPU) {
	oper, _ := cpu.aby()
	val := cpu.Read8(oper)
	las(cpu, val)
	cpu.PC += 3
}

// BC
func LDYabx(cpu *CPU) {
	oper, _ := cpu.abx()
	ldy(cpu, cpu.Read8(oper))
	cpu.PC += 3
}

// BD
func LDAabx(cpu *CPU) {
	oper, _ := cpu.abx()
	lda(cpu, cpu.Read8(oper))
	cpu.PC += 3
}

// BE
func LDXaby(cpu *CPU) {
	oper, _ := cpu.aby()
	val := cpu.Read8(oper)
	ldx(cpu, val)
	cpu.PC += 3
}

// BF
func LAXaby(cpu *CPU) {
	oper, _ := cpu.aby()
	val := cpu.Read8(oper)
	lax(cpu, val)
	cpu.PC += 3
}

// C0
func CPYimm(cpu *CPU) {
	oper := cpu.imm()
	cpy(cpu, oper)
	cpu.PC += 2
}

// C1
func CMPizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	cmp_(cpu, val)
	cpu.PC += 2
}

// C3
func DCPizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	val--
	cpu.tick()
	cpu.Write8(oper, val)
	cmp_(cpu, val)
	cpu.PC += 2
}

// C4
func CPYzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	cpy(cpu, val)
	cpu.PC += 2
}

// C5
func CMPzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	cmp_(cpu, val)
	cpu.PC += 2
}

// C6
func DECzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	dec(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// C7
func DCPzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	dec(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cmp_(cpu, val)
	cpu.PC += 2
}

// C8
func INY(cpu *CPU) {
	inc(cpu, &cpu.Y)
	cpu.P.checkNZ(cpu.Y)
	cpu.PC += 1
}

// C9
func CMPimm(cpu *CPU) {
	cmp_(cpu, cpu.imm())
	cpu.PC += 2
}

// CA
func DEX(cpu *CPU) {
	dec(cpu, &cpu.X)
	cpu.P.checkNZ(cpu.X)
	cpu.PC += 1
}

// CB
func SBX(cpu *CPU) {
	sbx(cpu, cpu.imm())
	cpu.PC += 2
}

// CC
func CPYabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	cpy(cpu, val)
	cpu.PC += 3
}

// CD
func CMPabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	cmp_(cpu, val)
	cpu.PC += 3
}

// CE
func DECabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(uint16(oper))
	dec(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 3
}

// CF
func DCPabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(uint16(oper))
	dec(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cmp_(cpu, val)
	cpu.PC += 3
}

// D0
func BNE(cpu *CPU) {
	branch(cpu, !cpu.P.Z())
}

// D1
func CMPizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	if crossed == 1 {
		cpu.tick()
	}
	val := cpu.Read8(oper)
	cmp_(cpu, val)
	cpu.PC += 2
}

// D3
func DCPizy(cpu *CPU) {
	oper, _ := cpu.izy()
	cpu.tick()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
	cpu.PC += 2
}

// D5
func CMPzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	cmp_(cpu, val)
	cpu.PC += 2
}

// D6
func DECzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	dec(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// D7
func DCPzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	dec(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cmp_(cpu, val)
	cpu.PC += 2
}

// D8
func CLD(cpu *CPU) {
	cpu.P.clearBit(pbitD)
	cpu.tick()
	cpu.PC += 1
}

// D9
func CMPaby(cpu *CPU) {
	oper, _ := cpu.aby()
	val := cpu.Read8(oper)
	cmp_(cpu, val)
	cpu.PC += 3
}

// DB
func DCPaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	val := cpu.Read8(oper)
	if crossed == 0 {
		cpu.tick()
	}
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
	cpu.PC += 3
}

// DD
func CMPabx(cpu *CPU) {
	oper, _ := cpu.abx()
	val := cpu.Read8(oper)
	cmp_(cpu, val)
	cpu.PC += 3
}

// DE
func DECabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(uint16(oper))
	if crossed == 0 {
		cpu.tick()
	}
	dec(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 3
}

// DF
func DCPabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	val := cpu.Read8(oper)
	dec(cpu, &val)
	cpu.Write8(oper, val)
	cmp_(cpu, val)
	if crossed != 1 {
		cpu.tick()
	}
	cpu.PC += 3
}

// E0
func CPXimm(cpu *CPU) {
	oper := cpu.imm()
	cpx(cpu, oper)
	cpu.PC += 2
}

// E1
func SBCizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	sbc(cpu, val)
	cpu.PC += 2
}

// E3
func ISBizx(cpu *CPU) {
	oper := cpu.izx()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(oper, val)
	cpu.PC += 2
}

// E4
func CPXzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	cpx(cpu, val)
	cpu.PC += 2
}

// E5
func SBCzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	sbc(cpu, val)
	cpu.PC += 2
}

// E6
func INCzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// E7
func ISBzp(cpu *CPU) {
	oper := cpu.zp()
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// E8
func INX(cpu *CPU) {
	inc(cpu, &cpu.X)
	cpu.P.checkNZ(cpu.X)
	cpu.PC += 1
}

// E9
func SBCimm(cpu *CPU) {
	oper := cpu.imm()
	sbc(cpu, oper)
	cpu.PC += 2
}

// EA - NOP

// EC
func CPXabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	cpx(cpu, val)
	cpu.PC += 3
}

// ED
func SBCabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(oper)
	sbc(cpu, val)
	cpu.PC += 3
}

// EE
func INCabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 3
}

// EF
func ISBabs(cpu *CPU) {
	oper := cpu.abs()
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 3
}

// F0
func BEQ(cpu *CPU) {
	branch(cpu, cpu.P.Z())
}

// F1
func SBCizy(cpu *CPU) {
	oper, crossed := cpu.izy()
	val := cpu.Read8(oper)
	sbc(cpu, val)
	cpu.PC += 2
	if crossed == 1 {
		cpu.tick()
	}
}

// F3
func ISBizy(cpu *CPU) {
	oper, _ := cpu.izy()
	val := cpu.Read8(oper)
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.tick()
	cpu.Write8(oper, val)
	cpu.PC += 2
}

// F5
func SBCzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	sbc(cpu, val)
	cpu.PC += 2
}

// F6
func INCzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// F7
func ISBzpx(cpu *CPU) {
	oper := cpu.zpx()
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 2
}

// F8
func SED(cpu *CPU) {
	cpu.P.setBit(pbitD)
	cpu.tick()
	cpu.PC += 1
}

// F9
func SBCaby(cpu *CPU) {
	oper, _ := cpu.aby()
	val := cpu.Read8(oper)
	sbc(cpu, val)
	cpu.PC += 3
}

// FB
func ISBaby(cpu *CPU) {
	oper, crossed := cpu.aby()
	val := cpu.Read8(uint16(oper))
	val++
	cpu.tick()
	sbc(cpu, val)
	if crossed == 0 {
		cpu.tick()
	}
	cpu.Write8(uint16(oper), val)
	cpu.PC += 3
}

// FD
func SBCabx(cpu *CPU) {
	oper, _ := cpu.abx()
	val := cpu.Read8(oper)
	sbc(cpu, val)
	cpu.PC += 3
}

// FE
func INCabx(cpu *CPU) {
	oper := cpu.abx2()
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 3
}

// FF
func ISBabx(cpu *CPU) {
	oper, crossed := cpu.abx()
	if crossed == 0 {
		cpu.tick()
	}
	val := cpu.Read8(uint16(oper))
	inc(cpu, &val)
	sbc(cpu, val)
	cpu.Write8(uint16(oper), val)
	cpu.PC += 3
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
	addr, _ := cpu.abx()
	_ = cpu.Read8(addr)
	cpu.PC += 3
}

func NOPzpx(cpu *CPU) {
	_ = cpu.Read8(uint16(cpu.zpx()))
	cpu.PC += 2
}

func NOPzp(cpu *CPU) {
	_ = cpu.Read8(uint16(cpu.zp()))
	cpu.PC += 2
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

func branch(cpu *CPU, cond bool) {
	addr := reladdr(cpu)
	if cond {
		if pagecrossed(cpu.PC+2, addr) {
			cpu.tick()
		}
		cpu.tick()
		cpu.PC = addr
		return
	}

	cpu.PC += 2
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

func (cpu *CPU) imm() uint8  { return cpu.Read8(cpu.PC + 1) }
func (cpu *CPU) abs() uint16 { return cpu.Read16(cpu.PC + 1) }
func (cpu *CPU) zp() uint8   { return cpu.Read8(cpu.PC + 1) }

func (cpu *CPU) zpx() uint8 {
	cpu.tick()
	return cpu.zp() + cpu.X
}
func (cpu *CPU) zpy() uint8 {
	cpu.tick()
	return cpu.zp() + cpu.Y
}

// abolute indexed x. returns the destination address and a integer set to 1 if
// a page boundary was crossed.
func (cpu *CPU) abx() (uint16, uint8) {
	addr := cpu.abs()
	dst := addr + uint16(cpu.X)
	crossed := pagecrossed(addr, dst)
	if crossed {
		cpu.tick()
	}
	return dst, b2i(crossed)
}

func (cpu *CPU) abx2() uint16 {
	cpu.tick()
	return cpu.abs() + uint16(cpu.X)
}

// abolute indexed y. returns the destination address and a integer set to 1 if
// a page boundary was crossed.
func (cpu *CPU) aby() (uint16, uint8) {
	addr := cpu.abs()
	dst := addr + uint16(cpu.Y)
	crossed := pagecrossed(addr, dst)
	if crossed {
		cpu.tick()
	}
	return dst, b2i(crossed)
}

// zeropage indexed indirect (zp,x)
func (cpu *CPU) izx() uint16 {
	cpu.tick()
	oper := uint8(cpu.zp())
	oper += cpu.X
	return cpu.zpr16(uint16(oper))
}

// zeropage indexed indirect (zp),y. returns the destination address and a
// integer set to 1 if a page boundary was crossed.
func (cpu *CPU) izy() (uint16, uint8) {
	oper := cpu.zp()
	addr := cpu.zpr16(uint16(oper))
	dst := addr + uint16(cpu.Y)
	return dst, b2i(pagecrossed(addr, dst))
}

func (cpu *CPU) ind() uint16 {
	oper := cpu.Read16(cpu.PC + 1)
	lo := cpu.Read8(oper)
	// 2 bytes address wrap around
	hi := cpu.Read8((0xff00 & oper) | (0x00ff & (oper + 1)))
	return uint16(hi)<<8 | uint16(lo)
}
