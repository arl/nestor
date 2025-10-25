package ui

import "nestor/config"

func videoConfigPage(cfg *config.Config) *page {
	c := newPageContentContainer()

	c.SetBackgroundImage(ninesliceFromHex(0x0000ff))

	return &page{title: "Video", content: c}
}
