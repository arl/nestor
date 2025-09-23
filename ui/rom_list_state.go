package ui

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
)

// RomListState represents the state for browsing and selecting ROMs
type RomListState struct {
	context StateContext
	ui      *ebitenui.UI
	root    *widget.Container
}

// NewRomListState creates a new ROM list state
func NewRomListState(context StateContext) *RomListState {
	state := &RomListState{
		context: context,
	}

	// Initialize UI
	state.initUI()

	return state
}

// initUI creates the UI for this state
func (s *RomListState) initUI() {
	// Create root container
	s.root = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Gainsboro),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	// Create center panel with buttons - change to grid layout
	centerPanel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Darkgray),
		),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
		// Change to grid layout with single column and proper spacing
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Spacing(0, 20),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Top:    20,
				Bottom: 20,
			}),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, false}),
		)),
	)

	// Create start button - update to use grid layout data
	startButton := widget.NewButton(
		widget.ButtonOpts.TextFace(DefaultFont()),
		widget.ButtonOpts.TextColor(&widget.ButtonTextColor{
			Idle:    colornames.Gainsboro,
			Hover:   colornames.Gainsboro,
			Pressed: Mix(colornames.Gainsboro, colornames.Black, 0.4),
		}),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:         DefaultNineSlice(colornames.Darkslategray),
			Hover:        DefaultNineSlice(Mix(colornames.Darkslategray, colornames.Mediumseagreen, 0.4)),
			Disabled:     DefaultNineSlice(Mix(colornames.Darkslategray, colornames.Gainsboro, 0.8)),
			Pressed:      PressedNineSlice(Mix(colornames.Darkslategray, colornames.Black, 0.4)),
			PressedHover: PressedNineSlice(Mix(colornames.Darkslategray, colornames.Black, 0.4)),
		}),
		widget.ButtonOpts.TextLabel("Start ROM"),
		widget.ButtonOpts.WidgetOpts(
			// Update to use grid layout data
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(180, 48),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			const romPath = "/home/aurelien/dev/roms/nes/all.nes.roms.goodnes/USA/Super Mario Bros. + Duck Hunt (U) [!].nes"
			if err := s.context.LaunchEmulator(romPath); err == nil {
				s.context.ChangeState(s.context.(*App).GetState(StateRomRunning))
			}
		}),
	)

	// Create config button - update to use grid layout data
	configButton := widget.NewButton(
		widget.ButtonOpts.TextFace(DefaultFont()),
		widget.ButtonOpts.TextColor(&widget.ButtonTextColor{
			Idle:    colornames.Gainsboro,
			Hover:   colornames.Gainsboro,
			Pressed: Mix(colornames.Gainsboro, colornames.Black, 0.4),
		}),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:         DefaultNineSlice(colornames.Darkslategray),
			Hover:        DefaultNineSlice(Mix(colornames.Darkslategray, colornames.Dodgerblue, 0.4)),
			Disabled:     DefaultNineSlice(Mix(colornames.Darkslategray, colornames.Gainsboro, 0.8)),
			Pressed:      PressedNineSlice(Mix(colornames.Darkslategray, colornames.Black, 0.4)),
			PressedHover: PressedNineSlice(Mix(colornames.Darkslategray, colornames.Black, 0.4)),
		}),
		widget.ButtonOpts.TextLabel("Configuration"),
		widget.ButtonOpts.WidgetOpts(
			// Update to use grid layout data
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(180, 48),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			s.context.ChangeState(s.context.(*App).GetState(StateConfig))
		}),
	)

	centerPanel.AddChild(startButton)
	centerPanel.AddChild(configButton)

	s.root.AddChild(centerPanel)
	s.ui = &ebitenui.UI{Container: s.root}
}

// Enter implements State interface
func (s *RomListState) Enter(prevState State) {
	// Nothing to do on enter
}

// Exit implements State interface
func (s *RomListState) Exit(nextState State) {
	// Nothing to do on exit
}

// Update implements State interface
func (s *RomListState) Update() {
	s.ui.Update()
}

// Draw implements State interface
func (s *RomListState) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

// ID implements State interface
func (s *RomListState) ID() StateID {
	return StateRomList
}
