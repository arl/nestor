/*
	crt-beam
	for best results use integer scale 5x or more
*/

// #pragma parameter blur "Horizontal Blur/Beam shape" 0.6 0.0 1.0 0.1
// #pragma parameter Scanline "Scanline thickness" 0.2 0.0 1.0 0.05
// #pragma parameter weightr "Scanline Red brightness" 0.8 0.0 1.0 0.05
// #pragma parameter weightg "Scanline Green brightness" 0.8 0.0 1.0 0.05
// #pragma parameter weightb "Scanline Blue brightness" 0.8 0.0 1.0 0.05
// #pragma parameter bogus_msk " [ MASKS ] " 0.0 0.0 0.0 0.0
// #pragma parameter mask "Mask 0:CGWG,1-2:Lottes,3-4 Gray,5-6:CGWG slot,7 VGA" 3.0 -1.0 7.0 1.0
// #pragma parameter msk_size "Mask size" 1.0 1.0 2.0 1.0
// #pragma parameter scale "VGA Mask Vertical Scale" 2.0 2.00 10.00 1.0
// #pragma parameter MaskDark "Lottes Mask Dark" 0.7 0.00 2.00 0.10
// #pragma parameter MaskLight "Lottes Mask Light" 1.0 0.00 2.00 0.10
// #pragma parameter bogus_col " [ COLOR ] " 0.0 0.0 0.0 0.0
// #pragma parameter sat "Saturation" 1.0 0.00 2.00 0.05
// #pragma parameter bright "Boost bright" 1.0 1.00 2.00 0.05
// #pragma parameter dark "Boost dark" 1.45 1.00 2.00 0.05
// #pragma parameter glow "Glow Strength" 0.08 0.0 0.5 0.01


#define pi 3.14159


uniform vec2 TextureSize;
varying vec2 TEX0;
varying vec2 fragpos;

uniform mat4 MVPMatrix;
attribute vec4 VertexCoord;
attribute vec2 TexCoord;
uniform vec2 InputSize;
uniform vec2 OutputSize;

void main()
{
	TEX0 = TexCoord*1.0001;                    
	gl_Position = MVPMatrix * VertexCoord;  
	fragpos = TEX0.xy*OutputSize.xy*TextureSize.xy/InputSize.xy;   
}
