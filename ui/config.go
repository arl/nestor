package ui

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/kirsle/configdir"

	"nestor/hw"
)

var ConfigDir string = sync.OnceValue(func() string {
	dir := configdir.LocalConfig("nestor")
	if err := configdir.MakePath(dir); err != nil {
		modGUI.Fatalf("failed to create directory %s: %v", dir, err)
	}
	return dir
})()

const cfgFilename = "config.toml"

type Config struct {
	Input hw.InputConfig `toml:"input"`
}

func LoadConfigOrDefault() Config {
	var cfg Config
	_, err := toml.DecodeFile(filepath.Join(ConfigDir, cfgFilename), &cfg)
	if err != nil {
		// TODO: specify default config
		return Config{}
	}
	return cfg
}

func SaveConfig(cfg Config) error {
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(ConfigDir, cfgFilename), buf, 0644)
}
