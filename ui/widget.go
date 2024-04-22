package ui

import (
	"context"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
)

type C = layout.Context
type D = layout.Dimensions

// WidgetView allows to use layout.Widget as a view.
type WidgetView func(gtx C, th *material.Theme) D

// Run displays the widget with default handling.
func (view WidgetView) Run(ctx context.Context, w *Window) error {
	viewCtx, cancel := context.WithCancel(ctx)
	var ops op.Ops

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	go func() {
		<-viewCtx.Done()
		w.Perform(system.ActionClose)
	}()
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			cancel()
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			view(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}

func Center(label material.LabelStyle) material.LabelStyle {
	label.Alignment = text.Middle
	return label
}
