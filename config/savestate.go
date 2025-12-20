package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const savestateExtenstion = ".nss"

var SavestatesDir = sync.OnceValue(func() string {
	dir := filepath.Join(Dir(), "savestates")
	if err := os.MkdirAll(dir, 0755); err != nil {
		modCfg.ErrorZ("failed to create directory").String("path", dir).Error("err", err).End()
	}
	return dir
})

// AddSavestate persits data about a savestate on the filesystem.
func AddSavestate(romName string, slot int, state []byte) error {
	fname := fmt.Sprintf("%s.%d%s", removeExt(romName), slot+1, savestateExtenstion)
	path := filepath.Join(SavestatesDir(), fname)
	return os.WriteFile(path, state, 0644)
}

// SavestateInfo returns the modification time of a savestate slot.
// Returns zero time if the slot is empty or an error occurred.
func SavestateInfo(romName string, slot int) time.Time {
	fname := fmt.Sprintf("%s.%d%s", removeExt(romName), slot+1, savestateExtenstion)
	path := filepath.Join(SavestatesDir(), fname)

	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// LoadSavestate loads a savestate slot.
func LoadSavestate(romName string, slot int) ([]byte, error) {
	fname := fmt.Sprintf("%s.%d%s", removeExt(romName), slot+1, savestateExtenstion)
	path := filepath.Join(SavestatesDir(), fname)
	return os.ReadFile(path)
}
