package shader

import (
	_ "embed"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed basic_crt.kage
var basicCRT []byte

//go:embed none.kage
var none []byte

//go:embed gizmo_crt.kage
var gizmoCRT []byte

//go:embed scanline.kage
var scanline []byte

//go:embed crt_nes_mini.kage
var crtNESMini []byte

//go:embed crt_aperture.kage
var crtAperture []byte

//go:embed crt_easymode.kage
var crtEasymode []byte

//go:embed crt_lottes.kage
var crtLottes []byte

//go:embed film_grain.kage
var filmGrain []byte

//go:embed ntsc_simple.kage
var ntscSimple []byte

//go:embed sharp_bilinear_scanlines.kage
var sharpBilinearScanlines []byte

var shaders = map[string][]byte{
	"basic-crt":                basicCRT,
	"crt-aperture":             crtAperture,
	"crt-easymode":             crtEasymode,
	"crt-lottes":               crtLottes,
	"crt-nes-mini":             crtNESMini,
	"film-grain":               filmGrain,
	"gizmo-crt":                gizmoCRT,
	"none":                     none,
	"ntsc-simple":              ntscSimple,
	"scanline":                 scanline,
	"sharp-bilinear-scanlines": sharpBilinearScanlines,
}

var Names = sync.OnceValue(func() []string {
	return slices.Sorted(maps.Keys(shaders))
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
