package shaders

import (
	_ "embed"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed basic_crt.go
var BasicCRT []byte

func Load(name string) (*ebiten.Shader, error) {
	switch name {
	case "basic-crt":
		return ebiten.NewShader(BasicCRT)
	default:
		return nil, fmt.Errorf("unknown shader: %s", name)
	}
}
