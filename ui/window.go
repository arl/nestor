package ui

import (
	"image"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
)

func errorWindow(ui *ebitenui.UI, err error) {
	// Main container with panel background
	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(10),
		)),
	)

	// Header
	header := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.header.background),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(res.header.padding),
		)),
	)
	headerText := widget.NewText(
		widget.TextOpts.Text("Error", res.fonts.titleFace, hex2color(textErrorColor)),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	)
	header.AddChild(headerText)
	c.AddChild(header)

	// Content container with padding
	content := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(res.panel.padding),
			widget.RowLayoutOpts.Spacing(15),
		)),
	)

	const windowWidth = 500
	const windowHeight = 240
	textareaWidth := windowWidth - res.panel.padding.Left - res.panel.padding.Right

	// Error message text area with proper theme colors
	textArea := widget.NewTextArea(
		widget.TextAreaOpts.ContainerOpts(
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position:  widget.RowLayoutPositionCenter,
					Stretch:   true,
					MaxWidth:  textareaWidth,
					MaxHeight: 100,
				}),
				widget.WidgetOpts.MinSize(textareaWidth, 100),
			),
		),

		widget.TextAreaOpts.ControlWidgetSpacing(2),
		widget.TextAreaOpts.FontColor(res.text.idleColor),
		widget.TextAreaOpts.FontFace(res.textArea.face),
		widget.TextAreaOpts.TextPadding(*widget.NewInsetsSimple(10)),
		widget.TextAreaOpts.Text(err.Error()),
		widget.TextAreaOpts.ScrollContainerImage(res.base.scrollimg),
	)
	content.AddChild(textArea)

	var rw widget.RemoveWindowFunc

	// Close button
	cb := widget.NewButton(
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Text("Close", res.button.face, res.button.text),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			rw()
		}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)
	content.AddChild(cb)
	c.AddChild(content)

	w := widget.NewWindow(
		widget.WindowOpts.Modal(),
		widget.WindowOpts.Contents(c),
		widget.WindowOpts.CloseMode(widget.CLICK_OUT),
	)

	windowSize := input.GetWindowSize()
	r := image.Rect(0, 0, windowWidth, windowHeight)
	r = r.Add(image.Point{(windowSize.X - windowWidth) / 2, (windowSize.Y - windowHeight) / 2})
	w.SetLocation(r)

	rw = ui.AddWindow(w)
}
