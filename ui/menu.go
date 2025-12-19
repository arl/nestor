package ui

import (
	"fmt"
	goimage "image"
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"golang.org/x/image/colornames"

	"nestor/ui/input"
)

const numSavestateSlots = 8

type appMenu struct {
	container *widget.Container

	fileOpen       *widget.Button
	fileLoadState  *widget.Button
	loadStateSlots [numSavestateSlots]*widget.Button
	fileSaveState  *widget.Button
	saveStateSlots [numSavestateSlots]*widget.Button
	fileQuit       *widget.Button

	settingsGeneral   *widget.Button
	settingsInput     *widget.Button
	settingsVideo     *widget.Button
	settingsEmulation *widget.Button

	help *widget.Button
}

func newAppMenu(ui *ebitenui.UI, actions *actionRegistry) *appMenu {
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.Black)),

		widget.ContainerOpts.Layout(
			widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			),
		),

		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true}),
		),
	)

	//
	// "File" menu
	//
	file := newAppMenuButton("File")
	var (
		open      = newAppMenuEntry("Open ROM ...")
		loadState = newAppMenuEntry("Load State")
		saveState = newAppMenuEntry("Save State")
		quit      = newAppMenuEntry("Quit")
	)

	open.ClickedEvent.AddHandler(func(args any) {
		actions.trigger(input.ActionOpenROM)
	})
	quit.ClickedEvent.AddHandler(func(args any) {
		actions.trigger(input.ActionQuit)
	})

	var loadStateSlots [numSavestateSlots]*widget.Button
	for i := range loadStateSlots {
		slot := i
		loadStateSlots[i] = newAppMenuEntry(fmt.Sprintf("Slot %d", i+1))
		loadStateSlots[i].ClickedEvent.AddHandler(func(args any) {
			actions.trigger(input.ActionLoadSavestateSlot1 + input.Action(slot))
		})
	}
	loadState.ClickedEvent.AddHandler(event.WrapHandler(func(args *widget.ButtonClickedEventArgs) {
		openAppSubMenu(args.Button.GetWidget(), ui, loadStateSlots[:]...)
	}))

	var saveStateSlots [numSavestateSlots]*widget.Button
	for i := range saveStateSlots {
		slot := i
		saveStateSlots[i] = newAppMenuEntry(fmt.Sprintf("Slot %d", i+1))
		saveStateSlots[i].ClickedEvent.AddHandler(func(args any) {
			actions.trigger(input.ActionSaveSavestateSlot1 + input.Action(slot))
		})
	}
	saveState.ClickedEvent.AddHandler(event.WrapHandler(func(args *widget.ButtonClickedEventArgs) {
		openAppSubMenu(args.Button.GetWidget(), ui, saveStateSlots[:]...)
	}))

	file.ClickedEvent.AddHandler(event.WrapHandler(func(args *widget.ButtonClickedEventArgs) {
		openAppMenu(args.Button.GetWidget(), ui, open, loadState, saveState, quit)
	}))
	root.AddChild(file)

	//
	// "Settings" menu
	//
	settings := newAppMenuButton("Settings")
	var (
		general       = newAppMenuEntry("General")
		inputSettings = newAppMenuEntry("Input")
		video         = newAppMenuEntry("Video")
		emulation     = newAppMenuEntry("Emulation")
	)

	general.ClickedEvent.AddHandler(func(args any) {
		actions.trigger(input.ActionSettingsOpenGeneralConfig)
	})
	inputSettings.ClickedEvent.AddHandler(func(args any) {
		actions.trigger(input.ActionSettingsOpenInputConfig)
	})
	video.ClickedEvent.AddHandler(func(args any) {
		actions.trigger(input.ActionSettingsOpenVideoConfig)
	})
	emulation.ClickedEvent.AddHandler(func(args any) {
		actions.trigger(input.ActionSettingsOpenEmulationConfig)
	})

	settings.ClickedEvent.AddHandler(event.WrapHandler(func(args *widget.ButtonClickedEventArgs) {
		openAppMenu(args.Button.GetWidget(), ui, general, inputSettings, video, emulation)
	}))
	root.AddChild(settings)

	help := newAppMenuButton("Help")
	root.AddChild(help)

	return &appMenu{
		container:         root,
		fileOpen:          open,
		fileLoadState:     loadState,
		fileSaveState:     saveState,
		fileQuit:          quit,
		loadStateSlots:    loadStateSlots,
		saveStateSlots:    saveStateSlots,
		settingsGeneral:   general,
		settingsInput:     inputSettings,
		settingsVideo:     video,
		settingsEmulation: emulation,
		help:              help,
	}
}

func newAppMenuButton(label string) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    image.NewNineSliceColor(color.Transparent),
			Hover:   image.NewNineSliceColor(colornames.Darkgray),
			Pressed: image.NewNineSliceColor(colornames.White),
		}),
		widget.ButtonOpts.Text(label, res.fonts.face, &widget.ButtonTextColor{
			Idle:     color.White,
			Disabled: colornames.Gray,
			Hover:    color.White,
			Pressed:  color.Black,
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Top:    4,
			Left:   4,
			Right:  32,
			Bottom: 4,
		}),
	)
}

func newAppMenuEntry(label string) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    image.NewNineSliceColor(color.Transparent),
			Hover:   image.NewNineSliceColor(colornames.Darkgray),
			Pressed: image.NewNineSliceColor(colornames.White),
		}),
		widget.ButtonOpts.Text(label, res.fonts.face, &widget.ButtonTextColor{
			Idle:     color.White,
			Disabled: colornames.Gray,
			Hover:    color.White,
			Pressed:  color.Black,
		}),
		widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 16, Right: 64}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)
}

func openAppMenu(opener *widget.Widget, ui *ebitenui.UI, entries ...*widget.Button) {
	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(hex2color(menuBackground))),

		widget.ContainerOpts.Layout(
			widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Spacing(4),
				widget.RowLayoutOpts.Padding(&widget.Insets{Top: 4, Bottom: 4}),
			),
		),

		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
	)

	for _, entry := range entries {
		c.AddChild(entry)
	}

	w, h := c.PreferredSize()
	x := opener.Rect.Min.X
	y := opener.Rect.Max.Y

	window := widget.NewWindow(
		widget.WindowOpts.Contents(c),
		widget.WindowOpts.CloseMode(widget.CLICK_OUT),
		widget.WindowOpts.Location(goimage.Rect(x, y, x+w, y+h)),
	)

	ui.AddWindow(window)
}

func openAppSubMenu(opener *widget.Widget, ui *ebitenui.UI, entries ...*widget.Button) {
	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(hex2color(menuBackground))),

		widget.ContainerOpts.Layout(
			widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Spacing(4),
				widget.RowLayoutOpts.Padding(&widget.Insets{Top: 4, Bottom: 4}),
			),
		),

		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
	)

	for _, entry := range entries {
		c.AddChild(entry)
	}

	w, h := c.PreferredSize()
	x := opener.Rect.Max.X
	y := opener.Rect.Min.Y

	window := widget.NewWindow(
		widget.WindowOpts.Contents(c),
		widget.WindowOpts.CloseMode(widget.CLICK),
		widget.WindowOpts.Location(goimage.Rect(x, y, x+w, y+h)),
	)

	ui.AddWindow(window)
}
