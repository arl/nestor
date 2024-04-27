package debugger

import (
	"context"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"nestor/hw"
	"nestor/ui"
)

type DebuggerWindow struct {
	dbg *debugger

	theme *material.Theme

	csviewer callstackViewer
	ptviewer patternsTable

	start ui.SmallSquareIconButton
	pause ui.SmallSquareIconButton
	step  ui.SmallSquareIconButton
}

func NewDebuggerWindow(dbg hw.Debugger, ppu *hw.PPU) *DebuggerWindow {
	theme := material.NewTheme()
	return &DebuggerWindow{
		dbg:      dbg.(*debugger),
		ptviewer: patternsTable{ppu: ppu},
		theme:    theme,
		start:    ui.NewSmallSquareIconButton(theme, icons.AVPlayArrow, "Start"),
		pause:    ui.NewSmallSquareIconButton(theme, icons.AVPause, "Pause"),
		step:     ui.NewSmallSquareIconButton(theme, icons.NavigationArrowForward, "Step"),
	}
}

func (dw *DebuggerWindow) Run(ctx context.Context, w *ui.Window) error {
	// defer dw.dbg.detach()
	viewCtx, cancel := context.WithCancel(ctx)

	go func() {
		<-viewCtx.Done()
		w.Perform(system.ActionClose)
		dw.dbg.detach()
	}()

	var ops op.Ops

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	dw.dbg.active.Store(true)

	events := make(chan event.Event)
	acks := make(chan struct{})

	go func() {
		for {
			ev := w.Event()
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
				cancel()
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
	// listing := &listing{nes: dw.nes, list: &dw.list}

	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEnd, Alignment: layout.Start}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return dw.start.Layout(gtx, status.stat != running)
				}),
				layout.Rigid(func(gtx C) D { return layout.Spacer{Width: 5}.Layout(gtx) }),

				layout.Rigid(func(gtx C) D {
					return dw.pause.Layout(gtx, status.stat == running)
				}),
				layout.Rigid(func(gtx C) D { return layout.Spacer{Width: 5}.Layout(gtx) }),
				layout.Rigid(func(gtx C) D {
					return dw.step.Layout(gtx, status.stat != running)
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
