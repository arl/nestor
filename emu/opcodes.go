package emu

var ops = [256]func(cpu *CPU){
	0x00: BRK,
	0x04: NOP(2, 3),
	0x05: ORAzer,
	0x08: PHP,
	0x09: ORAimm,
	0x0C: NOP(3, 4),
	0x10: BPL,
	0x14: NOP(2, 4),
	0x18: CLC,
	0x1A: NOP(1, 2),
	0x20: JSR,
	0x24: BITzer,
	0x28: PLP,
	0x29: ANDimm,
	0x2C: BITabs,
	0x30: BMI,
	0x34: NOP(2, 4),
	0x38: SEC,
	0x3A: NOP(1, 2),
	0x44: NOP(2, 3),
	0x45: EORzer,
	0x48: PHA,
	0x49: EORimm,
	0x4C: JMPabs,
	0x50: BVC,
	0x54: NOP(2, 4),
	0x58: CLI,
	0x5A: NOP(1, 2),
	0x60: RTS,
	0x64: NOP(2, 3),
	0x66: RORzer,
	0x68: PLA,
	0x69: ADCimm,
	0x6A: RORacc,
	0x6C: JMPind,
	0x70: BVS,
	0x74: NOP(2, 4),
	0x78: SEI,
	0x7A: NOP(1, 2),
	0x80: NOP(2, 2),
	0x81: STAindx,
	0x82: NOP(2, 2),
	0x84: STYzer,
	0x85: STAzer,
	0x86: STXzer,
	0x89: NOP(2, 2),
	0x8A: TXA,
	0x8D: STAabs,
	0x8E: STXabs,
	0x90: BCC,
	0x91: STAindy,
	0x95: STAzerx,
	0x99: STAabsy,
	0x9A: TXS,
	0x9D: STAabsx,
	0xA0: LDYimm,
	0xA2: LDXimm,
	0xA9: LDAimm,
	0xAA: TAX,
	0xAD: LDAabs,
	0xB0: BCS,
	0xB8: CLV,
	0xBA: TSX,
	0xC0: CPYimm,
	0xC2: NOP(2, 2),
	0xC8: INY,
	0xC9: CMPimm,
	0xCA: DEX,
	0xD0: BNE,
	0xD4: NOP(2, 4),
	0xD8: CLD,
	0xDA: NOP(1, 2),
	0xE0: CPXimm,
	0xE2: NOP(2, 2),
	0xE6: INCzer,
	0xE8: INX,
	0xEA: NOP(1, 2),
	0xF0: BEQ,
	0xF4: NOP(2, 4),
	0xF8: SED,
	0xFA: NOP(1, 2),
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

// 05
func ORAzer(cpu *CPU) {
	oper := cpu.zeropage()
	val := cpu.Read8(oper)
	cpu.A |= val
	cpu.P.checkNZ(cpu.A)
	cpu.PC += 2
	cpu.Clock += 3
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
	cpu.A |= cpu.immediate()
	cpu.P.checkNZ(cpu.A)
	cpu.PC += 2
	cpu.Clock += 2
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

// 18
func CLC(cpu *CPU) {
	cpu.P.clearBit(pbitC)
	cpu.PC += 1
	cpu.Clock += 2
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

// 24
func BITzer(cpu *CPU) {
	oper := cpu.zeropage()
	val := cpu.Read8(oper)

	// Copy bits 7 and 6 (N and V)
	cpu.P &= 0b00111111
	cpu.P |= P(val & 0b11000000)

	cpu.P.checkZ(cpu.A & val)
	cpu.PC += 2
	cpu.Clock += 3
}

// 2C
func BITabs(cpu *CPU) {
	oper := cpu.absolute()
	val := cpu.Read8(oper)

	// Copy bits 7 and 6 (N and V)
	cpu.P &= 0b00111111
	cpu.P |= P(val & 0b11000000)

	cpu.P.checkZ(cpu.A & val)
	cpu.PC += 3
	cpu.Clock += 4
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
	cpu.A &= cpu.immediate()
	cpu.P.checkNZ(cpu.A)
	cpu.PC += 2
	cpu.Clock += 2
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

// 38
func SEC(cpu *CPU) {
	cpu.P.setBit(pbitC)
	cpu.PC += 1
	cpu.Clock += 2
}

// 45
func EORzer(cpu *CPU) {
	oper := cpu.zeropage()
	val := cpu.Read8(oper)
	cpu.A ^= val
	cpu.P.checkNZ(cpu.A)
	cpu.PC += 2
	cpu.Clock += 3
}

// 49
func EORimm(cpu *CPU) {
	cpu.A ^= cpu.immediate()
	cpu.P.checkNZ(cpu.A)
	cpu.PC += 2
	cpu.Clock += 2
}

// 4C
func JMPabs(cpu *CPU) {
	cpu.PC = cpu.absolute()
	cpu.Clock += 3
}

// 48
func PHA(cpu *CPU) {
	push8(cpu, cpu.A)

	cpu.PC += 1
	cpu.Clock += 3
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

// 58
func CLI(cpu *CPU) {
	cpu.P.clearBit(pbitI)
	cpu.PC += 1
	cpu.Clock += 2
}

// 60
func RTS(cpu *CPU) {
	cpu.PC = pull16(cpu)
	cpu.PC++
	cpu.Clock += 6
}

// 66
func RORzer(cpu *CPU) {
	oper := cpu.zeropage()
	val := cpu.Read8(oper)
	carry := val & 1 // carry will be set to bit 0
	val >>= 1
	// bit 7 is set to the carry
	if cpu.P.C() {
		val |= 1 << 7
	}

	cpu.Write8(oper, val)

	cpu.P.checkNZ(val)
	cpu.P.writeBit(pbitC, carry != 0)

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
	oper := cpu.immediate()

	adc(cpu, oper)

	cpu.PC += 2
	cpu.Clock += 2
}

// 6A
func RORacc(cpu *CPU) {
	val := cpu.A
	// bit 7 is set to the carry
	if cpu.P.C() {
		val |= 1 << 7
	}

	cpu.A = val
	cpu.P.checkNZ(cpu.A)
	cpu.P.writeBit(pbitC, val&0x01 != 0)

	cpu.PC += 1
	cpu.Clock += 2
}

// 6C
func JMPind(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	cpu.PC = cpu.Read16(oper)
	cpu.Clock += 5
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

// 78
func SEI(cpu *CPU) {
	cpu.P.setBit(pbitI)
	cpu.PC += 1
	cpu.Clock += 2
}

// 81
func STAindx(cpu *CPU) {
	addr := cpu.zpindindx()
	cpu.Write8(addr, cpu.A)
	cpu.PC += 2
	cpu.Clock += 6
}

// 84
func STYzer(cpu *CPU) {
	oper := cpu.zeropage()
	cpu.Write8(oper, cpu.Y)
	cpu.PC += 2
	cpu.Clock += 3
}

// 85
func STAzer(cpu *CPU) {
	oper := cpu.zeropage()
	cpu.Write8(oper, cpu.A)
	cpu.PC += 2
	cpu.Clock += 3
}

// 86
func STXzer(cpu *CPU) {
	oper := cpu.zeropage()
	cpu.Write8(oper, cpu.X)
	cpu.PC += 2
	cpu.Clock += 3
}

// 8A
func TXA(cpu *CPU) {
	cpu.A = cpu.X
	cpu.P.checkNZ(cpu.A)
	cpu.PC += 1
	cpu.Clock += 2
}

// 8D
func STAabs(cpu *CPU) {
	oper := cpu.absolute()
	cpu.Write8(oper, cpu.A)
	cpu.PC += 3
	cpu.Clock += 4
}

// 8E
func STXabs(cpu *CPU) {
	oper := cpu.absolute()
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
func STAindy(cpu *CPU) {
	addr := cpu.zpindindy()
	cpu.Write8(addr, cpu.A)
	cpu.PC += 2
	cpu.Clock += 6
}

// 95
func STAzerx(cpu *CPU) {
	// Read from the zero page
	oper := cpu.Read8(cpu.PC + 1)
	addr := uint16(oper + cpu.X)
	cpu.Write8(addr, cpu.A)
	cpu.PC += 2
	cpu.Clock += 4
}

// 99
func STAabsy(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	addr := oper + uint16(cpu.Y)
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
func STAabsx(cpu *CPU) {
	oper := cpu.Read16(cpu.PC + 1)
	addr := oper + uint16(cpu.X)
	cpu.Write8(addr, cpu.A)
	cpu.PC += 3
	cpu.Clock += 5
}

// A0
func LDYimm(cpu *CPU) {
	cpu.Y = cpu.immediate()
	cpu.P.checkNZ(cpu.Y)
	cpu.PC += 2
	cpu.Clock += 2
}

// A2
func LDXimm(cpu *CPU) {
	cpu.X = cpu.immediate()
	cpu.P.checkNZ(cpu.X)
	cpu.PC += 2
	cpu.Clock += 2
}

// A9
func LDAimm(cpu *CPU) {
	cpu.A = cpu.immediate()
	cpu.P.checkNZ(cpu.A)
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

// AD
func LDAabs(cpu *CPU) {
	oper := cpu.absolute()
	cpu.A = cpu.Read8(oper)
	cpu.P.checkNZ(cpu.A)
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

// B8
func CLV(cpu *CPU) {
	cpu.P.clearBit(pbitV)
	cpu.PC += 1
	cpu.Clock += 2
}

// BA
func TSX(cpu *CPU) {
	cpu.X = cpu.SP
	cpu.P.checkNZ(cpu.X)
	cpu.PC += 1
	cpu.Clock += 2
}

// C0
func CPYimm(cpu *CPU) {
	oper := cpu.immediate()
	cpu.P.checkNZ(cpu.Y - oper)
	cpu.P.writeBit(pbitC, cpu.Y >= oper)
	cpu.PC += 2
	cpu.Clock += 2
}

// C8
func INY(cpu *CPU) {
	cpu.Y++
	cpu.P.checkNZ(cpu.Y)
	cpu.PC += 1
	cpu.Clock += 2
}

// CA
func DEX(cpu *CPU) {
	cpu.X--
	cpu.P.checkNZ(cpu.X)
	cpu.PC += 1
	cpu.Clock += 2
}

// C9
func CMPimm(cpu *CPU) {
	oper := cpu.immediate()
	cpu.P.checkNZ(cpu.A - oper)
	cpu.P.writeBit(pbitC, oper <= cpu.A)
	cpu.PC += 2
	cpu.Clock += 2
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

// D8
func CLD(cpu *CPU) {
	cpu.P.clearBit(pbitD)
	cpu.PC += 1
	cpu.Clock += 2
}

// E0
func CPXimm(cpu *CPU) {
	oper := cpu.immediate()
	cpu.P.checkNZ(cpu.X - oper)
	cpu.P.writeBit(pbitC, cpu.X >= oper)
	cpu.PC += 2
	cpu.Clock += 2
}

// E6
func INCzer(cpu *CPU) {
	oper := cpu.zeropage()
	val := cpu.Read8(oper)
	val++
	cpu.P.checkNZ(val)
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

// EA
func NOP(nb uint16, nc int64) func(*CPU) {
	return func(cpu *CPU) {
		cpu.PC += nb
		cpu.Clock += nc
	}
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

// F8
func SED(cpu *CPU) {
	cpu.P.setBit(pbitD)
	cpu.PC += 1
	cpu.Clock += 2
}

/* common instruction implementation */

func adc(cpu *CPU, oper uint8) {
	carry := cpu.P.ibit(pbitC)
	sum := uint16(cpu.A) + uint16(oper) + uint16(carry)

	cpu.P.checkCV(cpu.A, oper, sum)
	cpu.A = uint8(sum)
	cpu.P.checkNZ(cpu.A)
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
	off := int8(cpu.Read8(cpu.PC + 1))
	reladdr := int32(cpu.PC+2) + int32(off)
	return uint16(reladdr)
}

func branch(cpu *CPU) {
	addr := reladdr(cpu)
	if pagecrossed(cpu.PC, addr) {
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

// addressing modes

func (cpu *CPU) immediate() uint8 {
	return cpu.Read8(cpu.PC + 1)
}

func (cpu *CPU) absolute() uint16 {
	return cpu.Read16(cpu.PC + 1)
}

func (cpu *CPU) zeropage() uint16 {
	return uint16(cpu.Read8(cpu.PC + 1))
}

func (cpu *CPU) zpindindx() uint16 {
	oper := uint8(cpu.zeropage())
	oper += cpu.X
	return cpu.zpr16(uint16(oper))
}

func (cpu *CPU) zpindindy() uint16 {
	oper := cpu.zeropage()
	addr := cpu.zpr16(oper)
	addr += uint16(cpu.Y)
	return addr
}

// read 16 bytes from the zero page, handling page wrap.
func (cpu *CPU) zpr16(addr uint16) uint16 {
	lo := cpu.bus.Read8(addr)
	hi := cpu.bus.Read8(uint16(uint8(addr) + 1))
	return uint16(hi)<<8 | uint16(lo)
}
