package ui

import (
	"fmt"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

type running struct {
	*Application
	paused  bool
	pauseUI *ebitenui.UI
	elapsed int64 // elapsed seconds (for FPS display)

	shouldQuit bool
}

func newRunningState(app *Application) *running {
	s := &running{
		Application: app,
		elapsed:     time.Now().Unix(),
	}
	s.initUI()
	return s
}

func (s *running) initUI() {
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(mixColors(colornames.Black, transparent, 0.5)),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	buttonsGroup := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Gray),
		),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(&widget.Insets{Top: 20, Left: 20, Right: 20, Bottom: 20}),
			widget.GridLayoutOpts.Spacing(10, 10),
			widget.GridLayoutOpts.Columns(1),
		)),
	)
	buttonsGroup.AddChild(widget.NewLabel(
		widget.LabelOpts.Text("<paused>", titleFont(), &widget.LabelColor{Idle: colornames.White}),
		widget.LabelOpts.TextOpts(widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter)),
	))

	buttonsGroup.AddChild(stdButton("Resume", func(_ *widget.ButtonClickedEventArgs) {
		s.resume()
	}))
	buttonsGroup.AddChild(stdButton("Reset", func(_ *widget.ButtonClickedEventArgs) {
		s.resume()
		s.emulator.Reset()
		s.audioPlayer.Play()
	}))
	buttonsGroup.AddChild(stdButton("Restart", func(_ *widget.ButtonClickedEventArgs) {
		s.resume()
		s.emulator.Restart()
		s.audioPlayer.Play()
	}))
	buttonsGroup.AddChild(stdButton("Stop", func(_ *widget.ButtonClickedEventArgs) {
		s.shouldQuit = true
	}))

	root.AddChild(buttonsGroup)
	s.pauseUI = &ebitenui.UI{Container: root}
}

func (s *running) pause() {
	ebiten.SetWindowTitle("Nestor <paused>")
	modUI.InfoZ("Pause emulator").End()
	s.paused = true
	s.emulator.Block()
}

func (s *running) resume() {
	ebiten.SetWindowTitle("Nestor")
	modUI.InfoZ("Resume emulator").End()
	s.paused = false
	s.emulator.Resume()
}

func (s *running) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if s.paused {
			s.resume()
			s.audioPlayer.Play()
		} else {
			s.pause()
			s.audioPlayer.Pause()
		}
	}
	if s.paused {
		s.pauseUI.Update()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		s.emulator.Reset()
	}
}

func (s *running) Draw(screen *ebiten.Image) {
	if s.paused {
		if s.shouldQuit {
			s.paused = false
			s.shouldQuit = false
			s.emulator.Stop()
			<-s.framech // discard frame
			s.Application.setState("rom_list")
			return
		}

		s.pauseUI.Draw(screen)
		return
	}

	// retrieve frame
	frame := <-s.framech

	// audio

	// With audio and vsync enabled, we use vsync to enforce
	// the correct emulation speed. In this case we want to
	// avoid queueing too much audio as it would desync from
	// video. So we send audio only if there's less than a
	// frame's worth in the buffer.
	if s.Application.samples.Len() < len(frame.Audio.Samples)/2 {
		buf := byteSliceFromInt16(frame.Audio.Samples)
		if _, err := s.samples.Write(buf); err != nil {
			panic(err)
		}
	}

	// video
	s.Application.frameimg.WritePixels(frame.Video)

	// TODO: precalculate screen bounds on resize only
	s.drawFrame(screen, s.Application.frameimg, float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy()))

	if now := time.Now().Unix(); now != s.elapsed {
		s.elapsed = now
		ebiten.SetWindowTitle(fmt.Sprintf("Nestor - FPS: %.1f", ebiten.ActualTPS()))
	}
}

func (s *running) drawFrame(screen *ebiten.Image, frameImg *ebiten.Image, targetW, targetH float64) {
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
