package ui

import (
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu"
)

type emulatioConfigPage struct {
	parent *gtk.Dialog
	cfg    *emu.EmulationConfig
}

func buildEmulationConfigPage(parent *gtk.Dialog, cfg *emu.EmulationConfig, builder *gtk.Builder) *emulatioConfigPage {
	page := &emulatioConfigPage{
		parent: parent,
		cfg:    cfg,
	}

	adjustment := build[gtk.Adjustment](builder, "run_ahead_adjustment")
	adjustment.SetValue(float64(cfg.RunAheadFrames))

	runAheadFrames := build[gtk.SpinButton](builder, "run_ahead_frames_spin")
	runAheadFrames.Connect("value-changed", func(_ *gtk.SpinButton) {
		cfg.RunAheadFrames = runAheadFrames.GetValueAsInt()
		modGUI.InfoZ("Setting run ahead frames to").Int("value", cfg.RunAheadFrames).End()
	})
	return page
}
