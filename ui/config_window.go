package ui

import (
	_ "embed"

	"nestor/hw/input"

	"github.com/gotk3/gotk3/gtk"
)

//go:embed config.glade
var configUI string

type configWindow struct {
	*gtk.Window
}

func showConfigWindow(cfg *input.Config) {
	builder := mustT(gtk.BuilderNewFromString(configUI))

	win := build[gtk.Dialog](builder, "config_dialog")

	buildInputConfigPage(win, cfg, builder)
	win.ShowAll()
	win.Run()
	win.Destroy()
}
