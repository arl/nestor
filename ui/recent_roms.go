package ui

import (
	"archive/zip"
	"bytes"
	"cmp"
	"fmt"
	"image"
	_ "image/png" // Import for decoding PNG screenshots
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"

	"nestor/config"
)

// --- Core Logic (Reused from recent_roms.go) ---
// This section contains the UI-agnostic logic for finding, loading, and saving recent ROMs.

const recentROMextension = ".nrr"

var RecentROMsDir = sync.OnceValue(func() string {
	dir := filepath.Join(config.Dir(), "recent-roms")
	// Using 0755 for directory permissions
	if err := os.MkdirAll(dir, 0755); err != nil {
		// In a real app, you'd use a proper logger.
		fmt.Printf("failed to create directory %s: %v\n", dir, err)
	}
	return dir
})

// recentROM holds the data for a single recently played ROM.
type recentROM struct {
	Name     string
	Path     string
	Image    []byte // PNG data for the screenshot
	LastUsed time.Time
}

func (r recentROM) IsValid() bool {
	return r.Path != "" && r.Image != nil && r.Name != "" && !r.LastUsed.IsZero()
}

// save writes the recent ROM data to a .nrr file.
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

// AddRecentROM is the new entry point for saving a ROM to the recent list.
// Call this function when you run a new ROM.
func AddRecentROM(rom recentROM) error {
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
		if err != nil || d.IsDir() || filepath.Ext(path) != recentROMextension {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return err
		}

		zr, err := zip.NewReader(f, info.Size())
		if err != nil {
			return err // Corrupted zip file
		}

		cur := recentROM{
			Name:     removeExt(info.Name()),
			LastUsed: info.ModTime(),
		}

		for _, zf := range zr.File {
			switch zf.Name {
			case "screenshot.png":
				buf, err := readZipFile(zf)
				if err != nil {
					fmt.Printf("warning: could not read screenshot from %s: %v\n", path, err)
					continue
				}
				cur.Image = buf
			case "infos.txt":
				buf, err := readZipFile(zf)
				if err != nil {
					fmt.Printf("warning: could not read info from %s: %v\n", path, err)
					continue
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
		fmt.Printf("error loading recent roms: %s\n", err)
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

type recentRomsWidget struct {
	run       func(string)
	container *widget.Container
}

func newRecentROMsWidget(width, height int, runROM func(path string)) *recentRomsWidget {
	rr := &recentRomsWidget{
		run: runROM,
	}

	rr.recalc(width, height)
	return rr
}

func (rr *recentRomsWidget) recalc(width, height int) {
	const imgside = 250
	const spacing = 10
	numcols := width / (imgside + spacing)
	maxrows := height / (imgside + spacing) // max rows to display
	if maxrows == 0 {
		maxrows = 1
	}

	colstretch := make([]bool, numcols)
	for i := range colstretch {
		colstretch[i] = true
	}

	bc := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(numcols),
			widget.GridLayoutOpts.Stretch(colstretch, nil),
			widget.GridLayoutOpts.Spacing(10, 10))))

	roms := loadRecentROMs()

	for i := range roms {
		rowidx := i / numcols
		if rowidx == maxrows {
			break
		}

		img, _, err := image.Decode(bytes.NewReader(roms[i].Image))
		if err != nil {
			panic(err)
		}

		eimg := ebiten.NewImageFromImage(img)
		eimg = resizeImage(eimg, imgside, imgside)

		btnimg := &widget.ButtonImage{
			Idle:    uiimage.NewFixedNineSlice(eimg),
			Pressed: uiimage.NewNineSliceSimple(eimg, 2, 2),
		}

		b := widget.NewButton(
			widget.ButtonOpts.Image(btnimg),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				rr.run(roms[i].Path)
			}),
			// widget.ButtonOpts.Text(roms[i].Name, &font, &widget.ButtonTextColor{
			// 	Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
			// }),
			// widget.ButtonOpts.TextPadding(&widget.Insets{
			// 	Top: 90,
			// }),
			// widget.ButtonOpts.TextPosition(widget.TextPositionCenter, widget.TextPositionEnd),
		)
		bc.AddChild(b)
	}

	rr.container = bc
}
