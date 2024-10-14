package hw

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	"nestor/resource"
)

const mapInputNone = "none"

// AskForKeybding shows the user a small window,
// capturing the next key or joystick press.
func AskForKeybding(btnName string) (string, error) {
	cmd := exec.Command(os.Args[0], "map-input", "--button="+btnName)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func MapInputMain(btnname string) {
	pressed, err := runMapInput(btnname)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if pressed == mapInputNone {
		os.Exit(0)
	}

	fmt.Println(pressed)
	os.Exit(0)
}

func runMapInput(btnName string) (pressed string, err error) {
	pressed = mapInputNone

	// Initialize SDL
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return pressed, fmt.Errorf("failed to initialize SDL: %s", err)
	}
	defer sdl.Quit()

	// Initialize SDL_ttf for text rendering
	if err := ttf.Init(); err != nil {
		return pressed, fmt.Errorf("failed to initialize SDL_ttf: %s", err)
	}
	defer ttf.Quit()

	// Create a small rectangular win (e.g., 400x300)
	win, err := sdl.CreateWindow(
		"Nestor input capture",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		400,
		300,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return pressed, fmt.Errorf("failed to create window: %s", err)
	}
	defer win.Destroy()

	renderer, err := sdl.CreateRenderer(win, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return pressed, fmt.Errorf("failed to create renderer: %s", err)
	}
	defer renderer.Destroy()

	font, err := fontFromMem(resource.DejaVuSansFont)
	if err != nil {
		return pressed, fmt.Errorf("failed to load font: %s", err)
	}

	message := "Press key or joystick button for:"

	// Function to split the text into lines that fit within the given width
	lines := wrapText(font, message, 380) // 380 is the maximum width of each line
	lines = append(lines, "")
	lines = append(lines, btnName)

	for quit := false; !quit; {
		// Poll for SDL events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				quit = true
			case *sdl.KeyboardEvent:
				if e.State == sdl.PRESSED {
					if e.Keysym.Scancode != sdl.SCANCODE_ESCAPE {
						pressed = sdl.GetScancodeName(e.Keysym.Scancode)
					}
					quit = true
				}
			}
		}

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		// Render the text line by line
		winw, _ := win.GetSize()
		col := sdl.Color{R: 255, G: 255, B: 255, A: 255}
		if err := renderText(renderer, font, lines, col, 50, winw); err != nil {
			return pressed, fmt.Errorf("renderText error: %s", err)
		}

		renderer.Present()
		sdl.Delay(16) // max out at 60fps
	}

	return pressed, nil
}

func fontFromMem(data []byte) (*ttf.Font, error) {
	rwops, err := sdl.RWFromMem(data)
	if err != nil {
		return nil, fmt.Errorf("ttf.RWFromMem error: %s", err)
	}

	font, err := ttf.OpenFontRW(rwops, 1, 18) // 18 is the font size
	if err != nil {
		return nil, fmt.Errorf("ttf.OpenFontRW error: %s", err)
	}
	return font, nil

}

// wrapText splits the text into multiple lines based on the max width
func wrapText(font *ttf.Font, text string, maxw int) []string {
	words := strings.Split(text, " ")
	var lines []string

	curline := ""
	for _, word := range words {
		if curline == "" {
			curline = word
			continue
		}

		w, _, _ := font.SizeUTF8(curline + " " + word)
		if w > maxw {
			// Start a new line.
			lines = append(lines, curline)
			curline = word
		} else {
			curline += " " + word
		}
	}

	if curline != "" {
		lines = append(lines, curline)
	}

	return lines
}

// renderText renders each line of text at the given y position, centering each line.
func renderText(renderer *sdl.Renderer, font *ttf.Font, lines []string, color sdl.Color, y int32, winw int32) error {
	vspacing := int32(font.LineSkip())

	for _, line := range lines {
		if line == "" {
			y += vspacing
			continue
		}
		surface, err := font.RenderUTF8Blended(line, color)
		if err != nil {
			return fmt.Errorf("RenderUTF8Blended error: %s", err)
		}
		defer surface.Free()

		texture, err := renderer.CreateTextureFromSurface(surface)
		if err != nil {
			return fmt.Errorf("CreateTextureFromSurface error: %s", err)
		}
		defer texture.Destroy()

		_, _, texw, texh, _ := texture.Query()
		centeredX := (winw - texw) / 2
		rect := &sdl.Rect{
			X: centeredX, Y: y,
			W: int32(texw), H: int32(texh),
		}
		renderer.Copy(texture, nil, rect)

		y += vspacing
	}
	return nil
}
