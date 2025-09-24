package ui

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

// RunningState represents the state when a ROM is running
type RunningState struct {
	app        *App
	fullScreen bool

	// pause fields
	paused    bool
	lastFrame *ebiten.Image
	pauseUI   *ebitenui.UI
}

// NewRunningState creates a new running state
func NewRunningState(app *App) *RunningState {
	s := &RunningState{
		app:        app,
		fullScreen: true,
	}
	s.initPauseUI()
	return s
}

// initPauseUI builds the overlay shown when paused
func (s *RunningState) initPauseUI() {
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(Mix(colornames.Black, Transparent, 0.5)),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	menu := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Gray),
		),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(&widget.Insets{Top: 20, Left: 20, Right: 20, Bottom: 20}),
			widget.GridLayoutOpts.Spacing(10, 10),
			widget.GridLayoutOpts.Columns(1),
		)),
	)
	menu.AddChild(widget.NewLabel(
		widget.LabelOpts.Text("PAUSED", TitleFont(), &widget.LabelColor{Idle: colornames.White}),
	))
	// resume
	menu.AddChild(stdButton("Resume", func(_ *widget.ButtonClickedEventArgs) {
		s.paused = false
	}))
	// reset
	menu.AddChild(stdButton("Reset", func(_ *widget.ButtonClickedEventArgs) {
		s.app.ResetEmulator()
		s.paused = false
	}))
	// restart
	menu.AddChild(stdButton("Restart", func(_ *widget.ButtonClickedEventArgs) {
		s.app.RestartEmulator()
		s.paused = false
	}))
	// stop
	menu.AddChild(stdButton("Stop", func(_ *widget.ButtonClickedEventArgs) {
		s.app.StopEmulator()
		s.app.ChangeState(s.app.GetState(StateRomList))
	}))

	root.AddChild(menu)
	s.pauseUI = &ebitenui.UI{Container: root}
}

// Enter implements State interface
func (s *RunningState) Enter(prevState State) {
	s.paused = false
}

// Exit implements State interface
func (s *RunningState) Exit(nextState State) {
}

// Update implements State interface
func (s *RunningState) Update() {
	// Handle F key to toggle fullscreen mode
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		s.fullScreen = !s.fullScreen
	}

	// Handle Escape key to pause
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.paused = !s.paused
	}

	// Update pause UI so buttons receive events
	if s.paused {
		s.pauseUI.Update()
	}
}

// Draw implements State interface
func (s *RunningState) Draw(screen *ebiten.Image) {
	// fetch frame if not paused
	if !s.paused {
		select {
		case frame := <-s.app.GetFrameChannel():
			s.lastFrame = ImageFromFrame(frame)
			// default:
		}
	}
	if s.lastFrame != nil {
		// draw viewport
		w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
		s.drawFrame(screen, s.lastFrame, float64(w), float64(h))
	}
	// draw pause overlay if needed
	if s.paused {
		s.pauseUI.Draw(screen)
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
