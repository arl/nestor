package main

import (
	"context"
	"os"
	"os/signal"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"nestor/ui"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 224
)

type C = layout.Context
type D = layout.Dimensions

var th = material.NewTheme()

type emulator struct {
	nes      *NES
	app      *ui.Application
	debugger *DebuggerWindow
}

func newEmulator(nes *NES) *emulator {
	return &emulator{
		nes: nes,
	}
}

func (e *emulator) showDebugger() {
	if e.debugger != nil {
		panic("debugger already open")
	}
	e.debugger = NewDebuggerWindow(e.nes)
	e.app.NewWindow("Debugger", e.debugger)
}

func (e *emulator) run() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	e.app = ui.NewApplication(ctx)

	go func() {
		screen := NewScreenWindow(e)

		minw := unit.Dp(2*ScreenWidth + 200)
		minh := unit.Dp(2 * ScreenHeight)
		e.app.NewWindow("Screen", screen,
			app.MinSize(minw, minh),
			app.Size(minw, minh),
		)

		e.app.Wait()
		os.Exit(0)
	}()

	app.Main()
}
