package input

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

const UnsetKey = ebiten.Key(-1)

type KeyboardMapping struct {
	A      ebiten.Key `toml:"a"`
	B      ebiten.Key `toml:"b"`
	Select ebiten.Key `toml:"select"`
	Start  ebiten.Key `toml:"start"`
	Up     ebiten.Key `toml:"up"`
	Down   ebiten.Key `toml:"down"`
	Left   ebiten.Key `toml:"left"`
	Right  ebiten.Key `toml:"right"`
}

func newKeyboardMapping() KeyboardMapping {
	return KeyboardMapping{
		A:      UnsetKey,
		B:      UnsetKey,
		Select: UnsetKey,
		Start:  UnsetKey,
		Up:     UnsetKey,
		Down:   UnsetKey,
		Left:   UnsetKey,
		Right:  UnsetKey,
	}
}

func (kb *KeyboardMapping) UnmarshalTOML(data any) error {
	d, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid data format for 'input.presets.keyboard'")
	}

	setfield := func(field *ebiten.Key, v any) {
		if keycode, ok := v.(string); ok {
			*field = KeyFromString(keycode)
		}
	}

	for k, v := range d {
		switch k {
		case "a":
			setfield(&kb.A, v)
		case "b":
			setfield(&kb.B, v)
		case "select":
			setfield(&kb.Select, v)
		case "start":
			setfield(&kb.Start, v)
		case "up":
			setfield(&kb.Up, v)
		case "down":
			setfield(&kb.Down, v)
		case "left":
			setfield(&kb.Left, v)
		case "right":
			setfield(&kb.Right, v)
		}
	}

	return nil
}

func (km *KeyboardMapping) assign(btn PaddleButton, code Code) {
	switch btn {
	case PadA:
		km.A = code.Scancode
	case PadB:
		km.B = code.Scancode
	case PadSelect:
		km.Select = code.Scancode
	case PadStart:
		km.Start = code.Scancode
	case PadUp:
		km.Up = code.Scancode
	case PadDown:
		km.Down = code.Scancode
	case PadLeft:
		km.Left = code.Scancode
	case PadRight:
		km.Right = code.Scancode
	}
}

const UnsetGamepadButton = ebiten.GamepadButton(-1)

type GamepadMapping struct {
	GamepadSDLID string               `toml:"gamepad_sdlid"`
	A            ebiten.GamepadButton `toml:"a"`
	B            ebiten.GamepadButton `toml:"b"`
	Select       ebiten.GamepadButton `toml:"select"`
	Start        ebiten.GamepadButton `toml:"start"`
	Up           ebiten.GamepadButton `toml:"up"`
	Down         ebiten.GamepadButton `toml:"down"`
	Left         ebiten.GamepadButton `toml:"left"`
	Right        ebiten.GamepadButton `toml:"right"`
}

func newGamepadMapping() GamepadMapping {
	return GamepadMapping{
		GamepadSDLID: "",
		A:            UnsetGamepadButton,
		B:            UnsetGamepadButton,
		Select:       UnsetGamepadButton,
		Start:        UnsetGamepadButton,
		Up:           UnsetGamepadButton,
		Down:         UnsetGamepadButton,
		Left:         UnsetGamepadButton,
		Right:        UnsetGamepadButton,
	}
}

func (gm *GamepadMapping) UnmarshalTOML(data any) error {
	d, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid data format for 'input.presets.gamepad'")
	}

	setfield := func(field *ebiten.GamepadButton, v any) {
		if btncode, ok := v.(int64); ok {
			*field = ebiten.GamepadButton(btncode)
		}
	}

	sdlid, ok := d["gamepad_sdlid"].(string)
	if !ok || sdlid == "" {
		return nil
	}
	gm.GamepadSDLID = sdlid

	for k, v := range d {
		switch k {
		case "a":
			setfield(&gm.A, v)
		case "b":
			setfield(&gm.B, v)
		case "select":
			setfield(&gm.Select, v)
		case "start":
			setfield(&gm.Start, v)
		case "up":
			setfield(&gm.Up, v)
		case "down":
			setfield(&gm.Down, v)
		case "left":
			setfield(&gm.Left, v)
		case "right":
			setfield(&gm.Right, v)
		}
	}

	return nil
}

func (gm *GamepadMapping) assign(btn PaddleButton, code Code) {
	switch btn {
	case PadA:
		gm.A = code.GamepadButton
	case PadB:
		gm.B = code.GamepadButton
	case PadSelect:
		gm.Select = code.GamepadButton
	case PadStart:
		gm.Start = code.GamepadButton
	case PadUp:
		gm.Up = code.GamepadButton
	case PadDown:
		gm.Down = code.GamepadButton
	case PadLeft:
		gm.Left = code.GamepadButton
	case PadRight:
		gm.Right = code.GamepadButton
	}
}
