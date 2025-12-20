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

type menuOptions struct {
	getROMName       func() string
	settingsDisabled bool
	openROMDisabled  bool
}

func newAppMenu(ui *ebitenui.UI, actions *actionRegistry, opts menuOptions) *appMenu {
	def := buildMenuDefinition(opts)

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

func buildMenuDefinition(opts menuOptions) menu.Definition[input.Action] {
	fileItems := func() []menu.Item[input.Action] {
		romName := opts.getROMName()
		noROM := romName == ""

		slotLabel := func(slot int) string {
			t := config.SavestateInfo(romName, slot)
			if t.IsZero() {
				return fmt.Sprintf("%d - <empty>", slot+1)
			}
			return fmt.Sprintf("%d - %s", slot+1, t.Format(time.DateTime))
		}

		loadStateItems := make([]menu.Item[input.Action], numSavestateSlots)
		for i := range loadStateItems {
			loadStateItems[i] = menu.Item[input.Action]{
				Label:    slotLabel(i),
				Disabled: config.SavestateInfo(romName, i).IsZero(),
				Action:   input.ActionLoadSavestateSlot1 + input.Action(i),
			}
		}

		saveStateItems := make([]menu.Item[input.Action], numSavestateSlots)
		for i := range saveStateItems {
			saveStateItems[i] = menu.Item[input.Action]{
				Label:  slotLabel(i),
				Action: input.ActionSaveSavestateSlot1 + input.Action(i),
			}
		}

		return []menu.Item[input.Action]{
			{Label: "Open ROM ...", Action: input.ActionOpenROM, Disabled: opts.openROMDisabled},
			{Label: "Load State", ID: menuIDLoadState, SubMenu: loadStateItems, Disabled: noROM},
			{Label: "Save State", ID: menuIDSaveState, SubMenu: saveStateItems, Disabled: noROM},
			{Label: "Quit", Action: input.ActionQuit},
		}
	}

	return menu.Definition[input.Action]{
		Menus: []menu.Menu[input.Action]{
			{
				Label:     "File",
				ItemsFunc: fileItems,
			},
			{
				Label: "Settings",
				Items: []menu.Item[input.Action]{
					{Label: "General", Action: input.ActionSettingsOpenGeneralConfig, Disabled: opts.settingsDisabled},
					{Label: "Input", Action: input.ActionSettingsOpenInputConfig, Disabled: opts.settingsDisabled},
					{Label: "Video", Action: input.ActionSettingsOpenVideoConfig, Disabled: opts.settingsDisabled},
					{Label: "Emulation", Action: input.ActionSettingsOpenEmulationConfig, Disabled: opts.settingsDisabled},
				},
			},
			{
				Label: "Help",
				Items: []menu.Item[input.Action]{},
			},
		},
	}
}
