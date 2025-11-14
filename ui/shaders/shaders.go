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

//go:embed pixel_art.go
var PixelArt []byte

var shaders = map[string][]byte{
	"basic-crt": BasicCRT,
	"pixel-art": PixelArt,
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
