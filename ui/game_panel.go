package ui

import (
	_ "embed"
	"image"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu/rpc"
)

//go:embed game_panel.glade
var gamePanelUI string

type gamePanel struct {
	*gtk.Window

	pause   *gtk.ToggleButton
	stop    *gtk.Button
	reset   *gtk.Button
	restart *gtk.Button

	img *image.RGBA
}

func showGamePanel(parent *gtk.Window) *gamePanel {
	builder := mustT(gtk.BuilderNewFromString(gamePanelUI))
	gp := &gamePanel{
		Window:  build[gtk.Window](builder, "game_panel_window"),
		pause:   build[gtk.ToggleButton](builder, "pause_button"),
		stop:    build[gtk.Button](builder, "stop_button"),
		reset:   build[gtk.Button](builder, "reset_button"),
		restart: build[gtk.Button](builder, "restart_button"),
	}
	gp.moveAndShow(parent)
	return gp
}

func (gp *gamePanel) moveAndShow(parent *gtk.Window) {
	gdkw := mustT(parent.GetWindow())
	display := mustT(gdk.DisplayGetDefault())

	monitor := mustT(display.GetMonitorAtWindow(gdkw))
	geom := monitor.GetGeometry()
	monx, mony, monw, monh := geom.GetX(), geom.GetY(), geom.GetWidth(), geom.GetHeight()

	// We know the emulator window starts at the center of the screen and has a
	// scale factor of 2. We want to show the panel on top of it (best effort).
	const emuh = 240
	const windecoh = 32 // window decoration bar height

	panelw, panelh := gp.GetSize()

	// panel coordinate if it was centered on the screen
	centerx := monx + (monw-panelw)/2
	centery := mony + (monh-panelh)/2

	// move the panel to the top of the emulator window
	centery -= emuh + windecoh + panelh/2
	gp.Move(centerx, centery)
	gp.SetVisible(true)
}

func (gp *gamePanel) connect(proxy *rpc.Client) {
	gp.Connect("destroy", proxy.Stop)
	gp.reset.Connect("clicked", proxy.Reset)
	gp.restart.Connect("clicked", proxy.Restart)
	gp.pause.Connect("toggled", func(btn *gtk.ToggleButton) {
		paused := btn.GetActive()
		proxy.SetPause(paused)
		if paused {
			btn.SetLabel("Resume")
		} else {
			btn.SetLabel("Pause")
		}
		gp.reset.SetSensitive(!paused)
		gp.restart.SetSensitive(!paused)
	})
	gp.stop.Connect("clicked", func() {
		gp.img = proxy.Stop()
		gp.Close()
	})
}
