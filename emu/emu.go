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

// Start configures controllers, loads up the rom and creates output streams. It
// returns a run function that starts the whole emulation, and that terminates
// when emulation stops by itself or the window is closed.
func Start(rom *ines.Rom, cfg Config) (func(), error) {
	pads := StdControllerPair{
		Pad1Connected: true,
	}

	nes, err := PowerUp(rom)
	if err != nil {
		return nil, fmt.Errorf("power up failed: %s", err)
	}

	nes.CPU.PlugInputDevice(&pads)

	// Output setup
	nes.Frames = make(chan image.RGBA)
	out := hw.NewOutput(hw.OutputConfig{
		Width:           256,
		Height:          240,
		NumVideoBuffers: 2,
	})
	if err := out.EnableVideo(true); err != nil {
		return nil, err
	}

	runLoop := func() {
		if cfg.TraceOut != nil {
			defer cfg.TraceOut.Close()
			nes.CPU.SetTraceOutput(cfg.TraceOut)
		}

		nes.Run(out)
	}

	return runLoop, nil
}
