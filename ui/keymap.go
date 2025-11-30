package ui

import (
	einput "github.com/quasilyte/ebitengine-input"

	"nestor/config"
	"nestor/emu/log"
)

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
	actionPauseEmulator  // pause and show menu
	actionResetEmulator  // reset the NES
	actionToggleShaderUI // toggle shader selection UI

	// Paused state actions
	actionResumeEmulator // resume emulation
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

var actionRegistry = map[string]struct {
	action einput.Action
	keymap einput.Keymap
}{
	"global.toggle_fullscreen":            {actionToggleFullscreen, globalKeymap},
	"menu.file_open_rom":                  {actionFileOpenROM, menuKeymap},
	"menu.file_quit":                      {actionFileQuit, menuKeymap},
	"menu.settings_open_video_config":     {actionSettingsOpenVideoConfig, menuKeymap},
	"menu.settings_open_input_config":     {actionSettingsOpenInputConfig, menuKeymap},
	"menu.settings_open_emulation_config": {actionSettingsOpenEmulationConfig, menuKeymap},
	"running.pause_emulator":              {actionPauseEmulator, runningKeymap},
	"running.reset_emulator":              {actionResetEmulator, runningKeymap},
	"running.toggle_shader_ui":            {actionToggleShaderUI, runningKeymap},
	"paused.resume_emulator":              {actionResumeEmulator, pausedKeymap},
}

func loadKeymaps(gcfg config.General) {
	for name, keyStr := range gcfg.KeyboardShortcuts {
		reg, ok := actionRegistry[name]
		if !ok {
			log.ModEmu.Warnf("unknown keyboard shortcut action: %s", name)
			continue
		}

		k, err := einput.ParseKey(keyStr)
		if err != nil {
			log.ModEmu.Warnf("failed to parse key %q for action %s: %v", keyStr, name, err)
			continue
		}

		reg.keymap[reg.action] = []einput.Key{k}
	}
}
