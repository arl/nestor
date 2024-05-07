package debugger

import (
	"fmt"
	"sync/atomic"

	"gioui.org/layout"

	"nestor/hw"
)

type C = layout.Context
type D = layout.Dimensions

// a debugger holds the state of the CPU debugger. In order to be able to debug
// a program at any moment, the debugger has to keep track of the CPU state,
// even when inactive. The state to keep track of is kept to the minimum, that
// is the current PC and stack frames.
type debugger struct {
	active atomic.Bool
	status atomic.Int32

	cpuBlock  chan debuggerState
	blockAcks chan struct{}

	cpu *hw.CPU

	prevPC     uint16
	prevOpcode uint8

	cstack callStack
}

type debuggerState struct {
	stat status
	pc   uint16
}

func (s debuggerState) String() string {
	str := ""
	switch s.stat {
	case running:
		str = "running"
	case paused:
		str = "paused"
	case stepping:
		str = "stepping"
	}
	return fmt.Sprintf("pc: $%04X, stat: %s", s.pc, str)
}

type status int32

const (
	running status = iota
	paused
	stepping
)

func NewDebugger(cpu *hw.CPU) *debugger {
	dbg := &debugger{
		cpu:       cpu,
		cpuBlock:  make(chan debuggerState),
		blockAcks: make(chan struct{}),
	}
	cpu.SetDebugger(dbg)
	dbg.setStatus(running)
	return dbg
}

func (dbg *debugger) getStatus() status {
	return status(dbg.status.Load())
}

func (dbg *debugger) setStatus(s status) {
	dbg.status.Store(int32(s))
}

func (dbg *debugger) detach() {
	dbg.active.Store(false)
	if dbg.getStatus() != running {
		dbg.blockAcks <- struct{}{}
		dbg.setStatus(running)
	}
}

// Trace must be called before each opcode is executed. This is the main entry
// point for debugging activity, as the debug can stop the CPU execution by
// making this function blocking until user interaction finishes.
func (d *debugger) Trace(pc uint16) {
	d.updateStack(pc, sffNone)
	if !d.active.Load() {
		return
	}

	// disasm := d.cpu.Disasm(pc)
	// d.prevOpSize = uint8(len(disasm.Bytes))
	opcode := d.cpu.Bus.Read8(pc)
	d.prevPC = pc
	d.prevOpcode = opcode

	switch st := d.getStatus(); st {
	case running:
		break
	case paused, stepping:
		d.cpuBlock <- debuggerState{
			pc:   pc,
			stat: st,
		}
		<-d.blockAcks
	}
}

func (d *debugger) updateStack(dstPc uint16, sff stackFrameFlag) {
	switch d.prevOpcode {
	case 0x20: // JSR
		d.cstack.push(d.prevPC, dstPc, (d.prevPC+3)&0xFFFF, sff)
	case 0x40, 0x60: // RTS RTI
		d.cstack.pop()
	}
}

func (d *debugger) FrameEnd() {
	d.cstack.reset()
}

func (d *debugger) Interrupt(prevpc, curpc uint16, isNMI bool) {
	flag := sffIRQ
	if isNMI {
		flag = sffNMI
	}
	d.updateStack(prevpc, flag)
	d.prevOpcode = 0xFF

	d.cstack.push(d.prevPC, curpc, prevpc, flag)
}

func (d *debugger) WatchRead(addr uint16) {

}

func (d *debugger) WatchWrite(addr uint16, val uint16) {

}

// Break can be called by the CPU core to force breaking into the debugger.
func (d *debugger) Break(msg string) {
	d.setStatus(paused)
}
