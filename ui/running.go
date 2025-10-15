package ui

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type runningState struct {
	*app
	paused  bool
	elapsed int64 // elapsed seconds (for FPS display)

	shouldQuit bool
}

func newRunningState(app *app) *runningState {
	s := &runningState{
		app:     app,
		elapsed: time.Now().Unix(),
	}
	s.createUI()
	return s
}

func (s *runningState) createUI() {}
func (s *runningState) enter()    {}
func (s *runningState) exit()     {}

func (s *runningState) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.app.setState("paused")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		s.emulator.Reset()
	}
}

func (s *runningState) draw(screen *ebiten.Image) {
	if s.paused {
		if s.shouldQuit {
			s.paused = false
			s.shouldQuit = false
			s.emulator.Stop()
			<-s.framech // discard frame
			s.app.setState("rom_list")
			return
		}

		s.ui.Draw(screen)
		return
	}

	// retrieve frame
	frame := <-s.framech

	// audio

	// With audio and vsync enabled, we use vsync to enforce the correct
	// emulation speed. In this case we want to avoid queueing too much audio as
	// it would desync from video. So we send audio only if there's less than a
	// frame's worth of audio samples already queued in the buffer.
	if s.app.samples.Len() < len(frame.Audio.Samples)/2 {
		buf := byteSliceFromInt16(frame.Audio.Samples)
		if _, err := s.samples.Write(buf); err != nil {
			panic(err)
		}
	}

	// video
	s.app.frameimg.WritePixels(frame.Video)

	// TODO: precalculate screen bounds on resize only
	s.drawFrame(screen, s.app.frameimg, float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy()))

	if now := time.Now().Unix(); now != s.elapsed {
		s.elapsed = now
		ebiten.SetWindowTitle(fmt.Sprintf("Nestor - FPS: %.1f", ebiten.ActualTPS()))
	}
}

func (s *runningState) drawFrame(screen *ebiten.Image, frameImg *ebiten.Image, targetW, targetH float64) {
	// TODO: precalculate this on resize only

	// Calculate scaling to fit the target area while preserving aspect ratio.
	fw, fh := float64(frameImg.Bounds().Dx()), float64(frameImg.Bounds().Dy())
	scaleX := targetW / fw
	scaleY := targetH / fh
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Draw the frame centered
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate((targetW-fw*scale)/2, (targetH-fh*scale)/2)
	screen.DrawImage(frameImg, op)
}
