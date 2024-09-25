package hw

//go:generate go run ./cpugen/gen_nes6502.go -out ./opcodes.go

import (
	"io"

	"nestor/emu/hwio"
	"nestor/emu/log"
)

// Locations reserved for vector pointers.
const (
	NMIVector   = uint16(0xFFFA) // Non-Maskable Interrupt
	ResetVector = uint16(0xFFFC) // Reset
	IRQVector   = uint16(0xFFFE) // Interrupt Request
)

type CPU struct {
	Bus   *hwio.Table
	Ram   [0x800]byte // Internal RAM
	Clock int64       // cycles

	selfjumps uint // infinite loop detection: count successive jumps to same PC.
	doHalt    bool

	ppu       *PPU
	ppuAbsent bool // allow to disconnect PPU during CPU tests
	ppuDMA    ppuDMA

	input InputPorts

	// cpu registers
	A, X, Y, SP uint8
	PC          uint16
	P           P

	// interrupt handling
	nmiFlag, prevNmiFlag bool
	needNmi, prevNeedNmi bool
	runIRQ, prevRunIRQ   bool
	irqFlag              bool

	dbg FwdDebugger

	// Non-nil when execution tracing is enabled.
	tracer *tracer
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
	cpu.ppuDMA.cpu = cpu
	return cpu
}

func (c *CPU) PlugInputDevice(ip *InputProvider) {
	c.input.provider = ip
}

func (c *CPU) SetTraceOutput(w io.Writer) {
	c.tracer = &tracer{w: w, d: c}
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

	// Map PPU OAMDMA register.
	c.ppuDMA.InitBus(c.Bus)
	c.Bus.MapBank(0x4014, &c.ppuDMA, 0)

	c.input.initBus()
	c.Bus.MapBank(0x4000, &c.input, 0)

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
	c.P.setIntDisable(true)
	c.runIRQ = false
	c.Clock = -1

	c.ppuDMA.reset()

	// Directly read from the bus to prevent ticking the clock.
	c.PC = hwio.Read16(c.Bus, ResetVector)
	c.dbg.Reset()

	// The CPU takes 8 cycles before it starts executing the ROM's code
	// after a reset/power up
	for i := 0; i < 8; i++ {
		c.tick()
	}
}

func (c *CPU) setNMIflag()   { c.nmiFlag = true }
func (c *CPU) clearNMIflag() { c.nmiFlag = false }

func (c *CPU) Run(ncycles int64) bool {
	until := c.Clock + ncycles
	for c.Clock < until {
		opcode := c.Read8(c.PC)
		if c.tracer != nil {
			c.tracer.write(cpuState{
				A:        c.A,
				X:        c.X,
				Y:        c.Y,
				P:        c.P,
				SP:       c.SP,
				Clock:    c.Clock,
				PPUCycle: c.ppu.Cycle,
				Scanline: c.ppu.Scanline,
				PC:       c.PC,
			})
		}
		c.dbg.Trace(c.PC)
		c.PC++
		ops[opcode](c)

		if c.doHalt {
			log.ModCPU.WarnZ("CPU halted").
				Hex16("PC", c.PC).
				Hex8("opcode", opcode).
				Uint("self jumps", c.selfjumps).
				End()
			return false
		}

		if c.prevRunIRQ || c.prevNeedNmi {
			c.IRQ()
		}
	}

	return true
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
	c.dmaTransfer()
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

/* DMA */

func (c *CPU) dmaTransfer() {
	c.ppuDMA.process(c.Clock)
}

/* interrupt handling */

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
	c.runIRQ = c.irqFlag && !c.P.intDisable()
}

func BRK(cpu *CPU) {
	cpu.tick()

	cpu.push16(cpu.PC + 1)

	p := cpu.P
	p.setB(true)
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
	c.tick()
	c.tick()

	prevpc := c.PC
	c.push16(c.PC)

	if c.needNmi {
		c.needNmi = false
		p := c.P
		p.setB(true)
		c.push8(uint8(p))
		c.P.setIntDisable(true)
		c.PC = c.Read16(NMIVector)
		c.dbg.Interrupt(prevpc, c.PC, true)
	} else {
		p := c.P
		p.setB(true)
		c.push8(uint8(p))
		c.P.setIntDisable(true)
		c.PC = c.Read16(IRQVector)
		c.dbg.Interrupt(prevpc, c.PC, false)
	}
}

func (cpu *CPU) SetDebugger(dbg Debugger) {
	cpu.dbg.fwd = dbg
}

func (cpu *CPU) Disasm(pc uint16) DisasmOp {
	opcode := cpu.Bus.Peek8(pc)
	return disasmOps[opcode](cpu, pc)
}

// FwdDebugger is a no-op Debugger if fwd is nil.
type FwdDebugger struct {
	fwd Debugger
}

func (d *FwdDebugger) Reset() {
	if d.fwd != nil {
		d.Reset()
	}
}
func (d *FwdDebugger) Trace(pc uint16) {
	if d.fwd != nil {
		d.Trace(pc)
	}
}
func (d *FwdDebugger) Interrupt(prevpc, curpc uint16, isNMI bool) {
	if d.fwd != nil {
		d.Interrupt(prevpc, curpc, isNMI)
	}
}
func (d *FwdDebugger) WatchRead(addr uint16) {
	if d.fwd != nil {
		d.WatchRead(addr)
	}
}
func (d *FwdDebugger) WatchWrite(addr uint16, val uint16) {
	if d.fwd != nil {
		d.WatchWrite(addr, val)
	}
}
func (d *FwdDebugger) Break(msg string) {
	if d.fwd != nil {
		d.Break(msg)
	}
}
func (d *FwdDebugger) FrameEnd() {
	if d.fwd != nil {
		d.FrameEnd()
	}
}
