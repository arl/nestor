package ui

import (
	"archive/zip"
	"bytes"
	"cmp"
	"fmt"
	_ "image/png" // Import for decoding PNG screenshots
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"nestor/config"
)

const recentROMExtension = ".nrr"

var RecentROMsDir = sync.OnceValue(func() string {
	dir := filepath.Join(config.Dir(), "recent-roms")
	if err := os.MkdirAll(dir, 0755); err != nil {
		modUI.ErrorZ("failed to create directory").String("path", dir).Error("err", err).End()
	}
	return dir
})

// recentROM holds the data for a single recently played ROM.
type recentROM struct {
	Name      string    // Base name of the ROM file.
	Path      string    // Full path to the ROM file.
	Image     []byte    // PNG data for the screenshot.
	SaveState []byte    // Saved state data.
	LastUsed  time.Time // Last time the ROM was played.
}

func (r recentROM) IsValid() bool {
	return r.Path != "" && r.Image != nil && r.Name != "" && !r.LastUsed.IsZero()
}

// save writes the recent ROM data to a .nrr file.
func (r recentROM) save() error {
	f, err := os.Create(filepath.Join(RecentROMsDir(), r.Name+recentROMExtension))
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	zfw, err := zw.Create("infos.txt")
	if err != nil {
		return err
	}
	if _, err := zfw.Write([]byte(r.Path)); err != nil {
		return err
	}

	zfw, err = zw.Create("screenshot.png")
	if err != nil {
		return err
	}
	if _, err := zfw.Write(r.Image); err != nil {
		return err
	}

	zfw, err = zw.Create("state.bin")
	if err != nil {
		return err
	}
	if _, err := zfw.Write(r.SaveState); err != nil {
		return err
	}

	return nil
}

// addRecentROM is the new entry point for saving a ROM to the recent list.
// Call this function when you run a new ROM.
func addRecentROM(rom recentROM) error {
	if err := rom.save(); err != nil {
		return fmt.Errorf("failed to save recent ROM: %w", err)
	}
	return nil
}

func readZipFile(zf *zip.File) ([]byte, error) {
	zfr, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer zfr.Close()
	return io.ReadAll(zfr)
}

func removeExt(path string) string {
	return path[:len(path)-len(filepath.Ext(path))]
}

func loadRecentROMs() []recentROM {
	var roms []recentROM

	err := filepath.WalkDir(RecentROMsDir(), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != recentROMExtension {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open recent ROM file: %w", err)
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat recent ROM file: %w", err)
		}

		zr, err := zip.NewReader(f, info.Size())
		if err != nil {
			return fmt.Errorf("failed to read recent ROM zip: %w", err)
		}

		loaded := make(map[string][]byte)
		for _, zf := range zr.File {
			buf, err := readZipFile(zf)
			if err != nil {
				fmt.Printf("warning: could not read %s from %s: %v\n", zf.Name, path, err)
				continue
			}
			loaded[zf.Name] = buf
		}

		cur := recentROM{
			Name:      removeExt(info.Name()),
			LastUsed:  info.ModTime(),
			Path:      string(bytes.TrimSpace(loaded["infos.txt"])),
			Image:     loaded["screenshot.png"],
			SaveState: loaded["state.bin"],
		}

		if cur.IsValid() {
			roms = append(roms, cur)
		}

		return nil
	})
	if err != nil {
		modUI.WarnZ("error loading recent roms").Error("err", err).End()
	}

	// Normalize: remove duplicates and sort by LastUsed date.
	m := make(map[string]recentROM, len(roms))
	for _, rom := range roms {
		m[rom.Name] = rom
	}
	roms = roms[:0]
	for _, rom := range m {
		roms = append(roms, rom)
	}
	slices.SortFunc(roms, func(a, b recentROM) int {
		return cmp.Compare(b.LastUsed.Unix(), a.LastUsed.Unix())
	})

	const maxRecentsRoms = 16 // Limit the number of recent roms to show
	if len(roms) > maxRecentsRoms {
		roms = roms[:maxRecentsRoms]
	}

	return roms
}
