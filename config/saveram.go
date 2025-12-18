package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"nestor/emu/log"
)

const saveRAMExtension = ".nrr"

var SaveRAMDir = sync.OnceValue(func() string {
	dir := filepath.Join(Dir(), "saveram")
	if err := os.MkdirAll(dir, dirMode); err != nil {
		modCfg.Fatalf("failed to create directory %s: %v", dir, err)
	}

	return dir
})

// getSaveRAMPath returns the full path of the save ram file for the given rom,
// and a boolean indicating whether that file exists.
func getSaveRAMPath(rompath string) (string, bool) {
	file := ""
	romname := filepath.Base(rompath)
	dot := strings.LastIndexByte(romname, '.')
	if dot == -1 {
		file = romname + ".sav"
	} else {
		file = romname[:dot] + ".sav"
	}

	path := filepath.Join(SaveRAMDir(), file)
	if _, err := os.Stat(path); err != nil {
		log.ModEmu.DebugZ("rom has no saveram file").String("rom", romname).String("path", path).End()
		return path, false
	}
	return path, true
}
