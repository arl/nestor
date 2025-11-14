package shader

import (
	_ "embed"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed basic_crt.go
var basicCRT []byte

//go:embed none.go
var none []byte

//go:embed gizmo_crt.go
var gizmoCRT []byte

var shaders = map[string][]byte{
	"basic-crt": basicCRT,
	"none":      none,
	"gizmo-crt": gizmoCRT,
}

var Names = sync.OnceValue(func() []string {
	return slices.Collect(maps.Keys(shaders))
})

var Default = "none"

var DefaultIndex = sync.OnceValue(func() int {
	return slices.Index(Names(), Default)
})

func Load(name string) (*ebiten.Shader, error) {
	if shader, ok := shaders[name]; ok {
		return ebiten.NewShader(shader)
	}

	return nil, fmt.Errorf("unknown shader: %s", name)
}
