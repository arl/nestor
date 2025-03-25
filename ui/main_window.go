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
	"strconv"
	"sync"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"nestor/emu/log"
	"nestor/emu/rpc"
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
}

func showMainWindow(cfg *Config) {
	builder := mustT(gtk.BuilderNewFromString(mainWindowUI))

	mw := &mainWindow{
		Window: build[gtk.Window](builder, "main_window"),
		cfg:    cfg,
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
		showConfig(mw.cfg, m.GetLabel())
		if err := saveConfig(mw.cfg); err != nil {
			modGUI.Warnf("failed to save config: %s", err)
		}
	}
	build[gtk.MenuItem](builder, "menu_input").Connect("activate", onConfig)
	build[gtk.MenuItem](builder, "menu_video").Connect("activate", onConfig)
	build[gtk.MenuItem](builder, "menu_audio").Connect("activate", onConfig)
}

func (mw *mainWindow) Close(err error) {
	if err != nil {
		modGUI.Warnf("closing UI with error: %s", err)
	}

	mw.wg.Wait()
	gtk.MainQuit()
}

func (mw *mainWindow) runROM(path string) {
	mw.SetSensitive(false)

	monidx := monitorIdx(mustT(mw.GetWindow()))

	panel := showGamePanel(mw.Window)
	client, wait, err := driveEmulator(path, monidx)
	if err != nil {
		modGUI.WarnZ("failed to start rom").Error("err", err).End()
		panel.Close()
		mw.SetSensitive(true)
		return
	}

	panel.connect(client)

	mw.wg.Add(1)
	go func() {
		defer mw.wg.Done()
		defer mw.SetSensitive(true)

		modGUI.DebugZ("waiting for emulator process to finish").End()
		wait()

		glib.IdleAdd(func() {
			panel.setGameStopped()
			modGUI.DebugZ("closing game panel").End()
			panel.Close()
			mw.onRomStopped(path, client.TempDir())
		})
	}()
}

func (mw *mainWindow) onRomStopped(rompath, tmpdir string) {
	f, err := os.Open(filepath.Join(tmpdir, "screenshot.png"))
	if err != nil {
		modGUI.Warnf("failed to read screenshot: %s", err)
		return
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		modGUI.Warnf("failed to decode screenshot: %s", err)
		return
	}

	if err := mw.addRecentROM(rompath, img); err != nil {
		modGUI.Warnf("failed to add recent ROM: %s", err)
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

type waitFunc func() error

func driveEmulator(rompath string, monidx int32) (*rpc.Client, waitFunc, error) {
	port := rpc.UnusedPort()
	args := []string{"run",
		"--monitor", strconv.Itoa(int(monidx)),
		"--port", strconv.Itoa(port),
		rompath}

	cmd := exec.Command(mustT(os.Executable()), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start emulator process: %w", err)
	}

	client, err := rpc.NewClient(port)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create emulator proxy: %w", err)
	}

	return client, cmd.Wait, nil
}
