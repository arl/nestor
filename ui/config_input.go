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

const (
	PURERED   = 0xff0000
	PUREGREEN = 0x00ff00
	PUREBLUE  = 0x0000ff
)

func inputConfigPage(cfg *config.Config) *page {
	c := newPageContentContainer()

	c.SetBackgroundImage(ninesliceFromHex(PUREBLUE))

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(40)),
		)),
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(PUREGREEN)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionEnd,
			Stretch:  true,
			MaxWidth: 1000,
		})),
	)

	// Main vertical layout
	mainLayout := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(10),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
		})),
	)

	// Paddle preset selectors (top section)
	presetsContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(PURERED)),
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

	// Middle section: instruction label + paddle display + assignments table
	middleContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(10),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
	)

	instructionLabel := widget.NewLabel(
		widget.LabelOpts.Text("Click on a Paddle button to assign it", res.label.face, res.label.text),
	)
	middleContainer.AddChild(instructionLabel)

	// Horizontal layout for paddle display and assignments table
	horizontalContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
	)

	// Interactive NES Paddle placeholder (white background, fixed size)
	paddleDisplay := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(0xffffff)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			MaxWidth:  400,
			MaxHeight: 200,
		})),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(400, 200)),
	)

	// Assignments table
	tableContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(10, 5),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
		widget.ContainerOpts.BackgroundImage(ninesliceFromHex(0x333333)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
	)

	// Table headers
	tableContainer.AddChild(
		widget.NewLabel(widget.LabelOpts.Text("Button", res.label.face, res.label.text)),
		widget.NewLabel(widget.LabelOpts.Text("Assigned to", res.label.face, res.label.text)),
	)

	// Table rows for each NES button
	buttons := []string{"Select", "Start", "B", "A", "UP", "DOWN", "LEFT", "RIGHT"}
	for _, btn := range buttons {
		tableContainer.AddChild(
			widget.NewLabel(widget.LabelOpts.Text(btn, res.label.face, res.label.text)),
			widget.NewLabel(widget.LabelOpts.Text("", res.label.face, res.label.text)),
		)
	}

	horizontalContainer.AddChild(paddleDisplay, tableContainer)
	middleContainer.AddChild(horizontalContainer)

	// Bottom section: "Currently configuring" combobox
	bottomContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
	)

	currentlyConfiguringLabel := widget.NewLabel(
		widget.LabelOpts.Text("Currently configuring:", res.label.face, res.label.text),
	)

	currentPresetCombo := newCombobox(presets, int(cfg.Input.Paddles[0].PaddlePreset), widget.RowLayoutData{
		MaxWidth: 200,
	}, func(idx int) {
		modUI.InfoZ("changed current preset").Int("preset", idx).End()
	})

	bottomContainer.AddChild(currentlyConfiguringLabel, currentPresetCombo.Widget)

	// Add all sections to main layout
	mainLayout.AddChild(presetsContainer, middleContainer, bottomContainer)
	root.AddChild(mainLayout)

	c.AddChild(root)

	return &page{title: "Input", content: c}
}
