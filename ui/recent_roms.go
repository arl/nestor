package ui

import (
	"archive/zip"
	"bytes"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kirsle/configdir"
)

var ConfigDir string = sync.OnceValue(func() string {
	dir := configdir.LocalConfig("nestor")
	if err := configdir.MakePath(dir); err != nil {
		log.Fatalf("failed to create directory %s: %v", dir, err)
	}
	return dir
})()

var RecentROMsDir string = sync.OnceValue(func() string {
	dir := filepath.Join(ConfigDir, "recent-roms")
	if err := configdir.MakePath(dir); err != nil {
		log.Fatalf("failed to create directory %s: %v", dir, err)
	}
	return dir
})()

func loadRecentRoms() []recentROM {
	var roms []recentROM

	err := filepath.WalkDir(RecentROMsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != RecentROMsDir {
			return filepath.SkipDir
		}
		if filepath.Ext(path) != recentROMextension {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		dirent, err := d.Info()
		if err != nil {
			return err
		}

		zr, err := zip.NewReader(f, dirent.Size())
		if err != nil {
			return err
		}

		cur := recentROM{
			Name:     removeExt(dirent.Name()),
			LastUsed: dirent.ModTime(),
		}

		for _, zf := range zr.File {
			if zf.Name == "screenshot.png" {
				zfr, err := zf.Open()
				if err != nil {
					return err
				}
				defer zfr.Close()

				buf, err := io.ReadAll(zfr)
				if err != nil {
					return err
				}
				cur.Image = buf
			}
			if zf.Name == "infos.txt" {
				zfr, err := zf.Open()
				if err != nil {
					return err
				}
				defer zfr.Close()

				buf, err := io.ReadAll(zfr)
				if err != nil {
					return err
				}
				cur.Path = string(bytes.TrimSpace(buf))
			}
		}

		if cur.IsValid() {
			roms = append(roms, cur)
		}

		return nil
	})

	if err != nil {
		log.Println("error loading recent roms:", err)
	}

	// Return anyway what we could load.
	return roms
}

type recentROM struct {
	Name     string
	Path     string `json:"path"`
	Image    []byte `json:"image"`
	LastUsed time.Time
}

func (r recentROM) IsValid() bool {
	return r.Path != "" &&
		r.Image != nil &&
		r.Name != "" &&
		!r.LastUsed.IsZero()
}

const recentROMextension = ".nrr"

func (r recentROM) save() error {
	f, err := os.Create(filepath.Join(RecentROMsDir, r.Name+recentROMextension))
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

	return nil
}

func removeExt(path string) string {
	return path[:len(path)-len(filepath.Ext(path))]
}
