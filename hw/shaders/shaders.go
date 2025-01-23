package shaders

import (
	"embed"
	"fmt"
	"io"
	"iter"
	"maps"
	"strconv"

	"github.com/go-gl/gl/v4.5-core/gl"
)

//go:embed *
var dir embed.FS

var shaderInfo = map[string][2]string{
	"No shader": {"base.vert", "base.frag"},
	"CRT Basic": {"base.vert", "basic-crt.frag"},
}

func Names() iter.Seq[string] {
	return maps.Keys(shaderInfo)
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
	panic("invalid shader type " + strconv.Itoa(int(t)))
}

func Compile(name string, typ Type) (uint32, error) {
	info, ok := shaderInfo[name]
	if !ok {
		return 0, fmt.Errorf("shader not found: %s", name)
	}

	buf, err := readAll(info[typ])
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
