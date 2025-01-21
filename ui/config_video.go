package ui

import (
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu"
)

type videoConfigPage struct {
	parent gtk.IWidget
	cfg    *emu.VideoConfig
}

func buildVideoConfigPage(parent gtk.IWidget, cfg *emu.VideoConfig, builder *gtk.Builder) *videoConfigPage {
	page := &videoConfigPage{
		parent: parent,
		cfg:    cfg,
	}

	shaders := build[gtk.ComboBoxText](builder, "shaders_combo")
	shaders.SetActiveID(cfg.Shader)
	shaders.Connect("changed", func(combo *gtk.ComboBoxText) {
		cfg.Shader = combo.GetActiveID()
	})

	vsync := build[gtk.Switch](builder, "vsync_switch")
	vsync.SetActive(!cfg.DisableVSync)
	vsync.Connect("state-set", func(_ *gtk.Switch, state bool) {
		cfg.DisableVSync = !state
	})

	return page
}
