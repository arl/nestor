package main

import (
	"fmt"
	"image/color"
	"log"
	"nestor/ppu"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

func startScreen(nes *NES) {
	myApp := app.New()
	w := myApp.NewWindow("NEStor")

	rect := canvas.NewRectangle(color.White)
	w.SetContent(rect)

	wsz := fyne.NewSize(float32(ppu.NTSC.Width), float32(ppu.NTSC.Height))
	w.Resize(wsz)
	log.Println("starting window")
	w.ShowAndRun()
}

func tidyUp() {
	fmt.Println("Exited")
}
