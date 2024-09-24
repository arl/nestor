package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"path/filepath"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu"
	"nestor/emu/log"
	"nestor/ines"
)

var modGUI = log.NewModule("gui")

// ShowMainWindow creates and shows the main window, blocking until it's closed.
func ShowMainWindow() error {
	win, err := newMainWindow()
	if err != nil {
		return err
	}
	_ = win

	gtk.Main()
	return nil
}

type mainWindow struct {
	*gtk.Window
	rrv *recentROMsView
	cfg emu.Config
}

func newMainWindow() (*mainWindow, error) {
	gtk.Init(nil)
	builder, err := gtk.BuilderNewFromString(gladeUI)
	if err != nil {
		return nil, fmt.Errorf("builder: can't load UI file: %s", err)
	}

	mw := &mainWindow{
		Window: build[gtk.Window](builder, "window1"),
		cfg:    LoadConfigOrDefault(),
	}

	mw.Connect("destroy", func() { mw.Close(nil) })
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
		openInputConfigDialog(&mw.cfg.Input)
		if err := SaveConfig(mw.cfg); err != nil {
			modGUI.Warnf("failed to save config: %s", err)
		}
	})

	return mw, nil
}

func (mw *mainWindow) Close(err error) {
	if err != nil {
		modGUI.Warnf("closing UI with error: %s", err)
	}
	gtk.MainQuit()
}

func (mw *mainWindow) runROM(path string) {
	mw.SetSensitive(false)

	rom, err := ines.ReadRom(path)
	if err != nil {
		modGUI.Warnf("failed to read ROM: %s", err)
		return
	}

	errc := make(chan error)
	go func() {
		defer mw.SetSensitive(true)

		emulator, err := emu.PowerUp(rom, mw.cfg)
		errc <- err // Release gtk thread asap.

		emulator.Run()

		screenshot := emulator.Screenshot()

		glib.IdleAdd(func() {
			if err := mw.addRecentROM(path, screenshot); err != nil {
				modGUI.Warnf("failed to add recent ROM: %s", err)
			}
		})
	}()

	if err := <-errc; err != nil {
		modGUI.Fatalf("failed to start emulator window: %v", err)
		gtk.MainQuit()
	}
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
