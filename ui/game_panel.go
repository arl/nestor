package ui

import (
	_ "embed"
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

//go:embed game_panel.glade
var gamePanelUI string

func showGamePanel() (*gamePanel, error) {
	builder, err := gtk.BuilderNewFromString(gamePanelUI)
	if err != nil {
		return nil, fmt.Errorf("builder: can't load UI file: %s", err)
	}

	gp := &gamePanel{
		Window: build[gtk.Window](builder, "game_panel_window"),
	}

	// We know the emulator window starts at the center of the screen and has a
	// scale factor of 2. We want to show the panel on top of it (best effort)
	panelw, panelh := gp.GetSize()
	screenw, screenh := getWorkArea()
	const emuh = 240 * 2
	const winmenuh = 28
	gp.Move((screenw-panelw)/2, (screenh-2*panelh-emuh)/2-winmenuh)
	gp.SetVisible(true)

	return gp, nil
}

type gamePanel struct {
	*gtk.Window
}
