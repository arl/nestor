package ui

import (
	"fmt"
	"sync"

	"nestor/config"
	"nestor/emu"
	"nestor/emu/log"
	"nestor/hw/input"
	"nestor/ines"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var modUI = log.NewModule("ui")

// App is the main application container that manages states
type App struct {
	currentState State
	states       map[StateID]State

	cfg     config.Config
	input   *input.Provider
	framech chan *emu.Frame

	mu       sync.Mutex
	emulator *emu.Emulator
}

// NewApp creates a new application instance
func NewApp(cfg config.Config) *App {
	app := &App{
		cfg:    cfg,
		input:  input.NewProvider(cfg.Input),
		states: make(map[StateID]State),
	}

	// Create all available states
	romListState := NewRomListState(app)
	runningState := NewRunningState(app)
	configState := NewConfigState(app)

	// Register states in the map for easy access
	app.states[StateRomList] = romListState
	app.states[StateRomRunning] = runningState
	app.states[StateConfig] = configState

	// Set initial state
	app.currentState = romListState
	app.currentState.Enter(nil)

	return app
}

// Update delegates to the current state
func (app *App) Update() error {
	app.currentState.Update()
	return nil
}

// Draw delegates to the current state
func (app *App) Draw(screen *ebiten.Image) {
	app.currentState.Draw(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %f", ebiten.ActualFPS()))
}

// Layout implements the ebiten.Game interface
func (app *App) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

// ChangeState switches to a new state
func (app *App) ChangeState(newState State) {
	modUI.InfoZ("changing state").Stringer("from", app.currentState.ID()).Stringer("to", newState.ID()).End()
	if app.currentState != nil {
		app.currentState.Exit(newState)
	}

	oldState := app.currentState
	app.currentState = newState
	app.currentState.Enter(oldState)
}

// GetState returns a state by ID
func (app *App) GetState(id StateID) State {
	return app.states[id]
}

// LaunchEmulator starts the emulator with the given ROM
func (app *App) LaunchEmulator(romPath string) error {
	rom, err := ines.ReadROM(romPath)
	if err != nil {
		return fmt.Errorf("failed to read ROM: %w", err)
	}

	app.framech = make(chan *emu.Frame)
	out := emu.NewOutput(app.framech,
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

	emulator, err := emu.Launch(rom, app.cfg.Config, out, app.input)
	if err != nil {
		return fmt.Errorf("failed to start emulator: %w", err)
	}

	app.emulator = emulator
	go emulator.Run()

	return nil
}

func (app *App) Config() config.Config {
	return app.cfg
}

func (app *App) GetInput() *input.Provider {
	return app.input
}

func (app *App) GetFrameChannel() chan *emu.Frame {
	return app.framech
}

func (app *App) ResetEmulator() {
	if app.emulator == nil {
		return
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	app.emulator.SoftReset()
}

func (app *App) RestartEmulator() {
	if app.emulator == nil {
		return
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	app.emulator.HardReset()
}

func (app *App) StopEmulator() {
	if app.emulator == nil {
		return
	}

	app.emulator.Stop()
}
