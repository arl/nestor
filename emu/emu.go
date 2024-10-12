package emu

import (
	"fmt"
	"image"
	"io"
	"time"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
)

type Output interface {
	io.Closer

	BeginFrame() hw.Frame
	EndFrame(hw.Frame)
	Poll() bool
	Screenshot() image.Image
}

type Emulator struct {
	NES *NES
	out Output

	userAction chan func()
	paused     bool
	pauseCh    chan struct{}
}

// Launch instantiates an emulator, setup controllers, output streams and window.
func Launch(rom *ines.Rom, cfg Config, monidx int32) (*Emulator, error) {
	nes, err := powerUp(rom)
	if err != nil {
		return nil, fmt.Errorf("power up failed: %s", err)
	}

	// Output setup.
	nes.Frames = make(chan image.RGBA)
	out := hw.NewOutput(hw.OutputConfig{
		Width:           256,
		Height:          240,
		NumVideoBuffers: 2,
		Title:           "Nestor",
		ScaleFactor:     2,
		DisableVSync:    cfg.Video.DisableVSync,
		Monitor:         monidx,
	})
	if err := out.EnableVideo(true); err != nil {
		return nil, err
	}

	input, err := hw.NewInputProvider(cfg.Input)
	if err != nil {
		return nil, fmt.Errorf("input provider: %s", err)
	}
	nes.CPU.PlugInputDevice(input)

	// CPU trace setup.
	if cfg.TraceOut != nil {
		nes.CPU.SetTraceOutput(cfg.TraceOut)
	}

	return &Emulator{
		NES:        nes,
		out:        out,
		userAction: make(chan func(), 1),
		paused:     false,
		pauseCh:    make(chan struct{}),
	}, nil
}

func (e *Emulator) handleUserAction() {
	select {
	case a := <-e.userAction:
		a()
	default:
	}
}
func (e *Emulator) Pause() (paused bool) {
	e.paused = !e.paused
	go func() {
		if !e.paused {
			e.play()
			return
		}
		e.userAction <- e.pause
	}()
	return e.paused
}

// pause blocks the emulator loop until either play()
// is called or the output window is closed.
func (e *Emulator) pause() {
	outpoll := time.NewTicker(100 * time.Millisecond)
	defer outpoll.Stop()
	e.pauseCh = make(chan struct{})

	for {
		select {
		case <-e.pauseCh:
			return
		case <-outpoll.C:
			// poll the output window.
			if !e.out.Poll() {
				return
			}
		}
	}
}

func (e *Emulator) play() { close(e.pauseCh) }

func (e *Emulator) Run() {
	for e.out.Poll() {
		e.RunOneFrame()
		if e.NES.CPU.IsHalted() {
			break
		}
		e.handleUserAction()
	}

	log.ModEmu.InfoZ("Emulation stopped").End()
	if err := e.out.Close(); err != nil {
		log.ModEmu.WarnZ("Error closing emulator window").Error("error", err).End()
	}
}

func (e *Emulator) RunOneFrame() {
	frame := e.out.BeginFrame()
	e.NES.RunOneFrame(frame)
	e.out.EndFrame(frame)
}

func (e *Emulator) Screenshot() image.Image {
	return e.out.Screenshot()
}
