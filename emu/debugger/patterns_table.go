package debugger

import (
	"image"
	"nestor/hw"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/widget"
)

type patternsTable struct {
	ppu *hw.PPU
}

func (pt patternsTable) render(ppu *hw.PPU) *image.RGBA {
	ptbuf := ppu.Bus.FetchPointer(0x0000)
	img := image.NewRGBA(image.Rect(0, 0, 128, 256))

	// A pattern table is 0x1000 bytes so 0x8000 bits.
	// One pixel requires 2 bits (4 colors), so there are 0x4000 pixels to draw.
	// That's a square of 128 x 128 pixels
	// Each tile is 8 x 8 pixels, that 's 16 x 16 tiles.
	for row := uint16(0); row < 256; row++ {
		for col := uint16(0); col < 128; col++ {
			addr := (row / 8 * 0x100) + (row % 8) + (col/8)*0x10
			pixel := uint8((ptbuf[addr]>>(7-(col%8)))&1) + ((ptbuf[addr+8]>>(7-(col%8)))&1)*2
			gray := pixel * 64
			img.Pix[(row*128*4)+(col*4)] = gray
			img.Pix[(row*128*4)+(col*4)+1] = gray
			img.Pix[(row*128*4)+(col*4)+2] = gray
			img.Pix[(row*128*4)+(col*4)+3] = 255
		}
	}
	return img
}

func (pt patternsTable) Layout(gtx C) D {
	size := image.Pt(128, 256)
	gtx.Constraints = layout.Exact(size)

	img := pt.render(pt.ppu)

	return widget.Image{
		Src:   paint.NewImageOp(img),
		Fit:   widget.Contain,
		Scale: 1,
	}.Layout(gtx)
}
