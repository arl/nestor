package ui

import (
	"fmt"
	"image"
	"slices"
	"time"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	einput "github.com/quasilyte/ebitengine-input"

	"nestor/ui/keymap"
	"nestor/ui/shader"
)

type runningState struct {
	*app
	paused  bool
	elapsed int64 // elapsed seconds (for FPS display)
	inputh  *einput.Handler

	shader     *ebiten.Shader
	debugui    debugui.DebugUI
	shaderuiOn bool

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

func (s *runningState) enter(hinput *einput.Handler, _ any) {
	s.inputh = hinput
	s.setShader(s.app.cfg.Video.Shader)
}

func (s *runningState) createUI() {}
func (s *runningState) exit()     {}

func (s *runningState) update() {
	if s.shaderuiOn {
		s.shaderUI()
	}

	if s.inputh.ActionIsJustPressed(keymap.ActionPauseEmulator) {
		s.app.setState("paused", nil)
	} else if s.inputh.ActionIsJustPressed(keymap.ActionResetEmulator) {
		s.emulator.Reset()
	} else if s.inputh.ActionIsJustPressed(keymap.ActionToggleShaderUI) {
		s.shaderuiOn = !s.shaderuiOn
	}
}

func (s *runningState) draw(screen *ebiten.Image) {
	if s.paused {
		if s.shouldQuit {
			s.paused = false
			s.shouldQuit = false
			s.emulator.Stop()
			<-s.framech // discard frame
			s.app.setState("main", nil)
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

	if s.shaderuiOn {
		s.debugui.Draw(screen)
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

	// Calculate the scaled dimensions
	scaledW := fw * scale
	scaledH := fh * scale

	// Draw the frame centered and apply shader.
	op := &ebiten.DrawRectShaderOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate((targetW-scaledW)/2, (targetH-scaledH)/2)
	op.Images[0] = frameImg

	screen.DrawRectShader(int(fw), int(fh), s.shader, op)
}

func (s *runningState) setShader(name string) {
	shader, err := shader.Load(name)
	if err != nil {
		modUI.FatalZ("can't load shader").String("name", name).Error("err", err).End()
	}
	if s.shader != nil {
		s.shader.Deallocate()
	}
	s.shader = shader
}

func (s *runningState) shaderUI() {
	shaders := shader.Names()
	idx := slices.Index(shaders, s.cfg.Video.Shader)
	_, err := s.debugui.Update(func(ctx *debugui.Context) error {
		ctx.Window("Shader", image.Rect(10, 10, 240, 90), func(layout debugui.ContainerLayout) {
			ctx.Text(fmt.Sprintf("%d/%d: %s", idx+1, len(shaders), shaders[idx]))
			ctx.SetGridLayout([]int{-1, -1}, nil)
			dec := func() {
				idx += len(shaders) - 1
				idx %= len(shaders)
				s.cfg.Video.Shader = shaders[idx]
				s.setShader(shaders[idx])
			}
			ctx.Button("Prev").On(dec)
			inc := func() {
				idx++
				idx %= len(shaders)
				s.cfg.Video.Shader = shaders[idx]
				s.setShader(shaders[idx])
			}
			ctx.Button("Next").On(inc)
		})
		return nil
	})
	if err != nil {
		panic(err)
	}
}
