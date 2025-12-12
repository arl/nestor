package keymap

import (
	"fmt"

	einput "github.com/quasilyte/ebitengine-input"

	"nestor/emu/log"
)

// Action constants
const (
	// Global actions (available in all states)
	ActionToggleFullscreen einput.Action = iota + 1

	// Menu actions (main state)
	ActionFileOpenROM
	ActionFileQuit
	ActionSettingsOpenGeneralConfig
	ActionSettingsOpenVideoConfig
	ActionSettingsOpenInputConfig
	ActionSettingsOpenEmulationConfig

	// Running state actions
	ActionPauseEmulator  // pause and show menu
	ActionResetEmulator  // reset the NES
	ActionToggleShaderUI // toggle shader selection UI
	ActionLoadSaveStateSlot1
	ActionLoadSaveStateSlot2
	ActionLoadSaveStateSlot3
	ActionLoadSaveStateSlot4
	ActionLoadSaveStateSlot5
	ActionSaveSaveStateSlot6
	ActionSaveSaveStateSlot7
	ActionSaveSaveStateSlot8

	// Paused state actions
	ActionResumeEmulator // resume emulation
)

// Action String IDs
const (
	IDGlobalToggleFullscreen = "global.toggle_fullscreen"

	IDMenuFileOpenROM                 = "menu.file_open_rom"
	IDMenuFileQuit                    = "menu.file_quit"
	IDMenuSettingsOpenGeneralConfig   = "menu.settings_open_general_config"
	IDMenuSettingsOpenVideoConfig     = "menu.settings_open_video_config"
	IDMenuSettingsOpenInputConfig     = "menu.settings_open_input_config"
	IDMenuSettingsOpenEmulationConfig = "menu.settings_open_emulation_config"

	IDRunningPauseEmulator      = "running.pause_emulator"
	IDRunningResetEmulator      = "running.reset_emulator"
	IDRunningToggleShaderUI     = "running.toggle_shader_ui"
	IDRunningLoadSaveStateSlot1 = "running.load_save_state_slot_1"
	IDRunningLoadSaveStateSlot2 = "running.load_save_state_slot_2"
	IDRunningLoadSaveStateSlot3 = "running.load_save_state_slot_3"
	IDRunningLoadSaveStateSlot4 = "running.load_save_state_slot_4"
	IDRunningLoadSaveStateSlot5 = "running.load_save_state_slot_5"
	IDRunningSaveSaveStateSlot6 = "running.save_save_state_slot_6"
	IDRunningSaveSaveStateSlot7 = "running.save_save_state_slot_7"
	IDRunningSaveSaveStateSlot8 = "running.save_save_state_slot_8"

	IDPausedResumeEmulator = "paused.resume_emulator"
)

// Keymaps
var (
	GlobalKeymap = einput.Keymap{
		ActionToggleFullscreen: {einput.KeyF11},
	}

	MenuKeymap = einput.Keymap{
		ActionFileOpenROM:                 {einput.KeyWithModifier(einput.KeyO, einput.ModControl)},
		ActionFileQuit:                    {einput.KeyWithModifier(einput.KeyQ, einput.ModControl)},
		ActionSettingsOpenVideoConfig:     {einput.KeyWithModifier(einput.KeyV, einput.ModControl)},
		ActionSettingsOpenInputConfig:     {einput.KeyWithModifier(einput.KeyI, einput.ModControl)},
		ActionSettingsOpenEmulationConfig: {einput.KeyWithModifier(einput.KeyE, einput.ModControl)},
	}

	RunningKeymap = einput.Keymap{
		ActionPauseEmulator:      {einput.KeyEscape},
		ActionResetEmulator:      {einput.KeyR},
		ActionToggleShaderUI:     {einput.KeyF10},
		ActionLoadSaveStateSlot1: {einput.KeyF1},
		ActionLoadSaveStateSlot2: {einput.KeyF2},
		ActionLoadSaveStateSlot3: {einput.KeyF3},
		ActionLoadSaveStateSlot4: {einput.KeyF4},
		ActionLoadSaveStateSlot5: {einput.KeyF5},
		ActionSaveSaveStateSlot6: {einput.KeyF6},
		ActionSaveSaveStateSlot7: {einput.KeyF7},
		ActionSaveSaveStateSlot8: {einput.KeyF8},
	}

	PausedKeymap = einput.Keymap{
		ActionResumeEmulator: {einput.KeyEscape},
	}
)

