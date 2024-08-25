package hw

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"

	"nestor/emu/log"
)

type OutputConfig struct {
	Width           int
	Height          int
	NumVideoBuffers int
	Title           string
}

type Output struct {
	framebufidx int
	framebuf    [][]byte

	framecounter int
	framech      chan frame
	stop         chan struct{}

	fpscounter  int
	fpsclock    uint64
	fpsticks    []time.Time
	fpsticksidx int

	videoEnabled bool
	window       *window

	quit bool
	wg   sync.WaitGroup // workers loops

	cfg OutputConfig
}

func NewOutput(cfg OutputConfig) *Output {
	o := newOutput(cfg)
	o.wg.Add(2)
	go o.render()
	go o.poll()
	return o
}

func NewHeadlessOutput(cfg OutputConfig) *Output {
	o := newOutput(cfg)
	o.wg.Add(1)
	go o.render()
	return o
}

func newOutput(cfg OutputConfig) *Output {
	vb := make([][]byte, cfg.NumVideoBuffers)
	for i := range vb {
		vb[i] = make([]byte, cfg.Width*cfg.Height*4)
	}
	return &Output{
		framebuf: vb,
		cfg:      cfg,
		framech:  make(chan frame),
		stop:     make(chan struct{}),
	}
}

func (out *Output) EnableVideo(enable bool) error {
	if enable && !out.videoEnabled {
		window, err := newWindow(out.cfg.Title, out.cfg.Width, out.cfg.Height)
		if err != nil {
			return fmt.Errorf("failed to create emulator window: %s", err)
		}
		out.window = window
		out.videoEnabled = true
	} else if !enable && out.videoEnabled {
		err := out.window.Close()
		if err != nil {
			return fmt.Errorf("failed to close emulator window: %s", err)
		}
		out.videoEnabled = false
	}

	return nil
}

type frame struct {
	video []byte
}

func (out *Output) BeginFrame() (video []byte) {
	out.framebufidx++
	if out.framebufidx == out.cfg.NumVideoBuffers {
		out.framebufidx = 0
	}

	return out.framebuf[out.framebufidx]
}

func (out *Output) EndFrame(video []byte) {
	out.framecounter++
	out.framech <- frame{video: video}
}

// Stop output flow.
func (out *Output) Close() error {
	log.ModEmu.DebugZ("Terminating output streams").End()
	close(out.stop)

	// Flow is stopped by now, but window may still be rendering.
	out.window.SetTitle("halted")
	out.wg.Wait()

	log.ModEmu.DebugZ("Closing SDL window").End()
	return out.window.Close()
}

func (out *Output) render() {
	defer out.wg.Done()
	for {
		select {
		case <-out.stop:
			return
		case frame := <-out.framech:
			if out.videoEnabled {
				sdl.Do(func() {
					out.renderVideo(frame.video)
				})
			}

			// Update FPS counter in title bar
			if out.videoEnabled {
				out.fpscounter++
				if out.fpsclock+1000 < sdl.GetTicks64() {
					out.window.SetTitle(fmt.Sprintf("%s - %d FPS", out.cfg.Title, out.fpscounter))
					out.fpscounter = 0
					out.fpsclock += 1000
				}
			}
		}
	}
}

func (out *Output) renderVideo(video []byte) {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(out.window.prog)
	gl.BindTexture(gl.TEXTURE_2D, out.window.texture)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(out.cfg.Width), int32(out.cfg.Height), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&video[0]))
	gl.BindVertexArray(out.window.vao)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
	out.window.GLSwap()
}

// Poll reports whether input polling is ongoing.
func (out *Output) Poll() bool {
	return !out.quit
}

func (out *Output) poll() {
	defer out.wg.Done()
	for !out.quit {
		sdl.Do(func() {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch e := event.(type) {
				case *sdl.KeyboardEvent:
					if e.Type == sdl.KEYDOWN && e.Keysym.Sym == sdl.K_ESCAPE {
						out.quit = true
						return
					}

					kbstate := sdl.GetKeyboardState()
					fmt.Println(kbstate)
				case *sdl.QuitEvent:
					out.quit = true
				case *sdl.JoyButtonEvent:
				case *sdl.WindowEvent:
					if e.Event == sdl.WINDOWEVENT_RESIZED {
						width, height := e.Data1, e.Data2
						gl.Viewport(0, 0, int32(width), int32(height))
					}
				}
			}
		})
	}
}
