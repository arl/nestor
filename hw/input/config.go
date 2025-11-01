package input

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type PaddleButton byte

const (
	PadA PaddleButton = iota
	PadB
	PadSelect
	PadStart
	PadUp
	PadDown
	PadLeft
	PadRight

	PadButtonCount
)

func (pd PaddleButton) String() string {
	var buttonNames = [PadButtonCount]string{
		"A", "B",
		"Select", "Start",
		"Up", "Down", "Left", "Right",
	}
	return buttonNames[pd]
}

const numPresets = 8

type Config struct {
	Paddles [2]PaddleConfig          `toml:"paddles"`
	Presets [numPresets]PaddlePreset `toml:"presets"`
}

func (cfg *Config) PostLoad() {
	if cfg.Paddles[0].PaddlePreset >= numPresets {
		cfg.Paddles[0].PaddlePreset = 0
	}
	if cfg.Paddles[1].PaddlePreset >= numPresets {
		cfg.Paddles[1].PaddlePreset = 0
	}
	cfg.Paddles[0].Preset = &cfg.Presets[cfg.Paddles[0].PaddlePreset]
	cfg.Paddles[1].Preset = &cfg.Presets[cfg.Paddles[1].PaddlePreset]
}

type PaddleConfig struct {
	Plugged      bool          `toml:"plugged"`
	PaddlePreset uint          `toml:"preset"`
	Preset       *PaddlePreset `toml:"-"` // TODO: remove
}

type PaddlePreset struct {
	Keyboard *KeyboardMapping `toml:"keyboard,omitempty"`
	Gamepad  *GamepadMapping  `toml:"gamepad,omitempty"`
}

const UnsetKey = ebiten.Key(-1)

var unsetKeyboardMapping = KeyboardMapping{
	A:      UnsetKey,
	B:      UnsetKey,
	Select: UnsetKey,
	Start:  UnsetKey,
	Up:     UnsetKey,
	Down:   UnsetKey,
	Left:   UnsetKey,
	Right:  UnsetKey,
}

func ptrTo[T any](v T) *T {
	return &v
}

func (pp *PaddlePreset) UnmarshalTOML(data any) error {
	d, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid data format for 'input.presets'")
	}

	switch {
	case d["keyboard"] != nil:
		pp.Keyboard = ptrTo(unsetKeyboardMapping)
		return pp.Keyboard.UnmarshalTOML(d["keyboard"])
	case d["gamepad"] != nil:
		pp.Gamepad = &GamepadMapping{}
		return pp.Gamepad.UnmarshalTOML(d["gamepad"])
	default:
		pp.Keyboard = nil
		pp.Gamepad = nil
	}

	return nil
}

func (p *PaddlePreset) ToButtons() [PadButtonCount]Code {
	if p.Keyboard != nil {
		return [PadButtonCount]Code{
			Code{Type: Keyboard, Scancode: p.Keyboard.A},
			Code{Type: Keyboard, Scancode: p.Keyboard.B},
			Code{Type: Keyboard, Scancode: p.Keyboard.Select},
			Code{Type: Keyboard, Scancode: p.Keyboard.Start},
			Code{Type: Keyboard, Scancode: p.Keyboard.Up},
			Code{Type: Keyboard, Scancode: p.Keyboard.Down},
			Code{Type: Keyboard, Scancode: p.Keyboard.Left},
			Code{Type: Keyboard, Scancode: p.Keyboard.Right},
		}
	}

	if p.Gamepad != nil {
		return [PadButtonCount]Code{
			Code{Type: PadButton, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.A},
			Code{Type: PadButton, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.B},
			Code{Type: PadButton, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Select},
			Code{Type: PadButton, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Start},
			Code{Type: PadButton, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Up},
			Code{Type: PadButton, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Down},
			Code{Type: PadButton, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Left},
			Code{Type: PadButton, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Right},
		}
	}

	// Unset mappings
	return [PadButtonCount]Code{}
}

func (p *PaddlePreset) AssignCode(btn PaddleButton, code Code) {
	switch code.Type {
	case Keyboard:
		if p.Keyboard == nil {
			p.Keyboard = &KeyboardMapping{}
		}
		p.Keyboard.assign(btn, code)

	case PadButton:
		if p.Gamepad == nil {
			p.Gamepad = &GamepadMapping{}
			p.Gamepad.GamepadSDLID = code.GamepadSDLID
		} else {
			// do not mix gamepad ids so start anew if different
			if p.Gamepad.GamepadSDLID != code.GamepadSDLID {
				p.Gamepad = &GamepadMapping{}
				p.Gamepad.GamepadSDLID = code.GamepadSDLID
			}
		}
		p.Gamepad.assign(btn, code)
	case UnsetType:
		// unassign mapping.
		switch {
		case p.Keyboard != nil:
			p.Keyboard.assign(btn, code)
		case p.Gamepad != nil:
			p.Gamepad.assign(btn, code)
		}
	}
}

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

func (km *GamepadMapping) UnmarshalTOML(data any) error {
	d, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid data format for 'input.presets.gamepad'")
	}

	fmt.Println(d)
	return nil
}

func (km *GamepadMapping) assign(btn PaddleButton, code Code) {
	switch btn {
	case PadA:
		km.A = code.GamepadButton
	case PadB:
		km.B = code.GamepadButton
	case PadSelect:
		km.Select = code.GamepadButton
	case PadStart:
		km.Start = code.GamepadButton
	case PadUp:
		km.Up = code.GamepadButton
	case PadDown:
		km.Down = code.GamepadButton
	case PadLeft:
		km.Left = code.GamepadButton
	case PadRight:
		km.Right = code.GamepadButton
	}
}
