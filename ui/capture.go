package ui

import (
	"fmt"

	"nestor/hw/input"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type captureState struct {
	*app

	btn       input.PaddleButton // configured button
	idxpreset int                // configured preset
	gamepads  []ebiten.GamepadID
}

func newCaptureState(app *app) *captureState {
	state := &captureState{
		app: app,
	}

	state.createUI()
	return state
}

func (s *captureState) enter(args ...any) {
	s.btn = args[0].(input.PaddleButton)
	s.idxpreset = args[1].(int)

	modUI.InfoZ("Capture state entered").End()
}

func (s *captureState) exit() {}

func (s *captureState) update() {
	// key press.
	keys := inpututil.AppendJustPressedKeys(nil)
	if len(keys) > 0 {
		fmt.Println("keys:", keys)
		var code *input.Code
		if keys[0] != ebiten.KeyEscape {
			code = &input.Code{Type: input.KeyboardCtrl, Scancode: keys[0]}
		}
		s.app.setState("config", s.btn, s.idxpreset, code)
		return
	}

	// pad button press.
	numpads := len(s.gamepads)
	s.gamepads = inpututil.AppendJustConnectedGamepadIDs(s.gamepads)
	if len(s.gamepads) != numpads {
		id := s.gamepads[len(s.gamepads)-1]
		modUI.InfoZ("gamepad connected").
			Int("id", int(id)).
			String("name", ebiten.GamepadName(id)).
			End()

		fmt.Println("gamepad axis count:", ebiten.GamepadAxisCount(id))
	}

	for _, id := range s.gamepads {
		buttons := inpututil.AppendJustPressedGamepadButtons(id, nil)
		if len(buttons) > 0 {
			btn := buttons[len(buttons)-1]
			sdlid := ebiten.GamepadSDLID(id)

			modUI.InfoZ("gamepad button pressed").
				Int("gamepad_id", int(id)).
				String("gamepad_sdlid", sdlid).
				Int("button", int(btn)).
				End()
			s.app.setState("config", input.Code{Type: input.ButtonCtrl, GamepadSDLID: sdlid, GamepadButton: btn})
			return
		}
	}

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
