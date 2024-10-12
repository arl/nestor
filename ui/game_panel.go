package ui

import (
	_ "embed"

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

	// We know the emulator window starts at the center of the screen and has a
	// scale factor of 2. We want to show the panel on top of it (best effort)
	win := build[gtk.Window](builder, "game_panel_window")
	panelw, panelh := win.GetSize()
	screenw, screenh := getWorkArea()
	const emuh = 240 * 2
	const winmenuh = 28
	win.SetParent(parent)
	win.Move((screenw-panelw)/2, (screenh-2*panelh-emuh)/2-winmenuh)
	win.SetVisible(true)
	return &gamePanel{
		Window:  win,
		builder: builder,
	}

}

}
