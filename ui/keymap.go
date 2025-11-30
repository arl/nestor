package ui

import einput "github.com/quasilyte/ebitengine-input"

const (
	// Global actions (available in all states)
	actionToggleFullscreen einput.Action = iota + 1

	// Menu actions (main state)
	actionFileOpenROM
	actionFileQuit
	actionSettingsOpenVideoConfig
	actionSettingsOpenInputConfig
	actionSettingsOpenEmulationConfig

	// Running state actions
	actionPauseEmulator  // ESC - pause and show menu
	actionResetEmulator  // R - reset the NES
	actionToggleShaderUI // F5 - toggle shader selection UI

	// Paused state actions
	actionResumeEmulator // ESC - resume emulation
)

var globalKeymap = einput.Keymap{
	actionToggleFullscreen: {einput.KeyF11},
}

var menuKeymap = einput.Keymap{
	actionFileOpenROM:                 {einput.KeyWithModifier(einput.KeyO, einput.ModControl)},
	actionFileQuit:                    {einput.KeyWithModifier(einput.KeyQ, einput.ModControl)},
	actionSettingsOpenVideoConfig:     {einput.KeyWithModifier(einput.KeyV, einput.ModControl)},
	actionSettingsOpenInputConfig:     {einput.KeyWithModifier(einput.KeyI, einput.ModControl)},
	actionSettingsOpenEmulationConfig: {einput.KeyWithModifier(einput.KeyE, einput.ModControl)},
}

var runningKeymap = einput.Keymap{
	actionPauseEmulator:  {einput.KeyEscape},
	actionResetEmulator:  {einput.KeyR},
	actionToggleShaderUI: {einput.KeyF5},
}

var pausedKeymap = einput.Keymap{
	actionResumeEmulator: {einput.KeyEscape},
}
