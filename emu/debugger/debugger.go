package debugger

import (
	"fmt"
	"image"
	"sync/atomic"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"nestor/hw"
	"nestor/ui"
)

type C = layout.Context
type D = layout.Dimensions

// a debugger is always present and associated to the CPU when running, however
// it is only active when the debugger window is opened.
type debugger struct {
	active atomic.Bool
	status atomic.Int32

	cpuBlock  chan status
	blockAcks chan struct{}

	cpu *hw.CPU

	prevPC     uint16
	prevOpcode uint8

	cstack callStack
}

type status struct {
	stat   dbgStatus
	pc     uint16
	frames []frameInfo
}

func (s status) String() string {
	str := ""
	switch s.stat {
	case running:
		str = "running"
	case paused:
		str = "paused"
	case stepping:
		str = "stepping"
	}
	return fmt.Sprintf("pc: $%04X, stat: %s, frames: %+v", s.pc, str, s.frames)
}

type dbgStatus int32

const (
	running dbgStatus = iota
	paused
	stepping
)

func NewDebugger(cpu *hw.CPU) *debugger {
	dbg := &debugger{
		cpu:       cpu,
		cpuBlock:  make(chan status),
		blockAcks: make(chan struct{}),
	}
	cpu.SetDebugger(dbg)
	dbg.setStatus(running)
	return dbg
}

func (dbg *debugger) getStatus() dbgStatus {
	return dbgStatus(dbg.status.Load())
}

func (dbg *debugger) setStatus(s dbgStatus) {
	dbg.status.Store(int32(s))
}

func (dbg *debugger) detach() {
	dbg.active.Store(false)
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
		return
	case paused, stepping:
		sta := status{
			pc:     pc,
			stat:   st,
			frames: d.cstack.build(pc),
		}
		d.cpuBlock <- sta
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

type DebuggerWindow struct {
	dbg *debugger

	theme *material.Theme

	csviewer callstackViewer
	ptviewer patternsTable

	start widget.Clickable
	pause widget.Clickable
	step  widget.Clickable
}

func NewDebuggerWindow(dbg hw.Debugger, ppu *hw.PPU) *DebuggerWindow {
	return &DebuggerWindow{
		dbg:      dbg.(*debugger),
		ptviewer: patternsTable{ppu: ppu},
		theme:    material.NewTheme(),
	}
}

func (dw *DebuggerWindow) Run(w *ui.Window) error {
	defer dw.dbg.detach()

	var ops op.Ops

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	dw.dbg.active.Store(true)

	go func() {
		<-w.App.Context.Done()
		w.Perform(system.ActionClose)
	}()

	events := make(chan event.Event)
	acks := make(chan struct{})

	go func() {
		for {
			ev := w.NextEvent()
			events <- ev
			<-acks
			if _, ok := ev.(app.DestroyEvent); ok {
				return
			}
		}
	}()

	stat := status{stat: running}
	for {
		select {
		case stat = <-dw.dbg.cpuBlock:
			w.Invalidate()
		case e := <-events:
			switch e := e.(type) {
			case app.DestroyEvent:
				acks <- struct{}{}
				return e.Err
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				switch stat.stat {
				case running:
					if dw.pause.Clicked(gtx) {
						dw.dbg.setStatus(paused)
					}
				case paused:
					if dw.start.Clicked(gtx) {
						dw.dbg.setStatus(running)
						stat.stat = running
						dw.dbg.blockAcks <- struct{}{}
					}
					if dw.step.Clicked(gtx) {
						dw.dbg.setStatus(stepping)
						stat.stat = stepping
						dw.dbg.blockAcks <- struct{}{}
					}
				case stepping:
					if dw.start.Clicked(gtx) {
						dw.dbg.setStatus(running)
						stat.stat = running
						dw.dbg.blockAcks <- struct{}{}
					}
					if dw.step.Clicked(gtx) {
						dw.dbg.setStatus(stepping)
						stat.stat = stepping
						dw.dbg.blockAcks <- struct{}{}
					}
				}

				dw.Layout(w, stat, gtx)
				e.Frame(gtx.Ops)
			}
			acks <- struct{}{}
		}
	}
}

func (dw *DebuggerWindow) Layout(w *ui.Window, status status, gtx C) {
	btnSize := layout.Exact(image.Point{X: 70, Y: 35})
	// listing := &listing{nes: dw.nes, list: &dw.list}

	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEnd, Alignment: layout.Start}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					gtx.Constraints = btnSize
					if status.stat == running {
						gtx = gtx.Disabled()
					}
					return material.Button(dw.theme, &dw.start, "Start").Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D { return layout.Spacer{Width: 5}.Layout(gtx) }),

				layout.Rigid(func(gtx C) D {
					gtx.Constraints = btnSize
					if status.stat != running {
						gtx = gtx.Disabled()
					}
					return material.Button(dw.theme, &dw.pause, "Pause").Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D { return layout.Spacer{Width: 5}.Layout(gtx) }),
				layout.Rigid(func(gtx C) D {
					gtx.Constraints = btnSize
					if status.stat == running {
						gtx = gtx.Disabled()
					}
					return material.Button(dw.theme, &dw.step, "Step").Layout(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx C) D {
			return material.H6(dw.theme, "Patterns table").Layout(gtx)
		}),
		layout.Flexed(1, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEnd}.Layout(gtx,
				layout.Rigid(dw.ptviewer.Layout),
				layout.Rigid(func(gtx C) D {
					return dw.csviewer.Layout(dw.theme, gtx, status)
				}),
				// layout.Flexed(1, listing.Layout),
			)
		}),
	)
}
