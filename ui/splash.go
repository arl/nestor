package ui

import (
	_ "embed"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

//go:embed logo.png
var logoPNG []byte

// shows splash screen for 2 seconds, or until the user clicks on it.
func splashScreen(w, h int) {
	win := mustT(gtk.WindowNew(gtk.WINDOW_TOPLEVEL))
	win.SetSizeRequest(w, h)
	win.SetModal(true)
	win.SetDecorated(false)
	win.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)
	win.SetResizable(false)
	win.SetTypeHint(gdk.WINDOW_TYPE_HINT_SPLASHSCREEN)

	pbuf := mustT(pixbufFromBytes(logoPNG))
	pbuf = mustT(pbuf.ScaleSimple(w, h, gdk.INTERP_NEAREST))
	img := mustT(gtk.ImageNewFromPixbuf(pbuf))
	win.Add(img)
	win.ShowAll()

	const splashTimeout = time.Second * 2

	destroy := func() { glib.IdleAdd(win.Destroy) }
	time.AfterFunc(splashTimeout, destroy)
	win.Connect("button-press-event", destroy)
}
