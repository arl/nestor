package hw

import (
	"image"
)

type OutputConfig struct {
	Width           int
	Height          int
	NumVideoBuffers int

	FrameOutCh chan image.RGBA
}

type Output struct {
	framebufidx int
	framebuf    [][]byte

	framecounter int
	framech      chan frame

	cfg OutputConfig
}

func NewOutput(cfg OutputConfig) *Output {
	vb := make([][]byte, cfg.NumVideoBuffers)
	for i := range vb {
		vb[i] = make([]byte, cfg.Width*cfg.Height*4)
	}
	o := &Output{
		framebuf: vb,
		cfg:      cfg,
		framech:  make(chan frame),
	}
	go o.render()
	return o
}

type frame struct {
	video []byte
}

func (o *Output) BeginFrame() (video []byte) {
	o.framebufidx++
	if o.framebufidx == o.cfg.NumVideoBuffers {
		o.framebufidx = 0
	}

	return o.framebuf[o.framebufidx]
}

func (o *Output) EndFrame(video []byte) {
	o.framecounter++
	o.framech <- frame{video: video}
}

func (o *Output) render() {
	if o.cfg.FrameOutCh == nil {
		for range o.framech {
			// We're headless, just discard all frames.
		}
	} else {
		rgba := image.RGBA{
			Stride: 4 * o.cfg.Width,
			Rect: image.Rectangle{
				Max: image.Point{
					X: o.cfg.Width,
					Y: o.cfg.Height,
				},
			},
		}

		for frame := range o.framech {
			rgba.Pix = frame.video
			o.cfg.FrameOutCh <- rgba
		}
	}
}
