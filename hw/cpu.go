package hw

//go:generate go run ./cpugen/gen_nes6502.go -out ./opcodes.go

import (
	"nestor/emu/hwio"
)

// Locations reserved for vector pointers.
const (
	CpuNMIvector   = uint16(0xFFFA) // Non-Maskable Interrupt
	CpuResetVector = uint16(0xFFFC) // Reset
	CpuIRQvector   = uint16(0xFFFE) // Interrupt Request
)

type CPU struct {
	Bus   *hwio.Table
	Ram   [0x800]byte // Internal RAM
	Clock int64       // cycles

	ppu       *PPU
	ppuAbsent bool // allow to disconnect PPU during CPU tests

	// cpu registers
	A, X, Y, SP uint8
	PC          uint16
	P           P

	// interrupt handling
	nmiFlag, prevNmiFlag bool
	needNmi, prevNeedNmi bool
	runIRQ, prevRunIRQ   bool
	irqFlag              bool
}

// NewCPU creates a new CPU at power-up state.
func NewCPU(ppu *PPU) *CPU {
	cpu := &CPU{
		Bus: hwio.NewTable("cpu"),
		A:   0x00,
		X:   0x00,
		Y:   0x00,
		// SP:  0xFD,
		P:   0x00,
		PC:  0x0000,
		ppu: ppu,
	}
	return cpu
}

func (c *CPU) InitBus() {
	// 0x0000-0x1FFF -> RAM, mirrored.
	c.Bus.MapMemorySlice(0x0000, 0x07FF, c.Ram[:], false)
	c.Bus.MapMemorySlice(0x0800, 0x0FFF, c.Ram[:], false)
	c.Bus.MapMemorySlice(0x1000, 0x17FF, c.Ram[:], false)
	c.Bus.MapMemorySlice(0x1800, 0x1FFF, c.Ram[:], false)

	// 0x2000-0x3FFF -> PPU registers, mirrored.
	for i := uint16(0x2000); i < 0x4000; i += 8 {
		c.Bus.MapBank(i, c.ppu, 1)
	}

	// 0x4000-0x4017 -> APU and IO registers.
	// TODO
	// 0x4018-0x401F -> APU and IO registers (test mode).
	// unused

	// 0x4020-0xFFFF -> Cartridge space (PRG-ROM, PRG-RAM, mapper registers).
	// performed by the mapper.
}

func (c *CPU) Reset() {
	c.A = 0x00
	c.X = 0x00
	c.Y = 0x00
	c.SP = 0xFD
	c.P = c.P.SetI(true)
	c.runIRQ = false
	c.Clock = -1
	// Directly read from the bus to prevent clock ticks.
	c.PC = hwio.Read16(c.Bus, CpuResetVector)

	// The CPU takes 8 cycles before it starts executing the ROM's code
	// after a reset/power up
	for i := 0; i < 8; i++ {
		c.tick()
	}
	c.tick()
}

func (c *CPU) setNMIflag()   { c.nmiFlag = true }
func (c *CPU) clearNMIflag() { c.nmiFlag = false }

func (c *CPU) Run(until int64) {
	for c.Clock < until {
		opcode := c.Read8(c.PC)
		c.PC++
		ops[opcode](c)

		if c.prevRunIRQ || c.prevNeedNmi {
			c.IRQ()
		}
	}
}

func (c *CPU) tick() {
	if c.ppuAbsent {
		c.Clock++
		return
	}

	c.ppu.Tick()
	c.ppu.Tick()
	c.ppu.Tick()
	c.Clock++
}

func (c *CPU) Read8(addr uint16) uint8 {
	c.tick()
	val := c.Bus.Read8(addr)
	c.handleInterrupts()
	return val
}

func (c *CPU) Write8(addr uint16, val uint8) {
	c.tick()
	c.Bus.Write8(addr, val)
	c.handleInterrupts()
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

/* stack operations */

func (c *CPU) push8(val uint8) {
	top := uint16(c.SP) + 0x0100
	c.Write8(top, val)
	c.SP -= 1
}

func (c *CPU) push16(val uint16) {
	c.push8(uint8(val >> 8))
	c.push8(uint8(val & 0xff))
}

func (c *CPU) pull8() uint8 {
	c.SP++
	top := uint16(c.SP) + 0x0100
	return c.Read8(top)
}

func (c *CPU) pull16() uint16 {
	lo := c.pull8()
	hi := c.pull8()
	return uint16(hi)<<8 | uint16(lo)
}

/* interrupt handling */

func (c *CPU) handleInterrupts() {
	// The internal signal goes high during φ1 of the cycle that follows the one
	// where the edge is detected and stays high until the NMI has been handled.
	c.prevNeedNmi = c.needNmi

	// This edge detector polls the status of the NMI line during φ2 of each CPU
	// cycle (i.e., during the second half of each cycle) and raises an internal
	// signal if the input goes from being high during one cycle to being low
	// during the next.
	if !c.prevNmiFlag && c.nmiFlag {
		c.needNmi = true
	}
	c.prevNmiFlag = c.nmiFlag

	// It's really the status of the interrupt lines at the end of the
	// second-to-last cycle that matters. Keep the IRQ lines values from the
	// previous cycle. The before-to-last cycle's values will be used.
	c.prevRunIRQ = c.runIRQ
	c.runIRQ = c.irqFlag && !c.P.I()
}

func BRK(cpu *CPU) {
	cpu.tick()

	cpu.push16(cpu.PC + 1)

	p := cpu.P.SetB(true).SetU(true)
	if cpu.needNmi {
		cpu.needNmi = false
		cpu.push8(uint8(p))
		cpu.P = cpu.P.SetI(true)
		cpu.PC = cpu.Read16(CpuNMIvector)
	} else {
		cpu.push8(uint8(p))
		cpu.P = cpu.P.SetI(true)
		cpu.PC = cpu.Read16(CpuIRQvector)
	}

	// Ensure we don't start an NMI right after running a BRK instruction (first
	// instruction in IRQ handler must run first - needed for nmi_and_brk test)
	cpu.prevNeedNmi = false
}

func (c *CPU) IRQ() {
	c.tick()
	c.tick()

	c.push16(c.PC)

	if c.needNmi {
		c.needNmi = false
		p := c.P.SetB(true)
		c.push8(uint8(p))
		c.P = c.P.SetI(true)
		c.PC = c.Read16(CpuNMIvector)
		// TODO inform the debugger we just had an NMI
	} else {
		p := c.P.SetB(true)
		c.push8(uint8(p))
		c.P = c.P.SetI(true)
		c.PC = c.Read16(CpuIRQvector)
		// TODO inform the debugger we just had an IRQ
	}
}
