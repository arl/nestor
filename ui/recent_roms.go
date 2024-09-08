package ui

import (
	"archive/zip"
	"bytes"
	"cmp"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/kirsle/configdir"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"nestor/emu/log"
)

const recentROMextension = ".nrr"

var ConfigDir string = sync.OnceValue(func() string {
	dir := configdir.LocalConfig("nestor")
	if err := configdir.MakePath(dir); err != nil {
		modGUI.Fatalf("failed to create directory %s: %v", dir, err)
	}
	return dir
})()

var RecentROMsDir string = sync.OnceValue(func() string {
	dir := filepath.Join(ConfigDir, "recent-roms")
	if err := configdir.MakePath(dir); err != nil {
		modGUI.Fatalf("failed to create directory %s: %v", dir, err)
	}
	return dir
})()

func loadRecentROMs() []recentROM {
	var roms []recentROM

	err := filepath.WalkDir(RecentROMsDir, func(path string, d fs.DirEntry, err error) error {
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

const maxRecentsRoms = 16

type recentROMsView struct {
	flowbox    *gtk.FlowBox
	recentROMs []recentROM
	runROM     func(string)
}

func newRecentRomsView(builder *gtk.Builder, runROM func(path string)) (*recentROMsView, error) {
	v := &recentROMsView{
		runROM:     runROM,
		recentROMs: loadRecentROMs(),
		flowbox:    build[gtk.FlowBox](builder, "flowbox1"),
	}

	v.updateView()
	return v, nil
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{255, 0, 0, 255}
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
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

// remove duplicates and sort the list by last usage.
func (v *recentROMsView) fixOrderAndDups() {
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

func (v *recentROMsView) updateView() {
	v.fixOrderAndDups()

	// Empty the flowbox.
	v.flowbox.GetChildren().Foreach(func(item any) {
		item.(*gtk.Widget).Destroy()
	})

	addItem := func(rom recentROM) error {
		loader := mustT(gdk.PixbufLoaderNewWithType("png"))
		defer loader.Close()

		bufimg := make([]byte, len(rom.Image))
		copy(bufimg, rom.Image)
		mustT(loader.Write(bufimg))

		buf := mustT(loader.GetPixbuf())
		buf = mustT(buf.ScaleSimple(256, 256, gdk.INTERP_BILINEAR))

		// Create a button to contain the image
		img := mustT(gtk.ImageNewFromPixbuf(buf))
		button := mustT(gtk.ButtonNew())

		// Set the image as the button content
		button.SetImage(img)
		button.SetAlwaysShowImage(true)

		// Create a box to contain the button and the label
		box := mustT(gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0))
		label := mustT(gtk.LabelNew(rom.Name))
		box.PackStart(button, false, false, 0)
		box.PackStart(label, false, false, 0)

		button.Connect("clicked", func() {
			v.runROM(rom.Path)
		})

		v.flowbox.Insert(box, int(v.flowbox.GetChildren().Length()))

		button.SetVisible(true)
		img.SetVisible(true)
		box.SetVisible(true)
		label.SetVisible(true)
		img.SetVisible(true)
		return nil
	}

	for _, rom := range v.recentROMs {
		if err := addItem(rom); err != nil {
			log.ModEmu.Warnf("failed to add recent ROM %q to view: %s", rom.Name, err)
		}
	}
}
