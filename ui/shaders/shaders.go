package shaders

import (
	_ "embed"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed basic_crt.go
var BasicCRT []byte

//go:embed passthrough.go
var Passthrough []byte

//go:embed gizmo_crt.go
var GizmoCRT []byte

var shaders = map[string][]byte{
	"basic-crt":   BasicCRT,
	"passthrough": Passthrough,
	"gizmo-crt":   GizmoCRT,
}

var Names = sync.OnceValue(func() []string {
	return slices.Collect(maps.Keys(shaders))
})

func Load(name string) (*ebiten.Shader, error) {
	if shader, ok := shaders[name]; ok {
		return ebiten.NewShader(shader)
	}

	return nil, fmt.Errorf("unknown shader: %s", name)
}
