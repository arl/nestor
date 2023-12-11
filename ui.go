package main

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

const (
	Width  = 256
	Height = 224
)

func startScreen(nes *NES) {
	myApp := app.New()
	w := myApp.NewWindow("NEStor")

	rect := canvas.NewRectangle(color.White)
	w.SetContent(rect)

	wsz := fyne.NewSize(float32(Width), float32(Height))
	w.Resize(wsz)
	log.Println("starting window")
	w.ShowAndRun()
}

func tidyUp() {
	fmt.Println("Exited")
}
