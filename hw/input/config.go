package input

import (
	"fmt"
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
		pp.Keyboard = ptrTo(newKeyboardMapping())
		return pp.Keyboard.UnmarshalTOML(d["keyboard"])
	case d["gamepad"] != nil:
		pp.Gamepad = ptrTo(newKGamepadMapping())
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
			Code{Type: Joystick, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.A},
			Code{Type: Joystick, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.B},
			Code{Type: Joystick, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Select},
			Code{Type: Joystick, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Start},
			Code{Type: Joystick, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Up},
			Code{Type: Joystick, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Down},
			Code{Type: Joystick, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Left},
			Code{Type: Joystick, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Right},
		}
	}

	// Unset mappings
	return [PadButtonCount]Code{}
}

// AssignCode maps input code for the given paddle button. On a given preset, we
// don't mix keyboard and gamepad mappings, nor mapping for different gamepad
// IDs, so assigning a code of a different type will reset existing mappings.
func (p *PaddlePreset) AssignCode(btn PaddleButton, code Code) {
	// First pass: if type is different, reset existing mapping.
	switch code.Type {
	case Keyboard:
		p.Gamepad = nil

		if p.Keyboard == nil {
			p.Keyboard = ptrTo(newKeyboardMapping())
		}

	case Joystick:
		p.Keyboard = nil

		if p.Gamepad == nil || p.Gamepad.GamepadSDLID != code.GamepadSDLID {
			p.Gamepad = ptrTo(newKGamepadMapping())
			p.Gamepad.GamepadSDLID = code.GamepadSDLID
		}
	}

	// Second pass: assign the code.
	switch code.Type {
	case Keyboard:
		p.Keyboard.assign(btn, code)
	case Joystick:
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
