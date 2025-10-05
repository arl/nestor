package ui

import (
	"bytes"
	"unsafe"

	"github.com/ebitenui/ebitenui/widget"
	"golang.org/x/image/colornames"
)

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
