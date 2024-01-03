package main

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

const (
	Width  = 256
	Height = 224
)

type gui struct {
	app fyne.App
	nes *NES
}

func newGUI(nes *NES) *gui {
	return &gui{
		nes: nes,
		app: app.New(),
	}
}

func (ui *gui) showMainWindow(screenCh <-chan *image.RGBA) {
	wnd := ui.app.NewWindow("NEStor")

	img := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, Width, Height)))
	wnd.SetContent(img)

	go func() {
		for {
			img.Image = <-screenCh
			img.Refresh()
		}
	}()

	wsz := fyne.NewSize(float32(Width), float32(Height))
	wnd.Resize(wsz)
	wnd.Show()
}

func (ui *gui) showPatternTables(ch <-chan struct{}) {

	wnd := ui.app.NewWindow("pattern tables")
	img := image.NewRGBA(image.Rect(0, 0, 128, 256))
	cimg := canvas.NewImageFromImage(img)
	wnd.SetContent(cimg)

	go func() {
		pt := ui.nes.Hw.PPU.Bus.FetchPointer(0x0000)
		for range ch {
			/*
				A pattern table is 0x1000 bytes so 0x8000 bits.
				One pixel requires 2 bits (4 colors), so there are 0x4000 pixels to draw.
				That's a square of 128 x 128 pixels
				Each tile is 8 x 8 pixels, that 's 16 x 16 tiles.
			*/
			for row := uint16(0); row < 256; row++ {
				for col := uint16(0); col < 128; col++ {
					addr := (row / 8 * 0x100) + (row % 8) + (col/8)*0x10
					pixel := uint8((pt[addr]>>(7-(col%8)))&1) + ((pt[addr+8]>>(7-(col%8)))&1)*2

					gray := pixel * 64
					img.Pix[(row*128*4)+(col*4)] = gray
					img.Pix[(row*128*4)+(col*4)+1] = gray
					img.Pix[(row*128*4)+(col*4)+2] = gray
					img.Pix[(row*128*4)+(col*4)+3] = 255
				}
			}

			cimg.Refresh()
		}
	}()

	wsz := fyne.NewSize(float32(img.Rect.Max.X), float32(img.Rect.Min.Y))
	wnd.Resize(wsz)
	wnd.Show()
}

func (ui *gui) run() {
	screenCh := ui.nes.AttachScreen()
	imgCh := make(chan *image.RGBA, 1)
	ptCh := make(chan struct{}, 1)

	go func() {
		for {
			imgCh <- <-screenCh
			ptCh <- struct{}{}
		}
	}()

	ui.showMainWindow(imgCh)
	ui.showPatternTables(ptCh)
	ui.app.Run()
}
