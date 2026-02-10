package emu

import (
	"io"
	"slices"

	"github.com/arl/nestor/emu/log"
	"github.com/arl/nestor/hw/hwinput"
	"github.com/arl/nestor/ui/shader"
)

type Config struct {
	Input     hwinput.Config  `toml:"input"`
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
	VSync           bool   `toml:"vsync"`
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
