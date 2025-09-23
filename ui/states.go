package ui

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"

	"nestor/config"
	"nestor/emu"
	"nestor/hw/input"
	"nestor/ines"
)

type AppState int

const (
	RomList AppState = 1 + iota
	RomRunning
	RomPaused
)

type App struct {
	state         AppState
	ui            *ebitenui.UI
	emulator      *emu.Emulator
	framech       chan *emu.Frame
	input         *input.Provider
	fullScreen    bool
	cfg           config.Config
	leftPanel     *widget.Container
	rightPanel    *widget.Container
	centerPanel   *widget.Container
	rootContainer *widget.Container
}

func NewApp(cfg config.Config) *App {
	app := &App{
		cfg:        cfg,
		input:      input.NewProvider(cfg.Input),
		fullScreen: true, // Start in fullscreen mode
		state:      RomList,
	}

	// Create panels
	app.leftPanel = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Indianred),
		),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				StretchVertical:    true,
			}),
			widget.WidgetOpts.MinSize(50, 50),
		),
	)

	app.rightPanel = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Mediumseagreen),
		),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				StretchVertical:    true,
			}),
			widget.WidgetOpts.MinSize(50, 50),
		),
	)

	// Create center panel with start button
	app.centerPanel = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Darkgray),
		),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	font := DefaultFont()

	// Create start button
	startButton := widget.NewButton(
		widget.ButtonOpts.TextFace(&font),
		widget.ButtonOpts.TextColor(&widget.ButtonTextColor{
			Idle:    colornames.Gainsboro,
			Hover:   colornames.Gainsboro,
			Pressed: Mix(colornames.Gainsboro, colornames.Black, 0.4),
		}),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:         DefaultNineSlice(colornames.Darkslategray),
			Hover:        DefaultNineSlice(Mix(colornames.Darkslategray, colornames.Mediumseagreen, 0.4)),
			Disabled:     DefaultNineSlice(Mix(colornames.Darkslategray, colornames.Gainsboro, 0.8)),
			Pressed:      PressedNineSlice(Mix(colornames.Darkslategray, colornames.Black, 0.4)),
			PressedHover: PressedNineSlice(Mix(colornames.Darkslategray, colornames.Black, 0.4)),
		}),
		widget.ButtonOpts.TextLabel("Start"),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(180, 48),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			const romPath = "/home/aurelien/dev/roms/nes/all.nes.roms.goodnes/USA/Super Mario Bros. + Duck Hunt (U) [!].nes"
			app.startEmulator(romPath)
		}),
	)

	app.centerPanel.AddChild(startButton)

	// Root container
	app.rootContainer = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(colornames.Gainsboro),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	app.rootContainer.AddChild(app.leftPanel)
	app.rootContainer.AddChild(app.rightPanel)
	app.rootContainer.AddChild(app.centerPanel)

	app.ui = &ebitenui.UI{Container: app.rootContainer}

	return app
}

func (app *App) startEmulator(romPath string) {
	rom, err := ines.ReadROM(romPath)
	if err != nil {
		// In a real app, you'd want to show this error in the UI
		return
	}

	app.framech = make(chan *emu.Frame)
	out := emu.NewOutput(app.framech,
		emu.OutputConfig{
			Width:          emu.NTSCWidth,
			Height:         emu.NTSCHeight,
			NumBackBuffers: 4,
			Title:          "Nestor",
			ScaleFactor:    2,
			DisableVSync:   app.cfg.Video.DisableVSync,
			Monitor:        app.cfg.Video.Monitor,
			Shader:         app.cfg.Video.Shader,
		},
	)

	emulator, err := emu.Launch(rom, app.cfg.Config, out, app.input)
	if err != nil {
		// In a real app, you'd want to show this error in the UI
		return
	}

	app.emulator = emulator
	app.state = RomRunning

	go emulator.Run()
}

func (app *App) State() AppState {
	return app.state
}

func (app *App) Update() error {
	// Handle F key to toggle fullscreen mode
	if app.state == RomRunning && inpututil.IsKeyJustPressed(ebiten.KeyF) {
		app.fullScreen = !app.fullScreen
	}

	// Only update UI if not in fullscreen mode or if emulator isn't running
	if !app.fullScreen || app.state != RomRunning {
		app.ui.Update()
	}

	return nil
}

func (app *App) Draw(screen *ebiten.Image) {
	if app.state == RomRunning {
		// Check if there's a frame to draw
		select {
		case frame := <-app.framech:
			frameImg := ImageFromFrame(frame)

			if app.fullScreen {
				// Draw in fullscreen
				w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
				app.drawFrame(screen, frameImg, float64(w), float64(h))
			} else {
				// Draw in center panel only
				app.ui.Draw(screen) // Draw UI first

				// Get center panel dimensions from layout
				bounds := app.centerPanel.GetWidget().Rect
				centerW := float64(bounds.Dx())
				centerH := float64(bounds.Dy())

				// Create a sub-screen for the center panel
				subScreen := screen.SubImage(bounds).(*ebiten.Image)
				app.drawFrame(subScreen, frameImg, centerW, centerH)
			}
			return
		default:
			// No frame available, just draw UI
		}
	}

	// Default drawing of UI
	app.ui.Draw(screen)
}

func (app *App) drawFrame(screen *ebiten.Image, frameImg *ebiten.Image, targetW, targetH float64) {
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

func (app *App) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}
