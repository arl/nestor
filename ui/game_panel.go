package ui

import (
	_ "embed"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu"
)

//go:embed game_panel.glade
var gamePanelUI string

type gamePanel struct {
	*gtk.Window
	builder *gtk.Builder
}

func showGamePanel(parent *gtk.Window) *gamePanel {
	builder := mustT(gtk.BuilderNewFromString(gamePanelUI))
	win := build[gtk.Window](builder, "game_panel_window")

	gdkw := mustT(parent.GetWindow())
	display := mustT(gdk.DisplayGetDefault())

	monitor := mustT(display.GetMonitorAtWindow(gdkw))
	geom := monitor.GetGeometry()
	monx, mony, monw, monh := geom.GetX(), geom.GetY(), geom.GetWidth(), geom.GetHeight()

	// We know the emulator window starts at the center of the screen and has a
	// scale factor of 2. We want to show the panel on top of it (best effort).
	const emuh = 240
	const windecoh = 32 // window decoration bar height

	panelw, panelh := win.GetSize()

	// panel coordinate if it was centered on the screen
	centerx := monx + (monw-panelw)/2
	centery := mony + (monh-panelh)/2

	// move the panel to the top of the emulator window
	centery -= emuh + windecoh + panelh/2
	win.SetParent(parent)
	win.Move(centerx, centery)
	win.SetVisible(true)
	return &gamePanel{
		Window:  win,
		builder: builder,
	}
}

func (gp *gamePanel) connect(emulator *emu.Emulator) {
	build[gtk.ToggleButton](gp.builder, "pause_button").Connect("pressed", func(btn *gtk.ToggleButton) {
		switch emulator.Pause() {
		case true:
			btn.SetLabel("Resume")
		case false:
			btn.SetLabel("Pause")
		}
	})
}
