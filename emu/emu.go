package emu

import (
	"fmt"
	"image"
	"io"
	"sync/atomic"
	"time"

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

	quit   atomic.Bool
	paused atomic.Bool
	reset  atomic.Bool
}

// Launch instantiates an emulator, setup controllers, output streams and window.
func Launch(rom *ines.Rom, cfg Config, monidx int32) (*Emulator, error) {
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
		Monitor:         monidx,
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

func (e *Emulator) Screenshot() image.Image {
	return e.out.Screenshot()
}

func (e *Emulator) RunOneFrame() {
	frame := e.out.BeginFrame()
	e.NES.RunOneFrame(frame)
	e.out.EndFrame(frame)
}

func (e *Emulator) Run() {
	for {
		// Handle pause.
		if !e.paused.Load() {
			e.RunOneFrame()
		} else {
			// Don't burn cpu.
			time.Sleep(100 * time.Millisecond)
		}

		// Stop conditions
		if e.quit.Load() ||
			!e.out.Poll() ||
			e.NES.CPU.IsHalted() {
			if err := e.out.Close(); err != nil {
				log.ModEmu.WarnZ("Error closing emulator window").Error("error", err).End()
			}
			break
		}

		if e.reset.Load() {
			e.NES.Reset(true)
			e.reset.Store(false)
		}
	}
	log.ModEmu.InfoZ("Emulation loop exited").End()
}

func (e *Emulator) SetPause(pause bool) {
	e.paused.CompareAndSwap(!pause, pause)
}

func (e *Emulator) Stop()  { e.quit.Store(true) }
func (e *Emulator) Reset() { e.reset.Store(true) }

func (e *Emulator) Restart() {}
func (e *Emulator) restart() {}
