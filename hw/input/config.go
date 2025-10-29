package input

import (
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

type PaddlePreset struct {
	Keyboard *KeyboardMapping `toml:"keyboard,omitempty"`
	Gamepad  *GamepadMapping  `toml:"gamepad,omitempty"`
}

func (p *PaddlePreset) ToButtons() [PadButtonCount]Code {
	var btns [PadButtonCount]Code

	if p.Keyboard != nil {
		btns[PadA] = Code{Type: KeyboardCtrl, Scancode: p.Keyboard.A}
		btns[PadB] = Code{Type: KeyboardCtrl, Scancode: p.Keyboard.B}
		btns[PadSelect] = Code{Type: KeyboardCtrl, Scancode: p.Keyboard.Select}
		btns[PadStart] = Code{Type: KeyboardCtrl, Scancode: p.Keyboard.Start}
		btns[PadUp] = Code{Type: KeyboardCtrl, Scancode: p.Keyboard.Up}
		btns[PadDown] = Code{Type: KeyboardCtrl, Scancode: p.Keyboard.Down}
		btns[PadLeft] = Code{Type: KeyboardCtrl, Scancode: p.Keyboard.Left}
		btns[PadRight] = Code{Type: KeyboardCtrl, Scancode: p.Keyboard.Right}
	}

	if p.Gamepad != nil {
		btns[PadA] = Code{Type: ButtonCtrl, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.A}
		btns[PadB] = Code{Type: ButtonCtrl, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.B}
		btns[PadSelect] = Code{Type: ButtonCtrl, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Select}
		btns[PadStart] = Code{Type: ButtonCtrl, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Start}
		btns[PadUp] = Code{Type: ButtonCtrl, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Up}
		btns[PadDown] = Code{Type: ButtonCtrl, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Down}
		btns[PadLeft] = Code{Type: ButtonCtrl, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Left}
		btns[PadRight] = Code{Type: ButtonCtrl, GamepadSDLID: p.Gamepad.GamepadSDLID, GamepadButton: p.Gamepad.Right}
	}

	return btns
}

func (p *PaddlePreset) AssignCode(btn PaddleButton, code Code) {
	switch code.Type {
	case KeyboardCtrl:
		if p.Keyboard == nil {
			p.Keyboard = &KeyboardMapping{}
		}
		p.Keyboard.assign(btn, code)

	case ButtonCtrl:
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
	}
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
	Preset       *PaddlePreset `toml:"-"`
}
