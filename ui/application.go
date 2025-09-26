package ui

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"

	"nestor/config"
	"nestor/emu"
	"nestor/hw/input"
)

type Application struct {
	cfg config.Config

	emuctx   context.Context
	emulator *emu.Emulator
	framech  chan *emu.Frame

	states       map[string]state
	currentState state
}

func newApplication(cfg config.Config) *Application {
	app := &Application{
		cfg:    cfg,
		states: make(map[string]state),
	}

	app.states["running"] = newRunningState(app)

	return app
}

func (app *Application) setState(name string) {
	modUI.InfoZ("Switching to state").String("to", name).End()
	app.currentState = app.states[name]
}

func (app *Application) runRom(romPath string) error {
	inputProvider := input.NewProvider(app.cfg.Input)

	emulator, framech, err := runEmulator(app.cfg, inputProvider, romPath)
	if err != nil {
		return err
	}

	app.emulator = emulator
	app.framech = framech
	go emulator.Run()

	return nil
}

func (app *Application) Update() error {
	app.currentState.Update()
	return nil
}

func (app *Application) Draw(screen *ebiten.Image) {
	app.currentState.Draw(screen)
}

func (app *Application) Layout(outw, outh int) (screenw, screenh int) {
	return outw, outh
}

type state interface {
	Update()
	Draw(screen *ebiten.Image)
}

type running struct {
	*Application
	paused  bool
	pauseUI *ebitenui.UI

	clickedPause   bool
	clickedReset   bool
	clickedRestart bool
}

func newRunningState(app *Application) *running {
	s := &running{
		Application: app,
	}
	s.initUI()
	return s
}

// initUI builds the overlay shown when paused
func (s *running) initUI() {
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
		s.clickedPause = true
		// s.togglePause()
	}))
	// reset
	menu.AddChild(stdButton("Reset", func(_ *widget.ButtonClickedEventArgs) {
		s.clickedReset = true
		// s.paused = false
		// s.emulator.Reset()
	}))
	// restart
	menu.AddChild(stdButton("Restart", func(_ *widget.ButtonClickedEventArgs) {
		s.clickedRestart = true
		// s.paused = false
		// s.emulator.Restart()
	}))
	// stop
	menu.AddChild(stdButton("Stop", func(_ *widget.ButtonClickedEventArgs) {
		// s.app.StopEmulator()
		// s.app.ChangeState(s.app.GetState(StateRomList))
	}))

	root.AddChild(menu)
	s.pauseUI = &ebitenui.UI{Container: root}
}

func (s *running) togglePause() {
	s.paused = !s.paused
	modUI.InfoZ("Toggling pause").Bool("paused", s.paused).End()
	s.emulator.SetPause(s.paused)
}

func (s *running) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		s.emulator.Reset()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || s.clickedPause {
		s.clickedPause = false
		s.togglePause()
	}
	if s.clickedReset {
		s.clickedReset = false
		s.togglePause()
		s.emulator.Reset()
	}
	if s.clickedRestart {
		s.clickedRestart = false
		s.togglePause()
		s.emulator.Restart()
	}

	if s.paused {
		s.pauseUI.Update()
	}
}

func (s *running) Draw(screen *ebiten.Image) {

	select {
	case frame := <-s.framech:
		ptr := unsafe.SliceData(frame.Video)
		fmt.Printf("Frame video ptr: %p\n", ptr)
		img := ImageFromFrame(frame)
		s.drawFrame(screen, img, float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy()))
	default:
	}

	if s.paused {
		s.pauseUI.Draw(screen)
	}

	// print debug info: fps
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %f", ebiten.ActualFPS()))

	// default:
	// }
}

func (s *running) drawFrame(screen *ebiten.Image, frameImg *ebiten.Image, targetW, targetH float64) {
	// Calculate scaling to fit the target area while preserving aspect ratio.
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
