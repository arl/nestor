package ui

func (s *configState) videoConfigPage() *page {
	c := newPageContentContainer()
	c.SetBackgroundImage(res.panel.image)

	c.AddChild(
		newCheckbox2States("Enable V-Sync", !s.app.cfg.Video.DisableVSync, func(enabled bool) {
			s.app.cfg.Video.DisableVSync = !enabled
			s.app.savecfg()
		}),
		newCheckbox2States("Start in fullscreen", s.app.cfg.Video.StartFullscreen, func(enabled bool) {
			s.app.cfg.Video.StartFullscreen = enabled
			s.app.savecfg()
		}),
	)
	return &page{title: "Video", content: c}
}
