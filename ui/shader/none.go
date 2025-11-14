//go:build ignore

//kage:unit pixels

package shader

// Passthrough shader - returns the original texture color unchanged
func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	return imageSrc0At(srcPos)
}
