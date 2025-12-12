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
	ActionLoadSavestateSlot1
	ActionLoadSavestateSlot2
	ActionLoadSavestateSlot3
	ActionLoadSavestateSlot4
	ActionLoadSavestateSlot5
	ActionLoadSavestateSlot6
	ActionLoadSavestateSlot7
	ActionLoadSavestateSlot8
	ActionSaveSavestateSlot1
	ActionSaveSavestateSlot2
	ActionSaveSavestateSlot3
	ActionSaveSavestateSlot4
	ActionSaveSavestateSlot5
	ActionSaveSavestateSlot6
	ActionSaveSavestateSlot7
	ActionSaveSavestateSlot8

	// Paused state actions
	ActionResumeEmulator // resume emulation
)

// Action String IDs
const (
	idGlobalToggleFullscreen = "global.toggle_fullscreen"

	idMenuFileOpenROM                 = "menu.file_open_rom"
	idMenuFileQuit                    = "menu.file_quit"
	idMenuSettingsOpenGeneralConfig   = "menu.settings_open_general_config"
	idMenuSettingsOpenVideoConfig     = "menu.settings_open_video_config"
	idMenuSettingsOpenInputConfig     = "menu.settings_open_input_config"
	idMenuSettingsOpenEmulationConfig = "menu.settings_open_emulation_config"

	idRunningPauseEmulator      = "running.pause_emulator"
	idRunningResetEmulator      = "running.reset_emulator"
	idRunningToggleShaderUI     = "running.toggle_shader_ui"
	idRunningSaveSavestateSlot1 = "running.save_savestate_slot_1"
	idRunningSaveSavestateSlot2 = "running.save_savestate_slot_2"
	idRunningSaveSavestateSlot3 = "running.save_savestate_slot_3"
	idRunningSaveSavestateSlot4 = "running.save_savestate_slot_4"
	idRunningSaveSavestateSlot5 = "running.save_savestate_slot_5"
	idRunningSaveSavestateSlot6 = "running.save_savestate_slot_6"
	idRunningSaveSavestateSlot7 = "running.save_savestate_slot_7"
	idRunningSaveSavestateSlot8 = "running.save_savestate_slot_8"
	idRunningLoadSavestateSlot1 = "running.load_savestate_slot_1"
	idRunningLoadSavestateSlot2 = "running.load_savestate_slot_2"
	idRunningLoadSavestateSlot3 = "running.load_savestate_slot_3"
	idRunningLoadSavestateSlot4 = "running.load_savestate_slot_4"
	idRunningLoadSavestateSlot5 = "running.load_savestate_slot_5"
	idRunningLoadSavestateSlot6 = "running.load_savestate_slot_6"
	idRunningLoadSavestateSlot7 = "running.load_savestate_slot_7"
	idRunningLoadSavestateSlot8 = "running.load_savestate_slot_8"

	idPausedResumeEmulator = "paused.resume_emulator"
)

var (
	GlobalKeymap  = einput.Keymap{}
	MenuKeymap    = einput.Keymap{}
	RunningKeymap = einput.Keymap{}
	PausedKeymap  = einput.Keymap{}
)

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
	idGlobalToggleFullscreen: "f11",

	idMenuFileOpenROM: "ctrl+o",
	idMenuFileQuit:    "ctrl+q",

	idMenuSettingsOpenGeneralConfig:   "ctrl+g",
	idMenuSettingsOpenVideoConfig:     "ctrl+v",
	idMenuSettingsOpenInputConfig:     "ctrl+i",
	idMenuSettingsOpenEmulationConfig: "ctrl+e",

	idRunningPauseEmulator:      "escape",
	idRunningResetEmulator:      "r",
	idRunningToggleShaderUI:     "f9",
	idRunningSaveSavestateSlot1: "f1",
	idRunningSaveSavestateSlot2: "f2",
	idRunningSaveSavestateSlot3: "f3",
	idRunningSaveSavestateSlot4: "f4",
	idRunningSaveSavestateSlot5: "f5",
	idRunningSaveSavestateSlot6: "f6",
	idRunningSaveSavestateSlot7: "f7",
	idRunningSaveSavestateSlot8: "f8",
	idRunningLoadSavestateSlot1: "shift+f1",
	idRunningLoadSavestateSlot2: "shift+f2",
	idRunningLoadSavestateSlot3: "shift+f3",
	idRunningLoadSavestateSlot4: "shift+f4",
	idRunningLoadSavestateSlot5: "shift+f5",
	idRunningLoadSavestateSlot6: "shift+f6",
	idRunningLoadSavestateSlot7: "shift+f7",
	idRunningLoadSavestateSlot8: "shift+f8",

	idPausedResumeEmulator: "escape",
}

