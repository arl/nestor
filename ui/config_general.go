package ui

import (
	"maps"
	"slices"

	"github.com/ebitenui/ebitenui/widget"
)

func (s *configState) generalConfigPage() *page {
	c := newPageContentContainer()
	c.SetBackgroundImage(res.panel.image)

	shortcuts := s.app.cfg.General.KeyboardShortcuts
	actions := slices.Sorted(maps.Keys(shortcuts))

	properties := make([]property, len(actions))
	for i, action := range actions {
		properties[i] = property{
			key:   action,
			value: string(shortcuts[action]),
		}
	}

	listcfg := plistConfig{
		headers:    []string{"Action", "Shortcut"},
		properties: properties,
		onClick: func(i int) {
			s.app.setState("capture", captureArgs{mode: captureModeUI, action: actions[i]})
		},
	}

	shortcutsTable := newPlist(listcfg)

	helpLabel := widget.NewLabel(
		widget.LabelOpts.Text("Click on a shortcut value to change it.", res.label.face, res.label.text))

	c.AddChild(
		helpLabel,
		shortcutsTable,
	)

	return &page{title: "General", content: c}
}
