package shaders

import (
	"embed"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

//go:embed *.vert *.frag
var dir embed.FS

const DefaultName = "Passthrough"

// Names returns the names (without extension) of all embedded shader files.
func Names() []string {
	dirents, err := dir.ReadDir(".")
	if err != nil {
		panic(err)
	}

	files := make(map[string]bool)
	for _, dirent := range dirents {
		if dirent.IsDir() {
			continue
		}
		name := dirent.Name()
		name = strings.TrimSuffix(name, filepath.Ext(name))
		files[name] = true
	}

	var names []string
	for name := range files {
		names = append(names, name)
	}

	slices.Sort(names)
	return names
}

func readAll(path string) ([]byte, error) {
	f, err := dir.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

type Type uint32

const (
	Vertex   Type = 0
	Fragment Type = 1
)

func (t Type) glType() uint32 {
	switch t {
	case Vertex:
		return gl.VERTEX_SHADER
	case Fragment:
		return gl.FRAGMENT_SHADER
	}
	panic("glType: invalid shader type " + strconv.Itoa(int(t)))
}

func (t Type) ext() string {
	switch t {
	case Vertex:
		return ".vert"
	case Fragment:
		return ".frag"
	}
	panic("ext: invalid shader type " + strconv.Itoa(int(t)))
}

func Compile(name string, typ Type) (uint32, error) {
	buf, err := readAll(name + typ.ext())
	if err != nil {
		return 0, err
	}
	csrc, free := gl.Strs(string(buf) + "\x00")
	sh := gl.CreateShader(typ.glType())
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

func LinkProgram(vert, frag uint32) (uint32, error) {
	prg := gl.CreateProgram()
	gl.AttachShader(prg, vert)
	gl.AttachShader(prg, frag)
	gl.LinkProgram(prg)

	var status int32
	if gl.GetProgramiv(prg, gl.LINK_STATUS, &status); status == gl.FALSE {
		var logLength int32
		var glLog [256]byte
		gl.GetProgramInfoLog(prg, int32(len(glLog)), &logLength, &glLog[0])
		return 0, fmt.Errorf("shader program link error: %v", string(glLog[:logLength]))
	}

	gl.DeleteShader(vert)
	gl.DeleteShader(frag)

	return prg, nil
}
