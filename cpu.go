package main

import (
	"encoding/hex"
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
	c.PC = uint16(c.bus.Read8(0xFFFD))<<8 | uint16(c.bus.Read8(0xFFFC))
	c.SP = 0xFD
}

func (c *CPU) Run() {
	buf := make([]byte, 32)
	for i := range buf {
		buf[i] = c.bus.Read8(uint32(c.PC))
		c.PC++
	}

	fmt.Println(hex.Dump(buf[:]))
}

// P is the 6502  Processor Status Register
// doc https://codebase64.org/doku.php?id=base:6502_registers
type P uint8

func (p *P) clear() {
	// only the unused bit is set
	*p = 0x40
}

const (
	pN P = 0x80 >> iota // Negative flag
	pV                  // oVerflow flag
	_                   // unused
	pB                  // Break flag
	pD                  // Decimal mode flag
	pI                  // Interrupt disable flag
	pZ                  // Zero flag
	pC                  // Carry flag

)

func (p P) N() bool { return p&pN != 0 }
func (p P) V() bool { return p&pV != 0 }
func (p P) B() bool { return p&pB != 0 }
func (p P) D() bool { return p&pD != 0 }
func (p P) I() bool { return p&pI != 0 }
func (p P) Z() bool { return p&pZ != 0 }
func (p P) C() bool { return p&pC != 0 }

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
