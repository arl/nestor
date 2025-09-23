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
	context StateContext
	ui      *ebitenui.UI
	root    *widget.Container
}

// NewPausedState creates a new paused state
func NewPausedState(context StateContext) *PausedState {
	state := &PausedState{
		context: context,
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

	// Add resume button
	resumeButton := widget.NewButton(
		widget.ButtonOpts.TextFace(DefaultFont()),
		widget.ButtonOpts.TextColor(&widget.ButtonTextColor{
			Idle:    colornames.White,
			Hover:   colornames.White,
			Pressed: Mix(colornames.White, colornames.Black, 0.4),
		}),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:         DefaultNineSlice(colornames.Forestgreen),
			Hover:        DefaultNineSlice(Mix(colornames.Forestgreen, colornames.White, 0.2)),
			Pressed:      PressedNineSlice(Mix(colornames.Forestgreen, colornames.Black, 0.4)),
			PressedHover: PressedNineSlice(Mix(colornames.Forestgreen, colornames.Black, 0.4)),
		}),
		widget.ButtonOpts.TextLabel("Resume"),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			s.context.ChangeState(s.context.(*App).GetState(StateRomRunning))
		}),
	)
	pauseMenu.AddChild(resumeButton)

	// Add main menu button
	menuButton := widget.NewButton(
		widget.ButtonOpts.TextFace(DefaultFont()),
		widget.ButtonOpts.TextColor(&widget.ButtonTextColor{
			Idle:    colornames.White,
			Hover:   colornames.White,
			Pressed: Mix(colornames.White, colornames.Black, 0.4),
		}),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:         DefaultNineSlice(colornames.Darkslategray),
			Hover:        DefaultNineSlice(Mix(colornames.Darkslategray, colornames.White, 0.2)),
			Pressed:      PressedNineSlice(Mix(colornames.Darkslategray, colornames.Black, 0.4)),
			PressedHover: PressedNineSlice(Mix(colornames.Darkslategray, colornames.Black, 0.4)),
		}),
		widget.ButtonOpts.TextLabel("Main Menu"),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			s.context.ChangeState(s.context.(*App).GetState(StateRomList))
		}),
	)
	pauseMenu.AddChild(menuButton)

	overlay.AddChild(pauseMenu)
	s.root.AddChild(overlay)

	s.ui = &ebitenui.UI{Container: s.root}
}

// Enter implements State interface
func (s *PausedState) Enter(prevState State) {
	// Nothing special to do on enter
}

// Exit implements State interface
func (s *PausedState) Exit(nextState State) {
	// Nothing special to do on exit
}

// Update implements State interface
func (s *PausedState) Update() {
	// Handle Escape key to resume
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.context.ChangeState(s.context.(*App).GetState(StateRomRunning))
	}

	s.ui.Update()
}

// Draw implements State interface
func (s *PausedState) Draw(screen *ebiten.Image) {
	// First draw the game frame in the background
	select {
	case frame := <-s.context.(*App).GetFrameChannel():
		frameImg := ImageFromFrame(frame)
		w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

		op := &ebiten.DrawImageOptions{}
		fw, fh := frameImg.Bounds().Dx(), frameImg.Bounds().Dy()
		scaleX := float64(w) / float64(fw)
		scaleY := float64(h) / float64(fh)
		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY
		}

		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate((float64(w)-float64(fw)*scale)/2, (float64(h)-float64(fh)*scale)/2)
		screen.DrawImage(frameImg, op)
	default:
		// No frame available
	}

	// Then draw the UI overlay
	s.ui.Draw(screen)
}

// ID implements State interface
func (s *PausedState) ID() StateID {
	return StateRomPaused
}
