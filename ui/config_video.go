package ui

import "github.com/ebitenui/ebitenui/widget"

func (s *configState) videoConfigPage() *page {
	c := newPageContentContainer()
	c.SetBackgroundImage(res.panel.image)

	c.AddChild(
		widget.NewCheckbox(
			widget.CheckboxOpts.Image(res.checkbox.image),
			widget.CheckboxOpts.InitialState(bool2state(!s.app.cfg.Video.DisableVSync)),
			widget.CheckboxOpts.Text("V-Sync", res.label.face, res.label.text),
			widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
				s.app.cfg.Video.DisableVSync = args.State == widget.WidgetUnchecked
				s.app.savecfg()
			}),
		),
		widget.NewCheckbox(
			widget.CheckboxOpts.Image(res.checkbox.image),
			widget.CheckboxOpts.InitialState(bool2state(s.app.cfg.Video.StartFullscreen)),
			widget.CheckboxOpts.Text("Start full-screen", res.label.face, res.label.text),
			widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
				s.app.cfg.Video.StartFullscreen = state2bool(args.State)
				s.app.savecfg()
			}),
		),
	)
	return &page{title: "Video", content: c}
}

func bool2state(b bool) widget.WidgetState {
	if b {
		return widget.WidgetChecked
	}
	return widget.WidgetUnchecked
}

func state2bool(s widget.WidgetState) bool {
	return s == widget.WidgetChecked
}
