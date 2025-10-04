package ui

import (
	"context"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"nestor/config"
	"nestor/emu/log"
)

var modUI = log.NewModule("ui")

const (
	startWidth  = 640
	startHeight = 480
)

func StartUI(ctx context.Context, cfg config.Config) error {
	return start(ctx, cfg, "")
}

func StartROM(ctx context.Context, cfg config.Config, romPath string) error {
	return start(ctx, cfg, romPath)
}

func start(ctx context.Context, cfg config.Config, romPath string) error {
	ebiten.SetWindowSize(startWidth, startHeight)
	ebiten.SetWindowTitle("Nestor")
	// ebiten.SetFullscreen(true) // add option
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebiten.SetVsyncEnabled(!cfg.Video.DisableVSync)
	var options = &ebiten.RunGameOptions{
		// SingleThread: true,

	}

	app := newApplication(ctx, cfg)
	if romPath == "" {
		app.setState("rom_list")
	} else {
		app.setState("running")
		if err := app.runRom(romPath); err != nil {
			return fmt.Errorf("can't run rom: %w", err)
		}
	}

	if err := ebiten.RunGameWithOptions(app, options); err != nil {
		return fmt.Errorf("ui failure: %w", err)
	}

	return nil
}
