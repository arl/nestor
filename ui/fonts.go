package ui

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/arl/nestor/assets"
)

const (
	fontFaceRegular = "fonts/notosans-regular.ttf"
	fontFaceBold    = "fonts/notosans-bold.ttf"
	fontDejaVu      = "fonts/DejaVuSans.ttf"
)

type fonts struct {
	small        *text.Face
	smallBold    *text.Face
	face         *text.Face
	boldFace     *text.Face
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
	smallFace := must(loadFont(fontFaceRegular, 12))
	smallBoldFace := must(loadFont(fontFaceBold, 12))
	fontFace := must(loadFont(fontFaceRegular, 14))
	boldFace := must(loadFont(fontFaceBold, 14))
	titleFontFace := must(loadFont(fontFaceRegular, 24))
	bigTitleFontFace := must(loadFont(fontFaceRegular, 28))
	toolTipFace := must(loadFont(fontFaceRegular, 15))

	return &fonts{
		small:        &smallFace,
		smallBold:    &smallBoldFace,
		face:         &fontFace,
		boldFace:     &boldFace,
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
