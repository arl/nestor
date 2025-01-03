package hw

import (
	"io"

	"nestor/emu/log"
	"nestor/hw/hwdefs"
	"nestor/hw/hwio"
	"nestor/hw/input"
)

// Locations reserved for vector pointers.
const (
	nmiVector   = uint16(0xfffa) // Non-Maskable Interrupt
	resetVector = uint16(0xfffc) // Reset
	irqVector   = uint16(0xfffe) // Interrupt Request
)

type CPU struct {
	Bus *hwio.Table

	RAM hwio.Mem `hwio:"bank=0,offset=0x0,size=0x800,vsize=0x2000"`

	PPU *PPU // non-nil when there's a PPU.
	DMA DMA
	APU *APU

	// Non-nil when execution tracing is enabled.
	tracer *tracer
	dbg    Debugger

	input InputPorts

	halted      bool
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
	irqFlag              hwdefs.IRQSource
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

type nopDebugger struct{}

func (nopDebugger) Reset()                                     {}
func (nopDebugger) Trace(pc uint16)                            {}
func (nopDebugger) Interrupt(prevpc, curpc uint16, isNMI bool) {}
func (nopDebugger) WatchRead(addr uint16)                      {}
func (nopDebugger) WatchWrite(addr uint16, val uint16)         {}
func (nopDebugger) Break(msg string)                           {}
func (nopDebugger) FrameEnd()                                  {}

func (c *CPU) PlugInputDevice(ip *input.Provider) {
	c.input.provider = ip
}

func (c *CPU) InitBus() {
	hwio.MustInitRegs(c)
	// CPU internal RAM, mirrored.
	c.Bus.MapBank(0x0000, c, 0)

	// Map the 8 PPU registers (bank 1) from 0x2000 to 0x3ffF.
	for off := uint16(0x2000); off < 0x4000; off += 8 {
		c.Bus.MapBank(off, c.PPU, 1)
	}

	// Map PPU OAMDMA register.
	c.DMA.InitBus(c)
	c.Bus.MapBank(0x4014, &c.DMA, 0)

	c.input.initBus()
	c.Bus.MapBank(0x4000, &c.input, 0)

	if c.APU != nil {
		c.Bus.MapBank(0x4000, c.APU, 0)
		c.Bus.MapBank(0x4000, &c.APU.Square1, 0)
		c.Bus.MapBank(0x4004, &c.APU.Square2, 0)
		c.Bus.MapBank(0x4000, &c.APU.Noise, 0)
		c.Bus.MapBank(0x4000, &c.APU.Triangle, 0)
		c.Bus.MapBank(0x4000, &c.APU.DMC, 0)
	}

	var reg4017 reg4017
	hwio.MustInitRegs(&reg4017)
	c.Bus.MapBank(0x4017, &reg4017, 0)
	reg4017.Read = c.input.ReadOUT
	reg4017.Peek = c.input.PeekOUT
	if c.APU != nil {
		reg4017.Write = c.APU.frameCounter.WriteFRAMECOUNTER
	}
}

// Used to disambiguate between:
// - read 0x4017 -> reads controller state (OUT register)
// - write 0x4017 -> writes to APU frame counter.
type reg4017 struct {
	Reg   hwio.Reg8 `hwio:"offset=0,pcb=Peek4017,rcb=Read4017,wcb=Write4017"`
	Write func(old, val uint8)
	Read  func(old uint8) uint8
	Peek  func(old uint8) uint8
}

func (r *reg4017) Peek4017(old uint8) uint8 { return r.Peek(old) }
func (r *reg4017) Read4017(old uint8) uint8 { return r.Read(old) }
func (r *reg4017) Write4017(old, val uint8) { r.Write(old, val) }

func (c *CPU) Reset(soft bool) {
	if soft {
		c.SP -= 0x03
	} else {
		c.A = 0x00
		c.X = 0x00
		c.Y = 0x00
		c.runIRQ = false

		c.SP = 0xFD
		c.P = 0x00
	}
	c.P.setFlags(Interrupt)

	c.DMA.reset()

	// Directly read from the bus to avoid side effects.
	c.PC = hwio.Read16(c.Bus, resetVector)
	c.dbg.Reset()

	c.Cycles = -1
	c.nmiFlag = false
	c.irqFlag = 0
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

// branch is the based for branch opcodes. jump to dst if P&flag == val
func (cpu *CPU) branch(dst uint16, flag, val P) {
	if cpu.P&flag == val {
		return // no branch
	}
	// A taken, non-page-crossing branch ignores IRQ/NMI during its last clock,
	// so that next instruction executes before the IRQ. Fixes
	// 'branch_delays_irq' test.
	if cpu.runIRQ && !cpu.prevRunIRQ {
		cpu.runIRQ = false
	}

	// dummy read.
	_ = cpu.Read8(cpu.PC)

	// extra cycle for page cross
	if 0xff00&(cpu.PC) != 0xff00&(dst) {
		// dummy read.
		_ = cpu.Read8(cpu.PC&0xff00 | dst&0x00ff)
	}

	cpu.PC = dst
}

func JSR(cpu *CPU) {
	pclo := cpu.fetch8()

	// dummy read.
	_ = cpu.Read8(uint16(cpu.SP) + 0x0100)
	cpu.PC++

	cpu.push16(cpu.PC - 1)
	pchi := cpu.Read8(cpu.PC - 1)
	cpu.PC = uint16(pchi)<<8 | uint16(pclo)
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

func (c *CPU) CurrentCycle() int64 {
	return c.Cycles
}

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

func (c *CPU) fetch8() uint8 {
	val := c.Read8(c.PC)
	c.PC++
	return val
}

func (c *CPU) fetch16() uint16 {
	lo := c.fetch8()
	hi := c.fetch8()
	return uint16(hi)<<8 | uint16(lo)
}

func (c *CPU) Read8(addr uint16) uint8 {
	c.DMA.processPending(addr)
	c.cycleBegin(true)
	defer c.cycleEnd(true)

	return c.Bus.Read8(addr)
}

func (c *CPU) Write8(addr uint16, val uint8) {
	c.cycleBegin(false)
	defer c.cycleEnd(false)

	c.Bus.Write8(addr, val)
}

func (c *CPU) Read16(addr uint16) uint16 {
	lo := c.Read8(addr)
	hi := c.Read8(addr + 1)
	return uint16(hi)<<8 | uint16(lo)
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

/* DMC */

func (c *CPU) StartDMCTransfer() {
	c.DMA.startDMCTransfer()
}

func (c *CPU) StopDMCTransfer() {
	c.DMA.stopDMCTransfer()
}

/* interrupt handling */

func (c *CPU) SetIrqSource(src hwdefs.IRQSource) {
	log.ModCPU.DebugZ("set IRQ source").
		Stringer("src", src).
		Stringer("prev", c.irqFlag).
		Stringer("new", c.irqFlag|src).
		End()

	c.irqFlag |= src
}

func (c *CPU) HasIrqSource(src hwdefs.IRQSource) bool {
	return (c.irqFlag & src) != 0
}

func (c *CPU) ClearIrqSource(src hwdefs.IRQSource) {
	log.ModCPU.DebugZ("clear IRQ source").
		Stringer("src", src).
		Stringer("prev", c.irqFlag).
		Stringer("new", c.irqFlag&^src).
		End()

	c.irqFlag &= ^src
}

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
	c.runIRQ = c.irqFlag != 0 && !c.P.hasFlag(Interrupt)
}

func BRK(cpu *CPU) {
	// dummy read.
	_ = cpu.Read8(cpu.PC)

	cpu.push16(cpu.PC + 1)

	p := cpu.P
	p.setFlags(Break | Reserved)
	if cpu.needNmi {
		cpu.needNmi = false
		cpu.push8(uint8(p))
		cpu.P.setFlags(Interrupt)
		cpu.PC = cpu.Read16(nmiVector)
	} else {
		cpu.push8(uint8(p))
		cpu.P.setFlags(Interrupt)
		cpu.PC = cpu.Read16(irqVector)
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
		p.setFlags(Break)
		c.push8(uint8(p))

		c.P.setFlags(Interrupt)
		c.PC = c.Read16(nmiVector)
		c.dbg.Interrupt(prevpc, c.PC, true)
	} else {
		p := c.P
		p.setFlags(Reserved)
		c.push8(uint8(p))

		c.P.setFlags(Interrupt)
		c.PC = c.Read16(irqVector)
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

/* helpers for opcodes */

func (cpu *CPU) setreg(reg *uint8, val uint8) {
	cpu.P.clearFlags(Zero | Negative)
	cpu.P.setNZ(val)
	*reg = val
}

// generalized add with carry and overflow.
func (cpu *CPU) add(val uint8) {
	var carry uint16
	if cpu.P.hasFlag(Carry) {
		carry = 1
	}

	sum := uint16(cpu.A) + uint16(val) + uint16(carry)
	cpu.P.clearFlags(Carry | Overflow)

	// signed overflow, can only happen if the sign of the sum differs
	// from that of both operands.
	v := (uint16(cpu.A) ^ sum) & (uint16(val) ^ sum) & 0x80
	if v != 0 {
		cpu.P.setFlags(Overflow)
	}
	if sum > 0xff {
		cpu.P.setFlags(Carry)
	}
	cpu.A = uint8(sum)
	cpu.P.clearFlags(Zero | Negative)
	cpu.P.setNZ(cpu.A)
}

func pageCrossed(a uint16, b uint8) bool {
	return ((a + uint16(b)) & 0xff00) != (a & 0xff00)
}

/* addressing modes */

func (cpu *CPU) acc() {
	_ = cpu.Read8(cpu.PC) // dummy read.
}

func (cpu *CPU) imp() {
	_ = cpu.Read8(cpu.PC) // dummy read.
}

func (cpu *CPU) ind() uint16 {
	addr := cpu.Read16(cpu.PC)

	// 2 bytes address wrap around
	lo := cpu.Read8(addr)
	hi := cpu.Read8((0xff00 & addr) | (0x00ff & (addr + 1)))
	return uint16(hi)<<8 | uint16(lo)
}

func (cpu *CPU) rel() uint16 {
	off := int16(int8(cpu.fetch8()))
	return uint16(int16(cpu.PC) + off)
}

func (cpu *CPU) abs() uint16 {
	return cpu.fetch16()
}

func (cpu *CPU) abx(dummyread bool) uint16 {
	addr := cpu.fetch16()
	operand := addr + uint16(cpu.X)
	crossed := pageCrossed(addr, cpu.X)

	if crossed || dummyread {
		var off uint16
		if crossed {
			off = 0x100
		}
		_ = cpu.Read8(operand - off) // dummy read.
	}
	return operand
}

func (cpu *CPU) aby(dummyread bool) uint16 {
	addr := cpu.fetch16()
	operand := addr + uint16(cpu.Y)
	crossed := pageCrossed(addr, cpu.Y)

	if crossed || dummyread {
		var off uint16
		if crossed {
			off = 0x100
		}
		_ = cpu.Read8(operand - off) // dummy read.
	}
	return operand
}

func (cpu *CPU) zpg() uint16 {
	return uint16(cpu.fetch8())
}

func (cpu *CPU) zpx() uint16 {
	addr := cpu.fetch8()
	_ = cpu.Read8(uint16(addr)) // dummy read.

	return (uint16(addr) + uint16(cpu.X)) & 0xff
}

func (cpu *CPU) zpy() uint16 {
	addr := cpu.fetch8()
	_ = cpu.Read8(uint16(addr)) // dummy read.

	return (uint16(addr) + uint16(cpu.Y)) & 0xff
}

func (cpu *CPU) izx() uint16 {
	addr := uint16(cpu.fetch8())
	_ = cpu.Read8(addr) // dummy read.

	addr = uint16(uint8(addr) + cpu.X)

	// read 16 bytes from the zero page, handling page wrap
	lo := cpu.Read8(addr)
	hi := cpu.Read8(uint16(uint8(addr) + 1))
	return uint16(hi)<<8 | uint16(lo)
}

func (cpu *CPU) izy(dummyread bool) uint16 {
	addr := uint16(cpu.fetch8())

	// read 16 bytes from the zero page, handling page wrap
	lo := cpu.Read8(addr)
	hi := cpu.Read8(uint16(uint8(addr) + 1))
	addr = uint16(hi)<<8 | uint16(lo)

	operand := addr + uint16(cpu.Y)
	crossed := pageCrossed(addr, cpu.Y)

	if crossed || dummyread {
		var off uint16
		if crossed {
			off = 0x100
		}
		_ = cpu.Read8(operand - off) // dummy read.
	}
	return operand
}
