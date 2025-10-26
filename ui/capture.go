package ui

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type captureState struct {
	*app
}

func newCaptureState(app *app) *captureState {
	state := &captureState{
		app: app,
	}

	state.createUI()
	return state
}

func (s *captureState) enter(args ...any) {}
func (s *captureState) exit()             {}

func (s *captureState) update() {
	keys := inpututil.AppendJustPressedKeys(nil)

	if len(keys) > 0 {
		fmt.Println("keys:", keys)
		s.app.setState("config", keys[0])
	}
	// inpututil.AppendJustPressedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton)

	s.ui.Update()
}

func (s *captureState) draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

func (s *captureState) createUI() {
	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	root.AddChild(widget.NewLabel(
		widget.LabelOpts.Text("Capture Mode - Press any key or button", res.fonts.boldFace, &widget.LabelColor{Idle: hex2color(0xFFFFFF)}),
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			})))))

	s.ui.Container = root
}
