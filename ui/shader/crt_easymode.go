//go:build ignore

//kage:unit pixels

package shader

/*
 * CRT EasyMode shader
 * Converted from crt-easymode.cg to Kage format
 * Original by EasyMode (GPL License)
 *
 * A flat CRT shader ideally for 1080p or higher displays.
 * Features: scanlines, RGB mask, sharpness control, and gamma correction.
 *
 * This is a simplified version for Kage's single-pass architecture.
 * Some advanced features (Lanczos filtering) are approximated.
 */

const (
	// Sharpness controls
	SHARPNESS_H = 0.5 // Horizontal sharpness (0.0-1.0)
	SHARPNESS_V = 1.0 // Vertical sharpness (0.0-1.0)

	// Mask parameters (Aperture Grille default)
	MASK_STRENGTH   = 0.3 // Strength of RGB mask (0.0-1.0)
	MASK_DOT_WIDTH  = 1.0 // Width of mask dots
	MASK_DOT_HEIGHT = 1.0 // Height of mask dots
	MASK_STAGGER    = 0.0 // Stagger pattern (0.0 for aperture grille)
	MASK_SIZE       = 1.0 // Overall mask scale

	// Scanline parameters
	SCANLINE_STRENGTH       = 1.0   // Overall scanline intensity
	SCANLINE_BEAM_WIDTH_MIN = 1.5   // Minimum beam width
	SCANLINE_BEAM_WIDTH_MAX = 1.5   // Maximum beam width
	SCANLINE_BRIGHT_MIN     = 0.35  // Minimum brightness for scanline
	SCANLINE_BRIGHT_MAX     = 0.65  // Maximum brightness for scanline
	SCANLINE_CUTOFF         = 400.0 // Resolution cutoff for scanlines

	// Gamma and brightness
	GAMMA_INPUT  = 2.0 // Input gamma
	GAMMA_OUTPUT = 1.8 // Output gamma
	BRIGHT_BOOST = 1.2 // Final brightness boost
	DILATION     = 1.0 // Color dilation effect

	PI = 3.141592653589
)

// Apply dilation effect to enhance colors
func dilate(col vec4) vec4 {
	x := mix(vec4(1.0), col, DILATION)
	return col * x
}

// Apply s-curve for sharper interpolation
func curveDistance(x float, sharp float) float {
	xStep := step(0.5, x)
	curve := 0.5 - sqrt(0.25-(x-xStep)*(x-xStep))*sign(0.5-x)
	return mix(x, curve, sharp)
}

// Sample with gamma correction
func sampleGamma(srcPos vec2) vec3 {
	col := dilate(imageSrc0At(srcPos))
	return pow(col.rgb, vec3(GAMMA_INPUT/(DILATION+1.0)))
}

// Bilinear interpolation with sharpness control
func filterBilinear(srcPos vec2, imgSize vec2) vec3 {
	// Calculate pixel coordinates
	pixCo := srcPos - vec2(0.5)
	texCo := floor(pixCo) + vec2(0.5)
	dist := pixCo - floor(pixCo)

	// Apply curve distance for sharpness
	curveX := curveDistance(dist.x, SHARPNESS_H)
	curveY := curveDistance(dist.y, SHARPNESS_V)

	// Sample four neighbors
	dx := vec2(1.0, 0.0)
	dy := vec2(0.0, 1.0)

	col1 := mix(sampleGamma(texCo), sampleGamma(texCo+dx), curveX)
	col2 := mix(sampleGamma(texCo+dy), sampleGamma(texCo+dx+dy), curveX)

	return mix(col1, col2, curveY)
}

// Calculate RGB mask pattern
func getMaskWeight(coords vec2, imgSize vec2) vec3 {
	mask := 1.0 - MASK_STRENGTH

	// Calculate which dot we're on
	modFac := floor(coords * imgSize / vec2(MASK_SIZE, MASK_DOT_HEIGHT*MASK_SIZE))

	// Apply stagger pattern
	stagger := mod(modFac.y, 2.0) * MASK_STAGGER
	dotNo := int(mod((modFac.x+stagger)/MASK_DOT_WIDTH, 3.0))

	// Return RGB mask weight
	if dotNo == 0 {
		return vec3(1.0, mask, mask) // Red
	} else if dotNo == 1 {
		return vec3(mask, 1.0, mask) // Green
	} else {
		return vec3(mask, mask, 1.0) // Blue
	}
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	imgSize := imageSrc0Size()
	coords := srcPos / imgSize

	// Apply filtering with sharpness
	col := filterBilinear(srcPos, imgSize)

	// Calculate luminance and brightness
	luma := dot(vec3(0.2126, 0.7152, 0.0722), col)
	bright := (max(col.r, max(col.g, col.b)) + luma) / 2.0

	// Scanline parameters based on brightness
	scanBright := clamp(bright, SCANLINE_BRIGHT_MIN, SCANLINE_BRIGHT_MAX)
	scanBeam := clamp(bright*SCANLINE_BEAM_WIDTH_MAX, SCANLINE_BEAM_WIDTH_MIN, SCANLINE_BEAM_WIDTH_MAX)

	// Calculate scanline weight
	scanWeight := 1.0 - pow(cos(coords.y*2.0*PI*imgSize.y)*0.5+0.5, scanBeam)*SCANLINE_STRENGTH

	// Apply scanline effect (skip if resolution is too high)
	col2 := col
	if imgSize.y < SCANLINE_CUTOFF {
		col *= scanWeight
		col = mix(col, col2, scanBright)
	}

	// Apply RGB mask
	maskWeight := getMaskWeight(coords, imgSize)
	col *= maskWeight

	// Apply output gamma and brightness
	col = pow(col, vec3(1.0/GAMMA_OUTPUT))
	col *= BRIGHT_BOOST

	return vec4(col, 1.0)
}
