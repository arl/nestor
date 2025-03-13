package ui

import (
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu"
	"nestor/hw/shaders"
)

type videoConfigPage struct {
	parent *gtk.Dialog
	cfg    *emu.VideoConfig
}

func buildVideoConfigPage(parent *gtk.Dialog, cfg *emu.VideoConfig, builder *gtk.Builder) *videoConfigPage {
	page := &videoConfigPage{
		parent: parent,
		cfg:    cfg,
	}

	shaderList := build[gtk.ComboBoxText](builder, "shaders_combo")
	for _, name := range shaders.Names() {
		shaderList.Append(name, name)
	}
	shaderList.SetActiveID(cfg.Shader)
	shaderList.Connect("changed", func(combo *gtk.ComboBoxText) {
		cfg.Shader = combo.GetActiveID()
	})

	vsync := build[gtk.Switch](builder, "vsync_switch")
	vsync.SetActive(!cfg.DisableVSync)
	vsync.Connect("state-set", func(_ *gtk.Switch, state bool) {
		cfg.DisableVSync = !state
	})

	return page
}
