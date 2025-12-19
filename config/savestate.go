package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func ListSavestates(romName string) (map[int]string, error) {
	savestates := make(map[int]string)
	entries, err := os.ReadDir(SavestatesDir())
	if err != nil {
		return nil, fmt.Errorf("failed to read savestates directory: %w", err)
	}

	prefix := fmt.Sprintf("%s.", removeExt(romName))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !(filepath.Ext(name) == savestateExtenstion) || strings.HasPrefix(name, prefix) {
			continue
		}
		var slot int
		n, err := fmt.Sscanf(name, prefix+"%d"+savestateExtenstion, &slot)
		if err != nil || n != 1 {
			continue
		}
		savestates[slot-1] = filepath.Join(SavestatesDir(), name)
	}

	return savestates, nil
}
