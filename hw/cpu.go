package hw

import (
	"io"

	"nestor/emu/log"
	"nestor/hw/hwio"
	"nestor/hw/input"
)

// Locations reserved for vector pointers.
const (
	NMIVector   = uint16(0xFFFA) // Non-Maskable Interrupt
	ResetVector = uint16(0xFFFC) // Reset
	IRQVector   = uint16(0xFFFE) // Interrupt Request
)

type CPU struct {
	Bus *hwio.Table

	RAM hwio.Mem `hwio:"bank=0,offset=0x0,size=0x800,vsize=0x2000"`

	PPU    *PPU // non-nil when there's a PPU.
	PPUDMA PPUDMA
	APU    *APU

	// Non-nil when execution tracing is enabled.
	tracer *tracer
	dbg    Debugger

	input InputPorts

	Cycles      int64 // CPU cycles
	masterClock int64

	// cpu registers
	A, X, Y, SP uint8
	PC          uint16
	P           P

	// interrupt handling
	nmiFlag, prevNmiFlag bool
	needNmi, prevNeedNmi bool
	runIRQ, prevRunIRQ   bool
	irqFlag              irqSource
	irqMask              irqSource

	halted bool
}

// NewCPU creates a new CPU at power-up state.
func NewCPU(ppu *PPU) *CPU {
	cpu := &CPU{
		Bus: hwio.NewTable("cpu"),
		A:   0x00,
		X:   0x00,
		Y:   0x00,
		SP:  0xFD,
		P:   0x00,
		PC:  0x0000,
		PPU: ppu,
		dbg: nopDebugger{},
	}
	if ppu != nil {
		ppu.CPU = cpu
	}
	return cpu
}

func (c *CPU) PlugInputDevice(ip *input.Provider) {
	c.input.provider = ip
}

func (c *CPU) InitBus() {
	hwio.MustInitRegs(c)
	// CPU internal RAM, mirrored.
	c.Bus.MapBank(0x0000, c, 0)

	// Map the 8 PPU registers (bank 1) from 0x2000 to 0x3FFF.
	for off := uint16(0x2000); off < 0x4000; off += 8 {
		c.Bus.MapBank(off, c.PPU, 1)
	}

	// Map PPU OAMDMA register.
	c.PPUDMA.InitBus(c)
	c.Bus.MapBank(0x4014, &c.PPUDMA, 0)

	c.input.initBus()
	c.Bus.MapBank(0x4000, &c.input, 0)

	if c.APU != nil {
		c.Bus.MapBank(0x4000, c.APU, 0)
		c.Bus.MapBank(0x4000, &c.APU.Square1, 0)
		c.Bus.MapBank(0x4004, &c.APU.Square2, 0)
		c.Bus.MapBank(0x4000, &c.APU.Noise, 0)
		c.Bus.MapBank(0x4000, &c.APU.Triangle, 0)
	}

	var reg4017 reg4017
	hwio.MustInitRegs(&reg4017)
	c.Bus.MapBank(0x4017, &reg4017, 0)
	reg4017.Read = c.input.ReadOUT
	if c.APU != nil {
		reg4017.Write = c.APU.WriteFRAMECOUNTER
	}
}

// Used to disambiguate between:
// - read 0x4017 -> reads controller state (OUT register)
// - write 0x4017 -> writes to APU frame counter.
type reg4017 struct {
	Reg   hwio.Reg8 `hwio:"offset=0,rcb,wcb"`
	Write func(old, val uint8)
	Read  func(old uint8, peek bool) uint8
}

func (r *reg4017) WriteREG(old, val uint8)            { r.Write(old, val) }
func (r *reg4017) ReadREG(old uint8, peek bool) uint8 { return r.Read(old, peek) }

func (c *CPU) Reset(soft bool) {
	if soft {
		c.SP -= 0x03
		c.P.setIntDisable(true)
	} else {
		c.A = 0x00
		c.X = 0x00
		c.Y = 0x00
		c.runIRQ = false

		c.SP = 0xFD
		c.P = 0x00
		c.P.setIntDisable(true)
	}

	c.PPUDMA.reset()

	// Directly read from the bus to avoid side effects.
	c.PC = hwio.Read16(c.Bus, ResetVector)
	c.dbg.Reset()

	c.Cycles = -1
	c.nmiFlag = false
	c.masterClock = ntscCPUDivider

	// After a reset/power up, the CPU takes burns 8 cycles
	// before going on with ROM execution.
	for i := 0; i < 8; i++ {
		c.cycleBegin(true)
		c.cycleEnd(true)
	}
}

func (c *CPU) traceOp() {
	if c.tracer != nil {
		state := cpuState{
			A:     c.A,
			X:     c.X,
			Y:     c.Y,
			P:     c.P,
			SP:    c.SP,
			Clock: c.Cycles,
			PC:    c.PC,
		}
		if c.PPU != nil {
			state.PPUCycle = c.PPU.Cycle
			state.Scanline = c.PPU.Scanline
		}
		c.tracer.write(state)
	}

	c.dbg.Trace(c.PC)
}

