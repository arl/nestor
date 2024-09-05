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

	slices.SortFunc(roms, func(a, b recentROM) int {
		return cmp.Compare(a.LastUsed.Unix(), b.LastUsed.Unix())
	})
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

	for _, rom := range v.recentROMs {
		v.addROM(rom)
	}
	v.refreshView()
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

// addROM adds a new ROM to the list of recent roms, at the first position.
func (v *recentROMsView) addROM(rom recentROM) {
	// Drop duplicates of the same ROM, and trailing elements if we have too many.
	v.recentROMs = slices.DeleteFunc(v.recentROMs, func(r recentROM) bool {
		return r.Name == rom.Name
	})
	v.recentROMs = append([]recentROM{rom}, v.recentROMs...)
	v.recentROMs = v.recentROMs[:min(len(v.recentROMs), maxRecentsRoms)]
}

func (v *recentROMsView) refreshView() {
	// Remove all children from the flowbox.
	v.flowbox.GetChildren().Foreach(func(item any) {
		item.(*gtk.Widget).Destroy()
	})

	addItem := func(rom recentROM) error {
		loader := mustT(gdk.PixbufLoaderNewWithType("png"))
		defer loader.Close()

		if _, err := loader.Write([]byte(rom.Image)); err != nil {
			return fmt.Errorf("failed to write image data: %s", err)
		}

		buf, err := loader.GetPixbuf()
		if err != nil {
			return fmt.Errorf("failed to get pixbuf from loader: %s", err)
		}

		buf, err = buf.ScaleSimple(256, 256, gdk.INTERP_BILINEAR)
		if err != nil {
			return fmt.Errorf("failed to get pixbuf from loader: %s", err)
		}

		// Create a button to contain the image
		img := mustT(gtk.ImageNewFromPixbuf(buf))
		button := mustT(gtk.ButtonNew())

		// Set the image as the button content
		button.SetImage(img)
		button.SetAlwaysShowImage(true)

		img.SetVisible(true)

		// Create a box to contain the button and the label
		box := mustT(gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0))
		// Create the label
		label := mustT(gtk.LabelNew(rom.Name))

		// Pack the button and the label into the box
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
		return nil
	}

	for _, rom := range v.recentROMs {
		if err := addItem(rom); err != nil {
			log.ModEmu.Warnf("failed to add recent romw to view: %s", err)
		}
	}
}
