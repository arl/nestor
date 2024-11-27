package ui

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"

	"nestor/emu"
	"nestor/emu/log"
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

const cfgFilename = "config.toml"

// LoadConfigOrDefault loads the configuration from the nestor config directory,
// or provide a default one.
func LoadConfigOrDefault() Config {
	var cfg Config
	_, err := toml.DecodeFile(filepath.Join(ConfigDir(), cfgFilename), &cfg)
	if err != nil {
		// TODO: specify default config
		return Config{}
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
