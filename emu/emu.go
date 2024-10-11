package emu

import (
	"fmt"
	"image"
	"io"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
)

type Output interface {
	io.Closer

	BeginFrame() hw.Frame
	EndFrame(hw.Frame)
	Poll() bool
	Screenshot() image.Image
}

type Emulator struct {
	NES *NES
	out Output
}

// Start instantiates an emulator, setup controllers, output streams and window.
func Start(rom *ines.Rom, cfg Config) (*Emulator, error) {
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
		DisableVSync:    cfg.Video.DisableVSync,
	})
	if err := out.EnableVideo(true); err != nil {
		return nil, err
	}

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
		out: out,
	}, nil
}

func (e *Emulator) Run() {
	for e.out.Poll() {
		e.RunOneFrame()
		if e.NES.CPU.IsHalted() {
			break
		}
	}
	log.ModEmu.InfoZ("Emulation stopped").End()
	if err := e.out.Close(); err != nil {
		log.ModEmu.WarnZ("Error closing emulator window").Error("error", err).End()
	}
}

func (e *Emulator) RunOneFrame() {
	frame := e.out.BeginFrame()
	e.NES.RunOneFrame(frame)
	e.out.EndFrame(frame)
}

func (e *Emulator) Screenshot() image.Image {
	return e.out.Screenshot()
}
