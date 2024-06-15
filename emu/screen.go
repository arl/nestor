package emu

import (
	"context"
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

	"nestor/emu/hwio"
	"nestor/ui"
)

type C = layout.Context
type D = layout.Dimensions

var keymap = map[key.Name]StdPadButton{
	"A":                PadA,
	"Z":                PadB,
	"E":                PadSelect,
	"R":                PadStart,
	key.NameUpArrow:    PadUp,
	key.NameDownArrow:  PadDown,
	key.NameLeftArrow:  PadLeft,
	key.NameRightArrow: PadRight,
}

type ScreenWindow struct {
	nes *NES

	theme    *material.Theme
	debugBtn widget.Clickable

	// bitmap for the state of all 16 bits of the input devices.
	buttons, prevButtons uint16
	inputch              chan uint16
}

func NewScreenWindow(nes *NES) *ScreenWindow {
	return &ScreenWindow{
		nes:     nes,
		theme:   material.NewTheme(),
		inputch: make(chan uint16),
	}
}

func (sw *ScreenWindow) UserInputs() <-chan uint16 {
	return sw.inputch
}

func (sw *ScreenWindow) Run(ctx context.Context, w *ui.Window) error {
	viewCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-viewCtx.Done()
		w.Perform(system.ActionClose)
	}()

	var ops op.Ops
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

	var frame image.RGBA
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
						key.Filter{Name: key.NameLeftArrow},
						key.Filter{Name: key.NameRightArrow},
						key.Filter{Name: key.NameUpArrow},
						key.Filter{Name: key.NameDownArrow},
						key.Filter{Name: "A"},
						key.Filter{Name: "Z"},
						key.Filter{Name: "E"},
						key.Filter{Name: "R"},
					)
					if !ok {
						break
					}
					if kevt, ok := event.(key.Event); ok {
						btnIdx := keymap[kevt.Name]
						switch kevt.State {
						case key.Press:
							hwio.SetBit16(&sw.buttons, uint(btnIdx))
						case key.Release:
							hwio.ClearBit16(&sw.buttons, uint(btnIdx))
						}
						continue
					}
					return nil
				}

				if sw.prevButtons != sw.buttons {
					select {
					case sw.inputch <- sw.buttons:
					default:
					}
					sw.prevButtons = sw.buttons
				}

				if sw.debugBtn.Clicked(gtx) {
					ShowDebuggerWindow(w.App, sw.nes)
				}

				sw.Layout(gtx, &frame)
				area.Pop()

				e.Frame(gtx.Ops)

			case app.DestroyEvent:
				acks <- struct{}{}
				cancel()
				return e.Err
			}
			acks <- struct{}{}

		case frame = <-sw.nes.Frames:
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
				return material.Button(sw.theme, &sw.debugBtn, "Debug").Layout(gtx)
			})
		}),
	)
}
