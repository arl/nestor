package ui

import (
	"errors"

	"nestor/config"
	"nestor/emu"
	"nestor/ui/input"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sqweek/dialog"
)

type actionRegistry struct {
	handlers map[input.Action]func()
}

func newActionRegistry() *actionRegistry {
	return &actionRegistry{
		handlers: make(map[input.Action]func()),
	}
}

func (r *actionRegistry) register(action input.Action, handler func()) {
	r.handlers[action] = handler
}

func (r *actionRegistry) trigger(action input.Action) {
	if handler, ok := r.handlers[action]; ok {
		handler()
	}
}

func (app *app) registerActions() {
	app.actions.register(input.ActionToggleFullscreen, app.toggleFullscreen)
	app.actions.register(input.ActionQuit, app.exit)
	app.actions.register(input.ActionOpenROM, app.openROMDialog)
	app.actions.register(input.ActionSettingsOpenGeneralConfig, func() {
		app.setState("config", configPageDest("general"))
	})
	app.actions.register(input.ActionSettingsOpenVideoConfig, func() {
		app.setState("config", configPageDest("video"))
	})
	app.actions.register(input.ActionSettingsOpenInputConfig, func() {
		app.setState("config", configPageDest("input"))
	})
	app.actions.register(input.ActionSettingsOpenEmulationConfig, func() {
		app.setState("config", configPageDest("emulation"))
	})

	app.actions.register(input.ActionPauseEmulator, func() { app.setState("paused", nil) })
	app.actions.register(input.ActionResumeEmulator, app.resumeEmulator)
	app.actions.register(input.ActionResetEmulator, app.resetEmulator)
	app.actions.register(input.ActionToggleShaderUI, app.toggleShaderUI)

	for i := range 8 {
		slot := i
		app.actions.register(input.ActionSaveSavestateSlot1+input.Action(slot), func() {
			app.savestateToSlot(slot)
		})
		app.actions.register(input.ActionLoadSavestateSlot1+input.Action(slot), func() {
			app.loadstateFromSlot(slot)
		})
	}
}

func (app *app) toggleFullscreen() {
	enable := !ebiten.IsFullscreen()
	ebiten.SetFullscreen(enable)

	if enable {
		app.displayWidth, app.displayHeight = ebiten.Monitor().Size()
	} else {
		app.displayWidth, app.displayHeight = ebiten.WindowSize()
	}
	app.curstate.createUI()
}

func (app *app) openROMDialog() {
	dlg := dialog.File().Title("Open NES ROM").Filter("NES rom", "nes")
	dlg.StartDir = app.cfg.General.FileLoadStartDir

	name, err := dlg.Load()
	if err != nil {
		if !errors.Is(err, dialog.ErrCancelled) {
			modUI.ErrorZ("dialog: failed to open").Error("err", err).End()
			errorWindow(&app.ui, err)
		}
		return
	}

	if err := app.runRom(name, nil); err != nil {
		modUI.ErrorZ("failed to run rom").Error("err", err).End()
		errorWindow(&app.ui, err)
	}
}

func (app *app) resumeEmulator() {
	if app.emulator == nil {
		return
	}
	app.emulator.Unblock()
	app.setState("running", nil)
	app.audioPlayer.Play()
}

func (app *app) resetEmulator() {
	if app.emulator == nil {
		return
	}
	app.emulator.Reset()
}

func (app *app) toggleShaderUI() {
	if rs, ok := app.curstate.(*runningState); ok {
		rs.shaderuiOn = !rs.shaderuiOn
	}
}

func (app *app) isStatePaused() bool {
	_, ok := app.curstate.(*pausedState)
	return ok
}

func (app *app) loadstateFromSlot(slot int) {
	// TODO: implement savestate loading from slot
	modUI.InfoZ("load state from slot").Int("slot", slot+1).End()
}

func (app *app) savestateToSlot(slot int) {
	var (
		state emu.ExecState
		err   error
	)

	save := func() {
		if err != nil {
			modUI.ErrorZ("failed to save state").Int("slot", slot+1).Error("err", err).End()
			return
		}
		config.AddSavestate(app.romName(), slot, state.SaveState)

		modUI.InfoZ("saved state").Int("slot", slot+1).End()
		_ = state
	}

	if app.currentStateName() == "paused" {
		state, err = app.emulator.SavestateUnsafe()
		save()
	} else {
		// Savestate is a blocking action. We must do it in a goroutine to avoid
		// blocking the interaction between the emulator loop and the UI.
		go func() { state, err = app.emulator.Savestate(); save() }()
	}
}

