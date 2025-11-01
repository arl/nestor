package ui

import (
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

type cell struct {
	text      string
	clickable bool
}

type tableConfig struct {
	headers    []string
	cells      [][]cell
	layoutData any
	onClick    func(i, j int)
}

func newStaticTable(cfg tableConfig) *Table {
	numcols := len(cfg.headers)
	for _, row := range cfg.cells {
		if len(row) != numcols {
			panic("inconsistent number of columns in table rows")
		}
	}

	root := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(cfg.layoutData),
			widget.WidgetOpts.MinSize(280, 0),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(numcols),
			widget.GridLayoutOpts.DefaultStretch(true, false),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(2)))))

	// Table headers
	for _, header := range cfg.headers {
		root.AddChild(headerCell(header))
	}

	for irow, row := range cfg.cells {
		for icell, cell := range row {
			var handler func(args *MouseClickArgs)
			if cell.clickable && cfg.onClick != nil {
				handler = func(args *MouseClickArgs) {
					cfg.onClick(irow, icell)
				}
			}
			root.AddChild(tableCell(cell.text, handler))
		}
	}

	return &Table{Container: root}
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
		})))

	cellContainer.AddChild(label)
	return cellContainer
}

func tableCell(text string, onclick func(*MouseClickArgs)) *widget.Container {
	widgetOpts := []widget.WidgetOpt{
		widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal:  true,
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
		})}

	if onclick != nil {
		widgetOpts = append(widgetOpts, widget.WidgetOpts.MouseButtonClickedHandler(onclick))
	}

	var textinput *widget.TextInput
	textinput = widget.NewTextInput(
		widget.TextInputOpts.Face(res.fonts.face),
		widget.TextInputOpts.Padding(res.textInput.padding),
		widget.TextInputOpts.Color(res.textInput.color),
		widget.TextInputOpts.Placeholder("<unset>"),
		widget.TextInputOpts.WidgetOpts(widgetOpts...))

	textinput.SetText(text)
	textinput.GetWidget().CustomData = textinput
	textinput.GetWidget().Disabled = true

	cellContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(nineSliceBorderFromHex(cellBorderWidth, cellBorderColor, tableCellColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			})))

	cellContainer.AddChild(textinput)
	return cellContainer
}
