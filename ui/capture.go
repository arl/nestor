package ui

import (
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
	if keys := inpututil.AppendJustPressedKeys(nil); len(keys) > 0 {
		if keys[0] == ebiten.KeyEscape {
			s.app.setState("config", s.btn, s.idxpreset, nil)
			return
		}

		modUI.InfoZ("key pressed").
			Int("code", int(keys[0])).
			String("key", keys[0].String()).
			End()

		code := input.Code{
			Type:     input.Keyboard,
			Scancode: keys[0],
		}
		s.app.setState("config", s.btn, s.idxpreset, &code)

		return
	}

	// pad button press.
	s.gamepads = ebiten.AppendGamepadIDs(s.gamepads[:0])
	for _, id := range s.gamepads {
		padbuttons := inpututil.AppendJustPressedGamepadButtons(id, nil)
		if len(padbuttons) > 0 {
			padbtn := padbuttons[len(padbuttons)-1]
			sdlid := ebiten.GamepadSDLID(id)

			modUI.InfoZ("gamepad button pressed").
				Int("id", int(id)).
				String("sdlid", sdlid).
				Int("button", int(padbtn)).
				End()

			code := input.Code{
				Type:          input.PadButton,
				GamepadSDLID:  sdlid,
				GamepadButton: padbtn,
			}
			s.app.setState("config", s.btn, s.idxpreset, &code)

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
