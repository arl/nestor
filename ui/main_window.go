package ui

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu/log"
)

var modGUI = log.NewModule("gui")

//go:embed main_window.glade
var mainWindowUI string

//go:embed nestor.css
var stylesUI string

// RunApp creates and shows the main window,
// blocking until it's closed.
func RunApp(cfg *Config) {
	gtk.Init(nil)
	css := mustT(gtk.CssProviderNew())
	must(css.LoadFromData(stylesUI))
	screen := mustT(gdk.ScreenGetDefault())
	gtk.AddProviderForScreen(screen, css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	showMainWindow(cfg)
	if cfg.General.ShowSplash {
		splashScreen(360, 360)
	}
	gtk.Main()
	modGUI.InfoZ("Exited gtk").End()
}

type mainWindow struct {
	*gtk.Window
	rrv *recentROMsView
	wg  sync.WaitGroup
	cfg *Config

	stopEmu func()
}

func showMainWindow(cfg *Config) {
	builder := mustT(gtk.BuilderNewFromString(mainWindowUI))

	mw := &mainWindow{
		Window:  build[gtk.Window](builder, "main_window"),
		cfg:     cfg,
		stopEmu: func() {},
	}

	mw.Connect("destroy", func() bool { mw.Close(nil); return true })
	mw.rrv = newRecentRomsView(builder, mw.runROM)

	build[gtk.MenuItem](builder, "menu_quit").Connect("activate", gtk.MainQuit)
	build[gtk.MenuItem](builder, "menu_open").Connect("activate", func(m *gtk.MenuItem) {
		workdir, ok := mw.rrv.mostRecentDir()
		if !ok {
			workdir = ""
		}
		path, ok := openFileDialog(mw.Window, workdir)
		if !ok {
			return
		}
		mw.runROM(path)
	})

	onConfig := func(m *gtk.MenuItem) {
		menu := m.GetLabel()
		showConfig(mw.cfg, menu)
		if err := SaveConfig(mw.cfg); err != nil {
			modGUI.Warnf("failed to save config: %s", err)
		}
	}
	build[gtk.MenuItem](builder, "menu_input").Connect("activate", onConfig)
	build[gtk.MenuItem](builder, "menu_video").Connect("activate", onConfig)
}

func (mw *mainWindow) Close(err error) {
	if err != nil {
		modGUI.Warnf("closing UI with error: %s", err)
	}

	if mw.stopEmu != nil {
		mw.stopEmu()
	}

	mw.wg.Wait()
	gtk.MainQuit()
}

func (mw *mainWindow) runROM(path string) {
	mw.SetSensitive(false)
	defer mw.SetSensitive(true)

	// TODO: add -monitor flag to nestor and pass it monitorIdx()
	executable, err := os.Executable()
	if err != nil {
		modGUI.Warnf("failed to get executable path: %s", err)
		return
	}

	fmt.Println("about to run", executable, "run", path)
	cmd := exec.Command(executable, "run", path)
	if err := cmd.Run(); err != nil {
		modGUI.Warnf("failed to run ROM: %s", err)
		return
	}

	// TODO: handle error
	// TODO: connect game panel with bitbucket.org/avd/go-ipc@v0.6.1
}

func (mw *mainWindow) monitorIdx() int32 {
	display := mustT(gdk.DisplayGetDefault())
	gdkw := mustT(mw.GetWindow())
	monitor := mustT(display.GetMonitorAtWindow(gdkw))
	if monitor.IsPrimary() {
		return 0
	}
	return 1
}

func (mw *mainWindow) addRecentROM(romPath string, screenshot image.Image) error {
	bb := bytes.Buffer{}
	if err := png.Encode(&bb, screenshot); err != nil {
		return fmt.Errorf("failed to encode screenshot: %v", err)
	}

	return mw.rrv.addROM(recentROM{
		Name:     filepath.Base(romPath),
		Image:    bb.Bytes(),
		Path:     romPath,
		LastUsed: time.Now(),
	})
}
