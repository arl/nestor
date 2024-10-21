package hw

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"

	"nestor/hw/hwio"
)

// A PaddleButton is one of the button of a standard NES controller/paddle.
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
	switch pd {
	case PadA:
		return "A"
	case PadB:
		return "B"
	case PadSelect:
		return "Select"
	case PadStart:
		return "Start"
	case PadUp:
		return "Up"
	case PadDown:
		return "Down"
	case PadLeft:
		return "Left"
	case PadRight:
		return "Right"
	}
	panic(fmt.Sprintf("unknown paddle button %d", pd))
}

// PaddlePreset holds the mapping configuration of a paddle.
type PaddlePreset struct {
	A      string `toml:"a"`
	B      string `toml:"b"`
	Select string `toml:"select"`
	Start  string `toml:"start"`
	Up     string `toml:"up"`
	Down   string `toml:"down"`
	Left   string `toml:"left"`
	Right  string `toml:"right"`
}

// SetMapping defines the mapping for a PaddleButton.
func (cfg *PaddlePreset) SetMapping(b PaddleButton, val string) {
	switch b {
	case PadA:
		cfg.A = val
	case PadB:
		cfg.B = val
	case PadSelect:
		cfg.Select = val
	case PadStart:
		cfg.Start = val
	case PadUp:
		cfg.Up = val
	case PadDown:
		cfg.Down = val
	case PadLeft:
		cfg.Left = val
	case PadRight:
		cfg.Right = val
	default:
		panic(fmt.Sprintf("unknown paddle button %d", b))
	}
}

// GetMapping returns the mapping for a PaddleButton.
func (cfg *PaddlePreset) GetMapping(pd PaddleButton) string {
	switch pd {
	case PadA:
		return cfg.A
	case PadB:
		return cfg.B
	case PadSelect:
		return cfg.Select
	case PadStart:
		return cfg.Start
	case PadUp:
		return cfg.Up
	case PadDown:
		return cfg.Down
	case PadLeft:
		return cfg.Left
	case PadRight:
		return cfg.Right
	default:
		panic(fmt.Sprintf("unknown paddle button %d", pd))
	}
}

func (cfg PaddlePreset) keycodes() ([8]sdl.Scancode, error) {
	var codes [8]sdl.Scancode
	codes[PadA] = sdl.GetScancodeFromName(cfg.A)
	codes[PadB] = sdl.GetScancodeFromName(cfg.B)
	codes[PadSelect] = sdl.GetScancodeFromName(cfg.Select)
	codes[PadStart] = sdl.GetScancodeFromName(cfg.Start)
	codes[PadUp] = sdl.GetScancodeFromName(cfg.Up)
	codes[PadDown] = sdl.GetScancodeFromName(cfg.Down)
	codes[PadLeft] = sdl.GetScancodeFromName(cfg.Left)
	codes[PadRight] = sdl.GetScancodeFromName(cfg.Right)

	for btn, c := range codes {
		if c == sdl.SCANCODE_UNKNOWN {
			pbtn := PaddleButton(btn)
			return codes, fmt.Errorf("unrecognized key for button %s: %q", pbtn, cfg.GetMapping(pbtn))
		}
	}

	return codes, nil
}

const numPresets = 8

type InputConfig struct {
	Paddles [2]PaddleConfig          `toml:"paddles"`
	Presets [numPresets]PaddlePreset `toml:"presets"`
}

func (cfg *InputConfig) Init() {
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

type InputProvider struct {
	keys     [2][8]sdl.Scancode
	keystate []uint8

	cfg InputConfig
}

func NewInputProvider(cfg InputConfig) (*InputProvider, error) {
	up := &InputProvider{cfg: cfg}
	sdl.Do(func() {
		up.keystate = sdl.GetKeyboardState()
	})

	var err error
	if cfg.Paddles[0].Plugged {
		if up.keys[0], err = cfg.Paddles[0].Preset.keycodes(); err != nil {
			return nil, fmt.Errorf("pad1: %s", err)
		}
	}
	if cfg.Paddles[1].Plugged {
		if up.keys[1], err = cfg.Paddles[1].Preset.keycodes(); err != nil {
			return nil, fmt.Errorf("pad2: %s", err)
		}
	}
	return up, nil
}

func (ui *InputProvider) paddleState(idx int) uint8 {
	padcfg := ui.cfg.Paddles[idx]
	if !padcfg.Plugged {
		// TODO: check this
		return 0
	}

	state := uint8(0)
	for btn, code := range ui.keys[idx] {
		if ui.keystate[code] != 0 {
			state |= 1 << uint(btn)
		}
	}
	return state
}

func (ui *InputProvider) LoadState() (uint8, uint8) {
	return ui.paddleState(0), ui.paddleState(1)
}

// InputPorts handles I/O with an InputDevice (such as standard NES controller
// for example).
type InputPorts struct {
	In  hwio.Reg8 `hwio:"offset=0x16,rcb,wcb"`
	Out hwio.Reg8 `hwio:"offset=0x17,rcb"`

	// XXX: this is just to pass nestest.nes test diff,
	// while we don't have an APU.
	Stub1 hwio.Reg8 `hwio:"offset=0x04"`
	Stub2 hwio.Reg8 `hwio:"offset=0x05"`
	Stub3 hwio.Reg8 `hwio:"offset=0x06"`
	Stub4 hwio.Reg8 `hwio:"offset=0x07"`

	provider *InputProvider // nil if no input device is connected.

	prevStrobe, strobe bool     // to observe strobe falling edge.
	state              [2]uint8 // state shift registers.
}

func (ip *InputPorts) initBus() {
	hwio.MustInitRegs(ip)

	// XXX: this is just to pass nestest.nes test diff,
	// while we don't have an APU.
	ip.Stub1.Value = 0x40
	ip.Stub2.Value = 0x40
	ip.Stub3.Value = 0x40
	ip.Stub4.Value = 0x40
}

func (ip *InputPorts) regval(port uint8) uint8 {
	ret := ip.state[port] & 1
	ip.state[port] >>= 1

	// After 8 bits are read, all subsequent bits will report 1 on a standard
	// NES controller, but third party and other controllers may report other
	// values here
	ip.state[port] |= 0x80

	// Emulate open bus behavior.
	return 0x40 | ret
}

// like regval but without side effects.
func (ip *InputPorts) regvalPeek(port uint8) uint8 {
	ret := ip.state[port] & 1

	// Emulate open bus behavior.
	return 0x40 | ret
}

// capture state of all connected input devices.
func (ip *InputPorts) loadstate() {
	if ip.provider == nil {
		// No controller is connected.
		// TODO: check this
		ip.state[0] = 0x40
		ip.state[1] = 0x40
		return
	}

	ip.state[0], ip.state[1] = ip.provider.LoadState()
}

// In: $4016
func (ip *InputPorts) WriteIN(old, val uint8) {
	ip.prevStrobe = ip.strobe
	ip.strobe = val&1 == 1
	if ip.prevStrobe && !ip.strobe {
		ip.loadstate()
	}
}

func (ip *InputPorts) ReadIN(_ uint8, peek bool) uint8 {
	if peek {
		return ip.regvalPeek(0)
	}
	if ip.strobe {
		ip.loadstate()
	}
	return ip.regval(0)
}

// Out: $4017
func (ip *InputPorts) ReadOUT(_ uint8, peek bool) uint8 {
	if peek {
		return ip.regvalPeek(1)
	}
	if ip.strobe {
		ip.loadstate()
	}

	return ip.regval(1)
}
