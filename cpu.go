package main

import (
	"fmt"
	"io"
)

// Reserved locations for vector pointers.
const (
	vecNMI = 0xFFFA // Non-Maskable Interrupt
	vecIRQ = 0xFFFE // Interrupt Request
	vecRES = 0xFFFC // Reset
)

// https://www.nesdev.org/wiki/CPU_memory_map
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

	disasm *disasm
}

// NewCPU creates a new CPU at power-up state.
func NewCPU(bus Bus) *CPU {
	cpu := &CPU{
		bus: bus,
		A:   0x00,
		X:   0x00,
		Y:   0x00,
		SP:  0xFD,
		P:   0x30, // bits 4 and 5 are set at startup.
		PC:  0x0000,
	}
	return cpu
}

func (c *CPU) setDisasm(w io.Writer, nestest bool) {
	c.disasm = &disasm{cpu: c, w: w, isNestest: nestest}
}

func (c *CPU) reset() {
	c.PC = c.Read16(vecRES)
	c.SP = 0xFD
}

func (c *CPU) Run(until int64) {
	c.disasm.loopinit()

	c.targetCycles = until
	for c.Clock < c.targetCycles {
		opcode := c.bus.Read8(uint16(c.PC))
		op := ops[opcode]
		if op == nil {
			panic(fmt.Sprintf("unsupported op code %02X (PC:$%04X)", opcode, c.PC))
		}

		c.disasm.op()
		op(c)
	}
}

func (c *CPU) Read8(addr uint16) uint8 {
	return c.bus.Read8(addr)
}

func (c *CPU) Write8(addr uint16, val uint8) {
	c.bus.Write8(addr, val)
}

func (c *CPU) Read16(addr uint16) uint16 {
	lo := c.bus.Read8(addr)
	hi := c.bus.Read8(addr + 1)
	return uint16(hi)<<8 | uint16(lo)
}

func (c *CPU) Write16(addr uint16, val uint16) {
	lo := uint8(val & 0xff)
	hi := uint8(val >> 8)
	c.bus.Write8(addr, lo)
	c.bus.Write8(addr+1, hi)
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

// sets N flag if bit 7 of v is set, clears it otherwise.
func (p *P) checkN(v uint8) {
	p.writeBit(pbitN, v&(1<<7) != 0)
}

// sets Z flag if v == 0, clears it otherwise.
func (p *P) checkZ(v uint8) {
	p.writeBit(pbitZ, v == 0)
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
