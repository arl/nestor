package ui

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
)

type Table struct {
	*widget.Container
}

type ColumnAlignment widget.TextPosition

const (
	LeftAlign   ColumnAlignment = ColumnAlignment(widget.TextPositionStart)
	CenterAlign ColumnAlignment = ColumnAlignment(widget.TextPositionCenter)
	RightAlign  ColumnAlignment = ColumnAlignment(widget.TextPositionEnd)
)

const (
	headerCellColor = widgetDisabledColor
	tableCellColor  = listSelectedBackground
	cellBorderWidth = 1
	cellBorderColor = 0x000000
)

type tableConfig struct {
	headers  []string
	colAlign []ColumnAlignment

	rows       [][]string
	layoutData any
}

func newStaticTable(cfg tableConfig, relayout func()) *Table {
	numcols := len(cfg.headers)
	if len(cfg.colAlign) != numcols {
		panic("number of column alignments does not match number of headers")
	}
	for _, row := range cfg.rows {
		if len(row) != numcols {
			panic("inconsistent number of columns in table rows")
		}
	}

	root := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(cfg.layoutData)),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(numcols),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(2)))))

	// Table headers
	for _, header := range cfg.headers {
		root.AddChild(headerCell(header))
	}

	for irow, row := range cfg.rows {
		for icell, cell := range row {
			align := widget.TextPosition(cfg.colAlign[icell])

			var container *widget.Container
			var label *widget.Label

			clickHandler := func(args *widget.WidgetMouseButtonClickedEventArgs) {
				fmt.Printf("Clicked on cell: %d %d value: %s widget: %T %+v\n", irow, icell, cell, args.Widget, args.Widget)
				label.Label = fmt.Sprintf("%s (clicked)", cell)
				relayout()
				relayout()
				// root.RequestRelayout()
			}

			label = widget.NewLabel(
				widget.LabelOpts.Text(cell, res.fonts.face, res.label.text),
				widget.LabelOpts.TextOpts(
					widget.TextOpts.Position(align, widget.TextPositionCenter),
					widget.TextOpts.Padding(&widget.Insets{Top: 2, Bottom: 2, Left: 7, Right: 2}),
					widget.TextOpts.WidgetOpts(widget.WidgetOpts.MouseButtonClickedHandler(clickHandler)),
				),
			)

			container = widget.NewContainer(
				widget.ContainerOpts.BackgroundImage(nineSliceBorderFromHex(cellBorderWidth, cellBorderColor, tableCellColor)),
				widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
				widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MouseButtonClickedHandler(clickHandler)))

			container.AddChild(label)
			root.AddChild(container)
		}
	}

	return &Table{Container: root}
}

func headerCell(text string) *widget.Container {
	label := widget.NewLabel(
		widget.LabelOpts.Text(text, res.fonts.boldFace, res.label.text),
		widget.LabelOpts.TextOpts(
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.Padding(&widget.Insets{Top: 2, Bottom: 7, Left: 7, Right: 7})))

	container := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(nineSliceBorderFromHex(cellBorderWidth, cellBorderColor, headerCellColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()))

	container.AddChild(label)
	return container
}
