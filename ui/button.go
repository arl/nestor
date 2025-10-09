package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2/vector"
)

var transparent = color.RGBA{0, 0, 0, 0}

func mixColors(a, b color.Color, percent float64) color.Color {
	rgba := func(c color.Color) (r, g, b, a uint8) {
		r16, g16, b16, a16 := c.RGBA()
		return uint8(r16 >> 8), uint8(g16 >> 8), uint8(b16 >> 8), uint8(a16 >> 8)
	}
	lerp := func(x, y uint8) uint8 {
		return uint8(math.Round(float64(x) + percent*(float64(y)-float64(x))))
	}
	r1, g1, b1, a1 := rgba(a)
	r2, g2, b2, a2 := rgba(b)

	return color.RGBA{
		R: lerp(r1, r2),
		G: lerp(g1, g2),
		B: lerp(b1, b2),
		A: lerp(a1, a2),
	}
}

func roundedRectPath(x, y, w, h, tl, tr, br, bl float32) *vector.Path {
	path := &vector.Path{}

	path.Arc(x+w-tr, y+tr, tr, 3*math.Pi/2, 0, vector.Clockwise)
	path.LineTo(x+w, y+h-br)
	path.Arc(x+w-br, y+h-br, br, 0, math.Pi/2, vector.Clockwise)
	path.LineTo(x+bl, y+h)
	path.Arc(x+bl, y+h-bl, bl, math.Pi/2, math.Pi, vector.Clockwise)
	path.LineTo(x, y+tl)
	path.Arc(x+tl, y+tl, tl, math.Pi, 3*math.Pi/2, vector.Clockwise)
	path.Close()

	return path
}
