package ebit

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"nestor/emu"
)

func ImageFromFrame(frame *emu.Frame) *ebiten.Image {
	img := &image.RGBA{
		Pix:    frame.Video,
		Stride: 4 * emu.NTSCWidth,
		Rect:   image.Rect(0, 0, emu.NTSCWidth, emu.NTSCHeight),
	}

	return ebiten.NewImageFromImage(img)
}
