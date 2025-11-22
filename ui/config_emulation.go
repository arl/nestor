package ui

import "github.com/ebitenui/ebitenui/widget"

func (s *configState) emulationConfigPage() *page {
	c := newPageContentContainer()
	c.SetBackgroundImage(res.panel.image)

	// run-ahead frames.
	runaheadFrames := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal))))

	runaheadPresets := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

	runaheadFrames.AddChild(
		widget.NewLabel(
			widget.LabelOpts.Text("Run-ahead frames", res.label.face, res.label.text),
			widget.LabelOpts.LabelPadding(&widget.Insets{Top: 5, Left: 10}),
			widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionEnd,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				})))),
		newCombobox(runaheadPresets, int(s.app.cfg.Emulation.RunAheadFrames),
			widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				MaxWidth:           200,
			},
			func(idxpreset int) {
				from := s.app.cfg.Emulation.RunAheadFrames
				modUI.InfoZ("changed run-ahead frames").
					Uint("from", from).
					Uint("to", uint(idxpreset)).
					End()

				s.app.cfg.Emulation.RunAheadFrames = uint(idxpreset)
				s.app.savecfg()
				s.createUI()
			}))

	c.AddChild(runaheadFrames)

	return &page{title: "Emulation", content: c}
}
