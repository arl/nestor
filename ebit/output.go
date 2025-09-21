package ebit

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/veandco/go-sdl2/sdl"

	"nestor/hw"
	"nestor/hw/apu"
)

const (
	NTSCWidth  = 256
	NTSCHeight = 240
)

type output struct {
	framebufidx  int
	framebuf     [][]byte
	framecounter int
	framech      chan *hw.Frame

	audioEnabled bool
	audiodev     sdl.AudioDeviceID
	audiobuf     [][]int16

	cfg hw.OutputConfig
}

func newOutput(framech chan *hw.Frame, cfg hw.OutputConfig) *output {
	if cfg.NumBackBuffers == 0 {
		cfg.NumBackBuffers = 2
	}

	videobuf := make([][]byte, cfg.NumBackBuffers)
	for i := range videobuf {
		videobuf[i] = make([]byte, cfg.Width*cfg.Height*4)
	}

	audiobuf := make([][]int16, cfg.NumBackBuffers)
	for i := range audiobuf {
		audiobuf[i] = make([]int16, hw.SamplesPerFrame)
	}

	return &output{
		framebuf: videobuf,
		audiobuf: audiobuf,
		cfg:      cfg,
		framech:  framech,
		// framech: make(chan *hw.Frame),
		// stop:    make(chan struct{}),
	}
}

func (out *output) BeginFrame() hw.Frame {
	out.framebufidx++
	if out.framebufidx == out.cfg.NumBackBuffers {
		out.framebufidx = 0
	}

	vbuf := out.framebuf[out.framebufidx]

	// Figure out how many APU samples this frame.
	//    ns0 = total samples up to frame N (frame fidx)
	//    ns1 = up to N+1 (fidx + 1)
	//    ns1 - ns0 = samples this frame
	fidx := out.framecounter % hw.FramesPerSecond
	ns0 := (hw.AudioSampleRate * fidx) / hw.FramesPerSecond
	ns1 := (hw.AudioSampleRate * (fidx + 1)) / hw.FramesPerSecond
	total := (ns1 - ns0) * apu.AudioChannels // interleaved stereo samples

	abuf := out.audiobuf[out.framebufidx]
	audioSlice := abuf[:total]

	return hw.Frame{
		Video: vbuf,
		Audio: apu.AudioBuffer{Samples: audioSlice},
	}
}

func (out *output) EndFrame(frame *hw.Frame) {
	out.framecounter++
	out.framech <- frame
}

func (out *output) Poll() bool { return true }

func (out *output) Close() {}

func (out *output) Screenshot() *image.RGBA { return nil }

func ImageFromFrame(frame *hw.Frame) *ebiten.Image {
	img := &image.RGBA{
		Pix:    frame.Video,
		Stride: 4 * NTSCWidth,
		Rect:   image.Rect(0, 0, NTSCWidth, NTSCHeight),
	}

	return ebiten.NewImageFromImage(img)
}
