package shaders

import (
	"embed"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/go-gl/gl/v4.5-core/gl"
)

//go:embed *
var dir embed.FS

var includeRegex = regexp.MustCompile(`(?m)^\s*#include\s+<([^>]+)>\s*$`)

type set[T comparable] map[T]bool

func Compile(name string, shaderType uint32) (uint32, error) {
	visited := make(set[string])

	source, err := expandSource(name, visited)
	if err != nil {
		return 0, err
	}
	source = preprocessSlang(source)

	sh := gl.CreateShader(shaderType)
	csrc, free := gl.Strs(source + "\x00")
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

func preprocessSlang(source string) string {
	lines := strings.Split(source, "\n")
	var processed []string
	for _, line := range lines {
		if strings.HasPrefix(line, "#pragma parameter") {
			continue // Ignore pragma parameter
		} else if strings.HasPrefix(line, "#pragma stage") {
			// Ignore stage pragmas as shader type is already set
			continue
		}
		processed = append(processed, line)
	}
	return strings.Join(processed, "\n")
}

func expandSource(filePath string, visited set[string]) (string, error) {
	if visited[filePath] {
		return "", fmt.Errorf("circular include detected for file: %s", filePath)
	}
	visited[filePath] = true

	f, err := dir.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	src := string(data)

	// Find matches for #include <filename>
	matches := includeRegex.FindAllStringSubmatch(src, -1)
	for _, match := range matches {
		includeFile := match[1] // the filename inside <>
		includeSource, err := expandSource(includeFile, visited)
		if err != nil {
			return "", err
		}
		// Replace the entire #include line with the file contents
		oldLine := match[0] // the entire matched line
		src = strings.Replace(src, oldLine, includeSource, 1)
	}

	return src, nil
}
