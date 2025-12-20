package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/ebiten/v2"

	"nestor/config"
	"nestor/emu/log"
	"nestor/hw/apu"
)

var modUI = log.NewModule("ui")

func StartUI(ctx context.Context, cfg config.Config) error {
	return entrypoint(ctx, cfg, "")
}

func StartROM(ctx context.Context, cfg config.Config, romPath string) error {
	return entrypoint(ctx, cfg, romPath)
}

const startwidth = 800
const startheight = 600

func entrypoint(ctx context.Context, cfg config.Config, romPath string) error {
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
	ebiten.SetRunnableOnUnfocused(true)

	app := newApp(ctx, samples, audioctx, cfg)

	if romPath != "" {
		app.setState("running", nil)
		if err := app.runRom(romPath, nil); err != nil {
			return fmt.Errorf("can't run rom: %w", err)
		}
	} else {
		app.setState("main", nil)
	}

	options := &ebiten.RunGameOptions{
		SingleThread: false,
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
