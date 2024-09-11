package hw

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	"nestor/resource"
)

/*
TODO: write a function that returns error and returns everything to stderr
actual output is written to stdout
*/

func InputMappingMain() {
	// Initialize SDL
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		fmt.Printf("SDL could not initialize! SDL Error: %s\n", err)
		return
	}
	defer sdl.Quit()

	// Initialize SDL_ttf for text rendering
	if err := ttf.Init(); err != nil {
		fmt.Printf("SDL_ttf could not initialize! SDL Error: %s\n", err)
		return
	}
	defer ttf.Quit()

	fmt.Println("SDL initialized")

	// Create a small rectangular window (e.g., 300x200)
	window, err := sdl.CreateWindow(
		"Input Capture Window", // Title of the window
		sdl.WINDOWPOS_CENTERED, // Center window on screen
		sdl.WINDOWPOS_CENTERED, // Center window on screen
		300,                    // Window width
		200,                    // Window height
		sdl.WINDOW_SHOWN,       // Window is shown when created
	)
	if err != nil {
		fmt.Printf("Window could not be created! SDL Error: %s\n", err)
		return
	}
	defer window.Destroy()

	// Create an SDL renderer
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Printf("Renderer could not be created! SDL Error: %s\n", err)
		return
	}
	defer renderer.Destroy()

	// Load the embedded font using RWops from the embedded data
	rwops, err := sdl.RWFromMem(resource.DejaVuSansFont)
	if err != nil {
		fmt.Printf("Failed to create RWops from embedded font! SDL Error: %s\n", err)
		return
	}

	// Load the font from memory (using ttf.OpenFontRW)
	const freeRWops = 1
	font, err := ttf.OpenFontRW(rwops, freeRWops, 16)
	if err != nil {
		fmt.Printf("Failed to load font from memory! SDL_ttf Error: %s\n", err)
		return
	}
	defer font.Close()
	// Text message to display
	message := "Press key or joystick button corresponding to BLABLA"

	// Create a surface with the text
	color := sdl.Color{R: 255, G: 255, B: 255, A: 255} // White color
	surface, err := font.RenderUTF8Solid(message, color)
	if err != nil {
		fmt.Printf("Unable to render text! SDL_ttf Error: %s\n", err)
		return
	}
	defer surface.Free()

	// Create a texture from the surface
	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		fmt.Printf("Unable to create texture from surface! SDL Error: %s\n", err)
		return
	}
	defer texture.Destroy()

	// Get the width and height of the texture (which matches the text size)
	_, _, textw, texth, err := texture.Query()
	if err != nil {
		fmt.Printf("Unable to query texture! SDL Error: %s\n", err)
		return
	}

	// Calculate position to center the text
	windowWidth, windowHeight := window.GetSize()
	textRect := sdl.Rect{
		X: (windowWidth - textw) / 2,
		Y: (windowHeight - texth) / 2,
		W: textw,
		H: texth,
	}

	fmt.Println("input loop")
	for {
		// Poll for SDL events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.KeyboardEvent:
				if e.State == sdl.PRESSED {
					// Send the pressed key to the main SDL application
					message := fmt.Sprintf("Key pressed: %s\n", sdl.GetScancodeName(e.Keysym.Scancode))
					os.Stdout.Write([]byte(message))
				}
			}
		}

		// Clear the screen
		renderer.SetDrawColor(0, 0, 0, 255) // Black background
		renderer.Clear()

		// Render the text in the center of the screen
		renderer.Copy(texture, nil, &textRect)

		// Present the updated window
		renderer.Present()

		sdl.Delay(16) // Reduce CPU usage (~60 FPS)
	}
}
