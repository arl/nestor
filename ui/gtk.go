package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func build[T glib.IObject, P *T](builder *gtk.Builder, name string) *T {
	gobj, err := builder.GetObject(name)
	if err != nil {
		panic(fmt.Sprintf("builder: can't get object %q: %s", name, err))
	}
	obj, ok := gobj.(P)
	if !ok {
		var zero T
		panic(fmt.Sprintf("builder: object is not a %T but a %T", zero, gobj))
	}
	return obj
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

//lint:ignore U1000 useful later
func mustf(err error, format string, args ...any) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		panic(msg + "\n" + err.Error())
	}
}

func mustT[T any](v T, err error) T {
	must(err)
	return v
}

// openFileDialog shows a file chooser dialog for selecting a nes ROM file.
func openFileDialog(parent *gtk.Window, workdir string) (string, bool) {
	dlg := mustT(gtk.FileChooserDialogNewWith1Button(
		"Open NES ROM",
		parent,
		gtk.FILE_CHOOSER_ACTION_OPEN,
		"Open",
		gtk.RESPONSE_OK,
	))
	defer dlg.Close()

	filter := mustT(gtk.FileFilterNew())
	filter.AddPattern("*.nes")
	filter.SetName("nes/famicom ROM Files")
	dlg.AddFilter(filter)
	dlg.SetCurrentFolder(workdir)
	if resp := dlg.Run(); resp != gtk.RESPONSE_OK {
		return "", false
	}
	return dlg.GetFilename(), true
}

func pixbufFromBytes(data []byte) (*gdk.Pixbuf, error) {
	loader := mustT(gdk.PixbufLoaderNew())
	if _, err := loader.Write(data); err != nil {
		return nil, err
	}
	// Finalize loading before getting the buffer.
	loader.Close()
	return loader.GetPixbuf()
}

// returns a gtk.Image from the bytes of an image file.
func imageFromBytes(data []byte) (*gtk.Image, error) {
	buf, err := pixbufFromBytes(data)
	if err != nil {
		return nil, err
	}
	return gtk.ImageNewFromPixbuf(buf)
}

// reports whether a child is currently visible in a scrolled window.
func isVisibleIn(child *gtk.Widget, scrolled *gtk.Widget) bool {
	dstx, dsty, err := child.TranslateCoordinates(scrolled, 0, 0)
	if err != nil {
		panic("unexpected")
	}
	childw := child.GetAllocation().GetWidth()
	childh := child.GetAllocation().GetHeight()
	scrolledw := scrolled.GetAllocation().GetWidth()
	scrolledh := scrolled.GetAllocation().GetHeight()

	return (dstx >= 0 && dsty >= 0) &&
		dstx+childw <= scrolledw &&
		dsty+childh <= scrolledh
}
