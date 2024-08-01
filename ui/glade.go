package ui

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

//go:embed nestor.glade
var gladeUI string

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

func mustf(err error, format string, args ...any) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		panic(msg + "\n" + err.Error())
	}
}

func must2[T any](v T, err error) T {
	must(err)
	return v
}

// openFileDialog shows a file chooser dialog for selecting a nes ROM file.
func openFileDialog(parent *gtk.Window) (string, bool) {
	dlg, err := gtk.FileChooserDialogNewWith1Button(
		"Open NES ROM",
		parent,
		gtk.FILE_CHOOSER_ACTION_OPEN,
		"Open",
		gtk.RESPONSE_OK,
	)
	must(err)
	defer dlg.Destroy()

	filter, err := gtk.FileFilterNew()
	if err != nil {
		log.Fatal("Unable to create file filter:", err)
	}
	filter.AddPattern("*.nes")
	filter.SetName("nes/famicom ROM Files")
	dlg.AddFilter(filter)

	if resp := dlg.Run(); resp != gtk.RESPONSE_OK {
		return "", false
	}
	return dlg.GetFilename(), true
}
