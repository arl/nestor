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
	stop         chan struct{}

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
		stop:     make(chan struct{}),
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

// Stop output flow.
func (o *Output) Close() {
	close(o.stop)
}

func (o *Output) render() {
	rgba := image.RGBA{
		Stride: 4 * o.cfg.Width,
		Rect: image.Rectangle{
			Max: image.Point{
				X: o.cfg.Width,
				Y: o.cfg.Height,
			},
		},
	}

	for {
		select {
		case <-o.stop:
			return
		case frame := <-o.framech:
			if o.cfg.FrameOutCh == nil {
				// Just discard frame if we're headless.
				break
			}
			rgba.Pix = frame.video
			o.cfg.FrameOutCh <- rgba
		}
	}
}
