package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"nestor/config"
	"nestor/emu"
	"nestor/hw/input"
	"nestor/ines"
)

const (
	startWidth  = 640
	startHeight = 480
)

func StartUI(cfg config.Config) error {
	return start(cfg, "")
}

func StartROM(cfg config.Config, romPath string) error {
	return start(cfg, romPath)
}

func start(cfg config.Config, romPath string) error {
	ebiten.SetWindowSize(startWidth, startHeight)
	ebiten.SetWindowTitle("Nestor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(ebiten.SyncWithFPS)

	var options = &ebiten.RunGameOptions{
		// SingleThread: true,
	}

	app := newApplication(cfg)
	if romPath != "" {
		app.setState("running")
		if err := app.runRom(romPath); err != nil {
			return fmt.Errorf("failed to run ROM: %w", err)
		}
	}

	if err := ebiten.RunGameWithOptions(app, options); err != nil {
		return fmt.Errorf("failed to start ebiten: %w", err)
	}
	return nil
}

func runEmulator(cfg config.Config, inputProvider *input.Provider, romPath string) (*emu.Emulator, chan *emu.Frame, error) {
	rom, err := ines.ReadROM(romPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read ROM: %w", err)
	}

	framech := make(chan *emu.Frame)
	out := emu.NewOutput(framech,
		emu.OutputConfig{
			Width:          emu.NTSCWidth,
			Height:         emu.NTSCHeight,
			NumBackBuffers: 4,
			Title:          "Nestor",
			ScaleFactor:    2,
			DisableVSync:   cfg.Video.DisableVSync,
			Monitor:        cfg.Video.Monitor,
			Shader:         cfg.Video.Shader,
		},
	)

	emu, err := emu.Launch(rom, cfg.Config, out, inputProvider)
	return emu, framech, nil
}
