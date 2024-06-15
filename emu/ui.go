package emu

import (
	"context"

	"gioui.org/app"
	"gioui.org/unit"

	"nestor/emu/debugger"
	"nestor/hw"
	"nestor/ui"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 224
)

func ShowDebuggerWindow(a *ui.Application, nes *NES) {
	a.NewWindow("Debugger", debugger.NewDebuggerWindow(nes.Debugger, nes.CPU, nes.PPU))
}

// TODO(arl) merge with NES struct
type Emulator struct {
	nes *NES
	app *ui.Application

	screen *ScreenWindow
}

func NewEmulator(nes *NES) *Emulator {
	screen := NewScreenWindow(nes)
	minw := unit.Dp(2*ScreenWidth + 200)
	minh := unit.Dp(2 * ScreenHeight)
	app := ui.NewApplication("NEStor", screen,
		app.MinSize(minw, minh),
		app.Size(minw, minh),
	)

	return &Emulator{
		nes:    nes,
		app:    app,
		screen: screen,
	}
}

func (e *Emulator) Run(ctx context.Context, nes *NES) {
	go func() {
		<-ctx.Done()
		e.app.Shutdown()
	}()

	e.app.Wait()
}

// Defer allows to defer functions to be called at the end of the program. Defer
// can be called multiple times, as `defer` the functions are called in reverse
// order.
func (e *Emulator) Defer(f func()) {
	e.app.Defer(f)
}

type UserInput interface {
	hw.InputDevice
	UserInputReader
}

func (e *Emulator) ConnectInputDevice(up UserInput) {
	e.nes.CPU.PlugInputDevice(up)
	// Connect inputs from window to emulated input ports.
	up.ReadUserInput(e.screen.UserInputs())
}
