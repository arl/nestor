package emu

import (
	"context"

	"nestor/hw"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 224
)

// TODO(arl) merge with NES struct
type Emulator struct {
	nes *NES

	// TODO: gtk3
	// screen *ScreenWindow
}

func NewEmulator(nes *NES) *Emulator {
	// TODO: gtk3
	// screen := NewScreenWindow(nes)
	// minw := unit.Dp(2*ScreenWidth + 200)
	// minh := unit.Dp(2 * ScreenHeight)
	// app := ui.NewApplication("NEStor", screen,
	// app.MinSize(minw, minh),
	// app.Size(minw, minh),
	// )

	return &Emulator{
		nes: nes,
		// TODO: gtk3
		// app:    app,
		// screen: screen,
	}
}

func (e *Emulator) Run(ctx context.Context, nes *NES) {
	// TODO: gtk3
	// 	go func() {
	// 	<-ctx.Done()
	// 	e.app.Shutdown()
	// }()

	// e.app.Wait()
}

// Defer allows to defer functions to be called at the end of the program. Defer
// can be called multiple times, as `defer` the functions are called in reverse
// order.
func (e *Emulator) Defer(f func()) {
	// TODO: gtk3
	// e.app.Defer(f)
}

type UserInput interface {
	hw.InputDevice
	UserInputReader
}

func (e *Emulator) ConnectInputDevice(up UserInput) {
	e.nes.CPU.PlugInputDevice(up)
	// TODO: gtk3
	// up.ReadUserInput(e.screen.UserInputs())
}
