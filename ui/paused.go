package ui

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

type pausedState struct {
	*app

	shouldQuit bool
}

func newPausedState(app *app) *pausedState {
	s := &pausedState{
		app: app,
	}
	s.createUI()
	return s
}

func (s *pausedState) enter() {
	ebiten.SetWindowTitle("Nestor <paused>")
	modUI.InfoZ("Pause emulator").End()
	s.emulator.Block()
	s.audioPlayer.Pause()
}

func (s *pausedState) resume() {
	ebiten.SetWindowTitle("Nestor")
	modUI.InfoZ("Resume emulator").End()
}

func (s *pausedState) exit() {
	s.audioPlayer.Play()
}

func (s *pausedState) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.app.setState("running")
		s.emulator.Resume()
		return
	}

	s.ui.Update()
}

func (s *pausedState) draw(screen *ebiten.Image) {
	if s.shouldQuit {
		s.shouldQuit = false
		s.emulator.Stop()
		<-s.framech // discard frame
		s.app.setState("rom_list")
		return
	}

	s.ui.Draw(screen)
	return
}

func (s *pausedState) createUI() {
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
		widget.LabelOpts.Text("<paused>", res.fonts.titleFace, &widget.LabelColor{Idle: colornames.White}),
		widget.LabelOpts.TextOpts(widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter)),
	))
	buttonsGroup.AddChild(stdButton("Resume", func(_ *widget.ButtonClickedEventArgs) {
		s.emulator.Resume()
		s.app.setState("running")
	}))
	buttonsGroup.AddChild(stdButton("Reset", func(_ *widget.ButtonClickedEventArgs) {
		s.emulator.Resume()
		s.emulator.Reset()
		s.app.setState("running")
	}))
	buttonsGroup.AddChild(stdButton("Restart", func(_ *widget.ButtonClickedEventArgs) {
		s.emulator.Resume()
		s.emulator.Restart()
		s.app.setState("running")
	}))
	buttonsGroup.AddChild(stdButton("Stop", func(_ *widget.ButtonClickedEventArgs) {
		s.shouldQuit = true
	}))

	root.AddChild(buttonsGroup)
	s.ui.Container = root
}
