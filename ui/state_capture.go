package ui

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	einput "github.com/quasilyte/ebitengine-input"

	"nestor/hw/input"
	"nestor/ui/keymap"
)

type captureState struct {
	*app

	inputh   *einput.Handler
	args     *captureArgs
	scanner  einput.KeyScanner
	gamepads []ebiten.GamepadID
}

func newCaptureState(app *app) *captureState {
	state := &captureState{
		app: app,
	}

	state.createUI()
	return state
}

type captureMode string

const (
	captureModeUI  = "ui"
	captureModeNes = "nes"
)

type captureArgs struct {
	mode captureMode // "nes"|"ui"

	// "ui" for ui keyboard shortcut
	action string // ID of the ui action to modify

	// "nes" for nes controllers
	idxpreset int                // preset to modify
	btn       input.PaddleButton // nes pad button mapped
}

func (s *captureState) enter(inputh *einput.Handler, arg any) {
	// Disable input handler to prevent it from catching events
	// generated during capture.
	s.app.disableInputHandler()
	s.args = ptrTo(arg.(captureArgs))
	s.inputh = inputh

	modUI.InfoZ("Capture state entered").String("mode", string(s.args.mode)).End()
}

func (s *captureState) exit() {}

func (s *captureState) update() {
	switch s.args.mode {
	case captureModeUI:
		s.captureForUI()
	case captureModeNes:
		s.captureForNES()
	default:
		panic("unexpected capture mode " + s.args.mode)
	}

	s.ui.Update()
}

func (s *captureState) captureForUI() {
	k, status := s.scanner.Scan()
	if status != einput.KeyScanCompleted {
		return
	}

	if k != einput.KeyEscape {
		s.app.cfg.General.KeyboardShortcuts[s.args.action] = keymap.Shortcut(k.String())
		s.app.cfg.General.KeyboardShortcuts.Apply()
		s.app.savecfg()
	}

	s.app.setState("config", nil)
}

func (s *captureState) captureForNES() {
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
		s.app.cfg.Input.Presets[s.args.idxpreset].AssignCode(s.args.btn, *code)

		s.app.savecfg()
		s.app.setState("config", nil)
	}
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
