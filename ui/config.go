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
)

type GeneralConfig struct {
	ShowSplash bool `toml:"show_splash"`
}

type Config struct {
	emu.Config
	General GeneralConfig `toml:"general"`
}

const DefaultFileMode = os.FileMode(0755)

var ConfigDir = sync.OnceValue(func() string {
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		log.ModEmu.Fatalf("failed to get user config directory: %v", err)
	}

	dir := filepath.Join(cfgdir, "nestor")
	if err := os.MkdirAll(dir, DefaultFileMode); err != nil {
		log.ModEmu.Fatalf("failed to create directory %s: %v", dir, err)
	}
	return dir
})

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
			Shader:       "No shader",
		},
		TraceOut: nil,
	},
	General: GeneralConfig{
		ShowSplash: true,
	},
}

const cfgFilename = "config.toml"

// LoadConfigOrDefault loads the configuration from the nestor config directory,
// or provide a default one.
func LoadConfigOrDefault() Config {
	var cfg Config
	_, err := toml.DecodeFile(filepath.Join(ConfigDir(), cfgFilename), &cfg)
	if err != nil {
		return defaultConfig
	}
	cfg.Input.Init()
	return cfg
}

// SaveConfig into nestor config directory.
func SaveConfig(cfg Config) error {
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(ConfigDir(), cfgFilename), buf, 0644)
}
