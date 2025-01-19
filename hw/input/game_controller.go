package input

import (
	"nestor/emu/log"

	"github.com/veandco/go-sdl2/sdl"
)

var Gamectrls *GameControllers

// threshold for joystick axis to be considered as 'pressed'.
// goes from -32768 to 32767
const JoyAxisThreshold = 32000

// returns -1 for [-32768, 0) and 1 for [0, 32767]
func axissign(v int16) int16 {
	return int16(1 - 2*(uint16(v)>>15))
}

// GameControllers manages the state of SDL game controllers.
type GameControllers struct {
	Guids map[string]*sdl.GameController         // GUID -> controller
	Ids   map[sdl.JoystickID]*sdl.GameController // joystick ID -> controller
}

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
