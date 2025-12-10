package ui

import (
	"github.com/ebitenui/ebitenui/widget"
)

type PropertyList struct {
	*widget.Container
}

type ColumnAlignment = widget.AnchorLayoutPosition

const (
	LeftAlign   ColumnAlignment = widget.AnchorLayoutPositionStart
	CenterAlign ColumnAlignment = widget.AnchorLayoutPositionCenter
	RightAlign  ColumnAlignment = widget.AnchorLayoutPositionEnd
)

// Property list colors - using theme-consistent colors for better readability
const (
	propertyListHeaderBg    = 0x2a3944   // darker header background, same as listFocusedBackground
	propertyListCellBg      = panelColor // same as panel background for consistency
	propertyListBorderColor = 0x3a4a5a   // subtle border color
	propertyListBorderWidth = 1
)

type property struct {
	key       string
	value     string
	clickable bool
}

type propertyListConfig struct {
	headers    []string
	properties []property
	layoutData any
	onClick    func(i int)
}

func newPropertyList(cfg propertyListConfig) *PropertyList {
	if len(cfg.headers) != 2 {
		panic("property list requires exactly 2 headers (key and value)")
	}

	root := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(cfg.layoutData),
			widget.WidgetOpts.MinSize(320, 0),
		),
		// Use the container background as the border color
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(propertyListBorderColor)),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.DefaultStretch(true, false),
			// Add padding for the outer border
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(propertyListBorderWidth)),
			// Add spacing for the inner borders
			widget.GridLayoutOpts.Spacing(propertyListBorderWidth, propertyListBorderWidth))))

	// Headers
	for i, header := range cfg.headers {
		isValueColumn := i == 1
		root.AddChild(propertyHeaderCell(header, isValueColumn))
	}

	// Properties (key-value pairs)
	for idx, prop := range cfg.properties {
		var handler func(args *MouseClickArgs)
		if prop.clickable && cfg.onClick != nil {
			idx := idx // capture for closure
			handler = func(args *MouseClickArgs) {
				cfg.onClick(idx)
			}
		}
		// Key cell (left-aligned)
		root.AddChild(propertyKeyCell(prop.key))
		// Value cell (centered, clickable)
		root.AddChild(propertyValueCell(prop.value, handler))
	}

	return &PropertyList{Container: root}
}

func propertyHeaderCell(text string, isValueColumn bool) *widget.Container {
	padding := widget.Insets{Top: 8, Bottom: 8, Left: 12, Right: 12}

	// Both header cells are centered
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
				}))))

	cellContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(propertyListHeaderBg)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{
			HorizontalPosition: widget.GridLayoutPositionCenter,
			VerticalPosition:   widget.GridLayoutPositionCenter,
		})))

	cellContainer.AddChild(label)
	return cellContainer
}

// propertyKeyCell creates a left-aligned cell for the property key
func propertyKeyCell(text string) *widget.Container {
	padding := widget.Insets{Top: 6, Bottom: 6, Left: 12, Right: 12}

	label := widget.NewLabel(
		widget.LabelOpts.Text(text, res.fonts.face, res.label.text),
		widget.LabelOpts.LabelPadding(&padding),
		widget.LabelOpts.TextOpts(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					StretchHorizontal:  true,
					StretchVertical:    true,
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}))))

	cellContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(propertyListCellBg)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			})))

	cellContainer.AddChild(label)
	return cellContainer
}

// propertyValueCell creates a centered cell for the property value, optionally clickable
func propertyValueCell(text string, onclick func(*MouseClickArgs)) *widget.Container {
	padding := widget.Insets{Top: 6, Bottom: 6, Left: 12, Right: 12}

	widgetOpts := []widget.WidgetOpt{
		widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal:  true,
			StretchVertical:    true,
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		})}

	// Make the cell visually interactive if clickable
	var cellBg uint32 = propertyListCellBg
	if onclick != nil {
		widgetOpts = append(widgetOpts, widget.WidgetOpts.MouseButtonClickedHandler(onclick))
		cellBg = listSelectedBackground // slightly different background for interactive cells
	}

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

	label := widget.NewLabel(
		widget.LabelOpts.Text(displayText, res.fonts.face, labelColor),
		widget.LabelOpts.LabelPadding(&padding),
		widget.LabelOpts.TextOpts(
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(widgetOpts...)))

	cellContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(cellBg)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionCenter,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			})))

	cellContainer.AddChild(label)
	return cellContainer
}
