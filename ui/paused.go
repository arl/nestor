package ui

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

type pausedState struct{ *app }

func newPausedState(app *app) *pausedState {
	s := &pausedState{app: app}
	s.createUI()
	return s
}

func (s *pausedState) enter(...any) {
	ebiten.SetWindowTitle("Nestor <paused>")
	modUI.InfoZ("Blocking emulator").End()
	s.emulator.Block()
	s.audioPlayer.Pause()
}

func (s *pausedState) exit() {}

func (s *pausedState) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.onResume()
	}

	s.ui.Update()
}

func (s *pausedState) draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

func (s *pausedState) onResume() {
	s.emulator.Resume()
	s.app.setState("running")
	s.audioPlayer.Play()
}

func (s *pausedState) onReset() {
	s.emulator.Resume()
	s.emulator.Reset()
	s.app.setState("running")
	s.audioPlayer.Play()
}

func (s *pausedState) onReload() {
	s.emulator.Resume()
	s.emulator.Restart()
	s.app.setState("running")
	s.audioPlayer.Play()
}

func (s *pausedState) onStop() {
	s.emulator.Stop()
	<-s.framech // discard frame
	s.app.setState("main")
}

func (s *pausedState) createUI() {
	root := widget.NewContainer(
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

	buttonsGroup.AddChild(
		widget.NewLabel(
			widget.LabelOpts.Text("<paused>", res.fonts.titleFace, &widget.LabelColor{Idle: colornames.White}),
			widget.LabelOpts.TextOpts(widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter)),
		),

		widget.NewButton(
			widget.ButtonOpts.Text("Resume", res.button.face, res.button.text),
			widget.ButtonOpts.TextPadding(res.button.padding),
			widget.ButtonOpts.Image(res.button.image),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				s.onResume()
			}),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			),
		),

		widget.NewButton(
			widget.ButtonOpts.Text("Press Reset", res.button.face, res.button.text),
			widget.ButtonOpts.TextPadding(res.button.padding),
			widget.ButtonOpts.Image(res.button.image),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				s.onReset()
			}),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			),
		),

		widget.NewButton(
			widget.ButtonOpts.Text("Reload ROM", res.button.face, res.button.text),
			widget.ButtonOpts.TextPadding(res.button.padding),
			widget.ButtonOpts.Image(res.button.image),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				s.onReload()
			}),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			),
		),

		widget.NewButton(
			widget.ButtonOpts.Text("Stop", res.button.face, res.button.text),
			widget.ButtonOpts.TextPadding(res.button.padding),
			widget.ButtonOpts.Image(res.button.image),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				s.onStop()
			}),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			),
		),
	)

	root.AddChild(buttonsGroup)
	s.ui.Container = root
}
