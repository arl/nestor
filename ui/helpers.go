package ui

import (
	"bytes"
	"image"
	"image/color"
	"io"
	"sync"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// sampleBuffer is a wrapper around bytes.Buffer that returns nil instead of
// EOF, since returning EOF would stop the audio playback.
type sampleBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func newSampleBuffer(size int) *sampleBuffer {
	return &sampleBuffer{}
}

func (s *sampleBuffer) Read(p []uint8) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n, _ := s.buf.Read(p)
	if n == 0 {
		return 0, nil
	}
	return n, nil
}

func (s *sampleBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buf.Write(p)
}

func (s *sampleBuffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buf.Len()
}

func byteSliceFromInt16(arr []int16) []uint8 {
	return unsafe.Slice((*uint8)(unsafe.Pointer(unsafe.SliceData(arr))), len(arr)*2)
}

// filImage ensures the returned image has the given width, scaling it if
// necessary to fit. The height is adjusted to maintain the original aspect
// ratio.
func fitImage(src *ebiten.Image, width float64) *ebiten.Image {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	scale := width / float64(sw)
	dst := ebiten.NewImage(int(width), int(float64(sh)*scale))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	dst.DrawImage(src, op)
	return dst
}

// mustDecodeImage decodes an image using registered decoders.
func mustDecodeImage(r io.Reader) *ebiten.Image {
	img, _, err := image.Decode(r)
	if err != nil {
		modUI.PanicZ("can't decode image").Error("err", err).End()
	}

	return ebiten.NewImageFromImage(img)
}

// frameImage returns a new image of the same dimensions than src, that contains
// src, just scaled down enough so that we can frame it inside a rect of
// 'thickness' pixels, of the given color.
//
// TODO: doing this often is not efficient (dixit ebiten.NewImage), we should
// instead clear an existing image. Ensure that's at least not leaking.
func frameImage(src *ebiten.Image, thickness int, col color.Color) *ebiten.Image {
	orgw := src.Bounds().Dx()
	orgh := src.Bounds().Dy()
	scalex := float64(orgw-thickness*2) / float64(orgw)
	scaley := float64(orgh-thickness*2) / float64(orgh)

	var op ebiten.DrawImageOptions
	op.GeoM.Scale(scalex, scaley)
	op.GeoM.Translate(float64(thickness), float64(thickness))

	framed := ebiten.NewImage(orgw, orgh)
	framed.DrawImage(src, &op)
	vector.StrokeRect(framed, 0, 0, float32(orgw), float32(orgh), float32(thickness*2), col, true)
	return framed
}

func ptrTo[T any](v T) *T {
	return &v
}
