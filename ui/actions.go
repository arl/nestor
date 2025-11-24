package ui

import (
	einput "github.com/quasilyte/ebitengine-input"
)

// main menu actions
const (
	actionFileOpenROM einput.Action = iota + 1
	actionFileQuit
	actionSettingsOpenVideoConfig
	actionSettingsOpenInputConfig
	actionSettingsOpenEmulationConfig
)

// menuKeymap defines the non-configurable keyboard shortcuts for menu actions.
var menuKeymap = einput.Keymap{
	actionFileOpenROM:                 {einput.KeyWithModifier(einput.KeyO, einput.ModControl)},
	actionFileQuit:                    {einput.KeyWithModifier(einput.KeyQ, einput.ModControl)},
	actionSettingsOpenVideoConfig:     {einput.KeyWithModifier(einput.KeyV, einput.ModControl)},
	actionSettingsOpenInputConfig:     {einput.KeyWithModifier(einput.KeyI, einput.ModControl)},
	actionSettingsOpenEmulationConfig: {einput.KeyWithModifier(einput.KeyE, einput.ModControl)},
}
