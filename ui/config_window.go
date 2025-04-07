package ui

import (
	_ "embed"

	"github.com/gotk3/gotk3/gtk"
)

//go:embed config.glade
var configUI string

type configWindow struct {
	*gtk.Window
}

// show configuration dialog, with the given page selected.
func showConfig(cfg *Config, page string) {
	builder := mustT(gtk.BuilderNewFromString(configUI))
	win := build[gtk.Dialog](builder, "config_dialog")
	stack := build[gtk.Stack](builder, "stack")

	buildInputConfigPage(win, &cfg.Input, builder)
	buildVideoConfigPage(win, &cfg.Video, builder)
	buildAudioConfigPage(win, &cfg.Audio, builder)
	buildEmulationConfigPage(win, &cfg.Emulation, builder)

	stack.SetVisibleChildName(page)
	win.ShowAll()
	win.Run()
	win.Destroy()
}