func (c *CPU) Run(ncycles int64) {
	until := c.Cycles + ncycles
	var opcode uint8
	for c.Cycles < until {
		opcode = c.Read8(c.PC)
		c.traceOp()
		c.PC++
		ops[opcode](c)

		if c.halted {
			break
		}

		if c.prevRunIRQ || c.prevNeedNmi {
			c.IRQ()
		}
	}

	if c.halted {
		log.ModCPU.WarnZ("CPU halted").
			Hex16("PC", c.PC).
			Hex8("opcode", opcode).
			End()
	}
}

func (c *CPU) halt() {
	c.halted = true
}

func (c *CPU) IsHalted() bool {
	return c.halted
}

const (
	ntscStartClockCount = 6
	ntscEndClockCount   = 6
	ntscCPUDivider      = 12

	ppuOffset = 1
)

func (c *CPU) cycleBegin(forRead bool) {
	if forRead {
		c.masterClock += ntscStartClockCount - 1
	} else {
		c.masterClock += ntscStartClockCount + 1
	}
	c.Cycles++

	if c.PPU != nil {
		c.PPU.Run(uint64(c.masterClock - ppuOffset))
	}
	if c.APU != nil && c.APU.enabled {
		c.APU.Tick()
	}
}

func (c *CPU) cycleEnd(forRead bool) {
	if forRead {
		c.masterClock += ntscEndClockCount + 1
	} else {
		c.masterClock += ntscEndClockCount - 1
	}

	if c.PPU != nil {
		c.PPU.Run(uint64(c.masterClock - ppuOffset))
	}

	c.handleInterrupts()
}

func (c *CPU) Read8(addr uint16) uint8 {
	c.dmaTransfer()
	c.cycleBegin(true)
	val := c.Bus.Read8(addr, false)
	c.cycleEnd(true)
	return val
}

func (c *CPU) Write8(addr uint16, val uint8) {
	c.cycleBegin(false)
	c.Bus.Write8(addr, val)
	c.cycleEnd(false)
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

/* DMA */

func (c *CPU) dmaTransfer() {
	c.PPUDMA.process()
}

/* interrupt handling */

type irqSource uint8

const (
	external irqSource = 1 << iota
	frameCounter
	dmc
)

func (c *CPU) setIrqSource(src irqSource)      { c.irqFlag |= src }
func (c *CPU) hasIrqSource(src irqSource) bool { return (c.irqFlag & src) != 0 }
func (c *CPU) clearIrqSource(src irqSource)    { c.irqFlag &= ^src }

func (c *CPU) setNMIflag()   { c.nmiFlag = true }
func (c *CPU) clearNMIflag() { c.nmiFlag = false }

func (c *CPU) handleInterrupts() {
	// The internal signal goes high during φ1 of the cycle that follows the one
	// where the edge is detected and stays high until the NMI has been handled.
	c.prevNeedNmi = c.needNmi

	// This edge detector polls the status of the NMI line during φ2 of each CPU
	// cycle (i.e. during the second half of each cycle) and raises an internal
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
	c.runIRQ = c.irqFlag != 0 && !c.P.intDisable()
}

func BRK(cpu *CPU) {
	// dummy read.
	_ = cpu.Read8(cpu.PC)

	cpu.push16(cpu.PC + 1)

	p := cpu.P
	p.setBrk(true)
	p.setUnused(true)
	if cpu.needNmi {
		cpu.needNmi = false
		cpu.push8(uint8(p))
		cpu.P.setIntDisable(true)
		cpu.PC = cpu.Read16(NMIVector)
	} else {
		cpu.push8(uint8(p))
		cpu.P.setIntDisable(true)
		cpu.PC = cpu.Read16(IRQVector)
	}

	// Ensure we don't start an NMI right after running a BRK instruction (first
	// instruction in IRQ handler must run first - needed for nmi_and_brk test)
	cpu.prevNeedNmi = false
}

func (c *CPU) IRQ() {
	c.Read8(c.PC) // dummy reads
	c.Read8(c.PC)

	prevpc := c.PC
	c.push16(c.PC)

	if c.needNmi {
		c.needNmi = false
		p := c.P
		p.setBrk(true)
		c.push8(uint8(p))

		c.P.setIntDisable(true)
		c.PC = c.Read16(NMIVector)
		c.dbg.Interrupt(prevpc, c.PC, true)
	} else {
		p := c.P
		p.setUnused(true)
		c.push8(uint8(p))

		c.P.setIntDisable(true)
		c.PC = c.Read16(IRQVector)
		c.dbg.Interrupt(prevpc, c.PC, false)
	}
}

/* tracing / debugging */

func (c *CPU) SetTraceOutput(w io.Writer) {
	c.tracer = &tracer{w: w, d: c}
}

func (cpu *CPU) SetDebugger(dbg Debugger) {
	cpu.dbg = dbg
}

func (cpu *CPU) Disasm(pc uint16) DisasmOp {
	opcode := cpu.Bus.Peek8(pc)
	return disasmOps[opcode](cpu, pc)
}

type nopDebugger struct{}

func (nopDebugger) Reset()                                     {}
func (nopDebugger) Trace(pc uint16)                            {}
func (nopDebugger) Interrupt(prevpc, curpc uint16, isNMI bool) {}
func (nopDebugger) WatchRead(addr uint16)                      {}
func (nopDebugger) WatchWrite(addr uint16, val uint16)         {}
func (nopDebugger) Break(msg string)                           {}
func (nopDebugger) FrameEnd()                                  {}
