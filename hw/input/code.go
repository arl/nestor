package input

import (
	"fmt"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type ControlType uint8

const (
	ControlNotSet ControlType = iota
	KeyboardCtrl
	ButtonCtrl
	AxisCtrl
)

func (t ControlType) String() string {
	switch t {
	case KeyboardCtrl:
		return "key"
	case ButtonCtrl:
		return "joy button"
	case AxisCtrl:
		return "joy axis"
	}
	return "not set"
}

// A Code describes the user input event (keyboard key, game controller
// button/axis). Only one of these is valid.
type Code struct {
	Scancode sdl.Scancode

	CtrlGUID    string
	CtrlButton  sdl.GameControllerButton
	CtrlAxis    sdl.GameControllerAxis
	CtrlAxisDir int16

	Type ControlType
}

// Name returns an user-friendly name for the input code.
func (mc Code) Name() string {
	switch mc.Type {
	case KeyboardCtrl:
		return sdl.GetScancodeName(mc.Scancode)
	case ButtonCtrl:
		return sdl.GameControllerGetStringForButton(mc.CtrlButton)
	case AxisCtrl:
		axis := sdl.GameControllerGetStringForAxis(mc.CtrlAxis)
		if mc.CtrlAxisDir >= 0 {
			axis += "+"
		} else {
			axis += "-"
		}
		return axis
	}

	return ""
}

func (mc Code) MarshalText() ([]byte, error) {
	s := ""
	name := mc.Name()
	switch mc.Type {
	case KeyboardCtrl:
		s = fmt.Sprintf("key %s", name)
	case ButtonCtrl:
		s = fmt.Sprintf("joybtn %s %s", name, mc.CtrlGUID)
	case AxisCtrl:
		s = fmt.Sprintf("joyaxis %s %s", name, mc.CtrlGUID)
	}

	return []byte(s), nil
}

func (mc *Code) UnmarshalText(text []byte) error {
	s := string(text)

	switch {
	case s == "":
		mc.Type = 0
	case strings.HasPrefix(s, "joybtn"):
		str := ""
		if _, err := fmt.Sscanf(s, "joybtn %s %s", &str, &mc.CtrlGUID); err != nil {
			return fmt.Errorf("malformed joybtn code: %s", s)
		}
		mc.CtrlButton = sdl.GameControllerGetButtonFromString(str)
		if mc.CtrlButton == sdl.CONTROLLER_BUTTON_INVALID {
			return fmt.Errorf("unrecognized button %q", str)
		}
		mc.Type = ButtonCtrl

	case strings.HasPrefix(s, "joyaxis"):
		str := ""
		dir := ""
		if _, err := fmt.Sscanf(s, "joyaxis %s %s", &str, &mc.CtrlGUID); err != nil {
			return fmt.Errorf("malformed joyaxis code: %s", s)
		}
		switch {
		case strings.HasSuffix(str, "+"):
			mc.CtrlAxisDir = 1
		case strings.HasSuffix(str, "-"):
			mc.CtrlAxisDir = -1
		default:
			return fmt.Errorf("malformed axis direction: %s", dir)
		}

		mc.CtrlAxis = sdl.GameControllerGetAxisFromString(str[:len(str)-1])
		if mc.CtrlAxis == sdl.CONTROLLER_AXIS_INVALID {
			return fmt.Errorf("unrecognized axis %q", str)
		}
		mc.Type = AxisCtrl

	case strings.HasPrefix(s, "key"):
		str := ""
		if _, err := fmt.Sscanf(s, "key %s", &str); err != nil {
			return fmt.Errorf("malformed key code: %s", s)
		}

		mc.Scancode = sdl.GetScancodeFromName(str)
		if mc.Scancode == sdl.SCANCODE_UNKNOWN {
			return fmt.Errorf("unrecognized scancode %q", s)
		}
		mc.Type = KeyboardCtrl

	default:
		return fmt.Errorf("unrecognized input code: %s", s)
	}

	return nil
}
