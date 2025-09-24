package emu

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"

	"nestor/emu/log"
	"nestor/hw/input"
	"nestor/hw/shaders"
	"nestor/ines"
)

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
	out *Output
	cfg EmulationConfig

	quit atomic.Bool

	tmpdir string
}

// Launch starts the various hardware subsystems, shows the window, setups the
// video and audio streams and plugs controllers. It doesn't start the emulation
// loop, call Run() for that.
func Launch(rom *ines.Rom, cfg Config, out *Output, inp *input.Provider) (*Emulator, error) {
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
		NES: nes,
		out: out,
		cfg: cfg.Emulation,
	}, nil
}

func (e *Emulator) RunOneFrame() {
	log.ModEmu.DebugZ("frame").Uint32("number", e.NES.PPU.FrameCount).End()

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
	for !e.quit.Load() {
		e.RunOneFrame()
	}

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

// Stop stops the emulator loop in a concurrent safe way.
func (e *Emulator) Stop() { e.quit.Store(true) }

// SoftReset performs a soft reset synchronously (should be called from UI thread)
func (e *Emulator) SoftReset() {
	log.ModEmu.InfoZ("Performing soft reset").End()
	e.NES.Reset(true)
}

// HardReset performs a hard reset synchronously (should be called from UI thread)
func (e *Emulator) HardReset() {
	log.ModEmu.InfoZ("Performing hard reset").End()
	e.NES.Reset(false)
}
