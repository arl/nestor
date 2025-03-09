package emu

import (
	"fmt"
	"image"
	"io"
	"slices"
	"sync/atomic"
	"time"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/hw/input"
	"nestor/hw/shaders"
	"nestor/ines"
)

type Output interface {
	BeginFrame() hw.Frame
	EndFrame(hw.Frame)
	Poll() bool
	Close()
	Screenshot() *image.RGBA
}

type Config struct {
	Input input.Config `toml:"input"`
	Video VideoConfig  `toml:"video"`

	TraceOut io.WriteCloser `toml:"-"`
}

type VideoConfig struct {
	DisableVSync bool   `toml:"disable_vsync"`
	Monitor      int32  `toml:"monitor"`
	Shader       string `toml:"shader"`
}

func (vcfg *VideoConfig) Init() {
	// Ensure we have a valid shader.
	if vcfg.Shader == "" {
		vcfg.Shader = shaders.DefaultName
	}
	if !slices.Contains(shaders.Names(), vcfg.Shader) {
		log.ModEmu.Warnf("Invalid shader name %q, fallback to %q", vcfg.Shader, shaders.DefaultName)
		vcfg.Shader = shaders.DefaultName
	}
}

type Emulator struct {
	NES *NES
	out Output

	quit    atomic.Bool
	paused  atomic.Bool
	reset   atomic.Bool
	restart atomic.Bool
}

// Launch starts the various hardware subsystems, shows the window, setups the
// video and audio streams and plugs controllers. It doesn't start the emulation
// loop, call Run() for that.
func Launch(rom *ines.Rom, cfg Config) (*Emulator, error) {
	nes, err := powerUp(rom)
	if err != nil {
		return nil, fmt.Errorf("power up failed: %s", err)
	}

	// Output setup.
	out := hw.NewOutput(hw.OutputConfig{
		Width:           hw.NTSCWidth,
		Height:          hw.NTSCHeight,
		NumVideoBuffers: 2,
		Title:           "Nestor",
		ScaleFactor:     2,
		DisableVSync:    cfg.Video.DisableVSync,
		Monitor:         cfg.Video.Monitor,
		Shader:          cfg.Video.Shader,
	})
	if err := out.EnableVideo(true); err != nil {
		return nil, err
	}
	if err := out.EnableAudio(true); err != nil {
		return nil, err
	}

	inprov := input.NewProvider(cfg.Input)
	nes.CPU.PlugInputDevice(inprov)

	// CPU execution trace setup.
	if cfg.TraceOut != nil {
		nes.CPU.SetTraceOutput(cfg.TraceOut)
	}

	return &Emulator{
		NES: nes,
		out: out,
	}, nil
}

// RaiseWindow raises the emulator window above others and sets the input focus.
func (e *Emulator) RaiseWindow() {
	if hwout, ok := e.out.(*hw.Output); ok {
		hwout.FocusWindow()
	}
}

func (e *Emulator) Screenshot() *image.RGBA {
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
func (e *Emulator) Reset()              { e.reset.Store(true) }
func (e *Emulator) Restart()            { e.restart.Store(true) }
func (e *Emulator) Stop() *image.RGBA {
	e.quit.Store(true)
	return e.Screenshot()
}

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
