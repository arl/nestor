package emu

import (
	"fmt"
	"image"

	"nestor/hw"
	"nestor/ines"
)

// Start configures controllers, loads up the rom and creates output streams. It
// returns a run function that starts the whole emulation, and that terminates
// when emulation stops by itself or the window is closed.
func Start(rom *ines.Rom) (func(), error) {
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

	return func() { nes.Run(out) }, nil
}
