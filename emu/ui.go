package emu

import (
	"context"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"

	"nestor/emu/debugger"
	"nestor/ui"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 224
)

type C = layout.Context
type D = layout.Dimensions

const (
	debuggerTitle = "Debugger"
	screenTitle   = "Screen"
)

func ShowDebuggerWindow(a *ui.Application, nes *NES) {
	a.NewWindow(debuggerTitle, debugger.NewDebuggerWindow(nes.Debugger, nes.PPU))
}

// TODO(arl) merge with NES struct
type Emulator struct {
	nes *NES
	app *ui.Application
}

func NewEmulator(nes *NES) *Emulator {
	return &Emulator{nes: nes}
}

func (e *Emulator) Run(ctx context.Context, nes *NES, dbgOn bool) {
	screen := NewScreenWindow(nes)
	minw := unit.Dp(2*ScreenWidth + 200)
	minh := unit.Dp(2 * ScreenHeight)
	app := ui.NewApplication("NEStor", screen,
		app.MinSize(minw, minh),
		app.Size(minw, minh),
	)

	go func() {
		<-ctx.Done()
		app.Shutdown()
	}()

	app.Wait()
}

// Defer allows to defer functions to be called at the end of the program. Defer
// can be called multiple times, as `defer` the functions are called in reverse
// order.
func (e *Emulator) Defer(f func()) {
	e.app.Defer(f)
}
