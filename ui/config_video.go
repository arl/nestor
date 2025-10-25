package ui

func videoConfigPage() *page {
	c := newPageContentContainer()

	c.SetBackgroundImage(ninesliceFromHex(0x0000ff))

	return &page{
		title:   "Video",
		content: c,
	}
}
