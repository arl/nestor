package ui

import (
	goimage "image"
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"golang.org/x/image/colornames"
)

type appMenu struct {
	container *widget.Container

	fileQuit *widget.Button

	settingsInput     *widget.Button
	settingsVideo     *widget.Button
	settingsEmulation *widget.Button

	help *widget.Button
}

func newAppMenu(ui *ebitenui.UI) *appMenu {
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
		save = newAppMenuEntry("Save")
		load = newAppMenuEntry("Load")
		quit = newAppMenuEntry("Quit")
	)

	file.ClickedEvent.AddHandler(event.WrapHandler(func(args *widget.ButtonClickedEventArgs) {
		openAppMenu(args.Button.GetWidget(), ui, save, load, quit)
	}))
	root.AddChild(file)

	//
	// "Settings" menu
	// This is the same thing as the "File" menu, just with more entries.
	//
	settings := newAppMenuButton("Settings")
	var (
		input     = newAppMenuEntry("Input")
		video     = newAppMenuEntry("Video")
		emulation = newAppMenuEntry("Emulation")
	)
	settings.ClickedEvent.AddHandler(event.WrapHandler(func(args *widget.ButtonClickedEventArgs) {
		openAppMenu(args.Button.GetWidget(), ui, input, video, emulation)
	}))
	root.AddChild(settings)

	help := newAppMenuButton("Help")
	root.AddChild(help)

	return &appMenu{
		container:         root,
		fileQuit:          quit,
		settingsInput:     input,
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
		// Set the background to a translucent black.
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.RGBA{R: 0, G: 0, B: 0, A: 125})),

		// Menu entries should be arranged vertically.
		widget.ContainerOpts.Layout(
			widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Spacing(4),
				widget.RowLayoutOpts.Padding(&widget.Insets{Top: 1, Bottom: 1}),
			),
		),

		// Set the minimum size for the menu.
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(64, 0)),
	)

	for _, entry := range entries {
		c.AddChild(entry)
	}

	w, h := c.PreferredSize()

	window := widget.NewWindow(
		widget.WindowOpts.Modal(),
		widget.WindowOpts.Contents(c),

		widget.WindowOpts.CloseMode(widget.CLICK),

		widget.WindowOpts.Location(
			goimage.Rect(
				opener.Rect.Min.X,
				opener.Rect.Min.Y+opener.Rect.Max.Y,
				opener.Rect.Min.X+w,
				opener.Rect.Min.Y+opener.Rect.Max.Y+opener.Rect.Min.Y+h,
			),
		),
	)

	ui.AddWindow(window)
}
