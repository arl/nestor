package input

type Action int

const (
	ActionNone Action = iota

	// Global actions (available in all states)
	ActionToggleFullscreen
	ActionQuit

	// Menu actions (main state)
	ActionOpenROM
	ActionSettingsOpenGeneralConfig
	ActionSettingsOpenVideoConfig
	ActionSettingsOpenInputConfig
	ActionSettingsOpenEmulationConfig

	// Running state actions
	ActionPauseEmulator
	ActionResetEmulator
	ActionToggleShaderUI
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
	ActionResumeEmulator
)

// Action string IDs for configuration
const (
	IDToggleFullscreen = "toggle_fullscreen"

	IDOpenROM                     = "open_rom"
	IDQuit                        = "quit"
	IDSettingsOpenGeneralConfig   = "settings_open_general_config"
	IDSettingsOpenVideoConfig     = "settings_open_video_config"
	IDSettingsOpenInputConfig     = "settings_open_input_config"
	IDSettingsOpenEmulationConfig = "settings_open_emulation_config"

	IDPauseEmulator  = "pause_emulator"
	IDResumeEmulator = "resume_emulator"
	IDResetEmulator  = "reset_emulator"

	IDToggleShaderUI     = "toggle_shader_ui"
	IDSaveSavestateSlot1 = "save_savestate_slot_1"
	IDSaveSavestateSlot2 = "save_savestate_slot_2"
	IDSaveSavestateSlot3 = "save_savestate_slot_3"
	IDSaveSavestateSlot4 = "save_savestate_slot_4"
	IDSaveSavestateSlot5 = "save_savestate_slot_5"
	IDSaveSavestateSlot6 = "save_savestate_slot_6"
	IDSaveSavestateSlot7 = "save_savestate_slot_7"
	IDSaveSavestateSlot8 = "save_savestate_slot_8"
	IDLoadSavestateSlot1 = "load_savestate_slot_1"
	IDLoadSavestateSlot2 = "load_savestate_slot_2"
	IDLoadSavestateSlot3 = "load_savestate_slot_3"
	IDLoadSavestateSlot4 = "load_savestate_slot_4"
	IDLoadSavestateSlot5 = "load_savestate_slot_5"
	IDLoadSavestateSlot6 = "load_savestate_slot_6"
	IDLoadSavestateSlot7 = "load_savestate_slot_7"
	IDLoadSavestateSlot8 = "load_savestate_slot_8"
)

var ActionByID = map[string]Action{
	IDToggleFullscreen: ActionToggleFullscreen,

	IDOpenROM:                     ActionOpenROM,
	IDQuit:                        ActionQuit,
	IDSettingsOpenGeneralConfig:   ActionSettingsOpenGeneralConfig,
	IDSettingsOpenVideoConfig:     ActionSettingsOpenVideoConfig,
	IDSettingsOpenInputConfig:     ActionSettingsOpenInputConfig,
	IDSettingsOpenEmulationConfig: ActionSettingsOpenEmulationConfig,

	IDPauseEmulator:  ActionPauseEmulator,
	IDResetEmulator:  ActionResetEmulator,
	IDToggleShaderUI: ActionToggleShaderUI,

	IDLoadSavestateSlot1: ActionLoadSavestateSlot1,
	IDLoadSavestateSlot2: ActionLoadSavestateSlot2,
	IDLoadSavestateSlot3: ActionLoadSavestateSlot3,
	IDLoadSavestateSlot4: ActionLoadSavestateSlot4,
	IDLoadSavestateSlot5: ActionLoadSavestateSlot5,
	IDLoadSavestateSlot6: ActionLoadSavestateSlot6,
	IDLoadSavestateSlot7: ActionLoadSavestateSlot7,
	IDLoadSavestateSlot8: ActionLoadSavestateSlot8,
	IDSaveSavestateSlot1: ActionSaveSavestateSlot1,
	IDSaveSavestateSlot2: ActionSaveSavestateSlot2,
	IDSaveSavestateSlot3: ActionSaveSavestateSlot3,
	IDSaveSavestateSlot4: ActionSaveSavestateSlot4,
	IDSaveSavestateSlot5: ActionSaveSavestateSlot5,
	IDSaveSavestateSlot6: ActionSaveSavestateSlot6,
	IDSaveSavestateSlot7: ActionSaveSavestateSlot7,
	IDSaveSavestateSlot8: ActionSaveSavestateSlot8,

	IDResumeEmulator: ActionResumeEmulator,
}

