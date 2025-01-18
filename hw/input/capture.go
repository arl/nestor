package input

import (
	"fmt"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	"nestor/emu/log"
	"nestor/resource"
)

// Capture waits for a next key or joystick button press and
// returns a MappingCodecode identifying it, or "" if the user pressed Escape.
func Capture(padbtn string) (Code, error) {
	var code Code

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_GAMECONTROLLER); err != nil {
		return code, fmt.Errorf("failed to initialize SDL: %s", err)
	}

	if err := ttf.Init(); err != nil {
		return code, fmt.Errorf("failed to initialize SDL_ttf: %s", err)
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
		return code, fmt.Errorf("failed to create window: %s", err)
	}
	defer win.Destroy()

	renderer, err := sdl.CreateRenderer(win, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return code, fmt.Errorf("failed to create renderer: %s", err)
	}
	defer renderer.Destroy()

	font, err := fontFromMem(resource.DejaVuSansFont)
	if err != nil {
		return code, fmt.Errorf("failed to load font: %s", err)
	}

	const message = "Press key or joystick button to assign to"
	const maxwidth = 380
	lines := wrapText(font, message, maxwidth)
	lines = append(lines, "")
	lines = append(lines, padbtn)

	gamectrls := NewGameControllers()

	// Drain the events queue before starting. This removes previous events
	// which could have been generated during the release of a joystick trigger
	// for example.
	drainEvents(200 * time.Millisecond)
pollLoop:
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case sdl.QuitEvent:
				break pollLoop

			case sdl.KeyboardEvent:
				if e.State == sdl.PRESSED {
					if e.Keysym.Scancode != sdl.SCANCODE_ESCAPE {
						code.Type = Keyboard
						code.Scancode = e.Keysym.Scancode
					}
					break pollLoop
				}
			case sdl.ControllerDeviceEvent:
				gamectrls.UpdateDevices(e)

			case sdl.ControllerButtonEvent:
				ctrl := gamectrls.Get(e.Which)
				if ctrl == nil {
					log.ModInput.Fatalf("controller %d not found", e.Which)
				}

				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					code.Type = ControllerButton
					code.CtrlButton = e.Button
					code.CtrlGUID = gamectrls.GetGUID(e.Which)
					break pollLoop
				}

			case sdl.ControllerAxisEvent:
				if gamectrls.Get(e.Which) == nil {
					log.ModInput.Fatalf("controller %d not found", e.Which)
				}

				if e.Value < -JoyAxisThreshold || e.Value > JoyAxisThreshold {
					code.Type = ControllerAxis
					code.CtrlAxis = e.Axis
					code.CtrlAxisDir = axissign(e.Value)
					code.CtrlGUID = gamectrls.GetGUID(e.Which)
					break pollLoop
				}
			}
		}

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		winw, _ := win.GetSize()
		col := sdl.Color{R: 255, G: 255, B: 255, A: 255}
		if err := renderText(renderer, font, lines, col, 50, winw); err != nil {
			return code, fmt.Errorf("renderText error: %s", err)
		}

		renderer.Present()
	}

	gamectrls.Close()

	return code, nil
}

// Drain the events queue before exiting. But since some joystick axes are
// noisy, wait just long enough to drain 'actual' events, like for example
// the events generated when releasing a joystick trigger.
func drainEvents(maxwait time.Duration) {
	deadline := time.Now().Add(maxwait)
	for {
		if event := sdl.PollEvent(); event == nil {
			break
		}
		if time.Now().After(deadline) {
			break
		}
	}
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
