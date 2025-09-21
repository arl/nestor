package ebit

import (
	"os"
	"path/filepath"
	"strings"

	"nestor/config"
	"nestor/emu/log"
)

const saveRAMExtension = ".nrr"

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

	path := filepath.Join(config.SaveRAMDir(), file)
	if _, err := os.Stat(path); err != nil {
		log.ModEmu.DebugZ("rom has no saveram file").String("rom", romname).String("path", path).End()
		return path, false
	}
	return path, true
}
