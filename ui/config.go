package ui

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/veandco/go-sdl2/sdl"

	"nestor/emu"
	"nestor/emu/log"
	"nestor/hw/input"
	"nestor/hw/shaders"
)

type GeneralConfig struct {
	ShowSplash bool `toml:"show_splash"`
}

type Config struct {
	emu.Config
	General GeneralConfig `toml:"general"`
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
					Buttons: [8]input.Code{
						// TODO: change this to QWERTY layout?
						{Scancode: sdl.SCANCODE_W, Type: input.KeyboardCtrl},
						{Scancode: sdl.SCANCODE_Q, Type: input.KeyboardCtrl},
						{Scancode: sdl.SCANCODE_A, Type: input.KeyboardCtrl},
						{Scancode: sdl.SCANCODE_S, Type: input.KeyboardCtrl},
						{Scancode: sdl.SCANCODE_UP, Type: input.KeyboardCtrl},
						{Scancode: sdl.SCANCODE_DOWN, Type: input.KeyboardCtrl},
						{Scancode: sdl.SCANCODE_LEFT, Type: input.KeyboardCtrl},
						{Scancode: sdl.SCANCODE_RIGHT, Type: input.KeyboardCtrl},
					},
				},
			},
		},
		Video: emu.VideoConfig{
			DisableVSync: false,
			Monitor:      0,
			Shader:       shaders.DefaultName,
		},
		Audio: emu.AudioConfig{
			DisableAudio: false,
		},
		TraceOut: nil,
	},
	General: GeneralConfig{
		ShowSplash: true,
	},
}

const dirMode = os.FileMode(0755)

var ConfigDir = sync.OnceValue(func() string {
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

var configPath = sync.OnceValue(func() string {
	return filepath.Join(ConfigDir(), cfgFilename)
})

// LoadConfigOrDefault loads the configuration from the nestor config directory,
// or provide a default one.
func LoadConfigOrDefault() Config {
	// Create a config based on the default one.
	cfg := defaultConfig

	// Load the config from the file, overwriting the default values.
	_, err := toml.DecodeFile(configPath(), &cfg)
	if err != nil {
		log.ModEmu.Warnf("Failed to load config, using default: %v", err)
	}

	// Apply post-load operations (fix invalid values, etc).
	cfg.Input.PostLoad()
	cfg.Video.Check()
	log.ModEmu.Infof("Configuration loaded from %s", configPath())
	return cfg
}

// saveConfig into nestor config directory.
func saveConfig(cfg *Config) error {
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath(), buf, 0644); err != nil {
		return err
	}

	log.ModEmu.Infof("Configuration saved to %s", configPath())
	return nil
}
