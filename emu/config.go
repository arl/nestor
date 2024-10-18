package emu

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"nestor/emu/log"
	"nestor/hw"

	"github.com/BurntSushi/toml"
	"github.com/kirsle/configdir"
)

type Config struct {
	Input   hw.InputConfig `toml:"input"`
	Video   VideoConfig    `toml:"video"`
	General GeneralConfig  `toml:"general"`

	TraceOut io.WriteCloser `toml:"-"`
}

type GeneralConfig struct {
	ShowSplash bool `toml:"show_splash"`
}

type VideoConfig struct {
	DisableVSync bool `toml:"disable_vsync"`
}

var ConfigDir string = sync.OnceValue(func() string {
	dir := configdir.LocalConfig("nestor")
	if err := configdir.MakePath(dir); err != nil {
		log.ModEmu.Fatalf("failed to create directory %s: %v", dir, err)
	}
	return dir
})()

const cfgFilename = "config.toml"

// LoadConfigOrDefault loads the configuration from the nestor config directory,
// or provide a default one.
func LoadConfigOrDefault() Config {
	var cfg Config
	_, err := toml.DecodeFile(filepath.Join(ConfigDir, cfgFilename), &cfg)
	if err != nil {
		// TODO: specify default config
		return Config{}
	}
	return cfg
}

// SaveConfig into nestor config directory.
func SaveConfig(cfg Config) error {
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(ConfigDir, cfgFilename), buf, 0644)
}
