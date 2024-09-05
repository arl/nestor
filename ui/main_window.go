package ui

import (
	"fmt"
	"os"
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

	mw.recentRomsView, err = newRecentRomsView(builder, mw.guiRunROM)
	if err != nil {
		return nil, err
	}

	build[gtk.MenuItem](builder, "menu_quit").Connect("activate", gtk.MainQuit)
	build[gtk.MenuItem](builder, "menu_open").Connect("activate", func(m *gtk.MenuItem) {
		path, ok := openFileDialog(mw.w)
		if !ok {
			return
		}

		mw.guiRunROM(path)
	})

	return mw, nil
}

func (mw *mainWindow) Close(err error) {
	if err != nil {
		modGUI.Warnf("closing UI with error: %s", err)
	}
	gtk.MainQuit()
}

func (mw *mainWindow) guiRunROM(path string) {
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

		// TODO: should Screenshot return an image.Image (or image.RGBA) instead
		// of writing to a file? We wouldn't need to crate the file, close it,
		// read it again, to pass the bytes to the recent roms view.
		screenshot := tempFile()
		emulator.Screenshot(screenshot)
		fmt.Println(screenshot)

		glib.IdleAdd(func() {
			if err := mw.addRecentROM(path, screenshot); err != nil {
				modGUI.Warnf("failed to add recent ROM: %s", err)
			}
		})
	}()

	if err := <-errc; err != nil {
		if err != nil {
			log.ModEmu.Fatalf("failed to start emulator window: %v", err)
			gtk.MainQuit()
		}
	}
}

func tempFile() string {
	f, err := os.CreateTemp("", "nestor_*")
	if err != nil {
		log.ModEmu.Fatalf("failed to create temporary file: %s", err)
	}
	f.Close()
	return f.Name()
}

func (mw *mainWindow) addRecentROM(path, screenshot string) error {
	nrr := recentROM{
		Name:     filepath.Base(path),
		Image:    mustT(os.ReadFile(screenshot)),
		Path:     path,
		LastUsed: time.Now(),
	}
	mw.recentRomsView.addROM(nrr)
	if err := nrr.save(); err != nil {
		return fmt.Errorf("failed to save recent rom: %v", err)
	}
	mw.recentRomsView.refreshView()
	return nil
}
