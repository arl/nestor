package main

import (
	"fmt"

	"nestor/ines"
)

type CPU struct {
	bus Bus

	A  uint8
	X  uint8
	Y  uint8
	SP uint8
	PC uint16
	P  P
}

func NewCPU() *CPU {
	return &CPU{
		A:  0x00,
		X:  0x00,
		Y:  0x00,
		SP: 0xFD,
		P:  0x34,
		PC: 0x0000,
	}
}

func (c *CPU) LoadCartridge(cart *ines.Rom) {
	c.reset()
}

func (c *CPU) reset() {
	// Load the reset vector.
	pclo := uint16(c.bus.Read8(0xFFFC))
	pchi := uint16(c.bus.Read8(0xFFFD))
	c.PC = pchi<<8 | pclo
}

type P uint8

func (p *P) clear() {
	// only the unused bit is set
	*p = 0x40
}

func (p P) N() bool      { return p&0x80 != 0 } // Negative flag
func (p P) V() bool      { return p&0x40 != 0 } // oVerflow flag
func (p P) unused() bool { return true }        // always 1
func (p P) B() bool      { return p&0x10 != 0 } // Break flag
func (p P) D() bool      { return p&0x08 != 0 } // Decimal mode flag
func (p P) I() bool      { return p&0x04 != 0 } // Interrupt disable flag
func (p P) Z() bool      { return p&0x02 != 0 } // Zero flag
func (p P) C() bool      { return p&0x01 != 0 } // Carry flag

func (p P) String() string {
	return fmt.Sprintf("0x%x N:%d V:%d B:%d D:%d I:%d Z:%d C:%d\n", uint8(p),
		b2i(p.N()), b2i(p.V()), b2i(p.B()), b2i(p.D()), b2i(p.I()), b2i(p.Z()), b2i(p.C()))
}

func b2i(b bool) byte {
	if b {
		return 1
	}
	return 0
}
