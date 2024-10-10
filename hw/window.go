package hw

import (
	"fmt"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

type window struct {
	*sdl.Window
	prog    uint32
	texture uint32
	vao     uint32
	context sdl.GLContext
}

// create opengl window with a full screen texture buffer of size (texw, texh).
// The window is scaled by wscale.
func newWindow(title string, texw, texh, wscale int) (*window, error) {
	type result struct {
		w   *window
		err error
	}
	errc := make(chan result, 1)
	sdl.Do(func() {
		w, err := _newWindow(title, texw, texh, wscale)
		errc <- result{w, err}
	})
	res := <-errc
	return res.w, res.err
}

func _newWindow(title string, texw, texh, wscale int) (*window, error) {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_JOYSTICK); err != nil {
		return nil, fmt.Errorf("failed to initialize SDL: %s", err)
	}

	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)

	winw := int32(texw * wscale)
	winh := int32(texh * wscale)
	w, err := sdl.CreateWindow(title,
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		int32(winw), int32(winh),
		sdl.WINDOW_OPENGL|sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %s", err)
	}

	context, err := w.GLCreateContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenGL context: %s", err)
	}

	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize opengl: %s", err)
	}

	// Create empty texture buffer.
	tbuf := make([]byte, winh*winw*4)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(texw), int32(texh), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&tbuf[0]))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	vert, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, fmt.Errorf("vertex shader compliation: %s", err)
	}

	frag, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, fmt.Errorf("fragment shader compilation: %s", err)
	}

	prog, err := linkProgram(vert, frag)
	if err != nil {
		return nil, fmt.Errorf("shader program link: %s", err)
	}

	var VBO, VAO, EBO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.GenBuffers(1, &EBO)

	gl.BindVertexArray(VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Position attributes
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 5*4, 0)
	gl.EnableVertexAttribArray(0)

	// Texture coordinate attributes.
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 5*4, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return &window{
		Window:  w,
		prog:    prog,
		texture: texture,
		vao:     VAO,
		context: context,
	}, nil
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

// Columns are position and texture coordinates.
// Rows are the quad vertices in clockwise order.
var vertices = []float32{
	// x, y, z, s, t
	1.0, 1.0, 0, 1, 0, // top right
	1.0, -1.0, 0, 1, 1, // bottom right
	-1.0, -1.0, 0, 0, 1, // bottom left
	-1.0, 1.0, 0, 0, 0, // top left
}

var indices = []uint32{
	0, 1, 3,
	1, 2, 3,
}

const vertexShaderSource = `
#version 330 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec2 aTexCoord;

out vec2 TexCoord;

void main() {
    gl_Position = vec4(aPos, 1.0);
    TexCoord = aTexCoord;
}
` + "\x00"

//lint:ignore U1000 keep that for now
const fragmentShaderSource = `
#version 330 core
out vec4 FragColor;
in vec2 TexCoord;

uniform sampler2D ourTexture;

void main() {
    FragColor = texture(ourTexture, TexCoord);
}
` + "\x00"

//lint:ignore U1000 keep that for now
const crtFragmentShaderSource = `
#version 330 core
out vec4 FragColor;
in vec2 TexCoord;

uniform sampler2D ourTexture;

void main() {
    vec3 color = texture(ourTexture, TexCoord).rgb;
    float scanline = sin(TexCoord.y * 1200.0) * 0.05;
    float vignette = 0.3 + 0.7 * pow(16.0 * TexCoord.x * TexCoord.y * (1.0 - TexCoord.x) * (1.0 - TexCoord.y), 0.5);
    color = color * vignette - scanline;
    FragColor = vec4(color, 1.0);
}
` + "\x00"

func compileShader(source string, shaderType uint32) (uint32, error) {
	sh := gl.CreateShader(shaderType)
	csrc, free := gl.Strs(source)
	gl.ShaderSource(sh, 1, csrc, nil)
	free()
	gl.CompileShader(sh)

	var status int32
	if gl.GetShaderiv(sh, gl.COMPILE_STATUS, &status); status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(sh, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetShaderInfoLog(sh, logLength, nil, &log[0])

		return 0, fmt.Errorf("shader compile error: %v", string(log))
	}

	return sh, nil
}

func linkProgram(vertexShader, fragmentShader uint32) (uint32, error) {
	prg := gl.CreateProgram()
	gl.AttachShader(prg, vertexShader)
	gl.AttachShader(prg, fragmentShader)
	gl.LinkProgram(prg)

	var status int32
	if gl.GetProgramiv(prg, gl.LINK_STATUS, &status); status == gl.FALSE {
		var logLength int32
		var glLog [256]byte
		gl.GetProgramInfoLog(prg, int32(len(glLog)), &logLength, &glLog[0])
		return 0, fmt.Errorf("shader program link error: %v", string(glLog[:logLength]))
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return prg, nil
}
