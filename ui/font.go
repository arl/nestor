package ui

import (
	"bytes"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

var fontFaces = map[float64]*text.Face{}

func loadFont(size float64) *text.Face {
	if face, ok := fontFaces[size]; ok {
		return face
	}

	s, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic("failed to load font: " + err.Error())
	}

	var face text.Face = &text.GoTextFace{
		Source: s,
		Size:   size,
	}
	fontFaces[size] = &face
	return &face
}

func defaultFont() *text.Face {
	return loadFont(20)
}

func titleFont() *text.Face {
	return defaultFont()
}
