package ui

import (
	"archive/zip"
	"bytes"
	"cmp"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/gotk3/gotk3/gtk"
)

const recentROMextension = ".nrr"

var RecentROMsDir = sync.OnceValue(func() string {
	dir := filepath.Join(ConfigDir(), "recent-roms")
	if err := os.MkdirAll(dir, DefaultFileMode); err != nil {
		modGUI.Fatalf("failed to create directory %s: %v", dir, err)
	}

	return dir
})

func loadRecentROMs() []recentROM {
	var roms []recentROM

	err := filepath.WalkDir(RecentROMsDir(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
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
		modGUI.Warnf("error loading recent roms: %s", err)
	}

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

func (r recentROM) save() error {
	f, err := os.Create(filepath.Join(RecentROMsDir(), r.Name+recentROMextension))
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

// TODO: limit the number of recent roms we can view?
// but not the number of saved roms.
//
//lint:ignore U1000 todo
const maxRecentsRoms = 16

type recentROMsView struct {
	flowbox    *gtk.FlowBox
	recentROMs []recentROM
	runROM     func(string)
}

func newRecentRomsView(builder *gtk.Builder, runROM func(path string)) *recentROMsView {
	v := &recentROMsView{
		runROM:     runROM,
		recentROMs: loadRecentROMs(),
		flowbox:    build[gtk.FlowBox](builder, "flowbox1"),
	}

	v.updateView()
	return v
}

// addROM adds a new ROM to the list of recent roms.
func (v *recentROMsView) addROM(rom recentROM) error {
	v.recentROMs = append(v.recentROMs, rom)
	if err := rom.save(); err != nil {
		return fmt.Errorf("ROM save: %v", err)
	}
	v.updateView()
	return nil
}

// normalize sorts the list by last usage and remove duplicates.
func (v *recentROMsView) normalize() {
	m := make(map[string]recentROM, len(v.recentROMs))
	for _, rom := range v.recentROMs {
		m[rom.Name] = rom
	}

	v.recentROMs = v.recentROMs[:0]
	for _, rom := range m {
		v.recentROMs = append(v.recentROMs, rom)
	}

	slices.SortFunc(v.recentROMs, func(a, b recentROM) int {
		return cmp.Compare(b.LastUsed.Unix(), a.LastUsed.Unix())
	})
}

func (v *recentROMsView) mostRecentDir() (string, bool) {
	if len(v.recentROMs) == 0 {
		return "", false
	}

	return filepath.Dir(v.recentROMs[0].Path), true
}

func (v *recentROMsView) updateView() {
	v.normalize()

	// Empty the flowbox.
	v.flowbox.GetChildren().Foreach(func(item any) {
		item.(*gtk.Widget).Destroy()
	})

	addItem := func(rom recentROM) error {
		img, err := imageFromBytes(rom.Image)
		if err != nil {
			return fmt.Errorf("failed to load image: %v", err)
		}

		// Create a button to contain the image
		button := mustT(gtk.ButtonNew())
		button.SetImage(img)
		button.SetAlwaysShowImage(true)

		box := mustT(gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0))
		label := mustT(gtk.LabelNew(rom.Name))
		box.PackStart(button, false, false, 0)
		box.PackStart(label, false, false, 0)

		v.flowbox.Insert(box, int(v.flowbox.GetChildren().Length()))

		button.SetVisible(true)
		img.SetVisible(true)
		box.SetVisible(true)
		label.SetVisible(true)
		img.SetVisible(true)

		button.Connect("clicked", func() { v.runROM(rom.Path) })
		return nil
	}

	for _, rom := range v.recentROMs {
		if err := addItem(rom); err != nil {
			modGUI.Warnf("failed to show recent ROM %q: %s", rom.Name, err)
		}
	}
}
