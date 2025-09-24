package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"nestor/config"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

// StartUI is the entry point of the GUI mode.
func StartUI(cfg config.Config) error {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Nestor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	app := NewApp(cfg)

	var options = &ebiten.RunGameOptions{
		SingleThread: true,
	}

	if err := ebiten.RunGameWithOptions(app, options); err != nil {
		return fmt.Errorf("failed to start ebiten: %w", err)
	}
	return nil
}

// StartROM starts the emulation of a ROM in a window.
func StartROM(cfg config.Config, romPath string) error {
	// We can reuse the same App structure but immediately transition to the running state
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Nestor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	modUI.InfoZ("test").End()
	app := NewApp(cfg)
	app.ChangeState(app.GetState(StateRomRunning))

	// Launch emulator and transition to running state
	if err := app.LaunchEmulator(romPath); err != nil {
		return fmt.Errorf("failed to load rom: %w", err)
	}

	var options = &ebiten.RunGameOptions{
		SingleThread: true,
	}

	if err := ebiten.RunGameWithOptions(app, options); err != nil {
		return fmt.Errorf("failed to start ebiten: %w", err)
	}
	return nil
}
