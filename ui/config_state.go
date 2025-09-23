package ui

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
)

// ConfigState represents the configuration screen
type ConfigState struct {
	context StateContext
	ui      *ebitenui.UI
	root    *widget.Container
}

// NewConfigState creates a new configuration state
func NewConfigState(context StateContext) *ConfigState {
	state := &ConfigState{
		context: context,
	}

	// Initialize UI
	state.initUI()

	return state
}

// initUI creates the UI for this state
func (s *ConfigState) initUI() {
	// Create root container
	s.root = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Gainsboro),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	// Create content panel
	contentPanel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Lightgray),
		),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Top:    20,
				Left:   20,
				Right:  20,
				Bottom: 20,
			}),
			widget.GridLayoutOpts.Spacing(20, 20),
			widget.GridLayoutOpts.Columns(1),
		)),
	)

	// Add a title
	titleLabel := widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:           600,
				HorizontalPosition: widget.GridLayoutPositionCenter,
			}),
		)),
		widget.LabelOpts.Text("Configuration", TitleFont(), &widget.LabelColor{
			Idle:     colornames.Black,
			Disabled: Mix(colornames.Black, colornames.White, 0.4),
		}),
	)
	contentPanel.AddChild(titleLabel)

	// Add some sample configuration options
	videoCfg := widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth: 600,
			}),
		)),
		widget.LabelOpts.Text("Video Settings", DefaultFont(), &widget.LabelColor{
			Idle:     colornames.Black,
			Disabled: Mix(colornames.Black, colornames.White, 0.4),
		}),
	)
	contentPanel.AddChild(videoCfg)

	// Add back button
	backButton := widget.NewButton(
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
		widget.ButtonOpts.TextLabel("Back to Menu"),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:           200,
				HorizontalPosition: widget.GridLayoutPositionCenter,
			}),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			s.context.ChangeState(s.context.(*App).GetState(StateRomList))
		}),
	)
	contentPanel.AddChild(backButton)

	s.root.AddChild(contentPanel)
	s.ui = &ebitenui.UI{Container: s.root}
}

// Enter implements State interface
func (s *ConfigState) Enter(prevState State) {
	// Nothing to do on enter
}

// Exit implements State interface
func (s *ConfigState) Exit(nextState State) {
	// Nothing to do on exit
}

// Update implements State interface
func (s *ConfigState) Update() {
	s.ui.Update()
}

// Draw implements State interface
func (s *ConfigState) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

// ID implements State interface
func (s *ConfigState) ID() StateID {
	return StateConfig
}
