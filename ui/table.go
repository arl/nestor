package ui

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
)

type Table struct {
	*widget.Container
}

type ColumnAlignment = widget.AnchorLayoutPosition

const (
	LeftAlign   ColumnAlignment = widget.AnchorLayoutPositionStart
	CenterAlign ColumnAlignment = widget.AnchorLayoutPositionCenter
	RightAlign  ColumnAlignment = widget.AnchorLayoutPositionEnd
)

const (
	headerCellColor = widgetDisabledColor
	tableCellColor  = listSelectedBackground
	cellBorderWidth = 1
	cellBorderColor = 0x000000
)

type tableConfig struct {
	headers []string

	rows       [][]string
	layoutData any
}

func newStaticTable(cfg tableConfig) *Table {
	numcols := len(cfg.headers)
	for _, row := range cfg.rows {
		if len(row) != numcols {
			panic("inconsistent number of columns in table rows")
		}
	}

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(0xff0000)),

		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(cfg.layoutData),
			widget.WidgetOpts.MinSize(300, 0),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(numcols),
			widget.GridLayoutOpts.DefaultStretch(true, false),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(2)))))

	// Table headers
	for _, header := range cfg.headers {
		root.AddChild(headerCell(header))
	}

	for irow, row := range cfg.rows {
		for icell, cell := range row {
			handler := func(args *MouseClickArgs) {
				textinput := args.Widget.CustomData.(*widget.TextInput)
				textinput.SetText(fmt.Sprintf("(clicked) col %d row %d", icell, irow))
			}
			root.AddChild(tableCell(cell, handler))
		}
	}

	return &Table{Container: root}
}

func tableCell(text string, onclick func(*MouseClickArgs)) *widget.Container {
	clickHandler := func(args *widget.WidgetMouseButtonClickedEventArgs) {
		if onclick != nil {
			onclick(args)
		}
	}

	var textinput *widget.TextInput
	textinput = widget.NewTextInput(
		widget.TextInputOpts.Face(res.fonts.face),
		widget.TextInputOpts.Padding(res.textInput.padding),
		widget.TextInputOpts.Color(res.textInput.color),
		widget.TextInputOpts.Placeholder("<unset>"),
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MouseButtonClickedHandler(clickHandler),
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal:  true,
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
			})))

	textinput.SetText(text)
	textinput.GetWidget().CustomData = textinput
	textinput.GetWidget().Disabled = true

	cellContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(nineSliceBorderFromHex(cellBorderWidth, cellBorderColor, tableCellColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			// widget.WidgetOpts.MouseButtonClickedHandler(clickHandler),
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			})))

	cellContainer.AddChild(textinput)
	return cellContainer
}

func headerCell(text string) *widget.Container {
	padding := widget.Insets{Bottom: 10}

	label := widget.NewLabel(
		widget.LabelOpts.Text(text, res.fonts.boldFace, res.label.text),
		widget.LabelOpts.LabelPadding(&padding),
		widget.LabelOpts.TextOpts(
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					StretchHorizontal:  true,
					StretchVertical:    true,
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		))

	cellContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(nineSliceBorderFromHex(cellBorderWidth, cellBorderColor, headerCellColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{
			HorizontalPosition: widget.GridLayoutPositionCenter,
			VerticalPosition:   widget.GridLayoutPositionCenter,
		})),
	)

	cellContainer.AddChild(label)
	return cellContainer
}
