package hw

//go:generate go run ./cpugen/gen_nes6502.go -out ./opcodes.go

import (
	"bytes"
	"fmt"
	"io"
	"nestor/emu/hwio"
	"strings"
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

	ppu       *PPU
	ppuAbsent bool // allow to disconnect PPU during CPU tests
	ppuDMA    ppuDMA

	// cpu registers
	A, X, Y, SP uint8
	PC          uint16
	P           P

	// interrupt handling
	nmiFlag, prevNmiFlag bool
	needNmi, prevNeedNmi bool
	runIRQ, prevRunIRQ   bool
	irqFlag              bool

	dbg Debugger

	// Non-nil when execution tracing is enabled.
	tracer *tracer
}

// NewCPU creates a new CPU at power-up state.
func NewCPU(ppu *PPU) *CPU {
	return &CPU{
		Bus: hwio.NewTable("cpu"),
		A:   0x00,
		X:   0x00,
		Y:   0x00,
		// SP:  0xFD,
		P:   0x00,
		PC:  0x0000,
		ppu: ppu,
	}
}

func (c *CPU) SetTraceOutput(w io.Writer) {
	c.tracer = &tracer{w: w, cpu: c}
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
	c.ppuDMA.InitBus(c.Bus, c.ppu.oamMem[:])
	c.Bus.MapBank(0x4014, &c.ppuDMA, 0)

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
	c.tick()
}

func (c *CPU) setNMIflag()   { c.nmiFlag = true }
func (c *CPU) clearNMIflag() { c.nmiFlag = false }

func (c *CPU) Run(ncycles int64) {
	until := c.Clock + ncycles
	for c.Clock < until {
		opcode := c.Read8(c.PC)

		if c.tracer != nil {
			c.tracer.write()
		}
		c.dbg.Trace(c.PC)
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

	c.processDMA()

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

/* DMA */

func (c *CPU) processDMA() {
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
	cpu.dbg = dbg
}

func (cpu *CPU) Disasm(pc uint16) DisasmOp {
	opcode := cpu.Bus.Peek8(pc)
	return disasmOps[opcode](cpu, pc)
}

// cpuState stores the CPU state for the execution trace.
type cpuState struct {
	A, X, Y uint8
	P       P
	SP      uint8
	PC      uint16

	Clock    int64
	PPUCycle int
	Scanline int
}

type tracer struct {
	cpu *CPU
	w   io.Writer
	buf bytes.Buffer
}

// write the execution trace for current cycle.
func (t *tracer) write() {
	state := cpuState{
		A:        t.cpu.A,
		X:        t.cpu.X,
		Y:        t.cpu.Y,
		P:        t.cpu.P,
		SP:       t.cpu.SP,
		Clock:    t.cpu.Clock,
		PPUCycle: t.cpu.ppu.Cycle,
		Scanline: t.cpu.ppu.Scanline,
		PC:       t.cpu.PC,
	}

	dis := t.cpu.Disasm(state.PC)
	fmt.Fprintf(&t.buf, "%-30s A:%02X X:%02X Y:%02X P:%02X SP:%02X PPU:%3d,%3d CYC:%d\n",
		dis.String(), state.A, state.X, state.Y, byte(state.P), state.SP,
		state.Scanline, state.PPUCycle, state.Clock)

	t.buf.WriteTo(t.w) // WriteTo also resets the buffer.
}

type DisasmOp struct {
	Opcode string
	Oper   string
	Bytes  []byte
	PC     uint16
}

func (d DisasmOp) String() string {
	// C000  4C F5 C5  JMP $C5F5
	var sb strings.Builder
	fmt.Fprintf(&sb, "%04X ", d.PC)
	for _, b := range d.Bytes {
		fmt.Fprintf(&sb, " %02X ", b)
	}
	fmt.Fprintf(&sb, "%*s", 17-sb.Len(), "")
	sb.WriteString(d.Opcode)
	sb.WriteByte(' ')
	sb.WriteString(d.Oper)
	return sb.String()
}
