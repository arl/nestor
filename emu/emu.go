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

// Start configures controllers, loads up the rom and creates the output
// streams. It returns a NES emulator ready to run.
func Start(rom *ines.Rom, cfg Config) (*NES, error) {
	nes, err := PowerUp(rom)
	if err != nil {
		return nil, fmt.Errorf("power up failed: %s", err)
	}

	// Controller setup.
	paddles := StdControllerPair{
		Pad1Connected: true,
	}
	nes.CPU.PlugInputDevice(&paddles)

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

	// CPU trace setup.
	if cfg.TraceOut != nil {
		defer cfg.TraceOut.Close()
		nes.CPU.SetTraceOutput(cfg.TraceOut)
	}

	return nes, nil
}
