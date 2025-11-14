//go:build ignore

//kage:unit pixels

package shaders

/*
   Pixel art AA shader by DariusG 2023
   This program is free software; you can redistribute it and/or modify it
   under the terms of the GNU General Public License as published by the Free
   Software Foundation; either version 2 of the License, or (at your option)
   any later version.

   Converted from GLSL to Kage shader format
*/

// Luminance weight vector for calculating brightness
var lumweight = vec3(0.3, 0.6, 0.1)

// Calculate luminance of a color
func lum(c vec3) float {
	return c.x*lumweight.x + c.y*lumweight.y + c.z*lumweight.z
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	imgSize := imageSrc0Size()

	// Calculate pixel offsets
	dx := 1.0 / imgSize.x
	dy := 1.0 / imgSize.y

	// Sample 3x3 grid around current pixel
	/*
	   A B C
	   D E F
	   G H I
	*/
	A := imageSrc0At(srcPos + vec2(dx, -dy)).rgb
	B := imageSrc0At(srcPos + vec2(0.0, -dy)).rgb
	C := imageSrc0At(srcPos + vec2(-dx, -dy)).rgb
	D := imageSrc0At(srcPos + vec2(dx, 0.0)).rgb
	E := imageSrc0At(srcPos).rgb
	F := imageSrc0At(srcPos + vec2(-dx, 0.0)).rgb
	G := imageSrc0At(srcPos + vec2(dx, dy)).rgb
	H := imageSrc0At(srcPos + vec2(0.0, dy)).rgb
	I := imageSrc0At(srcPos + vec2(-dx, dy)).rgb

	// Pattern type 1 - Edge detection for antialiasing
	L := 0.0
	if D == B && B == C && E != D && B != A {
		L = 1.0
	}

	R := 0.0
	if A == B && A == F && E != F && B != C {
		R = 1.0
	}

	DL := 0.0
	if D == H && D == I && E != D && H != G {
		DL = 1.0
	}

	DR := 0.0
	if F == G && F == H && E != F && H != I {
		DR = 1.0
	}

	// Apply pattern type 1
	if (L == 1.0 && lum(E) < lum(D)) || (DL == 1.0 && lum(E) < lum(D)) {
		E = (E + D) / 2.0
	}
	if (R == 1.0 && lum(E) < lum(F)) || (DR == 1.0 && lum(E) < lum(F)) {
		E = (E + F) / 2.0
	}

	// Pattern type 2 - Different edge detection pattern
	GL := 0.0
	if E == H && E == F && E != D {
		GL = 1.0
	}

	GR := 0.0
	if E == D && E == H && E != F {
		GR = 1.0
	}

	GDL := 0.0
	if E == B && E == F && E != D {
		GDL = 1.0
	}

	GDR := 0.0
	if E == D && E == B && E != F {
		GDR = 1.0
	}

	// Apply pattern type 2
	if (GL == 1.0 && lum(E) > lum(D)) || (GDL == 1.0 && lum(E) > lum(D)) {
		E = (E + D) / 2.0
	}
	if (GR == 1.0 && lum(E) > lum(F)) || (GDR == 1.0 && lum(E) > lum(F)) {
		E = (E + F) / 2.0
	}

	// Pattern type 3 - Another edge detection pattern
	SL := 0.0
	if B == D && B == G && E != D {
		SL = 1.0
	}

	SR := 0.0
	if B == F && B == I && E != F {
		SR = 1.0
	}

	SDL := 0.0
	if H == D && H == A && E != D {
		SDL = 1.0
	}

	SDR := 0.0
	if H == F && H == C && E != F {
		SDR = 1.0
	}

	// Apply pattern type 3
	if (SL == 1.0 && lum(E) < lum(D)) || (SDL == 1.0 && lum(E) < lum(D)) {
		E = (E + D) / 2.0
	}
	if (SR == 1.0 && lum(E) < lum(F)) || (SDR == 1.0 && lum(E) < lum(F)) {
		E = (E + F) / 2.0
	}

	return vec4(E, 1.0)
}
