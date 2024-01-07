package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"nestor/emu/log"
	"nestor/hw"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 224
)

type C = layout.Context
type D = layout.Dimensions

var th = material.NewTheme()

type gui struct {
	w   *app.Window
	nes *NES
}

func newGUI(nes *NES) *gui {
	return &gui{
		nes: nes,
	}
}

func (ui *gui) run() {
	go func() {
		ui.w = app.NewWindow(
			app.Title("NEStor"),
			app.Size(512, 512),
		)
		if err := ui.loop(); err != nil {
			log.ModEmu.Fatalf("can't show window: %s", err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func (ui *gui) loop() error {
	var ops op.Ops

	events := make(chan event.Event)
	acks := make(chan struct{})

	nesFrame := ui.nes.FrameEvents()

	go func() {
		for {
			ev := ui.w.NextEvent()
			events <- ev
			<-acks
			if _, ok := ev.(system.DestroyEvent); ok {
				return
			}
		}
	}()

	for {
		select {
		case e := <-events:
			switch e := e.(type) {
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)

				// Register a global key listener for the escape key wrapping
				// our entire UI.
				area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
				key.InputOp{
					Tag:  ui.w,
					Keys: key.NameEscape,
				}.Add(gtx.Ops)

				// check for presses of the escape key and close the window if we find them.
				for _, event := range gtx.Events(ui.w) {
					switch event := event.(type) {
					case key.Event:
						if event.Name == key.NameEscape {
							return nil
						}
					}
				}
				// render and handle UI.
				ui.Layout(gtx)
				area.Pop()

				// Pass drawing operations to the gpu
				e.Frame(gtx.Ops)

			case system.DestroyEvent:
				fmt.Println("destroy event")
				acks <- struct{}{}
				return e.Err
			}
			acks <- struct{}{}

		case img := <-nesFrame:
			_ = img
			ui.w.Invalidate()
		}
	}
}

func (ui *gui) Layout(gtx C) D {
	pt := patternsTable{ppu: ui.nes.Hw.PPU}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Alignment(layout.NW), Spacing: layout.SpaceEnd}.
		Layout(gtx,
			layout.Rigid(nesScreen{}.Layout),
			layout.Rigid(pt.Layout),
		)
}

type nesScreen struct{}

func (ns nesScreen) Layout(gtx C) D {
	size := image.Pt(ScreenWidth, ScreenHeight)
	gtx.Constraints = layout.Exact(size)

	// Paint the shape with a green color.
	paint.FillShape(gtx.Ops, color.NRGBA{G: 0xFF, A: 0xFF}, clip.Rect{Max: gtx.Constraints.Min}.Op())

	return D{Size: size}
}

type patternsTable struct {
	ppu *hw.PPU
}

func (pt patternsTable) render(ptbuf []byte) *image.RGBA {
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

	ptbuf := pt.ppu.Bus.FetchPointer(0x0000)
	img := pt.render(ptbuf)

	return widget.Image{
		Src:   paint.NewImageOp(img),
		Fit:   widget.Contain,
		Scale: 0.5,
	}.Layout(gtx)
}
