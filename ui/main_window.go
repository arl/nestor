package ui

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"path/filepath"
	"sync"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu"
	"nestor/emu/log"
	"nestor/ines"
)

var modGUI = log.NewModule("gui")

//go:embed main_window.glade
var mainWindowUI string

//go:embed nestor.css
var stylesUI string

// ShowMainWindow creates and shows the main window, blocking until it's closed.
func ShowMainWindow() error {
	win, err := newMainWindow()
	if err != nil {
		return err
	}
	_ = win

	css := mustT(gtk.CssProviderNew())
	must(css.LoadFromData(stylesUI))
	screen := mustT(gdk.ScreenGetDefault())
	gtk.AddProviderForScreen(screen, css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	gtk.Main()
	return nil
}

type mainWindow struct {
	*gtk.Window
	rrv *recentROMsView
	wg  sync.WaitGroup
	cfg emu.Config

	stopEmu func()
}

func newMainWindow() (*mainWindow, error) {
	gtk.Init(nil)
	builder, err := gtk.BuilderNewFromString(mainWindowUI)
	if err != nil {
		return nil, fmt.Errorf("builder: can't load UI file: %s", err)
	}

	mw := &mainWindow{
		Window:  build[gtk.Window](builder, "main_window"),
		cfg:     emu.LoadConfigOrDefault(),
		stopEmu: func() {},
	}

	mw.Connect("destroy", func() bool { mw.Close(nil); return true })
	mw.rrv = newRecentRomsView(builder, mw.runROM)

	build[gtk.MenuItem](builder, "menu_quit").Connect("activate", gtk.MainQuit)
	build[gtk.MenuItem](builder, "menu_open").Connect("activate", func(m *gtk.MenuItem) {
		path, ok := openFileDialog(mw.Window)
		if !ok {
			return
		}
		mw.runROM(path)
	})
	build[gtk.MenuItem](builder, "menu_controls").Connect("activate", func(m *gtk.MenuItem) {
		showControllerConfig(&mw.cfg.Input)
		if err := emu.SaveConfig(mw.cfg); err != nil {
			modGUI.Warnf("failed to save config: %s", err)
		}
	})

	return mw, nil
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

	rom, err := ines.ReadRom(path)
	if err != nil {
		modGUI.Warnf("failed to read ROM: %s", err)
		return
	}

	emulator, err := emu.Launch(rom, mw.cfg, mw.monitorIndex())
	if err != nil {
		modGUI.Fatalf("failed to start emulator window: %v", err)
		gtk.MainQuit()
	}
	mw.stopEmu = emulator.Stop

	panel := showGamePanel(mw.Window)
	panel.connect(emulator)

	mw.wg.Add(1)
	go func() {
		defer mw.SetSensitive(true)
		defer mw.wg.Done()

		emulator.Run()
		mw.stopEmu = func() {}
		panel.Close()

		screenshot := emulator.Screenshot()
		glib.IdleAdd(func() {
			if err := mw.addRecentROM(path, screenshot); err != nil {
				modGUI.Warnf("failed to add recent ROM: %s", err)
			}
		})
	}()
}

func (mw *mainWindow) monitorIndex() int32 {
	gdkw := mustT(mw.GetWindow())
	display := mustT(gdk.DisplayGetDefault())
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
