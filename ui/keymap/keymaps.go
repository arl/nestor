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
	idToggleFullscreen = "global.toggle_fullscreen"

	idOpenROM                     = "open_rom"
	idQuit                        = "quit"
	idSettingsOpenGeneralConfig   = "settings_open_general_config"
	idSettingsOpenVideoConfig     = "settings_open_video_config"
	idSettingsOpenInputConfig     = "settings_open_input_config"
	idSettingsOpenEmulationConfig = "settings_open_emulation_config"

	idPauseEmulator  = "pause_emulator"
	idResumeEmulator = "resume_emulator"
	idResetEmulator  = "reset_emulator"

	idToggleShaderUI     = "toggle_shader_ui"
	idSaveSavestateSlot1 = "save_savestate_slot_1"
	idSaveSavestateSlot2 = "save_savestate_slot_2"
	idSaveSavestateSlot3 = "save_savestate_slot_3"
	idSaveSavestateSlot4 = "save_savestate_slot_4"
	idSaveSavestateSlot5 = "save_savestate_slot_5"
	idSaveSavestateSlot6 = "save_savestate_slot_6"
	idSaveSavestateSlot7 = "save_savestate_slot_7"
	idSaveSavestateSlot8 = "save_savestate_slot_8"
	idLoadSavestateSlot1 = "load_savestate_slot_1"
	idLoadSavestateSlot2 = "load_savestate_slot_2"
	idLoadSavestateSlot3 = "load_savestate_slot_3"
	idLoadSavestateSlot4 = "load_savestate_slot_4"
	idLoadSavestateSlot5 = "load_savestate_slot_5"
	idLoadSavestateSlot6 = "load_savestate_slot_6"
	idLoadSavestateSlot7 = "load_savestate_slot_7"
	idLoadSavestateSlot8 = "load_savestate_slot_8"
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
	idToggleFullscreen: "f11",

	idOpenROM: "ctrl+o",
	idQuit:    "ctrl+q",

	idSettingsOpenGeneralConfig:   "ctrl+g",
	idSettingsOpenVideoConfig:     "ctrl+v",
	idSettingsOpenInputConfig:     "ctrl+i",
	idSettingsOpenEmulationConfig: "ctrl+e",

	idPauseEmulator:      "space",
	idResetEmulator:      "ctrl+r",
	idToggleShaderUI:     "f9",
	idSaveSavestateSlot1: "f1",
	idSaveSavestateSlot2: "f2",
	idSaveSavestateSlot3: "f3",
	idSaveSavestateSlot4: "f4",
	idSaveSavestateSlot5: "f5",
	idSaveSavestateSlot6: "f6",
	idSaveSavestateSlot7: "f7",
	idSaveSavestateSlot8: "f8",
	idLoadSavestateSlot1: "shift+f1",
	idLoadSavestateSlot2: "shift+f2",
	idLoadSavestateSlot3: "shift+f3",
	idLoadSavestateSlot4: "shift+f4",
	idLoadSavestateSlot5: "shift+f5",
	idLoadSavestateSlot6: "shift+f6",
	idLoadSavestateSlot7: "shift+f7",
	idLoadSavestateSlot8: "shift+f8",

	idResumeEmulator: "space",
}

// Apply updates the keymaps based on the configuration.
func (s Shortcuts) Apply() {
	for name, key := range s {
		def, ok := actionRegistry[name]
		if !ok {
			log.ModEmu.Warnf("unknown keyboard shortcut action: %s", name)
			continue
		}

		k, err := einput.ParseKey(string(key))
		if err != nil {
			log.ModEmu.Warnf("failed to parse key %q for action %s: %v", key, name, err)
			continue
		}

		for _, km := range def.keymaps {
			km[def.action] = []einput.Key{k}
		}
	}
}

func init() {
	DefaultShortcuts.Apply()
}

type actionDef struct {
	action  einput.Action
	keymaps []einput.Keymap
}

func action(a einput.Action, keymaps ...einput.Keymap) actionDef {
	return actionDef{action: a, keymaps: keymaps}
}

var actionRegistry = map[string]actionDef{
	idToggleFullscreen: action(ActionToggleFullscreen, GlobalKeymap),

	idOpenROM:                     action(ActionFileOpenROM, MenuKeymap),
	idQuit:                        action(ActionFileQuit, MenuKeymap),
	idSettingsOpenGeneralConfig:   action(ActionSettingsOpenGeneralConfig, MenuKeymap),
	idSettingsOpenVideoConfig:     action(ActionSettingsOpenVideoConfig, MenuKeymap),
	idSettingsOpenInputConfig:     action(ActionSettingsOpenInputConfig, MenuKeymap),
	idSettingsOpenEmulationConfig: action(ActionSettingsOpenEmulationConfig, MenuKeymap),

	idPauseEmulator:  action(ActionPauseEmulator, RunningKeymap),
	idResetEmulator:  action(ActionResetEmulator, RunningKeymap),
	idToggleShaderUI: action(ActionToggleShaderUI, RunningKeymap),

	idLoadSavestateSlot1: action(ActionLoadSavestateSlot1, MenuKeymap, RunningKeymap),
	idLoadSavestateSlot2: action(ActionLoadSavestateSlot2, MenuKeymap, RunningKeymap),
	idLoadSavestateSlot3: action(ActionLoadSavestateSlot3, MenuKeymap, RunningKeymap),
	idLoadSavestateSlot4: action(ActionLoadSavestateSlot4, MenuKeymap, RunningKeymap),
	idLoadSavestateSlot5: action(ActionLoadSavestateSlot5, MenuKeymap, RunningKeymap),
	idLoadSavestateSlot6: action(ActionLoadSavestateSlot6, MenuKeymap, RunningKeymap),
	idLoadSavestateSlot7: action(ActionLoadSavestateSlot7, MenuKeymap, RunningKeymap),
	idLoadSavestateSlot8: action(ActionLoadSavestateSlot8, MenuKeymap, RunningKeymap),
	idSaveSavestateSlot1: action(ActionSaveSavestateSlot1, RunningKeymap),
	idSaveSavestateSlot2: action(ActionSaveSavestateSlot2, RunningKeymap),
	idSaveSavestateSlot3: action(ActionSaveSavestateSlot3, RunningKeymap),
	idSaveSavestateSlot4: action(ActionSaveSavestateSlot4, RunningKeymap),
	idSaveSavestateSlot5: action(ActionSaveSavestateSlot5, RunningKeymap),
	idSaveSavestateSlot6: action(ActionSaveSavestateSlot6, RunningKeymap),
	idSaveSavestateSlot7: action(ActionSaveSavestateSlot7, RunningKeymap),
	idSaveSavestateSlot8: action(ActionSaveSavestateSlot8, RunningKeymap),

	idResumeEmulator: action(ActionResumeEmulator, PausedKeymap),
}
