/* 
 * gizmo98 crt shader
 * Copyright (C) 2023 gizmo98
 *
 *   This program is free software; you can redistribute it and/or modify it
 *   under the terms of the GNU General Public License as published by the Free
 *   Software Foundation; either version 2 of the License, or (at your option)
 *   any later version.
 *
 * version 0.41, 01.05.2023
 * ---------------------------------------------------------------------------------------
 * - add more references for used sources 
 *
 * version 0.40, 29.04.2023
 * ---------------------------------------------------------------------------------------
 * - fix aspect ratio issue 
 * - fix screen centering issue
 * - use CRT/PI curvator
 * - add noise intensity value
 *
 * version 0.35, 29.04.2023
 * ---------------------------------------------------------------------------------------
 * - initial slang port
 * - remove NTSC and INTERLACE effects
 * 
 * version 0.3, 28.04.2023
 * ---------------------------------------------------------------------------------------
 * - unify shader in one file
 * - replace fixed macros and defines with pragmas
 * - add BLUR_OFFSET setting. This setting can be used to set the strength of a bad signal
 * - add ANAMORPH setting for megadrive and snes
 * 
 * https://github.com/gizmo98/gizmo-crt-shader
 *
 * This shader tries to mimic a CRT without extensive use of scanlines and rgb pattern emulation.
 * It uses horizontal subpixel scaling and adds brightness dependent scanline patterns and allows 
 * fractional scaling. 
 *
 * HORIZONTAL_BLUR simulates a bad composite signal which is neede for consoles like megadrive 
 * VERTICAL_BLUR vertical blur simulates N64 vertical blur 
 * BGR_LCD_PATTERN most LCDs have a RGB pixel pattern. Enable BGR pattern with this switch
 * BRIGHTNESS makes scanlines more or less visible
 * SHRINK scale screen in X direction
 * SNR noise intensity 
 *
 * uses parts curvator of CRT-PI shader from davej https://github.com/libretro/glsl-shaders/blob/master/crt/shaders/crt-pi.glsl
 * uses parts of texture anti-aliasing shader from Ikaros https://www.shadertoy.com/view/ldsSRX
 * uses gold noise shader from dcerisano https://www.shadertoy.com/view/ltB3zD
 */

attribute vec4 VertexCoord;
attribute vec4 COLOR;
attribute vec4 TexCoord;
varying vec4 COL0;
varying vec4 TEX0;
varying vec2 screenScale;

uniform mat4 MVPMatrix;
uniform int FrameDirection;
uniform int FrameCount;
uniform vec2 OutputSize;
uniform vec2 TextureSize;
uniform vec2 InputSize;

#define CURVATURE_X 0.1
#define CURVATURE_Y 0.15
#define BRIGHTNESS 0.5
#define HORIZONTAL_BLUR 0.0
#define VERTICAL_BLUR 0.0
#define BLUR_OFFSET 0.5
#define BGR_LCD_PATTERN 0.0
#define SHRINK 0.0
#define SNR 1.0

void main()
{
    screenScale = TextureSize / InputSize;
    gl_Position = VertexCoord.x * MVPMatrix[0] + VertexCoord.y * MVPMatrix[1] + VertexCoord.z * MVPMatrix[2] + VertexCoord.w * MVPMatrix[3];
    TEX0.xy = TexCoord.xy;
}