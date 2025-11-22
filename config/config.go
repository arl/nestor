package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/hajimehoshi/ebiten/v2"

	"nestor/emu"
	"nestor/emu/log"
	"nestor/hw/input"
)

type General struct {
	ShowSplash       bool   `toml:"show_splash"`
	FileLoadStartDir string `toml:"file_load_start_dir"`
}

type Config struct {
	emu.Config
	General General `toml:"general"`
}

var defaultConfig = Config{
	Config: emu.Config{
		Input: input.Config{
			Paddles: [2]input.PaddleConfig{
				{
					Plugged:      true,
					PaddlePreset: 0,
				},
				{
					Plugged:      false,
					PaddlePreset: 1,
				},
			},
			Presets: [8]input.PaddlePreset{
				{
					Keyboard: &input.KeyboardMapping{
						A:      ebiten.KeyS,
						B:      ebiten.KeyA,
						Select: ebiten.KeyQ,
						Start:  ebiten.KeyW,
						Up:     ebiten.KeyUp,
						Down:   ebiten.KeyDown,
						Left:   ebiten.KeyLeft,
						Right:  ebiten.KeyRight,
					},
				},
			},
		},
		Video: emu.VideoConfig{
			DisableVSync:    false,
			StartFullscreen: false,
			Monitor:         0,
			Shader:          "passthrough",
		},
		Audio: emu.AudioConfig{
			DisableAudio: false,
		},
		Emulation: emu.EmulationConfig{
			RunAheadFrames: 0,
		},
		TraceOut: nil,
	},
	General: General{
		ShowSplash: true,
	},
}

const dirMode = os.FileMode(0755)

var Dir = sync.OnceValue(func() string {
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		log.ModEmu.Fatalf("failed to get user config directory: %v", err)
	}

	dir := filepath.Join(cfgdir, "nestor")
	if err := os.MkdirAll(dir, dirMode); err != nil {
		log.ModEmu.Fatalf("failed to create directory %s: %v", dir, err)
	}
	return dir
})

const cfgFilename = "config.toml"

var Path = sync.OnceValue(func() string {
	return filepath.Join(Dir(), cfgFilename)
})

// LoadOrDefault loads the configuration from the nestor config directory, or
// provide a default one.
func LoadOrDefault() Config {
	// Create a config based on the default one.
	cfg := defaultConfig

	// Load the config from the file, overwriting the default values.
	_, err := toml.DecodeFile(Path(), &cfg)
	if err != nil {
		log.ModEmu.Warnf("Failed to load config, using default: %v", err)
	}

	// Apply post-load operations (fix invalid values, etc).
	cfg.Video.Check()
	cfg.Emulation.Check()
	log.ModEmu.InfoZ("loaded configuration").String("path", Path()).End()
	return cfg
}

// Save saves configuration into the default nestor config path.
func Save(cfg *Config) error {
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(Path(), buf, 0644); err != nil {
		return err
	}

	log.ModEmu.Infof("configuration saved to %s", Path())
	return nil
}

var SaveRAMDir = sync.OnceValue(func() string {
	dir := filepath.Join(Dir(), "saveram")
	if err := os.MkdirAll(dir, dirMode); err != nil {
		log.ModEmu.Fatalf("failed to create directory %s: %v", dir, err)
	}

	return dir
})
