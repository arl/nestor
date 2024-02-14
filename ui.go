package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 224
)

type C = layout.Context
type D = layout.Dimensions

var th = material.NewTheme()

type gui struct {
	w   *app.Window
	nes *NES
}

func newGUI(nes *NES) *gui {
	return &gui{
		nes: nes,
	}
}

func (ui *gui) run() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		a := NewApplication(ctx)

		debugger := NewDebuggerWindow(ui.nes)
		screen := NewScreenWindow(ui.nes)

		a.NewWindow("Screen", screen)
		a.NewWindow("Debugger", debugger)
		a.Wait()
		os.Exit(0)
	}()

	app.Main()
}

// Application keeps track of all the windows and global state.
type Application struct {
	// Context is used to broadcast application shutdown.
	Context context.Context
	// Shutdown shuts down all windows.
	Shutdown func()
	// active keeps track the open windows, such that application
	// can shut down, when all of them are closed.
	active sync.WaitGroup
}

func NewApplication(ctx context.Context) *Application {
	ctx, cancel := context.WithCancel(ctx)
	return &Application{
		Context:  ctx,
		Shutdown: cancel,
	}
}

// Wait waits for all windows to close.
func (a *Application) Wait() {
	a.active.Wait()
}

// NewWindow creates a new tracked window.
func (a *Application) NewWindow(title string, view View, opts ...app.Option) {
	opts = append(opts, app.Title(title))
	w := &Window{
		App:    a,
		Window: app.NewWindow(opts...),
	}
	a.active.Add(1)
	go func() {
		defer a.active.Done()
		view.Run(w)
	}()
}

// Window holds window state.
type Window struct {
	App *Application
	*app.Window
}

// View describes .
type View interface {
	// Run handles the window event loop.
	Run(w *Window) error
}

// WidgetView allows to use layout.Widget as a view.
type WidgetView func(gtx layout.Context, th *material.Theme) layout.Dimensions

// Run displays the widget with default handling.
func (view WidgetView) Run(w *Window) error {
	var ops op.Ops

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	go func() {
		<-w.App.Context.Done()
		w.Perform(system.ActionClose)
	}()
	for {
		switch e := w.NextEvent().(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			view(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}
