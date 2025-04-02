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
	Input input.Config `toml:"input"`
	Video VideoConfig  `toml:"video"`
	Audio AudioConfig  `toml:"audio"`

	TraceOut io.WriteCloser `toml:"-"`
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
	}, nil
}

func (e *Emulator) RunOneFrame() {
	frame := e.out.BeginFrame()
	e.NES.RunOneFrame(frame)
	e.out.EndFrame(frame)
}

func (e *Emulator) loop() {
	for {
		// Handle pause.
		if e.isPaused() {
			// Don't burn cpu while paused.
			time.Sleep(100 * time.Millisecond)
		} else {
			e.RunOneFrame()
		}
		if !e.out.Poll() || e.shouldStop() {
			e.out.Close()
			break
		}
		e.handleReset()
	}
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
