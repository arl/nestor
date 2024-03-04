package main

import (
	"image"
	"sync/atomic"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"nestor/hw"
	"nestor/ui"
)

// a debugger is always present and associated to the CPU when running, however
// it is only active when the debugger window is opened.
type debugger struct {
	active   atomic.Bool
	state    atomic.Int32
	cpuBlock chan struct{}

	cpu *hw.CPU

	cstack callStack
}

type dbgState int32

const (
	running dbgState = iota
	paused
	stepping
)

func newDebugger(cpu *hw.CPU) *debugger {
	dbg := &debugger{
		cpu:      cpu,
		cpuBlock: make(chan struct{}),
	}
	cpu.SetDebugger(dbg)
	dbg.setState(running)
	return dbg
}

func (dbg *debugger) getState() dbgState {
	return dbgState(dbg.state.Load())
}

func (dbg *debugger) setState(s dbgState) {
	dbg.state.Store(int32(s))
}

func (dbg *debugger) unblock() {
	dbg.cpuBlock <- struct{}{}
}

func (dbg *debugger) detach() {
	dbg.active.Store(false)
	dbg.cpuBlock <- struct{}{}
}

// Trace must be called before each opcode is executed. This is the main entry
// point for debugging activity, as the debug can stop the CPU execution by
// making this function blocking until user interaction finishes.
func (d *debugger) Trace(pc uint16) {
	d.updateStack(pc, sffNone)
	if !d.active.Load() {
		return
	}
	switch d.getState() {
	case running:
		return
	case paused:
		<-d.cpuBlock
	case stepping:
		<-d.cpuBlock
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
func (d *debugger) Interrupt(prevpc, curpc uint16, isNMI bool) {
}
func (d *debugger) WatchRead(addr uint16) {
}
func (d *debugger) WatchWrite(addr uint16, val uint16) {
}
// Break can be called by the CPU core to force breaking into the debugger.
func (d *debugger) Break(msg string) {
}
type DebuggerWindow struct {
	nes     *NES
	emu     *emulator
	dbg     *debugger
	addLine chan string
	lines   []string

	callstack callStackViewer

	start widget.Clickable
	pause widget.Clickable
	step  widget.Clickable

	list widget.List
}

func NewDebuggerWindow(emu *emulator) *DebuggerWindow {
	return &DebuggerWindow{
		emu:     emu,
		nes:     emu.nes,
		dbg:     emu.nes.Debugger,
		addLine: make(chan string, 100),
		list:    widget.List{List: layout.List{Axis: layout.Vertical}},
ackViewer(),
	}
}

func (dw *Debugg	}
}

func (dw *DebuggerWindow) Run(w *ui.Window) error {
	defer dw.dbg.detach()

	var ops op.Ops

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	dw.dbg.active.Store(true)
	dw.dbg.setState(paused)

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
	for {
		select {
		// TODO: continue here by removing this and adding the disassembly

		// listen to new lines from Printf and add them to our lines.
		case line := <-dw.addLine:
			dw.lines = append(dw.lines, line)
			w.Invalidate()
		case e := <-events:
			switch e := e.(type) {
			case app.DestroyEvent:
				acks <- struct{}{}
				dw.emu.app.CloseWindow(debuggerTitle)
				return e.Err
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				dw.Layout(w, th, gtx)
				e.Frame(gtx.Ops)
			}
			acks <- struct{}{}
		}
	}
}

func (dw *DebuggerWindow) Layout(w *ui.Window, th *material.Theme, gtx C) {
	switch dw.dbg.getState() {
	case running:
		if dw.pause.Clicked(gtx) {
			dw.dbg.setState(paused)
		}
	case paused:
		if dw.start.Clicked(gtx) {
			dw.dbg.setState(running)
			dw.dbg.unblock()
		}
		if dw.step.Clicked(gtx) {
			dw.dbg.setState(stepping)
			dw.dbg.unblock()
		}
	case stepping:
		if dw.start.Clicked(gtx) {
			dw.dbg.setState(running)
			dw.dbg.unblock()
		}
		if dw.step.Clicked(gtx) {
			dw.dbg.unblock()
		}
	}

	btnSize := layout.Exact(image.Point{X: 70, Y: 35})

	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEnd, Alignment: layout.Start}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					gtx.Constraints = btnSize
					if dw.dbg.getState() == running {
						gtx = gtx.Disabled()
					}
					return material.Button(th, &dw.start, "Start").Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D { return layout.Spacer{Width: 5}.Layout(gtx) }),

				layout.Rigid(func(gtx C) D {
					gtx.Constraints = btnSize
					if dw.dbg.getState() != running {
						gtx = gtx.Disabled()
					}
					return material.Button(th, &dw.pause, "Pause").Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D { return layout.Spacer{Width: 5}.Layout(gtx) }),
				layout.Rigid(func(gtx C) D {
					gtx.Constraints = btnSize
					if dw.dbg.getState() == running {
						gtx = gtx.Disabled()
					}
					return material.Button(th, &dw.step, "Step").Layout(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx C) D {
			return material.H6(th, "Patterns table").Layout(gtx)
		}),
c(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEnd}.Layout(gtx,
				layout.Rigid(
					patternsTable{ppu: dw.nes.Hw.PPU}.Layout,
				),
				layout.Rigid(func(gtx C) D {
					return dw.callstack.Layout(gtx, dw.dbg.cpu.PC, dw.dbg.cstack.frames)
				}),
				// layout.Flexed(1patternsTab	)
}

type patternsTable struct {
	ppu *hw.PPU
}

func (pt patternsTable) render(ppu *hw.PPU) *image.RGBA {
	ptbuf := ppu.Bus.FetchPointer(0x0000)
	img := image.NewRGBA(image.Rect(0, 0, 128, 256))

	// A pattern table is 0x1000 bytes so 0x8000 bits.
	// One pixel requires 2 bits (4 colors), so there are 0x4000 pixels to draw.
	// That's a square of 128 x 128 pixels
	// Each tile is 8 x 8 pixels, that 's 16 x 16 tiles.
	for row := uint16(0); row < 256; row++ {
		for col := uint16(0); col < 128; col++ {
			addr := (row / 8 * 0x100) + (row % 8) + (col/8)*0x10
			pixel := uint8((ptbuf[addr]>>(7-(col%8)))&1) + ((ptbuf[addr+8]>>(7-(col%8)))&1)*2
			gray := pixel * 64
			img.Pix[(row*128*4)+(col*4)] = gray
			img.Pix[(row*128*4)+(col*4)+1] = gray
			img.Pix[(row*128*4)+(col*4)+2] = gray
			img.Pix[(row*128*4)+(col*4)+3] = 255
		}
	}
	return img
}

func (pt patternsTable) Layout(gtx C) D {
	size := image.Pt(128, 256)
	gtx.Constraints = layout.Exact(size)

	img := pt.render(pt.ppu)

	return widget.Image{
		Src:   paint.NewImageOp(img),
		Fit:   widget.Contain,
		Scale: 1,
	}.Layout(gtx)
}
 {
	frames []stackFrame
}

func (cs *callStack) push(src, dst, ret uint16, flag stackFrameFlag) {
	cs.frames = append(cs.frames, stackFrame{src: src, target: dst, ret: ret})
}

func (cs *callStack) pop() {
	if len(cs.frames) == 0 {
		return
	}
	popped := cs.frames[len(cs.frames)-1]
	_ = popped
	cs.frames = cs.frames[:len(cs.frames)-1]
}

func (cs *callStack) Print() string {
	var frames []string
	for _, f := range cs.frames {
		frames = append(frames, fmt.Sprintf("src: %04X, target: %04X, ret: %04X", f.src, f.target, f.ret))
	}
	return strings.Join(frames, "\n")
}

type stackFrame struct {
	src    uint16
	target uint16
	ret    uint16
	flags  stackFrameFlag
}

type stackFrameFlag uint8

const (
	sffNone stackFrameFlag = iota
	sffNMI
	sffIRQ
)

type callStackViewer struct {
	stack widget.List
	table ui.Table
}

func newCallStackViewer() callStackViewer {
	return callStackViewer{
		stack: widget.List{List: layout.List{Axis: layout.Vertical}},
		table: ui.Table{Cols: 2, ColBorder: 1, RowBorder: 1},
	}
}

func (cs *callStackViewer) Layout(gtx C, pc uint16, frames []stackFrame) D {
	type frameInfo [2]string
	var items []frameInfo

	var curf *stackFrame
	for i, f := range frames {
		if i > 0 {
			curf = &frames[i-1]
		}
		// item := fmt.Sprintf("func: %20s pc: %04X",
		// 	callStackViewer{}.entryPoint(curf), f.src)
		items = slices.Insert(items, 0, frameInfo{cs.entryPoint(curf), fmt.Sprintf("%04X", f.src)})
	}

	// Current frame
	curf = nil
	if len(frames) > 0 {
		curf = &frames[len(frames)-1]
	}

	// item := fmt.Sprintf("func: %20s pc: %04X",
	// 	callStackViewer{}.entryPoint(curf), pc)
	items = slices.Insert(items, 0, frameInfo{cs.entryPoint(curf), fmt.Sprintf("%04X", pc)})

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return material.H6(th, "Call Stack").Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return cs.table.Layout(gtx, len(items), func(gtx ui.C, i, j int) D {
				return layout.UniformInset(unit.Dp(1)).Layout(gtx, func(gtx C) D {
					return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx C) D {
						return material.Body1(th, items[j][i]).Layout(gtx)
					})
				})
			})
		}),
	)
}

func (callStackViewer) entryPoint(f *stackFrame) string {
	if f == nil {
		return "[bottom of stack]"
	}
	switch f.flags {
	case sffNMI:
		return "[nmi] $" + fmt.Sprintf("%04X", f.target)
	case sffIRQ:
		return "[irq] $" + fmt.Sprintf("%04X", f.target)
	}
	return fmt.Sprintf("$%04X", f.target)
}
                      