// Apply updates the keymaps based on the configuration.
func (s Shortcuts) Apply() {
	for name, key := range s {
		reg, ok := actionRegistry[name]
		if !ok {
			log.ModEmu.Warnf("unknown keyboard shortcut action: %s", name)
			continue
		}

		k, err := einput.ParseKey(string(key))
		if err != nil {
			log.ModEmu.Warnf("failed to parse key %q for action %s: %v", key, name, err)
			continue
		}

		reg.keymap[reg.action] = []einput.Key{k}
	}
}

func init() {
	DefaultShortcuts.Apply()
}

var actionRegistry = map[string]struct {
	action einput.Action
	keymap einput.Keymap
}{
	idGlobalToggleFullscreen: {ActionToggleFullscreen, GlobalKeymap},

	idMenuFileOpenROM:                 {ActionFileOpenROM, MenuKeymap},
	idMenuFileQuit:                    {ActionFileQuit, MenuKeymap},
	idMenuSettingsOpenGeneralConfig:   {ActionSettingsOpenGeneralConfig, MenuKeymap},
	idMenuSettingsOpenVideoConfig:     {ActionSettingsOpenVideoConfig, MenuKeymap},
	idMenuSettingsOpenInputConfig:     {ActionSettingsOpenInputConfig, MenuKeymap},
	idMenuSettingsOpenEmulationConfig: {ActionSettingsOpenEmulationConfig, MenuKeymap},

	idRunningPauseEmulator:  {ActionPauseEmulator, RunningKeymap},
	idRunningResetEmulator:  {ActionResetEmulator, RunningKeymap},
	idRunningToggleShaderUI: {ActionToggleShaderUI, RunningKeymap},

	idRunningLoadSavestateSlot1: {ActionLoadSavestateSlot1, RunningKeymap},
	idRunningLoadSavestateSlot2: {ActionLoadSavestateSlot2, RunningKeymap},
	idRunningLoadSavestateSlot3: {ActionLoadSavestateSlot3, RunningKeymap},
	idRunningLoadSavestateSlot4: {ActionLoadSavestateSlot4, RunningKeymap},
	idRunningLoadSavestateSlot5: {ActionLoadSavestateSlot5, RunningKeymap},
	idRunningLoadSavestateSlot6: {ActionLoadSavestateSlot6, RunningKeymap},
	idRunningLoadSavestateSlot7: {ActionLoadSavestateSlot7, RunningKeymap},
	idRunningLoadSavestateSlot8: {ActionLoadSavestateSlot8, RunningKeymap},
	idRunningSaveSavestateSlot1: {ActionSaveSavestateSlot1, RunningKeymap},
	idRunningSaveSavestateSlot2: {ActionSaveSavestateSlot2, RunningKeymap},
	idRunningSaveSavestateSlot3: {ActionSaveSavestateSlot3, RunningKeymap},
	idRunningSaveSavestateSlot4: {ActionSaveSavestateSlot4, RunningKeymap},
	idRunningSaveSavestateSlot5: {ActionSaveSavestateSlot5, RunningKeymap},
	idRunningSaveSavestateSlot6: {ActionSaveSavestateSlot6, RunningKeymap},
	idRunningSaveSavestateSlot7: {ActionSaveSavestateSlot7, RunningKeymap},
	idRunningSaveSavestateSlot8: {ActionSaveSavestateSlot8, RunningKeymap},

	idPausedResumeEmulator: {ActionResumeEmulator, PausedKeymap},
}
