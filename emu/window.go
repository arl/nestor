package emu

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

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

const fragmentShaderSource = `
#version 330 core
out vec4 FragColor;
in vec2 TexCoord;

uniform sampler2D ourTexture;

void main() {
    FragColor = texture(ourTexture, TexCoord);
}
` + "\x00"

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
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetShaderInfoLog(shader, logLength, nil, &log[0])

		return 0, fmt.Errorf("shader compile error: %v", string(log))
	}

	return shader, nil
}

func linkProgram(vertexShader, fragmentShader uint32) (uint32, error) {
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		var glLog [256]byte
		gl.GetProgramInfoLog(program, int32(len(glLog)), &logLength, &glLog[0])
		return 0, fmt.Errorf("shader program link error: %v", string(glLog[:logLength]))
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

type emuWindow struct {
	window        *sdl.Window
	texture       uint32
	shaderProgram uint32
}

func ShowWindow(errc chan<- error) {
	runtime.LockOSThread()

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_JOYSTICK); err != nil {
		errc <- fmt.Errorf("failed to initialize SDL: %s", err)
		return
	}
	defer sdl.Quit()

	// Set OpenGL attributes
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)

	window, err := sdl.CreateWindow("SDL Window", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_OPENGL|sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		errc <- fmt.Errorf("failed to create window: %s", err)
		return
	}
	defer window.Destroy()

	context, err := window.GLCreateContext()
	if err != nil {
		errc <- fmt.Errorf("failed to create OpenGL context: %s", err)
		return
	}
	defer sdl.GLDeleteContext(context)

	if err := gl.Init(); err != nil {
		errc <- fmt.Errorf("failed to initialize Glow: %s", err)
		return
	}

	// Initialize SDL_image for loading PNG images
	if err := img.Init(img.INIT_PNG); err != nil {
		errc <- fmt.Errorf("failed to initialize SDL_image: %s", err)
		return
	}
	defer img.Quit()

	// Load image into surface
	surface, err := img.Load("logo.png")
	if err != nil {
		errc <- fmt.Errorf("failed to load image: %s", err)
		return
	}
	defer surface.Free()

	// Create texture from surface
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(surface.W), int32(surface.H), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&surface.Pixels()[0]))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		errc <- fmt.Errorf("vertex shader: %s", err)
		return
	}

	fragmentShader, err := compileShader(crtFragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		errc <- fmt.Errorf("fragment shader: %s", err)
		return
	}

	shaderProgram, err := linkProgram(vertexShader, fragmentShader)
	if err != nil {
		errc <- fmt.Errorf("failed to link shader program: %s", err)
		return
	}

	// Everything has been initialized, unlock the caller.
	errc <- nil

	vertices := []float32{
		// positions   // texture coords
		0.5, 0.5, 0.0, 1.0, 0.0, // top right
		0.5, -0.5, 0.0, 1.0, 1.0, // bottom right
		-0.5, -0.5, 0.0, 0.0, 1.0, // bottom left
		-0.5, 0.5, 0.0, 0.0, 0.0, // top left
	}

	indices := []uint32{
		0, 1, 3,
		1, 2, 3,
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

	// position attribute
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 5*4, 0)
	gl.EnableVertexAttribArray(0)
	// texture coord attribute
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 5*4, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	nframes := 0
	start := time.Now()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.KeyboardEvent:
				kbstate := sdl.GetKeyboardState()
				fmt.Println(kbstate)
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseButtonEvent:
				log.Println("Mouse button event")
			case *sdl.JoyButtonEvent:
				log.Println("Joystick button event")
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_RESIZED {
					width, height := e.Data1, e.Data2
					gl.Viewport(0, 0, int32(width), int32(height))
				}
			}
		}

		// Clear screen
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Draw texture with shader
		gl.UseProgram(shaderProgram)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.BindVertexArray(VAO)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)

		window.GLSwap()

		nframes++
		// sdl.Delay(16) // Approximate 60 FPS
	}

	// TODO: cleanup

	fmt.Println("FPS: ", float64(nframes)/time.Since(start).Seconds())
}
