package ui

import (
	"os"
	"path/filepath"
	"sync"

	"nestor/emu"

	"github.com/BurntSushi/toml"
	"github.com/kirsle/configdir"
)

var ConfigDir string = sync.OnceValue(func() string {
	dir := configdir.LocalConfig("nestor")
	if err := configdir.MakePath(dir); err != nil {
		modGUI.Fatalf("failed to create directory %s: %v", dir, err)
	}
	return dir
})()

const cfgFilename = "config.toml"

func LoadConfigOrDefault() emu.Config {
	var cfg emu.Config
	_, err := toml.DecodeFile(filepath.Join(ConfigDir, cfgFilename), &cfg)
	if err != nil {
		// TODO: specify default config
		return emu.Config{}
	}
	return cfg
}

func SaveConfig(cfg emu.Config) error {
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(ConfigDir, cfgFilename), buf, 0644)
}
