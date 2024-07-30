package ui

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"strconv"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func builderObj[T glib.IObject, P *T](builder *gtk.Builder, name string) (*T, error) {
	gobj, err := builder.GetObject(name)
	if err != nil {
		return nil, fmt.Errorf("builder: can't get object %q: %s", name, err)
	}
	obj, ok := gobj.(P)
	if !ok {
		var zero T
		return nil, fmt.Errorf("builder: object is not a %T but a %T", zero, gobj)
	}
	return obj, nil
}

// ShowMainWindow creates and shows the main window, blocking until it's closed.
func ShowMainWindow() error {
	win, err := newMainWindow()
	if err != nil {
		return err
	}
	_ = win

	gtk.Main()
	return nil
}

type mainWindow struct {
	window         *gtk.ApplicationWindow
	recentRomsView *recentROMsView
}

//go:embed nestor.glade
var gladeUI string

func newMainWindow() (*mainWindow, error) {
	gtk.Init(nil)
	builder, err := gtk.BuilderNewFromString(gladeUI)
	if err != nil {
		return nil, fmt.Errorf("builder: can't load UI file: %s", err)
	}

	window, err := builderObj[gtk.ApplicationWindow](builder, "main_appwindow")
	if err != nil {
		return nil, err
	}
	window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	recentRomsView, err := newRecentRomsView(builder)
	if err != nil {
		return nil, err
	}

	return &mainWindow{
		window:         window,
		recentRomsView: recentRomsView,
	}, nil
}

const maxRecentsRoms = 16

type recentROMsView struct {
	flowbox    *gtk.FlowBox
	recentROMs []recentROM
}

var dummyCounter = 0

func newRecentRomsView(builder *gtk.Builder) (*recentROMsView, error) {
	v := &recentROMsView{
		recentROMs: loadRecentRoms(),
	}

	flowbox, err := builderObj[gtk.FlowBox](builder, "main_flowbox")
	if err != nil {
		return nil, err
	}
	v.flowbox = flowbox

	addbtn, err := builderObj[gtk.Button](builder, "add_btn")
	if err != nil {
		return nil, err
	}
	addbtn.Connect("clicked", func() {
		dummyCounter++

		rom := recentROM{
			Name:  strconv.Itoa(dummyCounter),
			Path:  "/some/path",
			Image: logoWithNumber(dummyCounter),
		}
		if err := v.addRom(rom); err != nil {
			log.Println("failed to add rom to view:", err)
			return
		}
		if err := rom.save(); err != nil {
			log.Println("failed to save recent rom:", err)
		}
	})

	v.recentROMs = loadRecentRoms()
	for _, rom := range v.recentROMs {
		if err := v.addRom(rom); err != nil {
			log.Println("failed to add rom to view:", err)
			return nil, err
		}
		dummyCounter++
	}
	return v, nil
}

func logoWithNumber(n int) []byte {
	var logo []byte
	rgba64, err := pngToRGBA(logo)
	if err != nil {
		panic(err)
	}
	addLabel(rgba64, 10, 10, strconv.Itoa(n))

	var bb bytes.Buffer
	if err := png.Encode(&bb, rgba64); err != nil {
		panic(err)
	}
	return bb.Bytes()
}

func pngToRGBA(buf []byte) (*image.RGBA, error) {
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %s", err)
	}

	// Convert image to RGBA
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	return rgba, nil
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

// addRom adds a new ROM to the list of recent roms, in the first position.
func (v *recentROMsView) addRom(rom recentROM) error {
	loader, err := gdk.PixbufLoaderNewWithType("png")
	if err != nil {
		return fmt.Errorf("failed to create pixbuf loader: %s", err)
	}

	if _, err := loader.Write([]byte(rom.Image)); err != nil {
		return fmt.Errorf("failed to write image data: %s", err)
	}

	if err := loader.Close(); err != nil {
		return fmt.Errorf("failed to close pixbuf loader: %s", err)
	}

	buf, err := loader.GetPixbuf()
	if err != nil {
		return fmt.Errorf("failed to get pixbuf from loader: %s", err)
	}

	buf, err = buf.ScaleSimple(256, 256, gdk.INTERP_BILINEAR)
	if err != nil {
		return fmt.Errorf("failed to get pixbuf from loader: %s", err)
	}

	img, err := gtk.ImageNewFromPixbuf(buf)
	if err != nil {
		return fmt.Errorf("failed to create image from pixbuf: %s", err)
	}

	// Create a button to contain the image
	button, err := gtk.ButtonNew()
	if err != nil {
		return fmt.Errorf("failed to create button: %s", err)
	}

	// Set the image as the button content
	button.SetImage(img)
	button.SetAlwaysShowImage(true)

	img.SetVisible(true)

	// Create a box to contain the button and the label
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return fmt.Errorf("failed to create box: %s", err)
	}

	// Create the label
	label, err := gtk.LabelNew("some name")
	if err != nil {
		return fmt.Errorf("failed to create label: %s", err)
	}

	// Pack the button and the label into the box
	box.PackStart(button, false, false, 0)
	box.PackStart(label, false, false, 0)

	// Connect the click event
	button.Connect("clicked", func() {
		fmt.Println("Image clicked!")
	})

	// Insert the box into the flowbox
	v.flowbox.Insert(box, 0)

	button.SetVisible(true)
	img.SetVisible(true)
	box.SetVisible(true)
	label.SetVisible(true)

	return nil
}
