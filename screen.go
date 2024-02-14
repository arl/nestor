package main

import (
	"fmt"
	"image"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
)

type ScreenWindow struct {
	nes *NES
}

func NewScreenWindow(nes *NES) *ScreenWindow {
	return &ScreenWindow{nes: nes}
}

func (sw *ScreenWindow) Run(w *Window) error {
	go func() {
		<-w.App.Context.Done()
		w.Perform(system.ActionClose)
	}()

	var ops op.Ops

	events := make(chan event.Event)
	acks := make(chan struct{})

	frameCh := sw.nes.FrameEvents()
	var frame *image.RGBA

	go func() {
		for {
			ev := w.NextEvent()
			events <- ev
			<-acks
			if _, ok := ev.(system.DestroyEvent); ok {
				return
			}
		}
	}()

	for {
		select {
		case e := <-events:
			switch e := e.(type) {
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)

				// Register a global key listener for the escape key wrapping
				// our entire window.
				area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
				key.InputOp{
					Tag:  w,
					Keys: key.NameEscape,
				}.Add(gtx.Ops)

				// check for presses of the escape key and close the window if we find them.
				for _, event := range gtx.Events(w) {
					switch event := event.(type) {
					case key.Event:
						if event.Name == key.NameEscape {
							return nil
						}
					}
				}
				size := image.Pt(ScreenWidth, ScreenHeight)
				gtx.Constraints = layout.Exact(size)

				widget.Image{
					Src:   paint.NewImageOp(frame),
					Fit:   widget.Contain,
					Scale: 0.5,
				}.Layout(gtx)

				// sw.Layout(gtx, frame)
				area.Pop()

				// Pass drawing operations to the gpu
				e.Frame(gtx.Ops)

			case system.DestroyEvent:
				fmt.Println("destroy event")
				acks <- struct{}{}
				return e.Err
			}
			acks <- struct{}{}

		case img := <-frameCh:
			frame = img
			w.Invalidate()
		}
	}
}
