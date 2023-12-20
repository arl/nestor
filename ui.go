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

func startGUI(screenCh <-chan *image.RGBA) {
	myApp := app.New()
	w := myApp.NewWindow("NEStor")

	img := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, Width, Height)))
	w.SetContent(img)

	go func() {
		for {
			img.Image = <-screenCh
			img.Refresh()
		}
	}()

	wsz := fyne.NewSize(float32(Width), float32(Height))
	w.Resize(wsz)
	w.ShowAndRun()
}
