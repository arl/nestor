package ui

import (
	"fmt"
	"strings"

	"github.com/ebitenui/ebitenui/widget"

	"github.com/arl/nestor/hw/hwinput"
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

			// swap presets if needed
			other := 0
			if paddle == 0 {
				other = 1
			}
			if s.app.cfg.Input.Paddles[other].PaddlePreset == uint(idxpreset) {
				s.app.cfg.Input.Paddles[other].PaddlePreset = from
			}

			s.app.cfg.Input.Paddles[paddle].PaddlePreset = uint(idxpreset)
			s.app.savecfg()
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
			widget.RowLayoutOpts.Padding(&widget.Insets{Top: 20}),
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(spacing))))

	line1.AddChild(
		presetPaddlesBlock,
		widget.NewGraphic(widget.GraphicOpts.Image(res.images.paddleimg)))

	// selection preset paddle 1.
	presetPaddle1 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal))))

	presetPaddlesBlock.AddChild(presetPaddle1)

	presetPaddle1.AddChild(
		widget.NewLabel(
			widget.LabelOpts.Text("Paddle 1", res.label.face, res.label.text),
			widget.LabelOpts.LabelPadding(&widget.Insets{Top: 5, Left: 10}),
			widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionEnd,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				})))),
		newCombobox(presets, int(s.app.cfg.Input.Paddles[0].PaddlePreset),
			widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				MaxWidth:           200,
			},
			onPresetChanged(0)))

	// selection preset paddle 2.
	presetPaddle2 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal))))

	presetPaddlesBlock.AddChild(presetPaddle2)

	presetPaddle2.AddChild(
		widget.NewLabel(
			widget.LabelOpts.Text("Paddle 2", res.label.face, res.label.text),
			widget.LabelOpts.LabelPadding(&widget.Insets{Top: 5, Left: 10}),
			widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionEnd,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				})))),
		newCombobox(presets, int(s.app.cfg.Input.Paddles[1].PaddlePreset),
			widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				MaxWidth:           200,
			},
			onPresetChanged(1)))

	// 2nd line - preset configuration and button assignments
	line2 := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(spacing))))

	// currently configured preset.
	var currentPresetCombo *combobox
	currentPresetCombo = newCombobox(presets, s.presetidx, widget.RowLayoutData{}, func(idx int) {
		modUI.InfoZ("changed current preset").Int("preset", idx).End()
		s.presetidx = idx
		s.createUI()
	})

	currentPresetBlock := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(10),
		)))

	currentPresetBlock.AddChild(
		widget.NewLabel(
			widget.LabelOpts.Text("Configuring preset:", res.label.face, res.label.text)),
		currentPresetCombo,
	)

	// Buttons assignment table
	codes := s.app.cfg.Input.Presets[s.presetidx].ToButtons()

	buttons := []hwinput.PaddleButton{
		hwinput.PadSelect,
		hwinput.PadStart,
		hwinput.PadB,
		hwinput.PadA,
		hwinput.PadUp,
		hwinput.PadDown,
		hwinput.PadLeft,
		hwinput.PadRight,
	}

	properties := make([]property, len(buttons))
	for i, btn := range buttons {
		properties[i] = property{
			key:   strings.ToUpper(btn.String()),
			value: codes[btn].Name(),
		}
	}

	listcfg := plistConfig{
		headers:    []string{"Button", "Assigned to"},
		properties: properties,
		onClick: func(i int) {
			s.app.setState("capture", captureArgs{mode: captureModeNes, btn: buttons[i], idxpreset: s.presetidx})
		},
	}

	buttonsTable := newPlist(listcfg)

	helpLabel := widget.NewLabel(
		widget.LabelOpts.Text("Click on a button assignment to change it.", res.label.face, res.label.text))

	line2.AddChild(currentPresetBlock, helpLabel, buttonsTable)

	c.AddChild(
		line1,
		newSeparator(widget.RowLayoutData{Stretch: true}),
		line2,
	)

	return &page{title: "Input", content: c}
}
