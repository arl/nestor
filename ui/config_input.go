package ui

import (
	"fmt"

	"nestor/hw/input"

	"github.com/ebitenui/ebitenui/widget"
)

func (s *configState) inputConfigPage() *page {
	c := newPageContentContainer()
	c.SetBackgroundImage(res.panel.image)

	// paddle preset selection
	var presets []string
	for i := 1; i <= 8; i++ {
		presets = append(presets, fmt.Sprintf("Preset %d", i))
	}

	onPresetChanged := func(paddle int) func(int) {
		return func(idxpreset int) {
			from := s.app.cfg.Input.Paddles[paddle].PaddlePreset
			modUI.InfoZ("changed paddle preset").
				Int("paddle", paddle).
				String("from", presets[from]).
				String("to", presets[idxpreset]).
				End()
			s.app.cfg.Input.Paddles[paddle].PaddlePreset = uint(idxpreset)
			s.presetidx = idxpreset
			s.createUI()
		}
	}

	const spacing = 20

	// 1st line.
	line1 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(spacing),
		)))

	// preset selection block.
	presetPaddlesBlock := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(spacing),
		)))

	line1.AddChild(
		presetPaddlesBlock,
		widget.NewGraphic(widget.GraphicOpts.Image(res.images.paddleimg)),
	)

	// selection preset paddle 1.
	presetPaddle1 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)))

	presetPaddlesBlock.AddChild(presetPaddle1)

	presetPaddle1.AddChild(
		widget.NewLabel(
			widget.LabelOpts.Text("Paddle 1", res.label.face, res.label.text),
			widget.LabelOpts.LabelPadding(&widget.Insets{Top: 5, Left: 10}),
			widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionEnd,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				}),
			))),
		newCombobox(presets, int(s.app.cfg.Input.Paddles[0].PaddlePreset),
			widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				MaxWidth:           200,
			},
			onPresetChanged(0),
		),
	)

	// selection preset paddle 2.
	presetPaddle2 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)))

	presetPaddlesBlock.AddChild(presetPaddle2)

	presetPaddle2.AddChild(
		widget.NewLabel(
			widget.LabelOpts.Text("Paddle 2", res.label.face, res.label.text),
			widget.LabelOpts.LabelPadding(&widget.Insets{Top: 5, Left: 10}),
			widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionEnd,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				}),
			))),
		newCombobox(presets, int(s.app.cfg.Input.Paddles[1].PaddlePreset),
			widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				MaxWidth:           200,
			},
			onPresetChanged(1),
		),
	)

	// 2nd line.
	line2 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(spacing*2),
		)))

	// currently configured preset..
	var currentPresetCombo *Combobox
	currentPresetCombo = newCombobox(presets, s.presetidx, widget.RowLayoutData{
		MaxWidth: 200,
	}, func(idx int) {
		modUI.InfoZ("changed current preset").Int("preset", idx).End()
		s.presetidx = idx
		s.createUI()
	})

	currentPresetBlock := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)))

	currentPresetBlock.AddChild(
		widget.NewLabel(
			widget.LabelOpts.Text("Configuring preset:", res.label.face, res.label.text),
			widget.LabelOpts.LabelPadding(&widget.Insets{Top: 5, Left: 10}),
			widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionEnd,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				}),
			))),
		currentPresetCombo,
	)

	// Buttons assignment table
	codes := s.app.cfg.Input.Presets[s.presetidx].ToButtons()

	cells := [][]cell{
		{{text: "A"}, {text: codes[input.PadA].Name(), clickable: true}},
		{{text: "B"}, {text: codes[input.PadB].Name(), clickable: true}},
		{{text: "SELECT"}, {text: codes[input.PadSelect].Name(), clickable: true}},
		{{text: "START"}, {text: codes[input.PadStart].Name(), clickable: true}},
		{{text: "UP"}, {text: codes[input.PadUp].Name(), clickable: true}},
		{{text: "DOWN"}, {text: codes[input.PadDown].Name(), clickable: true}},
		{{text: "LEFT"}, {text: codes[input.PadLeft].Name(), clickable: true}},
		{{text: "RIGHT"}, {text: codes[input.PadRight].Name(), clickable: true}},
	}

	tablecfg := tableConfig{
		headers:    []string{"Button", "Assigned to"},
		cells:      cells,
		layoutData: widget.RowLayoutData{Stretch: true},
		onClick: func(i, j int) {
			s.app.setState("capture", input.PaddleButton(i), s.presetidx)
		},
	}

	buttonsTable := newStaticTable(tablecfg)

	line2.AddChild(currentPresetBlock, buttonsTable)

	bottomlabel := widget.NewLabel(
		widget.LabelOpts.Text("Click in the table to define mappings", res.label.face, res.label.text),
		widget.LabelOpts.LabelPadding(&widget.Insets{Left: 220}))

	c.AddChild(
		line1,
		newSeparator(widget.RowLayoutData{Stretch: true}),
		line2,
		bottomlabel,
	)

	return &page{title: "Input", content: c}
}
