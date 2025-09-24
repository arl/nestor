package ui

import (
	"image"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"

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

func stdButton(text string, onclick func(args *widget.ButtonClickedEventArgs)) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.TextFace(DefaultFont()),
		widget.ButtonOpts.TextColor(&widget.ButtonTextColor{
			Idle:    colornames.White,
			Hover:   colornames.White,
			Pressed: Mix(colornames.White, colornames.Black, 0.4),
		}),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:         DefaultNineSlice(colornames.Forestgreen),
			Hover:        DefaultNineSlice(Mix(colornames.Forestgreen, colornames.White, 0.2)),
			Pressed:      PressedNineSlice(Mix(colornames.Forestgreen, colornames.Black, 0.4)),
			PressedHover: PressedNineSlice(Mix(colornames.Forestgreen, colornames.Black, 0.4)),
		}),
		widget.ButtonOpts.TextLabel(text),
		widget.ButtonOpts.ClickedHandler(onclick),
	)
}
