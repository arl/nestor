package emu

import (
	"image"

	"github.com/arl/nestor/hw/apu"
)

type OutputConfig struct {
	// Dimensions of the video buffer (in pixels).
	Width, Height int32

	// Number of video and audio buffers to allocate. Defaults to 2.
	NumBackBuffers int
}

type Output struct {
	framebufidx  int
	framebuf     [][]byte
	framecounter int
	framech      chan *Frame

	audioEnabled bool
	audiobuf     [][]int16

	cfg OutputConfig
}

// NewOutput creates an emulation Output device.
//
// If framech is nil, frames are discarded (headless mode).
func NewOutput(framech chan *Frame, cfg OutputConfig) *Output {
	if cfg.NumBackBuffers == 0 {
		cfg.NumBackBuffers = 2
	}

	videobuf := make([][]byte, cfg.NumBackBuffers)
	for i := range videobuf {
		videobuf[i] = make([]byte, cfg.Width*cfg.Height*4)
	}

	audiobuf := make([][]int16, cfg.NumBackBuffers)
	for i := range audiobuf {
		audiobuf[i] = make([]int16, SamplesPerFrame)
	}

	return &Output{
		framebuf: videobuf,
		audiobuf: audiobuf,
		cfg:      cfg,
		framech:  framech,
	}
}

func (out *Output) BeginFrame() Frame {
	out.framebufidx++
	if out.framebufidx == out.cfg.NumBackBuffers {
		out.framebufidx = 0
	}

	vbuf := out.framebuf[out.framebufidx]

	// Figure out how many APU samples this frame.
	//    ns0 = total samples up to frame N (frame fidx)
	//    ns1 = up to N+1 (fidx + 1)
	//    ns1 - ns0 = samples this frame
	fidx := out.framecounter % FramesPerSecond
	ns0 := (AudioSampleRate * fidx) / FramesPerSecond
	ns1 := (AudioSampleRate * (fidx + 1)) / FramesPerSecond
	total := (ns1 - ns0) * apu.AudioChannels // interleaved stereo samples

	abuf := out.audiobuf[out.framebufidx]
	audioSlice := abuf[:total]

	return Frame{
		Video: vbuf,
		Audio: apu.AudioBuffer{Samples: audioSlice},
	}
}

func (out *Output) EndFrame(frame *Frame) {
	out.framecounter++
	if out.framech == nil {
		return
	}
	out.framech <- frame
}

func (out *Output) Screenshot() *image.RGBA {
	fbidx := out.framebufidx - 1
	if fbidx < 1 {
		fbidx = out.cfg.NumBackBuffers - 1
	}
	return framebufImage(out.framebuf[fbidx], out.cfg.Width, out.cfg.Height)
}
