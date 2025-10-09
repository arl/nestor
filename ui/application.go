package ui

import (
	"context"
	"fmt"
	"image"
	"sync"
	"sync/atomic"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/ebiten/v2"

	"nestor/config"
	"nestor/emu"
	"nestor/hw/apu"
	"nestor/hw/input"
	"nestor/ines"
)

type state interface {
	Update()
	Draw(screen *ebiten.Image)
}

type Application struct {
	cfg config.Config

	emulator *emu.Emulator
	quit     atomic.Bool
	framech  chan *emu.Frame
	frameimg *ebiten.Image

	samples     *sampleBuffer
	audioPlayer *oto.Player

	states       map[string]state
	currentState state

	screenw, screenh int
}

func newApplication(ctx context.Context, cfg config.Config) *Application {
	app := &Application{
		cfg:     cfg,
		states:  make(map[string]state),
		screenw: minWidth,
		screenh: minHeight,
	}

	app.states["running"] = newRunningState(app)
	app.states["rom_list"] = newRomListState(app)

	go func() {
		<-ctx.Done()
		app.exit()
	}()

	return app
}

func (app *Application) exit() {
	app.quit.Store(true)
}

func (app *Application) setState(name string) {
	modUI.InfoZ("Switching to state").String("to", name).End()
	app.currentState = app.states[name]
}

func (app *Application) Update() error {
	if app.quit.Load() {
		return ebiten.Termination
	}
	app.currentState.Update()
	return nil
}

func (app *Application) Draw(screen *ebiten.Image) {
	app.currentState.Draw(screen)
}

func (app *Application) Layout(outw, outh int) (screenw, screenh int) {
	app.screenw, app.screenh = outw, outh
	return outw, outh
}

func (app *Application) runRom(romPath string) error {
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

	// init audio
	const audioBufferSize = 1024 // TODO: adjust based on latency.
	app.samples = newSampleBuffer(audioBufferSize)
	if app.audioPlayer != nil {
		app.audioPlayer.Close()
	}
	app.audioPlayer = otoContext().NewPlayer(app.samples)
	app.audioPlayer.SetVolume(1)
	app.audioPlayer.SetBufferSize(8192)
	app.audioPlayer.Play()

	go emulator.Run()

	return nil
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
