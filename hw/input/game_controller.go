package input

import (
	"fmt"
	"strings"

	"nestor/emu/log"

	"github.com/veandco/go-sdl2/sdl"
)

var Gamectrls *GameControllers

type ControllerType uint8

const (
	UnsetController ControllerType = iota
	Keyboard
	ControllerButton
	ControllerAxis
)

func (t ControllerType) String() string {
	switch t {
	case Keyboard:
		return "key"
	case ControllerButton:
		return "joy button"
	case ControllerAxis:
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

	Type ControllerType
}

// Name returns an user-friendly name for the input code.
func (mc Code) Name() string {
	switch mc.Type {
	case Keyboard:
		return sdl.GetScancodeName(mc.Scancode)
	case ControllerButton:
		return sdl.GameControllerGetStringForButton(mc.CtrlButton)
	case ControllerAxis:
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
	case Keyboard:
		s = fmt.Sprintf("key %s", name)
	case ControllerButton:
		s = fmt.Sprintf("joybtn %s %s", name, mc.CtrlGUID)
	case ControllerAxis:
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
		mc.Type = ControllerButton

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
		mc.Type = ControllerAxis

	case strings.HasPrefix(s, "key"):
		str := ""
		if _, err := fmt.Sscanf(s, "key %s", &str); err != nil {
			return fmt.Errorf("malformed key code: %s", s)
		}

		mc.Scancode = sdl.GetScancodeFromName(str)
		if mc.Scancode == sdl.SCANCODE_UNKNOWN {
			return fmt.Errorf("unrecognized scancode %q", s)
		}
		mc.Type = Keyboard

	default:
		return fmt.Errorf("unrecognized input code: %s", s)
	}

	return nil
}

// threshold for joystick axis to be considered as 'pressed'.
// goes from -32768 to 32767
const JoyAxisThreshold = 32000

type GameControllers struct {
	Guids map[string]*sdl.GameController         // GUID -> controller
	Ids   map[sdl.JoystickID]*sdl.GameController // joystick ID -> controller
}

// As soon as it's been created, update must be called for each controller event
// in order to remain in sync.
func NewGameControllers() *GameControllers {
	gcs := GameControllers{
		Guids: make(map[string]*sdl.GameController),
		Ids:   make(map[sdl.JoystickID]*sdl.GameController),
	}
	for i := range sdl.NumJoysticks() {
		if sdl.IsGameController(i) {
			c := sdl.GameControllerOpen(i)
			joy := c.Joystick()
			guid := sdl.JoystickGetGUIDString(joy.GUID())
			gcs.Guids[guid] = c
			id := joy.InstanceID()
			gcs.Ids[id] = c

			log.ModInput.DebugZ("found controller").
				Int32("id", int32(id)).
				String("guid", guid).
				End()
		}
	}
	return &gcs
}

// returns -1 for [-32768, 0) and 1 for [0, 32767]
func axissign(v int16) int16 {
	return int16(1 - 2*(uint16(v)>>15))
}

func (gcs *GameControllers) Get(id sdl.JoystickID) *sdl.GameController {
	return gcs.Ids[id]
}

func (gcs *GameControllers) GetGUID(id sdl.JoystickID) string {
	gc := gcs.Get(id)
	guid := sdl.JoystickGetGUIDString(gc.Joystick().GUID())
	return guid
}

func (gcs *GameControllers) getByGUID(guid string) *sdl.GameController {
	return gcs.Guids[guid]
}

func (gcs *GameControllers) UpdateDevices(e sdl.ControllerDeviceEvent) {
	switch e.Type {
	case sdl.CONTROLLERDEVICEADDED:
		c := sdl.GameControllerOpen(int(e.Which))
		guid := sdl.JoystickGetGUIDString(c.Joystick().GUID())
		id := c.Joystick().InstanceID()
		gcs.Guids[guid] = c
		gcs.Ids[id] = c

		log.ModInput.InfoZ("added controller").
			Int32("id", int32(id)).
			String("guid", guid).
			End()

	case sdl.CONTROLLERDEVICEREMOVED:
		c := gcs.Get(e.Which)
		if c == nil {
			log.ModInput.FatalZ("controller not found").
				Int32("id", int32(e.Which)).
				End()
		}
		guid := sdl.JoystickGetGUIDString(c.Joystick().GUID())
		delete(gcs.Guids, guid)
		delete(gcs.Ids, e.Which)
		c.Close()

		log.ModInput.InfoZ("removed controller").
			Int32("id", int32(e.Which)).
			String("guid", guid).
			End()
	}
}

func (gcs *GameControllers) Close() {
	for _, c := range gcs.Guids {
		c.Close()
		c = nil
	}
	clear(gcs.Guids)
}
