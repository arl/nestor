package input

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type Type uint8

const (
	UnsetType Type = iota
	Keyboard
	PadButton
)

// A Code describes the user input event (keyboard key, game controller
// button/axis). Only one of these is valid.
type Code struct {
	Scancode ebiten.Key

	GamepadSDLID  string
	GamepadButton ebiten.GamepadButton

	Type Type
}

// Name returns an user-friendly name for the input code.
func (mc Code) Name() string {
	switch mc.Type {
	case Keyboard:
		if mc.Scancode != unsetKey {
			return "[ " + mc.Scancode.String() + " ]"
		}
	case PadButton:
		return fmt.Sprintf("Paddle %d", mc.GamepadButton)
	case UnsetType:
		return "<not set>"
	default:
		panic(fmt.Sprintf("unexpected code type: %v", mc.Type))
	}

	return "<not set>"
}

// KeyFromString returns the ebiten.Key for s, or a negative value if s is
// unknown.
func KeyFromString(s string) ebiten.Key {
	var k ebiten.Key
	if err := k.UnmarshalText([]byte(s)); err != nil {
		return unsetKey
	}

	return k
}
