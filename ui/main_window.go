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
	"nestor/hw"
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
	w              *gtk.Window
	recentRomsView *recentROMsView
}

func newMainWindow() (*mainWindow, error) {
	gtk.Init(nil)
	builder, err := gtk.BuilderNewFromString(gladeUI)
	if err != nil {
		return nil, fmt.Errorf("builder: can't load UI file: %s", err)
	}

	w := build[gtk.Window](builder, "window1")
	mw := &mainWindow{
		w: w,
	}
	w.Connect("destroy", func() { mw.Close(nil) })

	mw.recentRomsView, err = newRecentRomsView(builder, mw.runROM)
	if err != nil {
		return nil, err
	}

	build[gtk.MenuItem](builder, "menu_quit").Connect("activate", gtk.MainQuit)
	build[gtk.MenuItem](builder, "menu_open").Connect("activate", func(m *gtk.MenuItem) {
		path, ok := openFileDialog(mw.w)
		if !ok {
			return
		}
		mw.runROM(path)
	})
	build[gtk.MenuItem](builder, "menu_controls").Connect("activate", func(m *gtk.MenuItem) {
		// TODO(arl): when we're in GUI mode, configuration should be stored and
		// saved back to ~/.config/nestor. However, for 'emu' only mode
		// configuration is read, but never modified nor saved.
		//
		// Hence configuration should be passed to the emulator when it powers
		// up, so the mainWindow as well?
		var cfg hw.InputConfig
		openInputConfigDialog(&cfg)
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
	mw.w.SetSensitive(false)

	rom, err := ines.ReadRom(path)
	if err != nil {
		modGUI.Warnf("failed to read ROM: %s", err)
		return
	}

	errc := make(chan error)
	go func() {
		defer mw.w.SetSensitive(true)

		emulator, err := emu.PowerUp(rom, emu.Config{})
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
		log.ModEmu.Fatalf("failed to start emulator window: %v", err)
		gtk.MainQuit()
	}
}

func (mw *mainWindow) addRecentROM(romPath string, screenshot image.Image) error {
	bb := bytes.Buffer{}
	if err := png.Encode(&bb, screenshot); err != nil {
		return fmt.Errorf("failed to encode screenshot: %v", err)
	}

	return mw.recentRomsView.addROM(recentROM{
		Name:     filepath.Base(romPath),
		Image:    bb.Bytes(),
		Path:     romPath,
		LastUsed: time.Now(),
	})
}
