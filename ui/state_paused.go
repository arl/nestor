package ui

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	einput "github.com/quasilyte/ebitengine-input"

	"nestor/ui/keymap"
)

type pausedState struct {
	*app
	inputh *einput.Handler
	menu   *appMenu
}

func newPausedState(app *app) *pausedState {
	s := &pausedState{
		app: app,
	}
	s.createUI()
	return s
}

func (s *pausedState) enter(inputh *einput.Handler, _ any) {
	ebiten.SetWindowTitle("Nestor <paused>")
	modUI.InfoZ("Blocking emulator").End()
	s.inputh = inputh
	s.emulator.Block()
	s.audioPlayer.Pause()
}

func (s *pausedState) exit() {}

func (s *pausedState) update() {
	if s.inputh.ActionIsJustPressed(keymap.ActionResumeEmulator) {
		s.onResume()
	}

	s.ui.Update()
}

func (s *pausedState) draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

func (s *pausedState) onResume() {
	s.emulator.Unblock()
	s.app.setState("running", nil)
	s.audioPlayer.Play()
}

func (s *pausedState) onReset() {
	s.emulator.Unblock()
	s.emulator.Reset()
	s.app.setState("running", nil)
	s.audioPlayer.Play()
}

func (s *pausedState) onReload() {
	s.emulator.Unblock()
	s.emulator.Restart()
	s.app.setState("running", nil)
	s.audioPlayer.Play()
}

func (s *pausedState) onStop() {
	s.emulator.Stop()
	<-s.framech
	s.app.setState("main", nil)
}

func (s *pausedState) createUI() {
	s.menu = newAppMenu(&s.ui)

	s.menu.fileQuit.ClickedEvent.AddHandler(func(args any) { s.onStop(); s.exit() })

	s.menu.fileOpen.GetWidget().Disabled = true
	s.menu.settingsGeneral.GetWidget().Disabled = true
	s.menu.settingsInput.GetWidget().Disabled = true
	s.menu.settingsVideo.GetWidget().Disabled = true
	s.menu.settingsEmulation.GetWidget().Disabled = true

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
		)),
	)

	root.AddChild(s.menu.container)

	content := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	buttonsGroup := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.image),
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
			widget.LabelOpts.Text("Paused", res.text.titleFace, &widget.LabelColor{Idle: res.text.idleColor}),
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

	content.AddChild(buttonsGroup)
	root.AddChild(content)
	s.ui.Container = root
}
