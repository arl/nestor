package main

import (
	"fmt"

	"nestor/ines"
)

// https://www.nesdev.org/wiki/CPU_memory_map
type CPU struct {
	bus *Bus
	A   uint8
	X   uint8
	Y   uint8
	SP  uint8
	PC  uint16
	P   P

	Clock        int64 // cycles
	targetCycles int64
}

func NewCPU() *CPU {
	return &CPU{
		bus: NewBus("cpu"),
		A:   0x00,
		X:   0x00,
		Y:   0x00,
		SP:  0xFD,
		P:   0x34,
		PC:  0x0000,
	}
}

func (c *CPU) MapMemory(cart *ines.Rom) {
	// RAM is 0x800 bytes, mirrored.
	ram := make([]byte, 0x0800)
	c.bus.MapSlice(0x0000, 0x07FF, ram)
	c.bus.MapSlice(0x0800, 0x0FFF, ram)
	c.bus.MapSlice(0x1000, 0x17FF, ram)
	c.bus.MapSlice(0x1800, 0x1FFF, ram)
}

func (c *CPU) reset() {
	// Load reset vector.
	c.PC = c.Read16(0xFFFC)
	c.SP = 0xFD
}

const disasm = true

func (c *CPU) Run(until int64) {
	prevP := c.P
	prevClock := c.Clock
	c.targetCycles = until
	for c.Clock < c.targetCycles {
		op := c.bus.Read8(uint16(c.PC))
		f := opCodes[op]
		if f == nil {
			panic(fmt.Sprintf("unsupported op code %02X (PC:$%04X)", op, c.PC))
		}

		if disasm {
			fmt.Printf("%X    %s  (%d)\n", c.PC, disasmCodes[op](c), c.Clock-prevClock)
			fmt.Printf("A:%02X X:%02X Y:%02X SP:%02X\n", c.A, c.X, c.Y, c.SP)
			if prevP != c.P {
				fmt.Printf("P:%s\n", c.P)
			}
			prevP = c.P
			prevClock = c.Clock
		}
		f(c)
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

// P is the 6502  Processor Status Register
// doc https://codebase64.org/doku.php?id=base:6502_registers
type P uint8

func (p *P) clear() {
	// only the unused bit is set
	*p = 0x40
}

const (
	pbitN = 7 - iota // Negative flag
	pbitV            // oVerflow flag
	_                // unused
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

// sets N flag if bit 7 of v is set
func (p *P) maybeSetN(v uint8) {
	if v&(1<<7) != 0 {
		*p |= P(1 << pbitN)
	} else {
		*p &= ^(1 << pbitN) & 0xff
	}
}

// sets Z flag if v is 0
func (p *P) maybeSetZ(v uint8) {
	if v == 0 {
		*p |= P(1 << pbitZ)
	} else {
		*p &= ^(1 << pbitZ) & 0xff
	}
}

func (p P) String() string {
	return fmt.Sprintf("0x%x N%d V%d _ B%d D%d I%d Z%d C%d", uint8(p),
		b2i(p.N()), b2i(p.V()), b2i(p.B()), b2i(p.D()), b2i(p.I()), b2i(p.Z()), b2i(p.C()))
}

func b2i(b bool) byte {
	if b {
		return 1
	}
	return 0
}
