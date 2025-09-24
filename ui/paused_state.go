package ui

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

// PausedState represents the state when a ROM is paused
type PausedState struct {
	ui   *ebitenui.UI
	root *widget.Container
	app  *App
}

// NewPausedState creates a new paused state
func NewPausedState(app *App) *PausedState {
	state := &PausedState{
		app: app,
	}

	// Initialize UI
	state.initUI()

	return state
}

// initUI creates the UI for this state
func (s *PausedState) initUI() {
	// Create root container
	s.root = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Gainsboro),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	// Create pause overlay
	overlay := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(Mix(colornames.Black, Transparent, 0.5)),
		),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	// Create pause menu
	pauseMenu := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Gray),
		),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				// MinWidth:           300,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Top:    20,
				Left:   20,
				Right:  20,
				Bottom: 20,
			}),
			widget.GridLayoutOpts.Spacing(10, 10),
			widget.GridLayoutOpts.Columns(1),
		)),
	)

	// Add a title
	titleLabel := widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionCenter,
			}),
		)),
		widget.LabelOpts.Text("PAUSED", TitleFont(), &widget.LabelColor{
			Idle:     colornames.White,
			Disabled: Mix(colornames.White, colornames.Black, 0.4),
		}),
	)
	pauseMenu.AddChild(titleLabel)

	bc := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)))
	pauseMenu.AddChild(bc)

	bc.AddChild(stdButton("Resume", func(args *widget.ButtonClickedEventArgs) {
		s.app.ChangeState(s.app.GetState(StateRomRunning))
	}))
	bc.AddChild(stdButton("Reset", func(args *widget.ButtonClickedEventArgs) {
		s.app.ResetEmulator()
		s.app.ChangeState(s.app.GetState(StateRomRunning))
	}))
	bc.AddChild(stdButton("Restart", func(args *widget.ButtonClickedEventArgs) {
		s.app.RestartEmulator()
		s.app.ChangeState(s.app.GetState(StateRomRunning))
	}))
	bc.AddChild(stdButton("Stop", func(args *widget.ButtonClickedEventArgs) {
		s.app.StopEmulator()
		s.app.ChangeState(s.app.GetState(StateRomList))
	}))

	overlay.AddChild(pauseMenu)
	s.root.AddChild(overlay)

	s.ui = &ebitenui.UI{Container: s.root}
}

// Enter implements State interface
func (s *PausedState) Enter(prevState State) {
	modUI.InfoZ("entering paused state").End()
	// s.app.emulator.SetPause(true)
}

// Exit implements State interface
func (s *PausedState) Exit(nextState State) {
	modUI.InfoZ("exiting paused state").End()
	// s.app.emulator.SetPause(false)
}

// Update implements State interface
func (s *PausedState) Update() {
	// Handle Escape key to resume
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.app.ChangeState(s.app.GetState(StateRomRunning))
	}

	s.ui.Update()
}

// Draw implements State interface
func (s *PausedState) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

// ID implements State interface
func (s *PausedState) ID() StateID {
	return StateRomPaused
}
