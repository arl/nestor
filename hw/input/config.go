package input

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
	Preset       *PaddlePreset `toml:"-"` // points to the current preset
}
