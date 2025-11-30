package ui

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/ebitenui/ebitenui"
	"github.com/hajimehoshi/ebiten/v2"
	einput "github.com/quasilyte/ebitengine-input"

	"nestor/config"
	"nestor/emu"
	"nestor/hw/input"
	"nestor/ines"
)

type state interface {
	createUI()
	update()
	draw(screen *ebiten.Image)
	enter(arg any)
	exit()
}

type app struct {
	cfg config.Config

	emulator *emu.Emulator
	quit     atomic.Bool
	framech  chan *emu.Frame
	frameimg *ebiten.Image

	audioctx    *oto.Context
	samples     *sampleBuffer
	audioPlayer *oto.Player

	ui            ebitenui.UI
	displayWidth  int
	displayHeight int

	states   map[string]state
	curstate state

	inputsys     einput.System
	inputhandler *einput.Handler
}

func newApp(ctx context.Context, samples *sampleBuffer, audioctx *oto.Context, cfg config.Config) *app {
	app := &app{
		cfg:           cfg,
		states:        map[string]state{},
		displayWidth:  startwidth,
		displayHeight: startheight,
		samples:       samples,
		audioctx:      audioctx,
	}

	app.inputsys.Init(einput.SystemConfig{
		DevicesEnabled: einput.AnyDevice,
	})

	app.inputhandler = app.inputsys.NewHandler(0, globalKeymap)

	app.states["running"] = newRunningState(app)
	app.states["paused"] = newPausedState(app)
	app.states["main"] = newMainState(app)
	app.states["config"] = newConfigState(app)
	app.states["capture"] = newCaptureState(app)

	go func() {
		<-ctx.Done()
		app.exit()
	}()

	return app
}

func (app *app) exit() {
	app.quit.Store(true)
}

// setState defines the new application state, calling exit on the current and
// enter on the new one. Not re-entrant. Args are passed to the enter function
// of the new state.
func (app *app) setState(name string, arg any) {
	modUI.InfoZ("Switching to state").String("to", name).End()
	if app.curstate != nil {
		app.curstate.exit()
	}
	to, ok := app.states[name]
	if !ok {
		modUI.PanicZ("unknown state").String("state", name).End()
		return
	}
	app.curstate = to
	app.curstate.enter(arg)
	app.curstate.createUI()
}

func (app *app) Update() error {
	if app.quit.Load() {
		return ebiten.Termination
	}

	app.inputsys.Update()

	// Handle global shortcuts (available in all states)
	if app.inputhandler.ActionIsJustPressed(actionToggleFullscreen) {
		enable := !ebiten.IsFullscreen()
		ebiten.SetFullscreen(enable)

		if enable {
			app.displayWidth, app.displayHeight = ebiten.Monitor().Size()
			app.curstate.createUI()
		} else {
			app.displayWidth, app.displayHeight = ebiten.WindowSize()
			app.curstate.createUI()
		}
	}

	app.curstate.update()
	return nil
}

func (app *app) Draw(screen *ebiten.Image) {
	app.curstate.draw(screen)
}

func (app *app) Layout(outw, outh int) (screenw, screenh int) {
	if app.displayWidth == outw && app.displayHeight == outh {
		return outw, outh
	}

	app.displayWidth, app.displayHeight = outw, outh
	app.curstate.createUI()

	return outw, outh
}

func (app *app) runRom(romPath string) error {
	ebitenInput := input.NewEbitenInput(app.cfg.Input)

	rom, err := ines.ReadROM(romPath)
	if err != nil {
		return fmt.Errorf("failed to read ROM: %s", err)
	}

	framech := make(chan *emu.Frame)
	out := emu.NewOutput(framech,
		emu.OutputConfig{
			Width:          emu.NTSCWidth,
			Height:         emu.NTSCHeight,
			NumBackBuffers: 4,
			Title:          "Nestor",
			ScaleFactor:    2,
			DisableVSync:   app.cfg.Video.DisableVSync,
			Shader:         app.cfg.Video.Shader,
		},
	)

	emulator, err := emu.Launch(rom, app.cfg.Config, out, ebitenInput)
	if err != nil {
		return fmt.Errorf("failed to launch emulator: %s", err)
	}

	app.emulator = emulator
	app.framech = framech
	app.frameimg = ebiten.NewImageWithOptions(
		image.Rect(0, 0, emu.NTSCWidth, emu.NTSCHeight),
		&ebiten.NewImageOptions{Unmanaged: true},
	)

	app.audioPlayer = app.audioctx.NewPlayer(app.samples)
	app.audioPlayer.SetVolume(1)
	app.audioPlayer.SetBufferSize(8192)
	app.audioPlayer.Play()

	ebiten.SetVsyncEnabled(!app.cfg.Video.DisableVSync)

	app.setState("running", nil)

	go func() {
		defer func() {
			ebiten.SetVsyncEnabled(true)
			ebiten.SetWindowTitle("Nestor")
		}()

		tmpdir, err := os.MkdirTemp("", "nestor-")
		if err != nil {
			modUI.WarnZ("failed to create temp dir").Error("err", err).End()
		} else {
			emulator.SetTempDir(tmpdir)
		}

		emulator.Run()

		screenshot, err := os.ReadFile(filepath.Join(tmpdir, "screenshot.png"))
		if err != nil {
			modUI.WarnZ("failed to read screenshot").Error("err", err).End()
		}

		if err := addRecentROM(recentROM{
			Path:     romPath,
			Name:     filepath.Base(romPath),
			Image:    screenshot,
			LastUsed: time.Now(),
		}); err != nil {
			modUI.WarnZ("failed to add recent ROM").Error("err", err).End()
		}
	}()

	return nil
}

func (app *app) savecfg() {
	if err := config.Save(&app.cfg); err != nil {
		modUI.ErrorZ("failed to save config").Error("err", err).End()
	} else {
		modUI.InfoZ("config saved").String("path", config.Path()).End()
	}
}
