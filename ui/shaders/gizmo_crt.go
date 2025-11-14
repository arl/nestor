//go:build ignore

//kage:unit pixels

package shaders

/*
 * gizmo98 crt shader
 * Copyright (C) 2023 gizmo98
 *
 *   This program is free software; you can redistribute it and/or modify it
 *   under the terms of the GNU General Public License as published by the Free
 *   Software Foundation; either version 2 of the License, or (at your option)
 *   any later version.
 *
 * Converted from GLSL to Kage shader format
 *
 * https://github.com/gizmo98/gizmo-crt-shader
 *
 * This shader tries to mimic a CRT without extensive use of scanlines and rgb pattern emulation.
 * It uses horizontal subpixel scaling and adds brightness dependent scanline patterns and allows
 * fractional scaling.
 */

// Shader parameters (constants for Kage)
const (
	CURVATURE_X     = 0.1
	CURVATURE_Y     = 0.15
	BRIGHTNESS      = 0.5
	HORIZONTAL_BLUR = 0.0
	VERTICAL_BLUR   = 0.0
	BLUR_OFFSET     = 0.5
	BGR_LCD_PATTERN = 0.0
	SHRINK          = 0.0
	SNR             = 1.0
)

const PHI = 1.61803398874989484820459 // Φ = Golden Ratio

// Gold noise function for adding CRT noise
func goldNoise(xy vec2, seed float) float {
	return fract(tan(distance(xy*PHI, xy)*seed) * xy.x)
}

func saturateA(x vec2) vec2 {
	return clamp(x, 0.0, 1.0)
}

func textureVertical(uv vec2) vec4 {
	if HORIZONTAL_BLUR == 1.0 {
		uv1 := uv + vec2(-0.5, -0.5)
		uv2 := uv + vec2(-0.5+BLUR_OFFSET, -0.5)
		col1 := imageSrc0UnsafeAt(uv1)
		col2 := imageSrc0UnsafeAt(uv2)
		col := (col1 + col2) / 2.0

		if VERTICAL_BLUR == 1.0 {
			uv3 := uv + vec2(-0.5, -0.5+BLUR_OFFSET)
			uv4 := uv + vec2(-0.5+BLUR_OFFSET, -0.5+BLUR_OFFSET)
			col3 := imageSrc0UnsafeAt(uv3)
			col4 := imageSrc0UnsafeAt(uv4)
			col = (((col3 + col4) / 2.0) + col) / 2.0
		}
		return col
	} else {
		return imageSrc0UnsafeAt(uv)
	}
}

func textureCRT(uvr vec2, uvg vec2, uvb vec2) vec4 {
	return vec4(textureVertical(uvr).r, textureVertical(uvg).g, textureVertical(uvb).b, 1.0)
}

func getFuv(uv vec2) float {
	texSize := imageSrc0Size()
	uv = uv*texSize + 0.5
	iuv := floor(uv)
	fuv := uv - iuv
	return abs((fuv * fuv * fuv * (fuv*(fuv*6.0-15.0) + 10.0)).y - 0.5)
}

func getIuv(uv vec2) vec2 {
	texSize := imageSrc0Size()
	uv = uv * texSize
	iuv := floor(uv)
	return iuv
}

func addNoise(col vec4, coord vec2, frameCount float) vec4 {
	// Add some subpixel noise which simulates small CRT color variations
	iGlobalTime := frameCount * 0.025
	snr := SNR * 0.03125
	return clamp(col+goldNoise(coord, sin(iGlobalTime))*snr-snr/2.0, 0.0, 1.0)
}

func addScanlines(col vec4, coord vec2) vec4 {
	// Add scanlines which are wider for dark colors
	texSize := imageSrc0Size()
	dstSize := imageDstSize()
	brightness := 1.0 / BRIGHTNESS * 0.05
	scale := (dstSize.y / texSize.y) * 0.5
	dim := brightness * scale
	col.rgb -= dim * (abs(1.5 * (1.0 - col.rgb) * abs(abs(getFuv(coord)-0.5))))
	return col
}

func distort(coord vec2, screenScale vec2) vec2 {
	CURVATURE_DISTORTION := vec2(CURVATURE_X, CURVATURE_Y)
	// Barrel distortion shrinks the display area a bit, this will allow us to counteract that
	barrelScale := 1.0 - (0.23 * CURVATURE_DISTORTION)
	coord *= screenScale
	coord -= vec2(0.5)
	rsq := coord.x*coord.x + coord.y*coord.y
	coord += coord * (CURVATURE_DISTORTION * rsq)
	coord *= barrelScale
	if abs(coord.x) >= 0.5 || abs(coord.y) >= 0.5 {
		coord = vec2(-1.0) // If out of bounds, return an invalid value
	} else {
		coord += vec2(0.5)
		coord /= screenScale
	}
	return coord
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	texSize := imageSrc0Size()

	// srcPos is already in texture pixel coordinates (0-texSize range)
	// Normalize to 0-1 range for distortion
	texcoord := srcPos / texSize

	// Apply screen scale for aspect ratio
	screenScale := vec2(1.0, 1.0)

	if SHRINK > 0.0 {
		texcoord.x -= 0.5
		texcoord.x *= 1.0 + SHRINK
		texcoord.x += 0.5
	}

	texcoord = distort(texcoord, screenScale)
	if texcoord.x < 0.0 {
		return vec4(0.0)
	}

	// Convert back to pixel coordinates
	pixelCoord := texcoord * texSize

	// For the color separation effect
	spread := 0.333
	xr := pixelCoord.x
	xg := pixelCoord.x
	xb := pixelCoord.x

	if BGR_LCD_PATTERN == 1.0 {
		xr += spread * 2.0
	} else {
		xb += spread * 2.0
	}
	xg += spread

	coord_r := vec2(xr, pixelCoord.y)
	coord_g := vec2(xg, pixelCoord.y)
	coord_b := vec2(xb, pixelCoord.y)

	fragColor := textureCRT(coord_r, coord_g, coord_b)
	fragColor = addNoise(fragColor, dstPos.xy, dstPos.w)
	fragColor = addScanlines(fragColor, pixelCoord/texSize)

	return fragColor
}
