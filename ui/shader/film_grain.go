//go:build ignore

//kage:unit pixels

package shader

/*
 * Film Grain shader
 * Converted from film-grain.slang to Kage format
 * Original by Martins Upitis (2013)
 * Ported by Hyllian (2024)
 *
 * Features:
 * - Perlin noise based grain
 * - Luminance based grain intensity
 * - Optional colored noise
 * - Rotating noise pattern
 */

var Time float

const (
	GRAIN_AMOUNT = 0.05 // Grain amount (0.01 - 0.2)
	COLOR_AMOUNT = 0.6  // Color amount (0.0 - 1.0)
	GRAIN_SIZE   = 1.6  // Grain particle size (1.5 - 2.5)
	LUM_AMOUNT   = 1.0  // Luminance amount (0.0 - 1.0)
	COLORED      = 0.0  // Use colored noise (0.0 or 1.0)

	PERM_TEX_UNIT      = 1.0 / 256.0
	PERM_TEX_UNIT_HALF = 0.5 / 256.0
)

func noise(x vec2) float {
	// NOISE(x) (sin(dot(x+vec2(timer,timer),vec2(12.9898,78.233)))*43758.5453)
	t := vec2(Time)
	return sin(dot(x+t, vec2(12.9898, 78.233))) * 43758.5453
}

func rnmA(tc vec2) float {
	// fract(NOISE(tc) * nRGBA.a) * 2.0 - 1.0
	// nRGBA.a is 1.3647
	return fract(noise(tc)*1.3647)*2.0 - 1.0
}

func rnmRG(tc vec2) vec2 {
	// fract(NOISE(tc) * nRGBA.rg) * 2.0 - vec2(1.0)
	// nRGBA.rg is vec2(1.0, 1.2154)
	n := noise(tc)
	return fract(vec2(n*1.0, n*1.2154))*2.0 - vec2(1.0)
}

func fade(t vec2) vec2 {
	// t * t * t * (t * (t * 6.0 - vec2(15.0)) + vec2(10.0))
	return t * t * t * (t*(t*6.0-vec2(15.0)) + vec2(10.0))
}

func noise2D_contr(pi, pf, pxy, ptu vec2) float {
	perm := rnmA(pi + ptu)
	grad := rnmRG(vec2(perm, 0.0))*4.0 - vec2(1.0)
	return dot(grad, pf-pxy)
}

func pnoise2D(p vec2) float {
	pi := PERM_TEX_UNIT*floor(p) + PERM_TEX_UNIT_HALF
	pf := fract(p)

	// Noise contributions
	n00 := noise2D_contr(pi, pf, vec2(0.0, 0.0), vec2(0.0, 0.0))
	n01 := noise2D_contr(pi, pf, vec2(0.0, 1.0), vec2(0.0, PERM_TEX_UNIT))
	n10 := noise2D_contr(pi, pf, vec2(1.0, 0.0), vec2(PERM_TEX_UNIT, 0.0))
	n11 := noise2D_contr(pi, pf, vec2(1.0, 1.0), vec2(PERM_TEX_UNIT, PERM_TEX_UNIT))

	fd := fade(pf)

	// Blend along x
	nx := mix(vec2(n00, n01), vec2(n10, n11), fd.x)

	// Blend along y
	return mix(nx.x, nx.y, fd.y)
}

func coordRot(tc vec2, angle, width, height float) vec2 {
	aspect := width / height
	rotX := ((tc.x*2.0 - 1.0) * aspect * cos(angle)) - ((tc.y*2.0 - 1.0) * sin(angle))
	rotY := ((tc.x*2.0 - 1.0) * aspect * sin(angle)) + ((tc.y*2.0 - 1.0) * cos(angle))
	rotX = ((rotX/aspect)*0.5 + 0.5)
	rotY = rotY*0.5 + 0.5
	return vec2(rotX, rotY)
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	imgSize := imageSrc0Size()
	width := imgSize.x
	height := imgSize.y
	texCoord := srcPos / imgSize

	rotOffset := vec3(1.425, 3.892, 5.835)

	rotCoordsR := coordRot(texCoord, Time+rotOffset.x, width, height)
	noiseVal := vec3(pnoise2D(rotCoordsR * vec2(width/GRAIN_SIZE, height/GRAIN_SIZE)))

	if COLORED == 1.0 {
		rotCoordsG := coordRot(texCoord, Time+rotOffset.y, width, height)
		rotCoordsB := coordRot(texCoord, Time+rotOffset.z, width, height)

		noiseG := pnoise2D(rotCoordsG * vec2(width/GRAIN_SIZE, height/GRAIN_SIZE))
		noiseB := pnoise2D(rotCoordsB * vec2(width/GRAIN_SIZE, height/GRAIN_SIZE))

		noiseVal.g = mix(noiseVal.r, noiseG, COLOR_AMOUNT)
		noiseVal.b = mix(noiseVal.r, noiseB, COLOR_AMOUNT)
	}

	col := imageSrc0At(srcPos).rgb

	// Luminance response
	lumcoeff := vec3(0.299, 0.587, 0.114)
	luminance := mix(0.0, dot(col, lumcoeff), LUM_AMOUNT)
	lum := smoothstep(0.2, 0.0, luminance)
	lum += luminance

	noiseVal = mix(noiseVal, vec3(0.0), pow(lum, 4.0))
	col = col + noiseVal*GRAIN_AMOUNT

	return vec4(col, 1.0)
}
