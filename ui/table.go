package ui

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

type textAlignment = widget.TextPosition

const (
	alignStart  textAlignment = widget.TextPositionStart
	alignCenter textAlignment = widget.TextPositionCenter
	alignEnd    textAlignment = widget.TextPositionEnd
)

type mouseClickArgs = widget.WidgetMouseButtonClickedEventArgs

const (
	propertyListHeaderBg    = 0x2a3944   // darker header background, same as listFocusedBackground
	propertyListCellBg      = panelColor // same as panel background for consistency
	propertyListBorderColor = 0x3a4a5a   // subtle border color
)

type property struct {
	key   string
	value string
}

type plist struct {
	*widget.Container
}

type plistConfig struct {
	headers    []string
	properties []property
	onClick    func(i int)
}

func newPlist(cfg plistConfig) *plist {
	if len(cfg.headers) != 2 {
		panic("property list requires exactly 2 headers (key and value)")
	}

	padding := widget.Insets{Top: 0, Bottom: 6, Left: 8, Right: 8}
	cellbg := ninesliceFromHex(propertyListCellBg)
	headerbg := ninesliceFromHex(propertyListHeaderBg)
	inset1 := widget.NewInsetsSimple(1)
	cellld := widget.AnchorLayoutData{
		StretchHorizontal:  true,
		StretchVertical:    true,
		HorizontalPosition: widget.AnchorLayoutPositionCenter,
		VerticalPosition:   widget.AnchorLayoutPositionCenter,
	}

	makecol := func() *widget.Container {
		return widget.NewContainer(
			// Use the container background as the border color
			widget.ContainerOpts.BackgroundImage(ninesliceFromHex(propertyListBorderColor)),
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(1),
				widget.GridLayoutOpts.DefaultStretch(true, true),
				widget.GridLayoutOpts.Padding(inset1), // outer border
				widget.GridLayoutOpts.Spacing(1, 1)))) // inner borders
	}

	makecellcontainer := func(bg *image.NineSlice) *widget.Container {
		return widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(bg),
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionCenter,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				})))
	}

	makeheader := func(text string) *widget.Container {
		label := widget.NewLabel(
			widget.LabelOpts.Text(text, res.fonts.boldFace, res.label.text),
			widget.LabelOpts.LabelPadding(&padding),
			widget.LabelOpts.TextOpts(
				widget.TextOpts.Position(alignCenter, alignCenter),
				widget.TextOpts.Padding(&widget.Insets{}),
				widget.TextOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(cellld))))

		cell := makecellcontainer(headerbg)

		cell.AddChild(label)
		return cell
	}

	makecell := func(text string, align textAlignment, onclick func(*mouseClickArgs)) *widget.Container {
		// Display placeholder for empty values
		displayText := text
		labelColor := res.label.text
		if displayText == "" {
			displayText = "<click to set>"
			labelColor = &widget.LabelColor{
				Idle:     hex2color(textDisabledColor),
				Disabled: hex2color(textDisabledColor),
			}
		}

		cell := makecellcontainer(cellbg)
		cell.AddChild(widget.NewLabel(
			widget.LabelOpts.Text(displayText, res.fonts.face, labelColor),
			widget.LabelOpts.LabelPadding(&padding),
			widget.LabelOpts.TextOpts(
				widget.TextOpts.Position(align, alignCenter),
				widget.TextOpts.WidgetOpts(
					widget.WidgetOpts.MouseButtonClickedHandler(onclick),
					widget.WidgetOpts.LayoutData(cellld),
				))))
		return cell
	}

	keys, values := makecol(), makecol()

	// Headers
	keys.AddChild(makeheader(cfg.headers[0]))
	values.AddChild(makeheader(cfg.headers[1]))

	for idx, prop := range cfg.properties {
		var handler func(args *mouseClickArgs)
		if cfg.onClick != nil {
			idx := idx
			handler = func(args *mouseClickArgs) {
				cfg.onClick(idx)
			}
		}
		keys.AddChild(makecell(prop.key, alignStart, handler))
		values.AddChild(makecell(prop.value, alignCenter, handler))
	}

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(0),
		)),
	)

	root.AddChild(keys, values)

	return &plist{Container: root}
}
