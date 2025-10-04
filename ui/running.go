package ui

import (
	"fmt"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

type running struct {
	*Application
	paused  bool
	pauseUI *ebitenui.UI
}

func newRunningState(app *Application) *running {
	s := &running{
		Application: app,
	}
	s.initUI()
	return s
}

func (s *running) initUI() {
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(Mix(colornames.Black, Transparent, 0.5)),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	menu := widget.NewContainer(
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
	menu.AddChild(widget.NewLabel(
		widget.LabelOpts.Text("<paused>", TitleFont(), &widget.LabelColor{Idle: colornames.White}),
	))

	menu.AddChild(stdButton("Resume", func(_ *widget.ButtonClickedEventArgs) {
		s.togglePause()
	}))
	menu.AddChild(stdButton("Reset", func(_ *widget.ButtonClickedEventArgs) {
		s.togglePause()
		s.emulator.Reset()
	}))
	menu.AddChild(stdButton("Restart", func(_ *widget.ButtonClickedEventArgs) {
		s.togglePause()
		s.emulator.Restart()
	}))
	menu.AddChild(stdButton("Stop", func(_ *widget.ButtonClickedEventArgs) {
		s.togglePause()
		s.emulator.Stop()
		s.Application.setState("rom_list")
	}))

	root.AddChild(menu)
	s.pauseUI = &ebitenui.UI{Container: root}
}

func (s *running) togglePause() {
	s.paused = !s.paused
	modUI.InfoZ("Toggling pause").Bool("paused", s.paused).End()
	s.emulator.SetPause(s.paused)
}

func (s *running) Update() {
	if s.paused {
		s.pauseUI.Update()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		s.emulator.Reset()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.togglePause()
	}
}

func (s *running) Draw(screen *ebiten.Image) {
	if s.paused {
		s.pauseUI.Draw(screen)
		return
	}

	frame := <-s.framech
	s.Application.frameimg.WritePixels(frame.Video)

	// TODO: precalculate screen bounds on resize only
	s.drawFrame(screen, s.Application.frameimg, float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy()))
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %f", ebiten.ActualFPS()))
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
