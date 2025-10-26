package ui

import "github.com/ebitenui/ebitenui/widget"

type Table struct {
	*widget.Container
}

type tableConfig struct {
	headers    []string
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

	table := &Table{
		Container: widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(ninesliceFromHex(0x333333)),
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(cfg.layoutData)),
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(numcols),
				widget.GridLayoutOpts.Spacing(10, 5),
				widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(10))))),
	}

	// Table headers
	for _, header := range cfg.headers {
		table.Container.AddChild(headerCell(header))
	}

	for _, row := range cfg.rows {
		for _, cell := range row {
			label := widget.NewLabel(
				widget.LabelOpts.Text(cell, res.label.face, res.label.text),
				widget.LabelOpts.TextOpts(
					widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter)))
			table.Container.AddChild(label)
		}
	}

	return table
}

func headerCell(text string) *widget.Container {
	label := widget.NewLabel(
		widget.LabelOpts.Text(text, res.label.face, res.label.text),
		widget.LabelOpts.TextOpts(
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter)))

	container := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(0x4a5f6f)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()))
	container.AddChild(label)
	return container
}
