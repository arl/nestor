package emu

import (
	"fmt"
	"image"

	"nestor/hw"
	"nestor/ines"
)

type Emulator struct {
	NES *NES
}

// PowerUp configures controllers, loads up the rom and creates the output
// streams. It returns a NES emulator ready to run.
func PowerUp(rom *ines.Rom, cfg Config) (*Emulator, error) {
	nes, err := powerUp(rom)
	if err != nil {
		return nil, fmt.Errorf("power up failed: %s", err)
	}

	// Output setup.
	nes.Frames = make(chan image.RGBA)
	out := hw.NewOutput(hw.OutputConfig{
		Width:           256,
		Height:          240,
		NumVideoBuffers: 2,
		Title:           "Nestor",
		ScaleFactor:     2,
	})
	if err := out.EnableVideo(true); err != nil {
		return nil, err
	}
	nes.SetOutput(out)

	input, err := hw.NewInputProvider(cfg.Input)
	if err != nil {
		return nil, fmt.Errorf("input provider: %s", err)
	}
	nes.CPU.PlugInputDevice(input)

	// CPU trace setup.
	if cfg.TraceOut != nil {
		nes.CPU.SetTraceOutput(cfg.TraceOut)
	}

	return &Emulator{
		NES: nes,
	}, nil
}

func (e *Emulator) Run() {
	e.NES.Run()
}

func (e *Emulator) Screenshot() image.Image {
	return e.NES.Out.Screenshot()
}
