package ui

import (
	"fmt"

	"github.com/ebitenui/ebitenui"

	"nestor/config"
	"nestor/ui/input"
	"nestor/ui/menu"
)

const numSavestateSlots = 8

// Menu item IDs for items without actions
const (
	MenuIDLoadState = "load_state"
	MenuIDSaveState = "save_state"
)

type appMenu struct {
	*menu.Bar[input.Action]
}

func newAppMenu(ui *ebitenui.UI, actions *actionRegistry, currentROMName string) *appMenu {
	def := buildMenuDefinition(currentROMName)

	style := menu.Style{
		Font:               res.fonts.face,
		BackgroundColor:    hex2color(0x000000FF),
		MenuBackground:     hex2color(menuBackground),
		TextColorIdle:      hex2color(0xFFFFFFFF),
		TextColorDisabled:  hex2color(0x808080FF),
		TextColorHover:     hex2color(0xFFFFFFFF),
		TextColorPressed:   hex2color(0x000000FF),
		ButtonHoverColor:   hex2color(0x404040FF),
		ButtonPressedColor: hex2color(0xFFFFFFFF),
	}

	bar := menu.New(ui, def, style, func(action input.Action) {
		actions.trigger(action)
	})

	return &appMenu{Bar: bar}
}

func buildMenuDefinition(currentROMName string) menu.Definition[input.Action] {
	slotLabelFunc := func(slot int) func() string {
		return func() string {
			t := config.SavestateInfo(currentROMName, slot)
			if t.IsZero() {
				return fmt.Sprintf("<empty> — %d", slot+1)
			}
			return fmt.Sprintf("%s — %d", t.Format("06/01/02 15:04:05"), slot+1)
		}
	}

	loadStateItems := make([]menu.Item[input.Action], numSavestateSlots)
	for i := range loadStateItems {
		loadStateItems[i] = menu.Item[input.Action]{
			LabelFunc: slotLabelFunc(i),
			Action:    input.ActionLoadSavestateSlot1 + input.Action(i),
		}
	}

	saveStateItems := make([]menu.Item[input.Action], numSavestateSlots)
	for i := range saveStateItems {
		saveStateItems[i] = menu.Item[input.Action]{
			LabelFunc: slotLabelFunc(i),
			Action:    input.ActionSaveSavestateSlot1 + input.Action(i),
		}
	}

	return menu.Definition[input.Action]{
		Menus: []menu.Menu[input.Action]{
			{
				Label: "File",
				Items: []menu.Item[input.Action]{
					{Label: "Open ROM ...", Action: input.ActionOpenROM},
					{Label: "Load State", ID: MenuIDLoadState, SubMenu: loadStateItems},
					{Label: "Save State", ID: MenuIDSaveState, SubMenu: saveStateItems},
					{Label: "Quit", Action: input.ActionQuit},
				},
			},
			{
				Label: "Settings",
				Items: []menu.Item[input.Action]{
					{Label: "General", Action: input.ActionSettingsOpenGeneralConfig},
					{Label: "Input", Action: input.ActionSettingsOpenInputConfig},
					{Label: "Video", Action: input.ActionSettingsOpenVideoConfig},
					{Label: "Emulation", Action: input.ActionSettingsOpenEmulationConfig},
				},
			},
			{
				Label: "Help",
				Items: []menu.Item[input.Action]{},
			},
		},
	}
}
