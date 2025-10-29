package ui

func (s *configState) videoConfigPage() *page {
	c := newPageContentContainer()
	c.SetBackgroundImage(res.panel.image)

	return &page{title: "Video", content: c}
}
