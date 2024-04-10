package emu

import (
	"image"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"nestor/ui"
)

type ScreenWindow struct {
	emu   *emulator
	theme *material.Theme

	debugBtn widget.Clickable
}

func NewScreenWindow(emu *emulator) *ScreenWindow {
	return &ScreenWindow{
		emu:   emu,
		theme: material.NewTheme(),
	}
}

func (sw *ScreenWindow) Run(w *ui.Window) error {
	quit := make(chan struct{})
	go func() {
		select {
		case <-quit:
			sw.emu.app.Shutdown()
		case <-w.App.Context.Done():
			w.Perform(system.ActionClose)
		}
	}()

	var ops op.Ops
	events := make(chan event.Event)
	acks := make(chan struct{})
	frameCh := sw.emu.nes.FrameEvents()
	var frame *image.RGBA

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
		case e := <-events:
			switch e := e.(type) {
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
				event.Op(gtx.Ops, w)
				for {
					event, ok := gtx.Event(
						key.Filter{
							Name: key.NameEscape,
						},
						key.Filter{
							Name: key.NameEnter,
						},
					)
					if !ok {
						break
					}
					if _, ok := event.(key.Event); !ok {
						continue
					}
					return nil
				}

				sw.Layout(gtx, frame)
				area.Pop()

				e.Frame(gtx.Ops)

			case app.DestroyEvent:
				acks <- struct{}{}
				close(quit)
				return e.Err
			}
			acks <- struct{}{}

		case img := <-frameCh:
			frame = img
			w.Invalidate()
		}
	}
}

func (sw *ScreenWindow) Layout(gtx C, frame *image.RGBA) D {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Start,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return ui.Surface{Gray: 24, FitSize: false}.Layout(gtx,
				widget.Image{
					Src:      paint.NewImageOp(frame),
					Fit:      widget.Contain,
					Position: layout.Center,
				}.Layout)
		}),

		layout.Rigid(func(gtx C) D {
			return layout.NW.Layout(gtx, func(gtx C) D {
				if sw.debugBtn.Clicked(gtx) {
					sw.emu.showDebuggerWindow()
				}
				return material.Button(sw.theme, &sw.debugBtn, "Debug").Layout(gtx)
			})
		}),
	)
}
