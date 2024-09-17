package emu

import (
	"fmt"
	"image"
	"io"

	"nestor/hw"
	"nestor/ines"
)

type Config struct {
	TraceOut io.WriteCloser
}

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
	})
	if err := out.EnableVideo(true); err != nil {
		return nil, err
	}
	nes.SetOutput(out)

	// Controller setup.
	var inputcfg hw.InputConfig
	inputcfg.Paddles[0].Plugged = true
	inputcfg.Paddles[1].Plugged = false

	input, err := hw.NewInputProvider(inputcfg)
	if err != nil {
		return nil, fmt.Errorf("input provider: %s", err)
	}
	nes.CPU.PlugInputDevice(input)

	// CPU trace setup.
	if cfg.TraceOut != nil {
		defer cfg.TraceOut.Close()
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
