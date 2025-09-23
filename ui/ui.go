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
	screenWidth  = 640
	screenHeight = 480
)

// StartUI is the entry point of the GUI mode.
func StartUI(cfg config.Config) error {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Nestor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	var options *ebiten.RunGameOptions

	app := NewApp(cfg)

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

	var options *ebiten.RunGameOptions

	modUI.InfoZ("test").End()
	app := NewApp(cfg)
	app.ChangeState(app.GetState(StateRomRunning))

	// Launch emulator and transition to running state
	if err := app.LaunchEmulator(romPath); err != nil {
		return fmt.Errorf("failed to load rom: %w", err)
	}

	if err := ebiten.RunGameWithOptions(app, options); err != nil {
		return fmt.Errorf("failed to start ebiten: %w", err)
	}
	return nil
}

// Game implements [ebiten.Game] interface
type Game struct {
	emulator *emu.Emulator
	out      *emu.Output
	framech  chan *emu.Frame
	input    *input.Provider

	outw, outh float64 // current output window size
}

func (g *Game) loadROM(path string, cfg config.Config) error {
	rom, err := ines.ReadROM(path)
	if err != nil {
		return fmt.Errorf("failed to read rom: %w", err)
	}

	g.framech = make(chan *emu.Frame)
	g.out = emu.NewOutput(g.framech,
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

	emulator, err := emu.Launch(rom, cfg.Config, g.out, g.input)
	if err != nil {
		return fmt.Errorf("failed to start emulator: %w", err)
	}
	g.emulator = emulator

	go emulator.Run()

	return nil
}

func (g *Game) Update() error {
	if g.emulator == nil {
		return nil
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.emulator == nil {
		return
	}

	select {
	case frame := <-g.framech:
		frameImg := ImageFromFrame(frame)

		// Draw the frame at the maxium size that fits the window while keeping the aspect ratio.
		fw, fh := frameImg.Bounds().Dx(), frameImg.Bounds().Dy()

		scaleX := g.outw / float64(fw)
		scaleY := g.outh / float64(fh)

		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate((g.outw-float64(fw)*scale)/2, (g.outh-float64(fh)*scale)/2)
		screen.DrawImage(frameImg, op)
	default:
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	g.outw, g.outh = float64(outsideWidth), float64(outsideHeight)
	return outsideWidth, outsideHeight
}
