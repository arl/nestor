#version 330 core
out vec4 FragColor;
in vec2 TexCoord;

uniform sampler2D ourTexture;

void main() {
    vec3 color = texture(ourTexture, TexCoord).rgb;
    float scanline = sin(TexCoord.y * 1200.0) * 0.05;
    float vignette = 0.3 + 0.7 * pow(16.0 * TexCoord.x * TexCoord.y * (1.0 - TexCoord.x) * (1.0 - TexCoord.y), 0.5);
    color = color * vignette - scanline;
    FragColor = vec4(color, 1.0);
}