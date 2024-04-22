package debugger

import (
	"image"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"nestor/hw"
	"nestor/ui"
)

func mustNewIcon(data []byte) *widget.Icon {
	icon, err := widget.NewIcon(data)
	if err != nil {
		panic(err)
	}
	return icon
}

type DebuggerWindow struct {
	dbg *debugger

	theme *material.Theme

	csviewer callstackViewer
	ptviewer patternsTable

	start widget.Clickable
	pause widget.Clickable
	step  widget.Clickable

	startIcon *widget.Icon
	pauseIcon *widget.Icon
	stepIcon  *widget.Icon
}

func NewDebuggerWindow(dbg hw.Debugger, ppu *hw.PPU) *DebuggerWindow {
	return &DebuggerWindow{
		dbg:       dbg.(*debugger),
		ptviewer:  patternsTable{ppu: ppu},
		theme:     material.NewTheme(),
		startIcon: mustNewIcon(icons.AVPlayArrow),
		pauseIcon: mustNewIcon(icons.AVPause),
		stepIcon:  mustNewIcon(icons.NavigationArrowForward),
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


store this tipArea somewhere else
var tipArea component.TipArea

func (dw *DebuggerWindow) Layout(w *ui.Window, status status, gtx C) {
	btnSize := layout.Exact(image.Point{X: 25, Y: 25})
	// listing := &listing{nes: dw.nes, list: &dw.list}

	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEnd, Alignment: layout.Start}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					gtx.Constraints = btnSize
					if status.stat == running {
						gtx = gtx.Disabled()
					}

					tooltip := component.DesktopTooltip(dw.theme, "some tooltip")
					return tipArea.Layout(gtx, tooltip, ui.SmallSquareIconButton(dw.theme, &dw.start, dw.startIcon, "Start").Layout)

					// return ui.SmallSquareIconButton(dw.theme, &dw.start, dw.startIcon, "Start").Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D { return layout.Spacer{Width: 5}.Layout(gtx) }),

				layout.Rigid(func(gtx C) D {
					gtx.Constraints = btnSize
					if status.stat != running {
						gtx = gtx.Disabled()
					}
					return ui.SmallSquareIconButton(dw.theme, &dw.pause, dw.pauseIcon, "Pause").Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D { return layout.Spacer{Width: 5}.Layout(gtx) }),
				layout.Rigid(func(gtx C) D {
					gtx.Constraints = btnSize
					if status.stat == running {
						gtx = gtx.Disabled()
					}
					return ui.SmallSquareIconButton(dw.theme, &dw.step, dw.stepIcon, "Step").Layout(gtx)
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
