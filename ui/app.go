package ui

import (
	"context"
	"sync"

	"gioui.org/app"
	"gioui.org/widget/material"
)

var Theme = material.NewTheme()

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

// A View handles the event loop for a Window.
type View interface {
	// Run handles the window event loop.
	Run(w *Window) error
}
