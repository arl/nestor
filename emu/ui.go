package emu

import (
	"context"
	"os"

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

var th = ui.Theme

const (
	debuggerTitle = "Debugger"
	screenTitle   = "Screen"
)

type emulator struct {
	nes *NES
	app *ui.Application
}

func (e *emulator) showDebuggerWindow() {
	if e.app.HasWindow(debuggerTitle) {
		return
	}
	e.app.NewWindow(debuggerTitle, debugger.NewDebuggerWindow(e.nes.Debugger, e.nes.Hw.PPU))
}

func RunEmulator(ctx context.Context, nes *NES, dbgOn bool) {
	ctx, cancel := context.WithCancel(ctx)

	e := &emulator{
		app: ui.NewApplication(ctx),
		nes: nes,
	}

	go func() {
		defer cancel()
		screen := NewScreenWindow(e)

		minw := unit.Dp(2*ScreenWidth + 200)
		minh := unit.Dp(2 * ScreenHeight)
		e.app.NewWindow(screenTitle, screen,
			app.MinSize(minw, minh),
			app.Size(minw, minh),
		)

		if dbgOn {
			e.showDebuggerWindow()
		}

		e.app.Wait()
	}()

	go func() {
		<-ctx.Done()
		e.stop()
	}()

	app.Main()
}

func (e *emulator) stop() {
	e.app.Shutdown()
	e.app.Wait()
	runDefered()
	os.Exit(0)
}

var deferred []func()

// Defer allows to defer functions to be called at the end of the program. Defer
// can be called multiple times, as `defer` the functions are called in reverse
// order.
func Defer(f func()) {
	deferred = append(deferred, f)
}

func runDefered() {
	for i := len(deferred) - 1; i >= 0; i-- {
		deferred[i]()
	}
}
