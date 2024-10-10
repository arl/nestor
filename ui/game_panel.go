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
		Window: build[gtk.Window](builder, "window1"),
	}

	// Show the panel on top of the emulator window, which always starts at the
	// center of the screen.
	// pos := gtk.Pos
	// mw.SetPosition(gtk.WIN_POS_CENTER)

	gp.SetVisible(true)
	return gp, nil
}

type gamePanel struct {
	*gtk.Window
}
