// Package input provides the input handling for NES controllers.
package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type EbitenInput struct {
	i0, i1 func() uint8
	cfg    Config
}

func NewEbitenInput(cfg Config) *EbitenInput {
	pr0 := cfg.Presets[cfg.Paddles[0].PaddlePreset]
	pr1 := cfg.Presets[cfg.Paddles[1].PaddlePreset]

	var i0, i1 func() uint8

	switch {
	case pr0.Keyboard != nil:
		i0 = kbstate(pr0.ToButtons())
	case pr0.Gamepad != nil:
		i0 = gamepadState(pr0.ToButtons())
	default:
		i0 = unplugged
	}

	switch {
	case pr1.Keyboard != nil:
		i1 = kbstate(pr1.ToButtons())
	case pr1.Gamepad != nil:
		i1 = gamepadState(pr1.ToButtons())
	default:
		i1 = unplugged
	}

	return &EbitenInput{
		cfg: cfg,
		i0:  i0,
		i1:  i1,
	}
}

func (ei *EbitenInput) LoadState() (uint8, uint8) {
	return ei.i0(), ei.i1()
}

func unplugged() uint8 { return 0 }

// compute input port state from keyboard.
func kbstate(codes [PadButtonCount]Code) func() uint8 {
	var scratch [256]ebiten.Key

	f := func() uint8 {
		keys := inpututil.AppendPressedKeys(scratch[:0])

		var state uint8
		for _, k := range keys {
			for i, code := range codes {
				if k == code.Scancode {
					state |= 1 << i
					break
				}
			}
		}

		return state
	}

	return f
}

// compute input port state from a gamepad.
func gamepadState(codes [PadButtonCount]Code) func() uint8 {
	var gpid ebiten.GamepadID = -1

	for _, id := range ebiten.AppendGamepadIDs(nil) {
		if ebiten.GamepadSDLID(id) == codes[0].GamepadSDLID {
			gpid = id
		}
	}

	if gpid == -1 {
		return unplugged
	}

	var scratch [256]ebiten.GamepadButton

	f := func() uint8 {
		btns := inpututil.AppendPressedGamepadButtons(gpid, scratch[:0])

		var state uint8
		for _, k := range btns {
			for i, code := range codes {
				if k == code.GamepadButton {
					state |= 1 << i
					break
				}
			}
		}

		return state
	}

	return f
}
