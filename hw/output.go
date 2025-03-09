package hw

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/veandco/go-sdl2/sdl"

	"nestor/emu/log"
	"nestor/hw/input"
)

const (
	NTSCWidth  = 256
	NTSCHeight = 240
)

const PrimaryMonitor = 0

const DefaultScale = 2

type OutputConfig struct {
	// Dimensions of the video buffer (in pixels).
	Width, Height int32

	// Number of video buffers to allocate. Defaults to 2.
	NumVideoBuffers int

	// Window title.
	Title string

	// Window scale factor (defaults to 2).
	ScaleFactor int32

	// Monitor on which to display the window.
	// 0: primary monitor, 1: secondary monitor, etc.
	Monitor int32

	// Do not synchronize updates with vertical retrace (i.e immediate updates).
	DisableVSync bool

	// Shader name for additional video processing effects.
	Shader string
}

// A Frame holds the audio/video buffers the emulator
// should fill for a single frame.
type Frame struct {
	Video []byte
	_     []byte // TODO: Audio
}

type Output struct {
	framebufidx  int
	framebuf     [][]byte
	framecounter int
	framech      chan Frame

	fpscounter int
	fpsclock   uint64

	videoEnabled bool
	window       *window

	audioEnabled bool

	quit atomic.Bool
	stop chan struct{}
	wg   sync.WaitGroup // workers loops

	cfg OutputConfig
}

func NewOutput(cfg OutputConfig) *Output {
	if cfg.NumVideoBuffers == 0 {
		cfg.NumVideoBuffers = 2
	}

	vb := make([][]byte, cfg.NumVideoBuffers)
	for i := range vb {
		vb[i] = make([]byte, cfg.Width*cfg.Height*4)
	}
	out := &Output{
		framebuf: vb,
		cfg:      cfg,
		framech:  make(chan Frame),
		stop:     make(chan struct{}),
	}

	input.Gamectrls = input.NewGameControllers()

	out.wg.Add(2)
	go out.render()
	go out.poll()

	return out
}

func (out *Output) EnableVideo(enable bool) error {
	switch {
	case enable && !out.videoEnabled:
		if out.cfg.ScaleFactor == 0 {
			out.cfg.ScaleFactor = DefaultScale
		}

		window, err := newWindow(out.cfg)
		if err != nil {
			return fmt.Errorf("failed to create emulator window: %s", err)
		}
		out.window = window
		out.videoEnabled = true

	case !enable && out.videoEnabled:
		err := out.window.Close()
		if err != nil {
			return fmt.Errorf("failed to close emulator window: %s", err)
		}
		out.videoEnabled = false
	}

	return nil
}

func (out *Output) FocusWindow() {
	if out.videoEnabled {
		out.window.Raise()
	}
}

// Global for now
var audioDeviceID sdl.AudioDeviceID
var audioSpec sdl.AudioSpec

func (out *Output) EnableAudio(enable bool) error {
	log.ModSound.InfoZ("Enabling audio").Bool("enable", enable).End()
	switch {
	case enable && !out.audioEnabled:
		if err := sdl.Init(sdl.INIT_AUDIO); err != nil {
			return err
		}

		desired := sdl.AudioSpec{
			Freq:     maxSampleRate,
			Format:   AudioFormat,
			Channels: AudioChannels,
			Silence:  0,
			Samples:  AudioBufferSize,
			Callback: nil,
		}

		var obtained sdl.AudioSpec
		deviceID, err := sdl.OpenAudioDevice("", false, &desired, &obtained, 0)
		if err != nil {
			return err
		}

		audioDeviceID = deviceID
		audioSpec = obtained

		sdl.PauseAudioDevice(deviceID, false)
		return nil

	case !enable && out.audioEnabled:
		if audioDeviceID != 0 {
			sdl.CloseAudioDevice(audioDeviceID)
		}
		sdl.QuitSubSystem(sdl.INIT_AUDIO)
	}

	return nil
}

func (out *Output) BeginFrame() Frame {
	out.framebufidx++
	if out.framebufidx == out.cfg.NumVideoBuffers {
		out.framebufidx = 0
	}

	return Frame{
		Video: out.framebuf[out.framebufidx],
	}
}

func (out *Output) EndFrame(frame Frame) {
	out.framecounter++
	out.framech <- frame
}

// Stop output flow.
func (out *Output) Close() {
	log.ModEmu.DebugZ("Terminating output streams").End()
	close(out.stop)
	out.quit.Store(true)

	// Flow is stopped by now, but window may still be rendering.
	if out.window != nil {
		out.window.SetTitle("halted")
	}

	out.wg.Wait()
	if out.window == nil {
		return
	}

	if err := out.window.Close(); err != nil {
		log.ModEmu.WarnZ("Error closing SDL window").Error("error", err).End()
		return
	}
	log.ModEmu.DebugZ("Closing SDL window").End()
}

func (out *Output) render() {
	defer out.wg.Done()
	for {
		select {
		case <-out.stop:
			log.ModEmu.DebugZ("Stopped rendering loop").End()
			return
		case frame := <-out.framech:
			if out.videoEnabled {
				sdl.Do(func() { out.window.render(frame.Video) })
			}

			// Update FPS counter in title bar.
			if out.videoEnabled {
				out.fpscounter++
				if out.fpsclock+1000 < sdl.GetTicks64() {
					title := fmt.Sprintf("%s - %d FPS", out.cfg.Title, out.fpscounter)
					out.window.SetTitle(title)
					out.fpscounter = 0
					out.fpsclock += 1000
				}
			}
		}
	}
}

// Poll reports whether input polling is ongoing.
// (i.e false if user requested to quit)
// Safe for concurrent use.
func (out *Output) Poll() bool {
	return !out.quit.Load()
}

func (out *Output) poll() {
	defer out.wg.Done()

	for out.Poll() {
		sdl.Do(func() {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch e := event.(type) {
				case sdl.QuitEvent:
					out.quit.Store(true)
				case sdl.KeyboardEvent:
					if e.Type == sdl.KEYDOWN && e.Keysym.Sym == sdl.K_ESCAPE {
						out.quit.Store(true)
						return
					}
				case sdl.WindowEvent:
					if e.Event == sdl.WINDOWEVENT_RESIZED {
						width, height := e.Data1, e.Data2
						out.window.scaleViewport(width, height)
					}
				case sdl.ControllerDeviceEvent:
					input.Gamectrls.UpdateDevices(e)
				}
			}
		})
	}
}

func (out *Output) Screenshot() *image.RGBA {
	var img *image.RGBA

	sdl.Do(func() {
		fbidx := out.framebufidx - 1
		if fbidx < 1 {
			fbidx = out.cfg.NumVideoBuffers - 1
		}
		img = FramebufImage(out.framebuf[fbidx], out.cfg.Width, out.cfg.Height)
	})
	return img
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
