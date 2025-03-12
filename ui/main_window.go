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

func init() {
	log.EnableDebugModules(modGUI.Mask())
	log.EnableDebugModules(log.ModEmu.Mask())
}

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

	stopEmu func() *image.RGBA
}

func showMainWindow(cfg *Config) {
	builder := mustT(gtk.BuilderNewFromString(mainWindowUI))

	mw := &mainWindow{
		Window:  build[gtk.Window](builder, "main_window"),
		cfg:     cfg,
		stopEmu: func() *image.RGBA { return nil },
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

	if mw.stopEmu != nil {
		mw.stopEmu()
	}

	mw.wg.Wait()
	gtk.MainQuit()
}

func (mw *mainWindow) runROM(path string) {
	mw.SetSensitive(false)

	executable, err := os.Executable()
	if err != nil {
		modGUI.Warnf("failed to get executable path: %s", err)
		return
	}

	port := rpc.UnusedPort()
	args := []string{"run",
		"--monitor", fmt.Sprint(mw.monitorIdx()),
		"--port", strconv.Itoa(port),
		path}

	cmd := exec.Command(executable, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		modGUI.Warnf("failed to start emulator process: %s", err)
		return
	}

	client, err := rpc.NewClient(port)
	if err != nil {
		modGUI.Warnf("failed to create emulator window proxy: %s", err)
		return
	}

	mw.stopEmu = client.Stop

	panel := showGamePanel(mw.Window)
	panel.connect(client)

	mw.wg.Add(1)
	go func() {
		defer mw.SetSensitive(true)
		defer mw.wg.Done()

		modGUI.DebugZ("waiting for emulator process to finish").End()
		cmd.Wait()
		modGUI.DebugZ("closing game panel").End()
		panel.Close()
		glib.IdleAdd(func() {
			if err := mw.addRecentROM(path, panel.img); err != nil {
				modGUI.Warnf("failed to add recent ROM: %s", err)
			}
		})
	}()
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
