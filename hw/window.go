package hw

import (
	"fmt"
	"unsafe"

	"nestor/hw/shaders"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

type window struct {
	*sdl.Window

	prog     uint32
	uniforms uniforms
	texture  uint32
	vao      uint32
	ubo      uint32
	context  sdl.GLContext
	cfg      OutputConfig
}

// create an opengl window that renders an unique texture
// which takes up the whole viewport.
func newWindow(cfg OutputConfig) (*window, error) {
	var (
		w   *window
		err error
	)
	sdl.Do(func() { w, err = _newWindow(cfg) })
	return w, err
}

func _newWindow(cfg OutputConfig) (*window, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, fmt.Errorf("failed to initialize SDL: %s", err)
	}

	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 4)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 5)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)

	x := sdl.WINDOWPOS_CENTERED_MASK | cfg.Monitor
	y := sdl.WINDOWPOS_CENTERED_MASK | cfg.Monitor
	winw := cfg.Width * cfg.ScaleFactor
	winh := cfg.Height * cfg.ScaleFactor
	const flags = sdl.WINDOW_OPENGL | sdl.WINDOW_SHOWN | sdl.WINDOW_RESIZABLE

	w, err := sdl.CreateWindow(cfg.Title, x, y, winw, winh, flags)
	if err != nil {
		return nil, fmt.Errorf("failed to create sdl window: %s", err)
	}

	context, err := w.GLCreateContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenGL context: %s", err)
	}

	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize opengl: %s", err)
	}

	if cfg.DisableVSync {
		sdl.GLSetSwapInterval(0)
	}

	// Create empty texture buffer.
	texbuf := make([]byte, winh*winw*4)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, cfg.Width, cfg.Height, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&texbuf[0]))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	vert, err := shaders.Compile(cfg.Shader, shaders.Vertex)
	if err != nil {
		return nil, fmt.Errorf("vertex shader %q compilation: %s", cfg.Shader, err)
	}

	frag, err := shaders.Compile(cfg.Shader, shaders.Fragment)
	if err != nil {
		return nil, fmt.Errorf("fragment shader %q compilation: %s", cfg.Shader, err)
	}

	prog, err := shaders.LinkProgram(vert, frag)
	if err != nil {
		return nil, fmt.Errorf("shader program link: %s", err)
	}

	var uniforms uniforms
	uniforms.getLocations(prog)

	var VBO, VAO, EBO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.GenBuffers(1, &EBO)

	gl.BindVertexArray(VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(&vertices[0]), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(&indices[0]), gl.STATIC_DRAW)

	// Position attributes
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 5*4, 0)
	gl.EnableVertexAttribArray(0)

	// Texture coordinate attributes.
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 5*4, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return &window{
		Window:   w,
		prog:     prog,
		uniforms: uniforms,
		texture:  texture,
		vao:      VAO,
		context:  context,
		cfg:      cfg,
	}, nil
}

type uniforms struct {
	// vertex+fragment
	textureSize int32
	inputSize   int32
	outputSize  int32

	// vertex
	mvpMatrix int32

	// fragment
	texture int32
}

func (u *uniforms) getLocations(prog uint32) {
	u.textureSize = gl.GetUniformLocation(prog, gl.Str("TextureSize\x00"))
	u.inputSize = gl.GetUniformLocation(prog, gl.Str("InputSize\x00"))
	u.outputSize = gl.GetUniformLocation(prog, gl.Str("OutputSize\x00"))
	u.mvpMatrix = gl.GetUniformLocation(prog, gl.Str("MVPMatrix\x00"))
	u.texture = gl.GetUniformLocation(prog, gl.Str("Texture\x00"))
}

func (w *window) updateUniforms() {
	gl.Uniform2f(w.uniforms.textureSize, float32(w.cfg.Width), float32(w.cfg.Height))
	gl.Uniform2f(w.uniforms.inputSize, float32(w.cfg.Width), float32(w.cfg.Height))

	// TODO: use viewport width and height instead of window width and height.
	winw, winh := w.GetSize()
	gl.Uniform2f(w.uniforms.outputSize, float32(winw), float32(winh))

	var identity = [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	gl.UniformMatrix4fv(w.uniforms.mvpMatrix, 1, false, &identity[0])
	gl.Uniform1i(w.uniforms.texture, 0)
}

func (w *window) render(video []byte) {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(w.prog)

	w.updateUniforms()
	gl.BindTexture(gl.TEXTURE_2D, w.texture)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, w.cfg.Width, w.cfg.Height, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&video[0]))
	gl.BindVertexArray(w.vao)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
	w.GLSwap()
}

// scaleViewport scales the viewport so as to maintain nes aspect ratio.
func (w *window) scaleViewport(winw, winh int32) {
	winRatio := float64(winw) / float64(winh)
	nesRatio := float64(w.cfg.Width) / float64(w.cfg.Height)

	var vpw, vph int32
	if winRatio > nesRatio {
		// Window is wider than nes screen.
		vph = winh
		vpw = int32(float64(winh) * nesRatio)
	} else {
		// Window is taller than nes screen.
		vpw = winw
		vph = int32(float64(winw) / nesRatio)
	}

	// Center the viewport within the window.
	offx := (winw - vpw) / 2
	offy := (winh - vph) / 2

	gl.Viewport(offx, offy, vpw, vph)
}

func (w *window) Close() error {
	errc := make(chan error, 1)
	sdl.Do(func() {
		if w.context != nil {
			sdl.GLDeleteContext(w.context)
		}
		err := w.Destroy()
		sdl.Quit()
		errc <- err
	})
	return <-errc
}

// Rows are the quad vertices in clockwise order.
// Columns are vertices position in (x y z) and texture coords (z t).
var vertices = [20]float32{
	// x, y, z, s, t
	1.0, 1.0, 0, 1, 0, // top right
	1.0, -1.0, 0, 1, 1, // bottom right
	-1.0, -1.0, 0, 0, 1, // bottom left
	-1.0, 1.0, 0, 0, 0, // top left
}

var indices = [6]uint32{
	0, 1, 3,
	1, 2, 3,
}
