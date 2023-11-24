package emu

import (
	"fmt"
	"io"
)

// Locations reserved for vector pointers.
const (
	NMIvector   = 0xFFFA // Non-Maskable Interrupt
	ResetVector = 0xFFFC // Reset
	IRQvector   = 0xFFFE // Interrupt Request
)

type CPU struct {
	bus Bus
	A   uint8
	X   uint8
	Y   uint8
	SP  uint8
	PC  uint16
	P   P

	Clock        int64 // cycles
	targetCycles int64

	t Ticker // tick callback

	disasm *disasm
}

type Ticker interface {
	Tick()
}

// NewCPU creates a new CPU at power-up state.
func NewCPU(bus Bus, ticker Ticker) *CPU {
	cpu := &CPU{
		bus: bus,
		A:   0x00,
		X:   0x00,
		Y:   0x00,
		SP:  0xFD,
		P:   0x00,
		PC:  0x0000,
		t:   ticker,
	}
	return cpu
}

func (c *CPU) SetDisasm(w io.Writer, nestest bool) {
	c.disasm = &disasm{cpu: c, w: w, isNestest: nestest}
}

func (c *CPU) Reset() {
	c.PC = c.Read16(ResetVector)
	c.SP = 0xFD
	c.P = 0x34
}

func (c *CPU) Run(until int64) {
	c.targetCycles = until
	for c.Clock < c.targetCycles {
		opcode := c.Read8(uint16(c.PC))
		c.PC++
		op := ops[opcode]
		if op == nil {
			panic(fmt.Sprintf("unsupported op code %02X (PC:$%04X)", opcode, c.PC))
		}
		op(c)
	}
}

func (c *CPU) RunDisasm(until int64) {
	c.targetCycles = until
	for c.Clock < c.targetCycles {
		c.disasm.loopinit()
		opcode := c.Read8(uint16(c.PC))
		c.PC++
		op := ops[opcode]
		if op == nil {
			panic(fmt.Sprintf("unsupported op code %02X (PC:$%04X)", opcode, c.PC))
		}

		c.disasm.op()
		op(c)
	}
}

func (c *CPU) tick() {
	c.t.Tick()
	c.Clock++
}

func (c *CPU) Read8(addr uint16) uint8 {
	c.tick()
	return c.bus.Read8(addr)
}

func (c *CPU) Write8(addr uint16, val uint8) {
	c.tick()
	c.bus.Write8(addr, val)
}

func (c *CPU) Read16(addr uint16) uint16 {
	lo := c.Read8(addr)
	hi := c.Read8(addr + 1)
	return uint16(hi)<<8 | uint16(lo)
}

func (c *CPU) Write16(addr uint16, val uint16) {
	lo := uint8(val & 0xff)
	hi := uint8(val >> 8)
	c.Write8(addr, lo)
	c.Write8(addr+1, hi)
}

// P is the 6502 Processor Status Register.
type P uint8

func (p *P) clear() {
	// only the unused bit is set
	*p = 0x40
}

const (
	pbitN = 7 - iota // Negative flag
	pbitV            // oVerflow flag
	pbitU            // Unused
	pbitB            // Break flag
	pbitD            // Decimal mode flag
	pbitI            // Interrupt disable flag
	pbitZ            // Zero flag
	pbitC            // Carry flag
)

func (p P) N() bool { return p&(1<<pbitN) != 0 }
func (p P) V() bool { return p&(1<<pbitV) != 0 }
func (p P) B() bool { return p&(1<<pbitB) != 0 }
func (p P) D() bool { return p&(1<<pbitD) != 0 }
func (p P) I() bool { return p&(1<<pbitI) != 0 }
func (p P) Z() bool { return p&(1<<pbitZ) != 0 }
func (p P) C() bool { return p&(1<<pbitC) != 0 }

func (p *P) checkNZ(v uint8) {
	p.writeBit(pbitN, v&0x80 != 0)
	p.writeBit(pbitZ, v == 0)
}

// sets N flag if bit 7 of v is set, clears it otherwise.
func (p *P) checkN(v uint8) {
	p.writeBit(pbitN, v&(1<<7) != 0)
}

// sets Z flag if v == 0, clears it otherwise.
func (p *P) checkZ(v uint8) {
	p.writeBit(pbitZ, v == 0)
}

func (p *P) checkCV(x, y uint8, sum uint16) {
	// forward carry or unsigned overflow.
	p.writeBit(pbitC, sum > 0xFF)

	// signed overflow, can only happen if the sign of the sum differs
	// from that of both operands.
	v := (uint16(x) ^ sum) & (uint16(y) ^ sum) & 0x80
	p.writeBit(pbitV, v != 0)
}

func (p *P) writeBit(i int, v bool) {
	if v {
		p.setBit(i)
	} else {
		p.clearBit(i)
	}
}

func (p *P) setBit(i int) {
	*p |= P(1 << i)
}

func (p *P) clearBit(i int) {
	*p &= ^(1 << i) & 0xff
}

func (p *P) ibit(i int) uint8 {
	return (uint8(*p) & (1 << i)) >> i
}

func (p P) bit(i int) bool {
	return p&(1<<i) != 0
}

func (p P) String() string {
	const bits = "nvubdizcNVUBDIZC"

	s := make([]byte, 8)
	for i := 0; i < 8; i++ {
		s[i] = bits[i+int(8*p.ibit(7-i))]
	}
	return string(s)
}

func b2i(b bool) byte {
	if b {
		return 1
	}
	return 0
}
