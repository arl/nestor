package emu

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/hw/input"
	"nestor/ines"
	"nestor/ui/shader"
)

type Config struct {
	Input     input.Config    `toml:"input"`
	Video     VideoConfig     `toml:"video"`
	Audio     AudioConfig     `toml:"audio"`
	Emulation EmulationConfig `toml:"emulation"`

	TraceOut io.WriteCloser `toml:"-"`
}

type EmulationConfig struct {
	RunAheadFrames uint `toml:"run_ahead_frames"`
}

func (ecfg *EmulationConfig) Check() {
	// Max out the number of run-ahead frames to 10.
	ecfg.RunAheadFrames = min(ecfg.RunAheadFrames, 10)
}

type VideoConfig struct {
	DisableVSync    bool   `toml:"disable_vsync"`
	StartFullscreen bool   `toml:"start_fullscreen"`
	Monitor         uint   `toml:"monitor"`
	Shader          string `toml:"shader"`
}

func (vcfg *VideoConfig) Check() {
	// Ensure we have a valid shader.
	if vcfg.Shader == "" {
		vcfg.Shader = shader.Default
	}
	if !slices.Contains(shader.Names(), vcfg.Shader) {
		log.ModEmu.Warnf("Invalid shader name %q, fallback to %q", vcfg.Shader, shader.Default)
		vcfg.Shader = shader.Default
	}
}

type AudioConfig struct {
	DisableAudio bool `toml:"disable_audio"`
}

type Emulator struct {
	NES *NES
	out *Output
	cfg EmulationConfig

	// These are accessed concurrently by the emulator loop and the UI.
	quit    atomic.Bool
	paused  atomic.Bool
	reset   atomic.Bool
	restart atomic.Bool

	blockch chan struct{}

	tmpdir string
}

// Launch starts the various hardware subsystems, shows the window, setups the
// video and audio streams and plugs controllers. It doesn't start the emulation
// loop, call Run() for that.
func Launch(rom *ines.Rom, cfg Config, out *Output, inp hw.InputStateLoader) (*Emulator, error) {
	nes, err := powerUp(rom)
	if err != nil {
		return nil, fmt.Errorf("power up failed: %s", err)
	}

	nes.CPU.PlugInputDevice(inp)

	// CPU execution trace setup.
	if cfg.TraceOut != nil {
		nes.CPU.SetTraceOutput(cfg.TraceOut)
	}

	return &Emulator{
		NES:     nes,
		out:     out,
		cfg:     cfg.Emulation,
		blockch: make(chan struct{}, 1),
	}, nil
}

func (e *Emulator) RunOneFrame() {
	if e.cfg.RunAheadFrames > 0 {
		e.RunFrameWithRunAhead()
	} else {
		frame := e.out.BeginFrame()
		e.NES.RunOneFrame(&frame)
		e.out.EndFrame(&frame)
	}
}

func (e *Emulator) RunFrameWithRunAhead() {
	frames := e.cfg.RunAheadFrames

	// Run a single frame, make a snapshot, but do not render video nor play
	// audio out of it.
	e.NES.isRunAheadFrame = true
	e.NES.CPU.EnableTrace(true)
	e.NES.CPU.Run(29781)
	e.NES.APU.EndFrame(nil)
	e.NES.CPU.EnableTrace(false)

	buf, err := e.NES.SaveSnapshot()
	if err != nil {
		log.ModEmu.PanicZ("failed run-ahead frame snapshot").Error("err", err).End()
	}

	for frames > 1 {
		e.NES.CPU.Run(29781)
		e.NES.APU.EndFrame(nil)
		frames--
	}
	e.NES.isRunAheadFrame = false

	// Run one frame normally.
	frame := e.out.BeginFrame()
	e.NES.RunOneFrame(&frame)
	e.out.EndFrame(&frame)

	e.NES.isRunAheadFrame = true
	if err := e.NES.LoadSnapshot(buf); err != nil {
		log.ModEmu.PanicZ("failed to load snapshot").Error("err", err).End()
	}
	e.NES.isRunAheadFrame = false
}

func (e *Emulator) Run() {
	for !e.shouldStop() {
		if e.isPaused() {
			<-e.blockch
		} else {
			e.RunOneFrame()
		}
		e.handleReset()
	}

	log.ModEmu.InfoZ("Emulation loop exited").End()

	if e.tmpdir != "" {
		e.save()
	}
}

func (e *Emulator) handleReset() {
	if e.reset.CompareAndSwap(true, false) {
		log.ModEmu.InfoZ("Performing soft reset").End()
		frame := e.out.BeginFrame()
		e.NES.RunResetFrame(&frame, true)
		e.out.EndFrame(&frame)
	} else if e.restart.CompareAndSwap(true, false) {
		log.ModEmu.InfoZ("Performing hard reset").End()
		frame := e.out.BeginFrame()
		e.NES.RunResetFrame(&frame, false)
		e.out.EndFrame(&frame)
	}
}

// Block, Resume, Reset, Restart and Stop allow to control
// the emulator loop in a concurrent-safe way.

func (e *Emulator) Block()   { e.paused.Store(true) }
func (e *Emulator) Reset()   { e.reset.Store(true) }
func (e *Emulator) Restart() { e.restart.Store(true) }
func (e *Emulator) Stop()    { e.quit.Store(true) }

func (e *Emulator) Resume() {
	e.paused.Store(false)
	select {
	case e.blockch <- struct{}{}:
	default:
		// avoid deadlock if we're not waiting blocked
	}
}

func (e *Emulator) isPaused() bool {
	return e.paused.Load()
}

func (e *Emulator) shouldStop() bool {
	return e.quit.Load() || e.NES.CPU.IsHalted()
}

func (e *Emulator) save() {
	// Save state
	state, err := e.NES.SaveSnapshot()
	if err != nil {
		log.ModEmu.WarnZ("Failed to save state").Error("err", err).End()
		return
	}

	// TODO: state not saved for now
	_ = state

	path := filepath.Join(e.tmpdir, "screenshot.png")

	if err := SaveAsPNG(e.out.Screenshot(), path); err != nil {
		log.ModEmu.WarnZ("Error while saving screenshot").String("path", path).End()
	} else {
		log.ModEmu.DebugZ("Saved screenshot").String("path", path).End()
	}

	if saveram := e.NES.Mapper.BatteryPackedRAM(); saveram != nil {
		path = filepath.Join(e.tmpdir, "battery.sav")
		if err := os.WriteFile(path, saveram, 0644); err != nil {
			log.ModEmu.WarnZ("Error while saving save ram").String("path", path).End()
		} else {
			log.ModEmu.DebugZ("Saved save ram").String("path", path).End()
		}
	}
}

func (e *Emulator) SetTempDir(path string) { e.tmpdir = path }
