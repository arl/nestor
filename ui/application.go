package ui

import (
	"context"
	"fmt"
	"image"
	"path/filepath"
	"sync/atomic"
	"time"

	"nestor/config"
	"nestor/emu"
	"nestor/emu/log"
	"nestor/hw/hwinput"
	"nestor/ines"
	"nestor/ui/input"

	"github.com/ebitengine/oto/v3"
	"github.com/ebitenui/ebitenui"
	"github.com/hajimehoshi/ebiten/v2"
)

type state interface {
	createUI()
	update()
	draw(screen *ebiten.Image)
	enter(arg any)
	exit()
}

type stateDef struct {
	state  state
	keymap input.Keymap
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

	states       map[string]stateDef
	curstate     state
	stateKeymap  input.Keymap
	inputEnabled bool
	actions      *actionRegistry

	pendingfunc atomic.Pointer[func()] // function to call in the main loop, use app.do()
}

func newApp(ctx context.Context, samples *sampleBuffer, audioctx *oto.Context, cfg config.Config) *app {
	app := &app{
		cfg:           cfg,
		states:        make(map[string]stateDef),
		displayWidth:  startwidth,
		displayHeight: startheight,
		samples:       samples,
		audioctx:      audioctx,
		inputEnabled:  true,
		actions:       newActionRegistry(),
	}

	app.registerActions()

	app.states["main"] = stateDef{state: newMainState(app), keymap: input.MenuKeymap}
	app.states["running"] = stateDef{state: newRunningState(app), keymap: input.RunningKeymap}
	app.states["paused"] = stateDef{state: newPausedState(app), keymap: input.PausedKeymap}
	app.states["config"] = stateDef{state: newConfigState(app), keymap: nil}
	app.states["capture"] = stateDef{state: newCaptureState(app), keymap: nil}

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

	app.stateKeymap = input.GlobalKeymap.Merge(to.keymap)
	app.inputEnabled = true

	app.curstate = to.state
	app.curstate.enter(arg)
	app.curstate.createUI()
}

func (app *app) currentStateName() string {
	for name, def := range app.states {
		if def.state == app.curstate {
			return name
		}
	}

	panic("current state not found!")
}

func (app *app) do(fn func()) {
	app.pendingfunc.Store(&fn)
}

func (app *app) Update() error {
	if app.quit.Load() {
		return ebiten.Termination
	}

	if fn := app.pendingfunc.Swap(nil); fn != nil {
		(*fn)()
	}

	if app.inputEnabled {
		if action := app.stateKeymap.JustPressed(); action != input.ActionNone {
			app.actions.trigger(action)
		}
	}

	app.curstate.update()
	return nil
}

// disableInputHandler disables input handler for current state's lifetime.
func (app *app) disableInputHandler() {
	app.inputEnabled = false
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

func (app *app) runRom(romPath string, savestate []byte) error {
	ebitenInput := hwinput.NewEbitenInput(app.cfg.Input)

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

	ebiten.SetVsyncEnabled(app.cfg.Video.VSync)

	app.setState("running", nil)

	go func() {
		defer func() {
			app.do(func() { app.setState("main", nil) })

			ebiten.SetVsyncEnabled(true)
			ebiten.SetWindowTitle("Nestor")
		}()

		if savestate != nil {
			if err := emulator.NES.LoadSnapshot(savestate); err != nil {
				modUI.ErrorZ("failed to load savestate").Error("err", err).End()
			}
		}

		execstate, err := emulator.Run()
		if err != nil {
			log.ModEmu.ErrorZ("emulation ended").Error("err", err).End()
			return
		}

		if err := config.AddRecentROM(config.RecentROM{
			Path:      romPath,
			Name:      filepath.Base(romPath),
			Image:     execstate.PNGBytes,
			SaveState: execstate.SaveState,
			LastUsed:  time.Now(),
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

func (app *app) savestateToSlot(slot int) {
	var (
		state emu.ExecState
		err   error
	)

	save := func() {
		if err != nil {
			modUI.ErrorZ("failed to save state").Int("slot", slot+1).Error("err", err).End()
			return
		}

		modUI.InfoZ("saved state").Int("slot", slot+1).End()
		_ = state
	}

	if app.currentStateName() == "paused" {
		state, err = app.emulator.SavestateUnsafe()
		save()
	} else {
		go func() {
			state, err = app.emulator.Savestate()
			save()
		}()
	}
}
