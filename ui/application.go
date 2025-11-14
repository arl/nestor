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
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"nestor/config"
	"nestor/emu"
	"nestor/emu/log"
	"nestor/hw/apu"
	"nestor/hw/input"
	"nestor/ines"
	"nestor/ui/shaders"
)

var modUI = log.NewModule("ui")

const (
	startwidth  = 800
	startheight = 600
)

func StartUI(ctx context.Context, cfg config.Config) error {
	return start(ctx, cfg, "")
}

func StartROM(ctx context.Context, cfg config.Config, romPath string) error {
	return start(ctx, cfg, romPath)
}

func start(ctx context.Context, cfg config.Config, romPath string) error {
	initResources()

	// Init audio.
	samples, audioctx, err := initAudio()
	if err != nil {
		return fmt.Errorf("initAudio failure: %s", err)
	}

	// Init video.
	setMonitor(cfg.Video.Monitor)
	ebiten.SetWindowTitle("Nestor")
	ebiten.SetWindowSize(startwidth, startheight)
	ebiten.SetWindowSizeLimits(startwidth, startheight, -1, -1)
	ebiten.SetRunnableOnUnfocused(false)
	if cfg.Video.StartFullscreen {
		ebiten.SetFullscreen(true)
	}
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebiten.SetVsyncEnabled(!cfg.Video.DisableVSync)
	var options = &ebiten.RunGameOptions{
		SingleThread: false,
	}

	app := newApp(ctx, samples, audioctx, cfg)

	if romPath != "" {
		app.setState("running")
		if err := app.runRom(romPath); err != nil {
			return fmt.Errorf("can't run rom: %w", err)
		}
	} else {
		app.setState("main")
	}

	if err := ebiten.RunGameWithOptions(app, options); err != nil {
		return fmt.Errorf("ui failure: %w", err)
	}

	modUI.InfoZ("ui quitted").End()
	return nil
}

func initAudio() (*sampleBuffer, *oto.Context, error) {
	const audioBufferSize = 1024 // TODO: adjust based on latency.
	samples := newSampleBuffer(audioBufferSize)

	audioctx, readych, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   apu.MaxSampleRate,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}

	const timeout = 5 * time.Second
	select {
	case <-readych:
		return samples, audioctx, nil
	case <-time.After(timeout):
		break
	}

	return nil, nil, fmt.Errorf("audio context not ready after %s", timeout)
}

// Can't fail, always fallback to primary/default monitor.
// Use 0 for primary monitor.
func setMonitor(idxmon uint) {
	monitors := ebiten.AppendMonitors(nil)
	selidx := 0
	for i, m := range monitors {
		modUI.InfoZ("Detected monitor").Int("idx", i).String("name", m.Name()).End()
		if i == int(idxmon) {
			selidx = i
		}
	}

	ebiten.SetMonitor(monitors[selidx])
	modUI.InfoZ("Using monitor").Int("idx", selidx).String("name", monitors[selidx].Name()).End()
}

type appState interface {
	createUI()
	update()
	draw(screen *ebiten.Image)
	enter(...any)
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

	ui       ebitenui.UI
	states   map[string]appState
	curstate appState

	screenw, screenh int
	shader           *ebiten.Shader
}

func newApp(ctx context.Context, samples *sampleBuffer, audioctx *oto.Context, cfg config.Config) *app {
	app := &app{
		cfg:      cfg,
		states:   map[string]appState{},
		screenw:  startwidth,
		screenh:  startheight,
		samples:  samples,
		audioctx: audioctx,
	}

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
func (app *app) setState(name string, args ...any) {
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
	app.curstate.enter(args...)
	app.curstate.createUI()
}

func (app *app) Update() error {
	if app.quit.Load() {
		return ebiten.Termination
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		enable := !ebiten.IsFullscreen()
		ebiten.SetFullscreen(enable)
	}

	if ebiten.IsFullscreen() {
		neww, newh := ebiten.Monitor().Size()
		if neww != app.screenw || newh != app.screenh {
			app.screenw, app.screenh = ebiten.Monitor().Size()
			app.curstate.createUI()
		}
	} else {
		neww, newh := ebiten.WindowSize()
		if neww != app.screenw || newh != app.screenh {
			app.screenw, app.screenh = ebiten.WindowSize()
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
	if app.screenw == outw && app.screenh == outh {
		return outw, outh
	}

	app.screenw, app.screenh = outw, outh
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

	const name = "gizmo-crt" // TODO: read from config
	shader, err := shaders.Load(name)
	if err != nil {
		modUI.FatalZ("can't load shader").String("name", name).Error("err", err).End()
	}
	app.shader = shader

	app.setState("running")

	go func() {
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
