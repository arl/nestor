package emu

import (
	"fmt"
	"image"
	"sync/atomic"
	"time"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
)

type Output interface {
	BeginFrame() hw.Frame
	EndFrame(hw.Frame)
	Poll() bool
	Close()
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
			// Don't burn cpu while paused.
			time.Sleep(100 * time.Millisecond)
		} else {
			e.RunOneFrame()
		}

		// handle stop conditions.
		if e.quit.Load() || !e.out.Poll() || e.NES.CPU.IsHalted() {
			e.out.Close()
			break
		}

		// handle reset.
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

func (e *Emulator) Stop()    { e.quit.Store(true) }
func (e *Emulator) Reset()   { e.reset.Store(true) }
func (e *Emulator) Restart() {}
func (e *Emulator) restart() {}
