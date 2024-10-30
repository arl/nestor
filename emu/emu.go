package emu

import (
	"fmt"
	"image"
	"io"
	"sync/atomic"
	"time"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/hw/input"
	"nestor/ines"
)

type Output interface {
	BeginFrame() hw.Frame
	EndFrame(hw.Frame)
	Poll() bool
	Close()
	Screenshot() image.Image
}

type Config struct {
	Input input.InputConfig `toml:"input"`
	Video VideoConfig       `toml:"video"`

	TraceOut io.WriteCloser `toml:"-"`
}

type VideoConfig struct {
	DisableVSync bool `toml:"disable_vsync"`
	Monitor      int  `toml:"monitor"`
}

type Emulator struct {
	NES *NES
	out Output

	quit    atomic.Bool
	paused  atomic.Bool
	reset   atomic.Bool
	restart atomic.Bool
}

// Launch instantiates an emulator, setup controllers, output streams and window.
func Launch(rom *ines.Rom, cfg Config) (*Emulator, error) {
	nes, err := powerUp(rom)
	if err != nil {
		return nil, fmt.Errorf("power up failed: %s", err)
	}

	// Output setup.
	out := hw.NewOutput(hw.OutputConfig{
		Width:           256,
		Height:          240,
		NumVideoBuffers: 2,
		Title:           "Nestor",
		ScaleFactor:     2,
		DisableVSync:    cfg.Video.DisableVSync,
		Monitor:         int32(cfg.Video.Monitor),
	})
	if err := out.EnableVideo(true); err != nil {
		return nil, err
	}

	input, err := input.NewProvider(cfg.Input)
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
		if e.isPaused() {
			// Don't burn cpu while paused.
			time.Sleep(100 * time.Millisecond)
		} else {
			e.RunOneFrame()
		}
		if e.shouldStop() {
			e.out.Close()
			break
		}
		e.handleReset()
	}
	log.ModEmu.InfoZ("Emulation loop exited").End()
}

// SetPause, Stop, Reset and Restart allows to control
// the emulator loop in a concurrent-safe way.

func (e *Emulator) SetPause(pause bool) { e.paused.CompareAndSwap(!pause, pause) }
func (e *Emulator) Stop()               { e.quit.Store(true) }
func (e *Emulator) Reset()              { e.reset.Store(true) }
func (e *Emulator) Restart()            { e.restart.Store(true) }

func (e *Emulator) isPaused() bool {
	return e.paused.Load()
}

func (e *Emulator) shouldStop() bool {
	return e.quit.Load() || !e.out.Poll() || e.NES.CPU.IsHalted()
}

func (e *Emulator) handleReset() {
	if e.reset.CompareAndSwap(true, false) {
		log.ModEmu.InfoZ("Performing soft reset").End()
		e.NES.Reset(true)
	} else if e.restart.CompareAndSwap(true, false) {
		log.ModEmu.InfoZ("Performing hard reset").End()
		e.NES.Reset(false)
	}
}
