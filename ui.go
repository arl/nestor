package main

import (
	"context"
	"os"
	"os/signal"

	"gioui.org/app"
	"gioui.org/layout"
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
	nes *NES
}

func newEmulator(nes *NES) *emulator {
	return &emulator{
		nes: nes,
	}
}

func (e *emulator) run() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		a := ui.NewApplication(ctx)

		debugger := NewDebuggerWindow(e.nes)
		screen := NewScreenWindow(e.nes)

		a.NewWindow("Screen", screen)
		a.NewWindow("Debugger", debugger)
		a.Wait()
		os.Exit(0)
	}()

	app.Main()
}
