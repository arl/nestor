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

type captureArgs struct {
	btn       input.PaddleButton
	idxpreset int
}

func (s *captureState) enter(arg any) {
	cargs := arg.(captureArgs)
	s.btn = cargs.btn
	s.idxpreset = cargs.idxpreset

	modUI.InfoZ("Capture state entered").End()
}

func (s *captureState) exit() {}

func (s *captureState) update() {
	var code *input.Code

	// key press.
	if keys := inpututil.AppendJustPressedKeys(nil); len(keys) > 0 {
		modUI.InfoZ("key pressed").
			Int("code", int(keys[0])).
			String("key", keys[0].String()).
			End()

		if keys[0] != ebiten.KeyEscape {
			code = &input.Code{
				Type:     input.Keyboard,
				Scancode: keys[0],
			}
		} else {
			code = &input.Code{
				Type:     input.UnsetType,
				Scancode: input.UnsetKey,
			}
		}
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

			code = &input.Code{
				Type:          input.Joystick,
				GamepadSDLID:  sdlid,
				GamepadButton: padbtn,
			}

		}
	}

	if code != nil {
		s.assign(s.btn, s.idxpreset, *code)
		s.app.savecfg()
		s.app.setState("config", nil)
	}

	s.ui.Update()
}

func (s *captureState) assign(btn input.PaddleButton, idxpreset int, code input.Code) {
	s.app.cfg.Input.Presets[idxpreset].AssignCode(btn, code)
}

func (s *captureState) draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

func (s *captureState) createUI() {
	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	root.AddChild(widget.NewLabel(
		widget.LabelOpts.Text("Capture Mode - Press any key or button", res.fonts.boldFace, res.label.text),
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			})))))

	s.ui.Container = root
}
