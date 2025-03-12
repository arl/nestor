package ui

import (
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu"
	"nestor/hw/shaders"
)

type audioConfigPage struct {
	parent gtk.IWidget
	cfg    *emu.AudioConfig
}

func buildAudioConfigPage(parent gtk.IWidget, cfg *emu.AudioConfig, builder *gtk.Builder) *audioConfigPage {
	page := &audioConfigPage{
		parent: parent,
		cfg:    cfg,
	}

	shaderList := build[gtk.ComboBoxText](builder, "shaders_combo")
	for _, name := range shaders.Names() {
		shaderList.Append(name, name)
	}

	enabled := build[gtk.Switch](builder, "audio_enabled_switch")
	enabled.SetActive(!cfg.DisableAudio)
	enabled.Connect("state-set", func(_ *gtk.Switch, state bool) {
		cfg.DisableAudio = !state
	})

	return page
}
