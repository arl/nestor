package ui

import (
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu"
)

type audioConfigPage struct {
	parent *gtk.Dialog
	cfg    *emu.AudioConfig
}

func buildAudioConfigPage(parent *gtk.Dialog, cfg *emu.AudioConfig, builder *gtk.Builder) *audioConfigPage {
	page := &audioConfigPage{
		parent: parent,
		cfg:    cfg,
	}

	enabled := build[gtk.Switch](builder, "audio_enabled_switch")
	enabled.SetActive(!cfg.DisableAudio)
	enabled.Connect("state-set", func(_ *gtk.Switch, state bool) {
		cfg.DisableAudio = !state
	})

	return page
}
