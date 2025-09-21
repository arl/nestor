package ebit

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"nestor/config"
	"nestor/emu"
	"nestor/hw/input"
	"nestor/ines"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

// StartUI is the entry point of the GUI mode.
func StartUI() {
	panic("not implemented")
}

// StartROM starts the emulation of a ROM in a window.
func StartROM(cfg config.Config, romPath string) error {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Nestor")

	eui := new(UI)
	eui.input = input.NewProvider(cfg.Input)

	if err := eui.loadROM(romPath, cfg); err != nil {
		return fmt.Errorf("failed to load rom: %w", err)
	}

	var options *ebiten.RunGameOptions

	if err := ebiten.RunGameWithOptions(eui, options); err != nil {
		return fmt.Errorf("failed to start ebiten: %w", err)
	}
	return nil
}

// UI implements [ebiten.Game] interface
type UI struct {
	emulator *emu.Emulator
	out      *emu.Output
	framech  chan *emu.Frame
	input    *input.Provider
}

func (ui *UI) loadROM(path string, cfg config.Config) error {
	rom, err := ines.ReadROM(path)
	if err != nil {
		return fmt.Errorf("failed to read rom: %w", err)
	}

	ui.framech = make(chan *emu.Frame)
	ui.out = emu.NewOutput(ui.framech,
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

	emulator, err := emu.Launch(rom, cfg.Config, ui.out, ui.input)
	if err != nil {
		return fmt.Errorf("failed to start emulator: %w", err)
	}
	ui.emulator = emulator

	go emulator.Run()

	return nil
}

func (ui *UI) Update() error {
	if ui.emulator == nil {
		return nil
	}

	return nil
}

// Draw draws the game screen by one frame.
//
// The give argument represents a screen image. The updated content is adopted as the game screen.
//
// The frequency of Draw calls depends on the user's environment, especially the monitors refresh rate.
// For portability, you should not put your game logic in Draw in general.
func (ui *UI) Draw(screen *ebiten.Image) {
	if ui.emulator == nil {
		return
	}

	select {
	case frame := <-ui.framech:
		frameImg := ImageFromFrame(frame)
		screen.DrawImage(frameImg, nil)
	default:
	}
}

// Layout accepts a native outside size in device-independent pixels and returns the game's logical screen
// size in pixels. The logical size is used for 1) the screen size given at Draw and 2) calculation of the
// scale from the screen to the final screen size.
//
// On desktops, the outside is a window or a monitor (fullscreen mode). On browsers, the outside is a body
// element. On mobiles, the outside is the view's size.
//
// Even though the outside size and the screen size differ, the rendering scale is automatically adjusted to
// fit with the outside.
//
// Layout is called almost every frame.
//
// It is ensured that Layout is invoked before Update is called in the first frame.
//
// If Layout returns non-positive numbers, the caller can panic.
//
// You can return a fixed screen size if you don't care, or you can also return a calculated screen size
// adjusted with the given outside size.
//
// If the game implements the interface LayoutFer, Layout is never called and LayoutF is called instead.
func (ui *UI) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}
