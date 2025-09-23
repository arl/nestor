package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// RunningState represents the state when a ROM is running
type RunningState struct {
	app        *App
	fullScreen bool
}

// NewRunningState creates a new running state
func NewRunningState(app *App) *RunningState {
	return &RunningState{
		app:        app,
		fullScreen: true,
	}
}

// Enter implements State interface
func (s *RunningState) Enter(prevState State) {
	// Nothing special to do on enter
}

// Exit implements State interface
func (s *RunningState) Exit(nextState State) {
	// Nothing special to do on exit
}

// Update implements State interface
func (s *RunningState) Update() {
	// Handle F key to toggle fullscreen mode
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		s.fullScreen = !s.fullScreen
	}

	// Handle Escape key to pause
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.app.ChangeState(s.app.GetState(StateRomPaused))
	}
}

// Draw implements State interface
func (s *RunningState) Draw(screen *ebiten.Image) {
	// Get frame from emulator
	select {
	case frame := <-s.app.GetFrameChannel():
		frameImg := ImageFromFrame(frame)

		// Draw in fullscreen
		w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
		s.drawFrame(screen, frameImg, float64(w), float64(h))
	default:
		// No frame available
	}
}

// drawFrame draws the emulator frame with proper scaling
func (s *RunningState) drawFrame(screen *ebiten.Image, frameImg *ebiten.Image, targetW, targetH float64) {
	// Calculate scaling to fit the target area while preserving aspect ratio
	fw, fh := float64(frameImg.Bounds().Dx()), float64(frameImg.Bounds().Dy())
	scaleX := targetW / fw
	scaleY := targetH / fh
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Draw the frame centered
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate((targetW-fw*scale)/2, (targetH-fh*scale)/2)
	screen.DrawImage(frameImg, op)
}

// ID implements State interface
func (s *RunningState) ID() StateID {
	return StateRomRunning
}
