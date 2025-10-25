package ui

import (
	"fmt"

	"nestor/config"

	"github.com/ebitenui/ebitenui/widget"
)

/*
|-----------------------------------------------------------------------------|
| paddle 1: <Listbox selection preset> | paddle 2: <Listbox selection preset> |
|-----------------------------------------------------------------------------|
|         Click on a Paddle button to assign it                               |
|                                                      |----------------------|
|          |-----------------------------------|       |  button | assigned to|
|          |                                   |       |----------------------|
|          |     Interactive NES Paddle        |       | Select  |     F1     |
|          |                                   |       | Start   |     F2     |
|          |                                   |       |   B     |   space    |
|          |-----------------------------------|       |   A     |            |
|                                                      |  UP     |            |
|                                                      |  DOWN   |            |
|                                                      |  LEFT   |            |
|        Currently configuring:                        |  RIGHT  |            |
|        <Listbox selection preset>                    |         |            |
|                                                      |         |            |
*/

func inputConfigPage(cfg *config.Config) *page {
	c := newPageContentContainer()

	c.SetBackgroundImage(ninesliceFromHex(0x0000ff))

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(40)),
		)),
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(0x00ff00)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionEnd,
			Stretch:  true,
			MaxWidth: 1000,
		})),
	)

	// Paddle preset selectors
	presetsContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal:  true,
			StretchVertical:    true,
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
		})),
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(0xff0000)),
	)

	var presets []string
	for i := 1; i <= 8; i++ {
		presets = append(presets, fmt.Sprintf("Preset %d", i))
	}

	onSelectPreset := func(val string) {
		fmt.Println("selected", val)
	}

	presetPaddle1 := newCombobox(presets, widget.RowLayoutData{
		Position: widget.RowLayoutPositionEnd,
		Stretch:  true,
	}, onSelectPreset)
	presetPaddle2 := newCombobox(presets, widget.RowLayoutData{
		Position: widget.RowLayoutPositionEnd,
		Stretch:  true,
	}, onSelectPreset)

	presetsContainer.AddChild(
		widget.NewLabel(widget.LabelOpts.Text("paddle 1", res.label.face, res.label.text)),
		presetPaddle1.Widget,
		widget.NewLabel(widget.LabelOpts.Text("paddle 2", res.label.face, res.label.text)),
		presetPaddle2.Widget,
	)
	presetsContainer.Validate()

	root.AddChild(presetsContainer)

	c.AddChild(root)

	return &page{
		title:   "Input",
		content: c,
	}
}
