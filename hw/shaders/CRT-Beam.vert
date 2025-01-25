/*
	crt-beam
	for best results use integer scale 5x or more
*/

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
