package ui

import (
	"maps"
	"slices"

	"github.com/ebitenui/ebitenui/widget"
)

func (s *configState) generalConfigPage() *page {
	c := newPageContentContainer()
	c.SetBackgroundImage(res.panel.image)

	const spacing = 20

	shortcuts := s.app.cfg.General.KeyboardShortcuts
	actions := slices.Sorted(maps.Keys(shortcuts))

	cells := make([][]cell, len(actions))
	for i, action := range actions {
		cells[i] = []cell{
			{text: action},
			{text: string(shortcuts[action]), clickable: true},
		}
	}

	tablecfg := tableConfig{
		headers:    []string{"Action", "Keyboard Shortcut"},
		cells:      cells,
		layoutData: widget.RowLayoutData{Stretch: true},
		onClick: func(i, j int) {
			s.app.setState("capture", captureArgs{mode: captureModeUI, action: actions[i]})
		},
	}

	shortcutsTable := newStaticTable(tablecfg)

	bottomlabel := widget.NewLabel(
		widget.LabelOpts.Text("Click in the table to define shortcuts", res.label.face, res.label.text),
		widget.LabelOpts.LabelPadding(&widget.Insets{Left: 275}))

	c.AddChild(
		shortcutsTable,
		bottomlabel,
	)

	return &page{title: "General", content: c}
}
