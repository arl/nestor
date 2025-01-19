package input

import "github.com/veandco/go-sdl2/sdl"

// A PaddleButton identifies a button of a standard NES controller/paddle.
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

// PaddlePreset holds the mapping configuration of a paddle.
type PaddlePreset struct {
	Buttons [PadButtonCount]Code `toml:"buttons"`
}

const numPresets = 8

type Config struct {
	Paddles [2]PaddleConfig          `toml:"paddles"`
	Presets [numPresets]PaddlePreset `toml:"presets"`
}

func (cfg *Config) Init() {
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
	Preset       *PaddlePreset `toml:"-"` // points to the current preset
}

type Provider struct {
	keys     [2][8]sdl.Scancode
	keystate []uint8

	cfg Config
}

func NewProvider(cfg Config) *Provider {
	var keystate []uint8
	sdl.Do(func() { keystate = sdl.GetKeyboardState() })
	return &Provider{keystate: keystate, cfg: cfg}
}

func (ui *Provider) paddleState(idx int) uint8 {
	padcfg := ui.cfg.Paddles[idx]
	if !padcfg.Plugged {
		// TODO: check this
		return 0
	}

	preset := ui.cfg.Paddles[idx].Preset

	state := uint8(0)
	for i, code := range preset.Buttons {
		pressed := uint8(0)
		switch code.Type {
		case KeyboardCtrl:
			pressed = ui.keystate[code.Scancode]
		case ButtonCtrl:
			ctrl := Gamectrls.getByGUID(code.CtrlGUID)
			if ctrl != nil {
				pressed = ctrl.Button(code.CtrlButton)
			}
		case AxisCtrl:
			ctrl := Gamectrls.getByGUID(code.CtrlGUID)
			if ctrl != nil {
				if ctrl.Axis(code.CtrlAxis) >= JoyAxisThreshold {
					pressed = 1
				}
			}
		}
		state |= pressed << i
	}
	return state
}

func (ui *Provider) LoadState() (uint8, uint8) {
	return ui.paddleState(0), ui.paddleState(1)
}