var actionRegistry = map[string]struct {
	action einput.Action
	keymap einput.Keymap
}{
	IDGlobalToggleFullscreen: {ActionToggleFullscreen, GlobalKeymap},

	IDMenuFileOpenROM:                 {ActionFileOpenROM, MenuKeymap},
	IDMenuFileQuit:                    {ActionFileQuit, MenuKeymap},
	IDMenuSettingsOpenGeneralConfig:   {ActionSettingsOpenGeneralConfig, MenuKeymap},
	IDMenuSettingsOpenVideoConfig:     {ActionSettingsOpenVideoConfig, MenuKeymap},
	IDMenuSettingsOpenInputConfig:     {ActionSettingsOpenInputConfig, MenuKeymap},
	IDMenuSettingsOpenEmulationConfig: {ActionSettingsOpenEmulationConfig, MenuKeymap},

	IDRunningPauseEmulator:      {ActionPauseEmulator, RunningKeymap},
	IDRunningResetEmulator:      {ActionResetEmulator, RunningKeymap},
	IDRunningToggleShaderUI:     {ActionToggleShaderUI, RunningKeymap},
	IDRunningLoadSaveStateSlot1: {ActionLoadSaveStateSlot1, RunningKeymap},
	IDRunningLoadSaveStateSlot2: {ActionLoadSaveStateSlot2, RunningKeymap},
	IDRunningLoadSaveStateSlot3: {ActionLoadSaveStateSlot3, RunningKeymap},
	IDRunningLoadSaveStateSlot4: {ActionLoadSaveStateSlot4, RunningKeymap},
	IDRunningLoadSaveStateSlot5: {ActionLoadSaveStateSlot5, RunningKeymap},
	IDRunningSaveSaveStateSlot6: {ActionSaveSaveStateSlot6, RunningKeymap},
	IDRunningSaveSaveStateSlot7: {ActionSaveSaveStateSlot7, RunningKeymap},
	IDRunningSaveSaveStateSlot8: {ActionSaveSaveStateSlot8, RunningKeymap},

	IDPausedResumeEmulator: {ActionResumeEmulator, PausedKeymap},
}

type Shortcut string

func (s *Shortcut) UnmarshalText(text []byte) error {
	str := string(text)
	if _, err := einput.ParseKey(str); err != nil {
		return fmt.Errorf("invalid key %q: %w", str, err)
	}
	*s = Shortcut(str)
	return nil
}

// Shortcuts defines the keyboard shortcuts configuration. Keys are action
// string ids and values are ebiten-input key combination string.
type Shortcuts map[string]Shortcut

// DefaultShortcuts defines the default key bindings.
var DefaultShortcuts = Shortcuts{
	IDGlobalToggleFullscreen: "f11",

	IDMenuFileOpenROM: "ctrl+o",
	IDMenuFileQuit:    "ctrl+q",

	IDMenuSettingsOpenGeneralConfig:   "ctrl+g",
	IDMenuSettingsOpenVideoConfig:     "ctrl+v",
	IDMenuSettingsOpenInputConfig:     "ctrl+i",
	IDMenuSettingsOpenEmulationConfig: "ctrl+e",

	IDRunningPauseEmulator:      "escape",
	IDRunningResetEmulator:      "r",
	IDRunningToggleShaderUI:     "f5",
	IDRunningLoadSaveStateSlot1: "f1",
	IDRunningLoadSaveStateSlot2: "f2",
	IDRunningLoadSaveStateSlot3: "f3",
	IDRunningLoadSaveStateSlot4: "f4",
	IDRunningLoadSaveStateSlot5: "f5",
	IDRunningSaveSaveStateSlot6: "f6",
	IDRunningSaveSaveStateSlot7: "f7",
	IDRunningSaveSaveStateSlot8: "f8",

	IDPausedResumeEmulator: "escape",
}

// Apply updates the keymaps based on the configuration.
func (s Shortcuts) Apply() {
	for name, keyStr := range s {
		reg, ok := actionRegistry[name]
		if !ok {
			log.ModEmu.Warnf("unknown keyboard shortcut action: %s", name)
			continue
		}

		k, err := einput.ParseKey(string(keyStr))
		if err != nil {
			log.ModEmu.Warnf("failed to parse key %q for action %s: %v", keyStr, name, err)
			continue
		}

		reg.keymap[reg.action] = []einput.Key{k}
	}
}
