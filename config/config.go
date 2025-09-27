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
						{Scancode: ebiten.KeyW, Type: input.KeyboardCtrl},
						{Scancode: ebiten.KeyQ, Type: input.KeyboardCtrl},
						{Scancode: ebiten.KeyA, Type: input.KeyboardCtrl},
						{Scancode: ebiten.KeyS, Type: input.KeyboardCtrl},
						{Scancode: ebiten.KeyUp, Type: input.KeyboardCtrl},
						{Scancode: ebiten.KeyDown, Type: input.KeyboardCtrl},
						{Scancode: ebiten.KeyLeft, Type: input.KeyboardCtrl},
						{Scancode: ebiten.KeyRight, Type: input.KeyboardCtrl},
					},
				},
			},
		},
		Video: emu.VideoConfig{
			DisableVSync: false,
			Monitor:      0,
			Shader:       "",
		},
		Audio: emu.AudioConfig{
			DisableAudio: false,
		},
		Emulation: emu.EmulationConfig{
			RunAheadFrames: 0,
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

var SaveRAMDir = sync.OnceValue(func() string {
	dir := filepath.Join(ConfigDir(), "saveram")
	if err := os.MkdirAll(dir, dirMode); err != nil {
		log.ModEmu.Fatalf("failed to create directory %s: %v", dir, err)
	}

	return dir
})
