package ui

import (
	"fmt"

	"nestor/config"

	"github.com/ebitenui/ebitenui/widget"
)

/*
|-------------------------------------------------------------------------------|
| paddle 1: <Combobox selection preset> | paddle 2: <Combobox selection preset> |
|-------------------------------------------------------------------------------|
|         Currently configuring:           <Combobox preset>                    |
|         Click on a Paddle button to assign it                                 |
|                                                      |------------------------|
|          |-----------------------------------|       |  button | assigned to  |
|          |                                   |       |------------------------|
|          |     Interactive NES Paddle        |       | Select  |     F1       |
|          |                                   |       | Start   |     F2       |
|          |                                   |       |   B     |   space      |
|          |-----------------------------------|       |   A     |              |
|                                                      |  UP     |              |
|                                                      |  DOWN   |              |
|                                                      |  LEFT   |              |
|        Currently configuring:                        |  RIGHT  |              |
|        <Combobox selection preset>                   |         |              |
|                                                      |         |              |
*/

const (
	PURERED   = 0xff0000
	PUREGREEN = 0x00ff00
	PUREBLUE  = 0x0000ff
)

func inputConfigPage(cfg *config.Config) *page {
	c := newPageContentContainer()
	c.SetBackgroundImage(res.background)

	presetsContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.BackgroundImage(res.background),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(4),
			widget.GridLayoutOpts.DefaultStretch(false, true),
			widget.GridLayoutOpts.Spacing(10, 0))))

	var presets []string
	for i := 1; i <= 8; i++ {
		presets = append(presets, fmt.Sprintf("Preset %d", i))
	}

	onPresetChanged := func(paddle int) func(int) {
		return func(idx int) {
			from := cfg.Input.Paddles[paddle].PaddlePreset
			to := idx
			modUI.InfoZ("changed paddle preset").
				Int("paddle", paddle).
				String("from", presets[from]).
				String("to", presets[to]).
				End()
			cfg.Input.Paddles[paddle].PaddlePreset = uint(to)
		}
	}

	presetsContainer.AddChild(
		widget.NewLabel(
			widget.LabelOpts.Text("Paddle 1", res.label.face, res.label.text),
			widget.LabelOpts.LabelPadding(&widget.Insets{Top: 5, Left: 10}),
			widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionEnd,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				}),
			))),
		newCombobox(presets, int(cfg.Input.Paddles[0].PaddlePreset),
			widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				MaxWidth:           200,
			},
			onPresetChanged(0),
		),
		widget.NewLabel(
			widget.LabelOpts.Text("Paddle 2", res.label.face, res.label.text),
			widget.LabelOpts.LabelPadding(&widget.Insets{Top: 5, Left: 10}),
			widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionEnd,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				}),
			))),
		newCombobox(presets, int(cfg.Input.Paddles[1].PaddlePreset),
			widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				MaxWidth:           200,
			},
			onPresetChanged(1),
		),
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
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(400, 200),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				MaxWidth:  400,
				MaxHeight: 200,
			})))

	buttonsTable := newStaticTable(tableConfig{
		headers:  []string{"Button", "Assigned to"},
		colAlign: []ColumnAlignment{CenterAlign, CenterAlign},
		rows: [][]string{
			{"Select", "F1"},
			{"Start", "F2"},
			{"B", "Space"},
			{"A", ""},
			{"Up", ""},
			{"Down", ""},
			{"Left", ""},
			{"Right", ""},
		},
		layoutData: widget.RowLayoutData{
			Stretch: true,
		},
	}, func() {
		c.RequestRelayout()
	})

	horizontalContainer.AddChild(paddleDisplay, buttonsTable)
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

	bottomContainer.AddChild(currentlyConfiguringLabel, currentPresetCombo)

	// Add all sections to main layout
	c.AddChild(
		presetsContainer,
		middleContainer,
		bottomContainer,
	)

	return &page{title: "Input", content: c}
}
