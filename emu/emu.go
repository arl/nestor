package emu

import (
	"fmt"
	"image"
	"io"
	"path/filepath"
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
	Input     input.Config    `toml:"input"`
	Video     VideoConfig     `toml:"video"`
	Audio     AudioConfig     `toml:"audio"`
	Emulation EmulationConfig `toml:"emulation"`

	TraceOut io.WriteCloser `toml:"-"`
}

type EmulationConfig struct {
	RunAheadFrames int `toml:"run_ahead_frames"`
}

type VideoConfig struct {
	DisableVSync bool   `toml:"disable_vsync"`
	Monitor      int32  `toml:"monitor"`
	Shader       string `toml:"shader"`
}

func (vcfg *VideoConfig) Check() {
	// Ensure we have a valid shader.
	if vcfg.Shader == "" {
		vcfg.Shader = shaders.DefaultName
	}
	if !slices.Contains(shaders.Names(), vcfg.Shader) {
		log.ModEmu.Warnf("Invalid shader name %q, fallback to %q", vcfg.Shader, shaders.DefaultName)
		vcfg.Shader = shaders.DefaultName
	}
}

type AudioConfig struct {
	DisableAudio bool `toml:"disable_audio"`
}

type Emulator struct {
	NES *NES
	out Output
	cfg EmulationConfig

	// These are accessed concurrently by the emulator loop and the UI.
	quit    atomic.Bool
	paused  atomic.Bool
	reset   atomic.Bool
	restart atomic.Bool

	tmpdir string
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
		Width:          hw.NTSCWidth,
		Height:         hw.NTSCHeight,
		NumBackBuffers: 2,
		Title:          "Nestor",
		ScaleFactor:    2,
		DisableVSync:   cfg.Video.DisableVSync,
		Monitor:        cfg.Video.Monitor,
		Shader:         cfg.Video.Shader,
	})
	if err := out.EnableVideo(true); err != nil {
		return nil, err
	}

	if cfg.Audio.DisableAudio {
		log.ModEmu.WarnZ("Audio disabled").End()
	} else {
		if err := out.EnableAudio(true); err != nil {
			return nil, err
		}
		log.ModEmu.InfoZ("Audio enabled").End()
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
		cfg: cfg.Emulation,
	}, nil
}

func (e *Emulator) RunOneFrame() {
	if e.cfg.RunAheadFrames > 0 {
		e.RunFrameWithRunAhead()
	} else {
		frame := e.out.BeginFrame()
		e.NES.RunOneFrame(frame)
		e.out.EndFrame(frame)
	}
}

func (e *Emulator) RunFrameWithRunAhead() {
	frames := e.cfg.RunAheadFrames

	// Run a single frame, make a snapshot, but do not render video nor play
	// audio out of it.
	e.NES.isRunAheadFrame = true
	e.NES.CPU.Run(29781)
	e.NES.APU.EndFrame()

	buf, err := e.NES.SaveSnapshot()
	if err != nil {
		log.ModEmu.PanicZ("failed run-ahead frame snapshot").Error("err", err).End()
	}

	for frames > 1 {
		e.NES.CPU.Run(29781)
		e.NES.APU.EndFrame()
		frames--
	}
	e.NES.isRunAheadFrame = false

	// Run one frame normally.
	frame := e.out.BeginFrame()
	e.NES.RunOneFrame(frame)
	e.out.EndFrame(frame)

	e.NES.isRunAheadFrame = true
	if err := e.NES.LoadSnapshot(buf); err != nil {
		log.ModEmu.PanicZ("failed to load snapshot").Error("err", err).End()
	}
	e.NES.isRunAheadFrame = false
}

func (e *Emulator) loop() {
	for e.out.Poll() {
		// Handle pause.
		if e.isPaused() {
			// Don't burn cpu while paused.
			time.Sleep(100 * time.Millisecond)
		} else {
			e.RunOneFrame()
		}
		if e.shouldStop() {
			break
		}
		e.handleReset()
	}

	e.out.Close()
}

// RaiseWindow raises the emulator window above others and sets the input focus.
func (e *Emulator) RaiseWindow() {
	if hwout, ok := e.out.(*hw.Output); ok {
		hwout.FocusWindow()
	}
}

func (e *Emulator) Run() {
	e.loop()
	log.ModEmu.InfoZ("Emulation loop exited").End()

	if e.tmpdir != "" {
		e.save()
	}
}

func (e *Emulator) save() {
	// Save state
	state, err := e.NES.SaveSnapshot()
	if err != nil {
		log.ModEmu.WarnZ("Failed to save state").Error("err", err).End()
		return
	}

	fmt.Printf("save state: %d bytes\n", len(state))

	path := filepath.Join(e.tmpdir, "screenshot.png")
	if err := hw.SaveAsPNG(e.out.Screenshot(), path); err != nil {
		log.ModEmu.WarnZ("Failed to save screenshot").String("path", path).End()
	}
}

func (e *Emulator) SetTempDir(path string) { e.tmpdir = path }

// SetPause, Stop, Reset and Restart allows to control
// the emulator loop in a concurrent-safe way.

func (e *Emulator) SetPause(pause bool) { e.paused.CompareAndSwap(!pause, pause) }
func (e *Emulator) Reset()              { e.reset.Store(true) }
func (e *Emulator) Restart()            { e.restart.Store(true) }
func (e *Emulator) Stop() {
	e.quit.Store(true)
}

func (e *Emulator) isPaused() bool {
	return e.paused.Load()
}

func (e *Emulator) shouldStop() bool {
	return e.quit.Load() || e.NES.CPU.IsHalted()
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
