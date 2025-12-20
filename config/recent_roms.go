package config

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
)

const recentROMExtension = ".nrr"

var RecentROMsDir = sync.OnceValue(func() string {
	dir := filepath.Join(Dir(), "recent-roms")
	if err := os.MkdirAll(dir, 0755); err != nil {
		modCfg.ErrorZ("failed to create directory").String("path", dir).Error("err", err).End()
	}
	return dir
})

// RecentROM holds the data for a single recently played ROM.
type RecentROM struct {
	Name      string    // Base name of the ROM file.
	Path      string    // Full path to the ROM file.
	Image     []byte    // PNG data for the screenshot.
	SaveState []byte    // Saved state data.
	LastUsed  time.Time // Last time the ROM was played.
}

// AddRecentROM persists a recent ROM to the filesystem.
func AddRecentROM(rom RecentROM) error {
	if err := rom.save(); err != nil {
		return fmt.Errorf("failed to save recent ROM: %w", err)
	}
	return nil
}

func LoadRecentROMs() []RecentROM {
	var roms []RecentROM

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
			return fmt.Errorf("failed to stat recent ROM file %s: %w", path, err)
		}

		zr, err := zip.NewReader(f, info.Size())
		if err != nil {
			return fmt.Errorf("failed to read recent ROM file %s: %w", path, err)
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

		cur := RecentROM{
			Name:      removeExt(info.Name()),
			LastUsed:  info.ModTime(),
			Path:      string(bytes.TrimSpace(loaded["infos.txt"])),
			Image:     loaded["screenshot.png"],
			SaveState: loaded["state.bin"],
		}

		if cur.isValid() {
			roms = append(roms, cur)
		}

		return nil
	})
	if err != nil {
		modCfg.WarnZ("error loading recent roms").Error("err", err).End()
	}

	// Normalize: remove duplicates and sort by LastUsed date.
	m := make(map[string]RecentROM, len(roms))
	for _, rom := range roms {
		m[rom.Name] = rom
	}
	roms = roms[:0]
	for _, rom := range m {
		roms = append(roms, rom)
	}
	slices.SortFunc(roms, func(a, b RecentROM) int {
		return cmp.Compare(b.LastUsed.Unix(), a.LastUsed.Unix())
	})

	const maxRecentsRoms = 32 // Limit the number of recent roms to show
	if len(roms) > maxRecentsRoms {
		roms = roms[:maxRecentsRoms]
	}

	return roms
}

func (r RecentROM) isValid() bool {
	return r.Path != "" && r.Image != nil && r.Name != "" && !r.LastUsed.IsZero()
}

// save writes the recent ROM data to a .nrr file.
func (r RecentROM) save() error {
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
