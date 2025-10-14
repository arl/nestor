package ui

import (
	"log"

	"nestor/assets"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	fontFaceRegular = "fonts/notosans-regular.ttf"
	fontFaceBold    = "fonts/notosans-bold.ttf"
	fontDejaVu      = "fonts/DejaVuSans.ttf"
)

type fonts struct {
	face         *text.Face
	titleFace    *text.Face
	bigTitleFace *text.Face
	toolTipFace  *text.Face
}

func must[T any](t T, err error) T {
	if err != nil {
		modUI.PanicZ(err.Error()).End()
	}
	return t
}

func loadFonts() *fonts {
	fontFace := must(loadFont(fontDejaVu, 14))
	titleFontFace := must(loadFont(fontDejaVu, 24))
	bigTitleFontFace := must(loadFont(fontDejaVu, 28))
	toolTipFace := must(loadFont(fontDejaVu, 15))

	return &fonts{
		face:         &fontFace,
		titleFace:    &titleFontFace,
		bigTitleFace: &bigTitleFontFace,
		toolTipFace:  &toolTipFace,
	}
}

func loadFont(path string, size float64) (text.Face, error) {
	fontFile, err := assets.FS.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := text.NewGoTextFaceSource(fontFile)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &text.GoTextFace{
		Source: s,
		Size:   size,
	}, nil
}
