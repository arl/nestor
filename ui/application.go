package ui

import (
	"context"

	"github.com/hajimehoshi/ebiten/v2"

	"nestor/config"
	"nestor/emu"
	"nestor/hw/input"
)

type Application struct {
	cfg config.Config

	emuctx   context.Context
	emulator *emu.Emulator
	framech  chan *emu.Frame

	states       map[string]state
	currentState state
}

func newApplication(cfg config.Config) *Application {
	app := &Application{
		cfg:    cfg,
		states: make(map[string]state),
	}

	app.states["running"] = &running{
		Application: app,
	}

	return app
}

func (app *Application) setState(name string) {
	modUI.InfoZ("Switching to state").String("to", name).End()
	app.currentState = app.states[name]
}

func (app *Application) runRom(romPath string) error {
	inputProvider := input.NewProvider(app.cfg.Input)

	emulator, framech, err := runEmulator(app.cfg, inputProvider, romPath)
	if err != nil {
		return err
	}

	app.emulator = emulator
	app.framech = framech
	go emulator.Run()

	return nil
}

func (app *Application) stopEmulator() {
	app.emulator.Stop()
	for app.emulator.IsRunning() {
		// busy wait
	}
}

func (app *Application) Update() error {
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

type running struct {
	*Application
}

func (s *running) Update() {}
func (s *running) Draw(screen *ebiten.Image) {
	frame := <-s.framech
	img := ImageFromFrame(frame)
	s.drawFrame(screen, img, float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy()))
}

func (s *running) drawFrame(screen *ebiten.Image, frameImg *ebiten.Image, targetW, targetH float64) {
	// Calculate scaling to fit the target area while preserving aspect ratio
	fw, fh := float64(frameImg.Bounds().Dx()), float64(frameImg.Bounds().Dy())
	scaleX := targetW / fw
	scaleY := targetH / fh
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Draw the frame centered
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate((targetW-fw*scale)/2, (targetH-fh*scale)/2)
	screen.DrawImage(frameImg, op)
}
