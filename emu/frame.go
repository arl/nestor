package emu

import (
	"image"
	"image/png"
	"os"
	"slices"

	"nestor/hw/apu"
)

const (
	NTSCWidth  = 256
	NTSCHeight = 240

	FramesPerSecond = 60
	AudioSampleRate = apu.MaxSampleRate // 96_000 kHz
	AudioBufferSize = 1024              // TODO: adjust based on latency.

	// How many audio samples per frame, per channel.
	SamplesPerFrame = apu.AudioChannels * AudioSampleRate / FramesPerSecond
)

// A Frame holds the audio/video buffers for a single frame.
type Frame struct {
	Video []byte
	Audio apu.AudioBuffer
}

// FramebufImage returns an image.RGBA from a frame buffer.
func FramebufImage(framebuf []byte, w, h int32) *image.RGBA {
	return &image.RGBA{
		Pix:    slices.Clone(framebuf),
		Stride: 4 * int(w),
		Rect:   image.Rect(0, 0, int(w), int(h)),
	}
}

func SaveAsPNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		return err
	}
	return f.Close()
}
