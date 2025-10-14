package ui

import (
	"context"
	"fmt"
	"image"
	"sync"
	"sync/atomic"

	"github.com/ebitengine/oto/v3"
	"github.com/ebitenui/ebitenui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"nestor/config"
	"nestor/emu"
	"nestor/hw/apu"
	"nestor/hw/input"
	"nestor/ines"
)

type state interface {
	createUI()
	update()
	draw(screen *ebiten.Image)
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

	ui           ebitenui.UI
	states       map[string]state
	currentState state

	screenw, screenh int
}

var res = newUIResources()

func newApp(ctx context.Context, cfg config.Config) *app {
	app := &app{
		cfg:     cfg,
		states:  map[string]state{},
		screenw: minWidth,
		screenh: minHeight,
	}

	app.initAudio()

	app.states["running"] = newRunningState(app)
	app.states["rom_list"] = newRomListState(app)

	go func() {
		<-ctx.Done()
		app.exit()
	}()

	return app
}

func (app *app) initAudio() {
	// init audio
	const audioBufferSize = 1024 // TODO: adjust based on latency.
	app.samples = newSampleBuffer(audioBufferSize)

	audioctx, readych, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   apu.MaxSampleRate,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readych

	app.audioctx = audioctx
}

var otoContext = sync.OnceValue(func() *oto.Context {
	context, readyChan, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   apu.MaxSampleRate,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan
	return context
})

func (app *app) exit() {
	app.quit.Store(true)
}

func (app *app) setState(name string) {
	modUI.InfoZ("Switching to state").String("to", name).End()
	app.currentState = app.states[name]
	app.currentState.createUI()
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
			app.currentState.createUI()
		}
	} else {
		neww, newh := ebiten.WindowSize()
		if neww != app.screenw || newh != app.screenh {
			app.screenw, app.screenh = ebiten.WindowSize()
			app.currentState.createUI()
		}
	}

	app.currentState.update()
	return nil
}

func (app *app) Draw(screen *ebiten.Image) {
	app.currentState.draw(screen)
}

func (app *app) Layout(outw, outh int) (screenw, screenh int) {
	if app.screenw == outw && app.screenh == outh {
		return outw, outh
	}

	app.screenw, app.screenh = outw, outh
	app.currentState.createUI()

	return outw, outh
}

func (app *app) runRom(romPath string) error {
	inputProvider := input.NewProvider(app.cfg.Input)

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
			Monitor:        app.cfg.Video.Monitor,
			Shader:         app.cfg.Video.Shader,
		},
	)

	emulator, err := emu.Launch(rom, app.cfg.Config, out, inputProvider)
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

	go func() {
		emulator.Run()
		app.audioPlayer.Close()
	}()

	return nil
}
