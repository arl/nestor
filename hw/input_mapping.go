package hw

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

func InputMappingMain() {
	// Initialize SDL
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		fmt.Printf("SDL could not initialize! SDL Error: %s\n", err)
		return
	}
	defer sdl.Quit()

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
		sdl.Delay(16) // Reduce CPU usage
	}
}
