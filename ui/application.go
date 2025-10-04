package ui

import (
	"context"
	"fmt"
	"image"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"

	"nestor/config"
	"nestor/emu"
	"nestor/hw/input"
	"nestor/ines"
)

type Application struct {
	cfg  config.Config
	quit atomic.Bool

	emuctx   context.Context
	emulator *emu.Emulator
	framech  chan *emu.Frame
	frameimg *ebiten.Image

	states       map[string]state
	currentState state
}

func newApplication(ctx context.Context, cfg config.Config) *Application {
	app := &Application{
		cfg:    cfg,
		states: make(map[string]state),
	}

	app.states["running"] = newRunningState(app)
	app.states["rom_list"] = newRomListState(app)
	app.states["config"] = newConfigState(app)

	go func() {
		<-ctx.Done()
		app.quit.Store(true)
	}()

	return app
}

func (app *Application) setState(name string) {
	modUI.InfoZ("Switching to state").String("to", name).End()
	app.currentState = app.states[name]
}

func (app *Application) runRom(romPath string) error {
	inputProvider := input.NewProvider(app.cfg.Input)

	rom, err := ines.ReadROM(romPath)
	if err != nil {
		return fmt.Errorf("failed to read ROM: %w", err)
	}

	framech := make(chan *emu.Frame)
	out := emu.NewOutput(framech,
		emu.OutputConfig{
			Width:          emu.NTSCWidth,
			Height:         emu.NTSCHeight,
			NumBackBuffers: 4,
			Title:          "Nestor",
			ScaleFactor:    2,
			DisableVSync:   app.cfg.Video.DisableVSync,
			Monitor:        app.cfg.Video.Monitor,
			Shader:         app.cfg.Video.Shader,
		},
	)

	emulator, err := emu.Launch(rom, app.cfg.Config, out, inputProvider)
	if err != nil {
		return fmt.Errorf("failed to launch emulator: %w", err)
	}

	app.emulator = emulator
	app.framech = framech
	app.frameimg = ebiten.NewImageWithOptions(
		image.Rect(0, 0, emu.NTSCWidth, emu.NTSCHeight),
		&ebiten.NewImageOptions{Unmanaged: true},
	)

	go emulator.Run()

	return nil
}

func (app *Application) Update() error {
	if app.quit.Load() {
		return ebiten.Termination
	}
	app.currentState.Update()
	return nil
}

func (app *Application) Draw(screen *ebiten.Image) {
	app.currentState.Draw(screen)
}

func (app *Application) Layout(outw, outh int) (screenw, screenh int) {
	return outw, outh
}

type state interface {
	Update()
	Draw(screen *ebiten.Image)
}
