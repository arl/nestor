//go:build ignore
// +build ignore

//kage:unit pixels

package shader

const intensity = 0.05

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	texColor := imageSrc0At(srcPos)
	imgSize := imageSrc0Size()
	texCoord := srcPos / imgSize

	// Scanline effect.
	scanline := sin(texCoord.y*1200.0) * intensity

	// Vignette effect (darken screen edges).
	edge := 16.0 * texCoord.x * texCoord.y * (1.0 - texCoord.x) * (1.0 - texCoord.y)
	vignette := 0.3 + 0.7*pow(edge, 0.5)

	finalColor := texColor.rgb*vignette - scanline

	return vec4(finalColor, 1.0)
}
