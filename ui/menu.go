package ui

import (
	"fmt"
	"time"

	"github.com/ebitenui/ebitenui"

	"nestor/config"
	"nestor/ui/input"
	"nestor/ui/menu"
)

const numSavestateSlots = 8

// Menu item IDs for items without actions
const (
	menuIDLoadState = "load_state"
	menuIDSaveState = "save_state"
)

type appMenu struct {
	*menu.Bar[input.Action]
}

func newAppMenu(ui *ebitenui.UI, actions *actionRegistry, currentROMName string) *appMenu {
	def := buildMenuDefinition(currentROMName)

	style := menu.Style{
		Font:               res.fonts.face,
		BackgroundColor:    hex2color(0x000000FF),     // rgba(0, 0, 0, 1)
		MenuBackground:     hex2color(menuBackground), // rgba(50,50,50,1)
		TextColorIdle:      hex2color(0xFFFFFFFF),     // rgba(255,255,255,1)
		TextColorDisabled:  hex2color(0x808080FF),     // rgba(128,128,128,1)
		TextColorHover:     hex2color(0xFFFFFFFF),     // rgba(255,255,255,1)
		TextColorPressed:   hex2color(0x000000FF),     // rgba(0,0,0,1)
		ButtonHoverColor:   hex2color(0x404040FF),     // rgba(64,64,64,1)
		ButtonPressedColor: hex2color(0xFFFFFFFF),     // rgba(255,255,255,1)
	}

	bar := menu.New(ui, def, style, func(action input.Action) {
		actions.trigger(action)
	})

	return &appMenu{Bar: bar}
}

func buildMenuDefinition(currentROMName string) menu.Definition[input.Action] {
	loadslot := func(slot int) (getlabel func() string, getdisabled func() bool) {
		getlabel = func() string {
			t := config.SavestateInfo(currentROMName, slot)
			if t.IsZero() {
				return fmt.Sprintf("<empty> — %d", slot+1)
			}
			return fmt.Sprintf("%s — %d", t.Format(time.DateTime), slot+1)
		}
		getdisabled = func() bool {
			t := config.SavestateInfo(currentROMName, slot)
			if t.IsZero() {
				return true
			}
			return false
		}

		return getlabel, getdisabled
	}

	loadStateItems := make([]menu.Item[input.Action], numSavestateSlots)
	for i := range loadStateItems {
		getlabel, getdisabled := loadslot(i)
		loadStateItems[i] = menu.Item[input.Action]{
			LabelFunc:    getlabel,
			DisabledFunc: getdisabled,
			Action:       input.ActionLoadSavestateSlot1 + input.Action(i),
		}
	}

	saveStateItems := make([]menu.Item[input.Action], numSavestateSlots)
	for i := range saveStateItems {
		getlabel, _ := loadslot(i)
		saveStateItems[i] = menu.Item[input.Action]{
			LabelFunc: getlabel,
			Action:    input.ActionSaveSavestateSlot1 + input.Action(i),
		}
	}

	return menu.Definition[input.Action]{
		Menus: []menu.Menu[input.Action]{
			{
				Label: "File",
				Items: []menu.Item[input.Action]{
					{Label: "Open ROM ...", Action: input.ActionOpenROM},
					{Label: "Load State", ID: menuIDLoadState, SubMenu: loadStateItems},
					{Label: "Save State", ID: menuIDSaveState, SubMenu: saveStateItems},
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
