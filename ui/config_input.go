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

	onPresetChanged := func(paddle int) func(int) {
		return func(idx int) {
			from := cfg.Input.Paddles[paddle].PaddlePreset
			to := idx
			modUI.InfoZ("changed paddle preset").Int("paddle", paddle).
				String("from", presets[from]).String("to", presets[to]).
				End()
			cfg.Input.Paddles[paddle].PaddlePreset = uint(to)
		}
	}

	presetPad1 := newCombobox(presets, int(cfg.Input.Paddles[0].PaddlePreset), widget.RowLayoutData{
		Position: widget.RowLayoutPositionEnd,
		Stretch:  true,
	}, onPresetChanged(0))
	presetPad2 := newCombobox(presets, int(cfg.Input.Paddles[1].PaddlePreset), widget.RowLayoutData{
		Position: widget.RowLayoutPositionEnd,
		Stretch:  true,
	}, onPresetChanged(1))

	presetsContainer.AddChild(
		widget.NewLabel(widget.LabelOpts.Text("Paddle 1", res.label.face, res.label.text)),
		presetPad1.Widget,
		widget.NewLabel(widget.LabelOpts.Text("Paddle 2", res.label.face, res.label.text)),
		presetPad2.Widget,
	)

	// XXX: seems to not be necessary anymore. For some reason, ebitenui crashed
	// in input layer initialization before.
	// presetsContainer.Validate()

	root.AddChild(presetsContainer)

	c.AddChild(root)

	return &page{title: "Input", content: c}
}
