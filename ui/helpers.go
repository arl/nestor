package ui

import (
	"bytes"
	"fmt"
	"image/color"
	"sync"
	"unsafe"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

var buttonImage = sync.OnceValue(func() *widget.ButtonImage {
	return &widget.ButtonImage{
		Idle: image.NewBorderedNineSliceColor(
			color.NRGBA{R: 170, G: 170, B: 180, A: 255},
			color.NRGBA{90, 90, 90, 255},
			3),
		Hover: image.NewBorderedNineSliceColor(
			color.NRGBA{R: 130, G: 130, B: 150, A: 255},
			color.NRGBA{70, 70, 70, 255},
			3),
		Pressed: image.NewAdvancedNineSliceColor(
			color.NRGBA{R: 130, G: 130, B: 150, A: 255},
			image.NewBorder(3, 2, 2, 2, color.NRGBA{70, 70, 70, 255}),
		),
	}
})

var fontFaces = map[float64]text.Face{}

func loadFont(size float64) text.Face {
	if face, ok := fontFaces[size]; ok {
		return face
	}

	s, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(fmt.Errorf("failed to load font: %s", err))
	}

	face := &text.GoTextFace{
		Source: s,
		Size:   size,
	}
	fontFaces[size] = face
	return face
}

func stdButton(text string, onclick func(args *widget.ButtonClickedEventArgs)) *widget.Button {
	var (
		font   = loadFont(20)
		button *widget.Button
	)

	button = widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ButtonOpts.Image(buttonImage()),
		widget.ButtonOpts.Text(text, &font, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),
		widget.ButtonOpts.TextProcessBBCode(false),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),
		widget.ButtonOpts.PressedHandler(func(args *widget.ButtonPressedEventArgs) {
			button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			button.GetWidget().CustomData = true
		}),
		widget.ButtonOpts.ReleasedHandler(func(args *widget.ButtonReleasedEventArgs) {
			button.Text().SetPadding(&widget.Insets{})
			button.GetWidget().CustomData = false
		}),
		widget.ButtonOpts.CursorEnteredHandler(func(args *widget.ButtonHoverEventArgs) {
			if button.GetWidget().CustomData == true {
				button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			}
		}),
		widget.ButtonOpts.CursorExitedHandler(func(args *widget.ButtonHoverEventArgs) {
			button.Text().SetPadding(&widget.Insets{})
		}),
		widget.ButtonOpts.ClickedHandler(onclick),
		widget.ButtonOpts.DisableDefaultKeys(),
	)
	return button
}

// sampleBuffer is a wrapper around bytes.Buffer that returns nil instead of
// EOF, since returning EOF would stop the audio playback.
type sampleBuffer struct {
	*bytes.Buffer
}

func newSampleBuffer(size int) *sampleBuffer {
	return &sampleBuffer{bytes.NewBuffer(make([]byte, 0, size))}
}

func (s *sampleBuffer) Read(p []uint8) (int, error) {
	n, _ := s.Buffer.Read(p)
	if n == 0 {
		return 0, nil
	}
	return n, nil
}

func byteSliceFromInt16(arr []int16) []uint8 {
	return unsafe.Slice((*uint8)(unsafe.Pointer(unsafe.SliceData(arr))), len(arr)*2)
}
