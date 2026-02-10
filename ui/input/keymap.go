package input

import (
	"fmt"

	"github.com/arl/nestor/emu/log"
)

var modInput = log.NewModule("INPUT")

type Keymap map[Action]KeyCombo

// JustPressed returns the action that was just pressed, or ActionNone if none
func (km Keymap) JustPressed() Action {
	for action, combo := range km {
		if combo.IsJustPressed() {
			return action
		}
	}
	return ActionNone
}

// Merge returns a new keymap that combines this keymap with others.
// Later keymaps override earlier ones for the same action.
func (km Keymap) Merge(others ...Keymap) Keymap {
	result := make(Keymap)
	for action, combo := range km {
		result[action] = combo
	}
	for _, other := range others {
		for action, combo := range other {
			result[action] = combo
		}
	}
	return result
}

type Shortcut string

func (s *Shortcut) UnmarshalText(text []byte) error {
	str := string(text)
	if _, err := Parse(str); err != nil {
		return fmt.Errorf("invalid key %q: %w", str, err)
	}
	*s = Shortcut(str)
	return nil
}

// Shortcuts defines the keyboard shortcuts configuration.
// Keys are action string IDs and values are key combination strings.
type Shortcuts map[string]Shortcut

// DefaultShortcuts defines the default key bindings.
var DefaultShortcuts = Shortcuts{
	IDToggleFullscreen: "f11",

	IDOpenROM: "ctrl+o",
	IDQuit:    "ctrl+a",

	IDSettingsOpenGeneralConfig:   "ctrl+g",
	IDSettingsOpenVideoConfig:     "ctrl+v",
	IDSettingsOpenInputConfig:     "ctrl+i",
	IDSettingsOpenEmulationConfig: "ctrl+e",

	IDPauseEmulator:      "space",
	IDResetEmulator:      "ctrl+r",
	IDToggleShaderUI:     "f9",
	IDSaveSavestateSlot1: "f1",
	IDSaveSavestateSlot2: "f2",
	IDSaveSavestateSlot3: "f3",
	IDSaveSavestateSlot4: "f4",
	IDSaveSavestateSlot5: "f5",
	IDSaveSavestateSlot6: "f6",
	IDSaveSavestateSlot7: "f7",
	IDSaveSavestateSlot8: "f8",
	IDLoadSavestateSlot1: "shift+f1",
	IDLoadSavestateSlot2: "shift+f2",
	IDLoadSavestateSlot3: "shift+f3",
	IDLoadSavestateSlot4: "shift+f4",
	IDLoadSavestateSlot5: "shift+f5",
	IDLoadSavestateSlot6: "shift+f6",
	IDLoadSavestateSlot7: "shift+f7",
	IDLoadSavestateSlot8: "shift+f8",

	IDResumeEmulator: "space",
}

var (
	GlobalKeymap  = Keymap{}
	MenuKeymap    = Keymap{}
	RunningKeymap = Keymap{}
	PausedKeymap  = Keymap{}
)

type actionDef struct {
	action  Action
	keymaps []Keymap
}

func actiondef(a Action, keymaps ...Keymap) actionDef {
	return actionDef{action: a, keymaps: keymaps}
}

var actionKeymaps = map[string]actionDef{
	IDToggleFullscreen: actiondef(ActionToggleFullscreen, GlobalKeymap),

	IDOpenROM:                     actiondef(ActionOpenROM, MenuKeymap),
	IDQuit:                        actiondef(ActionQuit, GlobalKeymap),
	IDSettingsOpenGeneralConfig:   actiondef(ActionSettingsOpenGeneralConfig, MenuKeymap),
	IDSettingsOpenVideoConfig:     actiondef(ActionSettingsOpenVideoConfig, MenuKeymap),
	IDSettingsOpenInputConfig:     actiondef(ActionSettingsOpenInputConfig, MenuKeymap),
	IDSettingsOpenEmulationConfig: actiondef(ActionSettingsOpenEmulationConfig, MenuKeymap),

	IDPauseEmulator:  actiondef(ActionPauseEmulator, RunningKeymap),
	IDResetEmulator:  actiondef(ActionResetEmulator, RunningKeymap),
	IDToggleShaderUI: actiondef(ActionToggleShaderUI, RunningKeymap),

	IDLoadSavestateSlot1: actiondef(ActionLoadSavestateSlot1, MenuKeymap, RunningKeymap),
	IDLoadSavestateSlot2: actiondef(ActionLoadSavestateSlot2, MenuKeymap, RunningKeymap),
	IDLoadSavestateSlot3: actiondef(ActionLoadSavestateSlot3, MenuKeymap, RunningKeymap),
	IDLoadSavestateSlot4: actiondef(ActionLoadSavestateSlot4, MenuKeymap, RunningKeymap),
	IDLoadSavestateSlot5: actiondef(ActionLoadSavestateSlot5, MenuKeymap, RunningKeymap),
	IDLoadSavestateSlot6: actiondef(ActionLoadSavestateSlot6, MenuKeymap, RunningKeymap),
	IDLoadSavestateSlot7: actiondef(ActionLoadSavestateSlot7, MenuKeymap, RunningKeymap),
	IDLoadSavestateSlot8: actiondef(ActionLoadSavestateSlot8, MenuKeymap, RunningKeymap),
	IDSaveSavestateSlot1: actiondef(ActionSaveSavestateSlot1, RunningKeymap, PausedKeymap),
	IDSaveSavestateSlot2: actiondef(ActionSaveSavestateSlot2, RunningKeymap, PausedKeymap),
	IDSaveSavestateSlot3: actiondef(ActionSaveSavestateSlot3, RunningKeymap, PausedKeymap),
	IDSaveSavestateSlot4: actiondef(ActionSaveSavestateSlot4, RunningKeymap, PausedKeymap),
	IDSaveSavestateSlot5: actiondef(ActionSaveSavestateSlot5, RunningKeymap, PausedKeymap),
	IDSaveSavestateSlot6: actiondef(ActionSaveSavestateSlot6, RunningKeymap, PausedKeymap),
	IDSaveSavestateSlot7: actiondef(ActionSaveSavestateSlot7, RunningKeymap, PausedKeymap),
	IDSaveSavestateSlot8: actiondef(ActionSaveSavestateSlot8, RunningKeymap, PausedKeymap),

	IDResumeEmulator: actiondef(ActionResumeEmulator, PausedKeymap),
}

// Apply updates the keymaps based on the shortcuts configuration.
func (s Shortcuts) Apply() {
	for name, key := range s {
		def, ok := actionKeymaps[name]
		if !ok {
			modInput.Warnf("unknown keyboard shortcut action: %s", name)
			continue
		}

		k, err := Parse(string(key))
		if err != nil {
			modInput.Warnf("failed to parse key %q for action %s: %v", key, name, err)
			continue
		}

		for _, km := range def.keymaps {
			km[def.action] = k
		}
	}
}

func init() {
	DefaultShortcuts.Apply()
}
