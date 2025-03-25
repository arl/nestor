package ui

import (
	_ "embed"
	"sync"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu/rpc"
)

//go:embed game_panel.glade
var gamePanelUI string

type gamePanel struct {
	win *gtk.Window

	pause   *gtk.ToggleButton
	stop    *gtk.Button
	reset   *gtk.Button
	restart *gtk.Button

	emuStopped bool
	emuStop    func()
}

func showGamePanel(parent *gtk.Window) *gamePanel {
	builder := mustT(gtk.BuilderNewFromString(gamePanelUI))
	gp := &gamePanel{
		win:     build[gtk.Window](builder, "game_panel_window"),
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

	panelw, panelh := gp.win.GetSize()

	// panel coordinate if it was centered on the screen
	centerx := monx + (monw-panelw)/2
	centery := mony + (monh-panelh)/2

	// move the panel to the top of the emulator window
	centery -= emuh + windecoh + panelh/2
	gp.win.Move(centerx, centery)
	gp.win.SetVisible(true)
	gp.win.SetSensitive(false)
	gp.win.ShowAll()
}

func (gp *gamePanel) connect(proxy *rpc.Client) {
	gp.emuStop = sync.OnceFunc(proxy.Stop)

	gp.win.Connect("destroy", func() { gp.emuStop() })
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
		gp.emuStop()
		gp.Close()
	})

	gp.win.SetSensitive(true)
}

func (gp *gamePanel) setGameStopped() {
	gp.emuStop = func() {}
}

func (gp *gamePanel) Close() {
	gp.win.Close()
	gp.emuStop()
}
