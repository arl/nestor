package ui

import (
	"bytes"
	"fmt"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

type romList struct {
	*Application

	selidx  int
	numcols int
	roms    []recentROM

	winw, winh int
	ui         *ebitenui.UI
	sc         *widget.ScrollContainer
}

func newRomListState(app *Application) *romList {
	state := &romList{
		Application: app,
		winw:        app.screenw,
		winh:        app.screenh,
		ui:          &ebitenui.UI{},
		roms:        loadRecentROMs(),
	}

	state.initUI()

	return state
}

func (s *romList) onClickedROM(path string) {
	modUI.InfoZ("selected ROM").String("path", path).End()
	if err := s.runRom(path); err == nil {
		s.setState("running")
	} else {
		modUI.ErrorZ("failed to run ROM").String("path", path).Error("err", err).End()
	}
}

// use a grid layout (look at the ebitenui demo example (grid layout))
func (s *romList) initUI() {
	s.ui.Container = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(&widget.Insets{}),
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch(
				// This is what contains the root container grid (1 column, 2 rows):
				// [    menu     ]
				// [ recent roms ]
				[]bool{true}, // our column stretches horizontally:
				[]bool{
					false, // the menu height cell stays fixed
					true,  // the recent roms cell streches vertically
				},
			),
		)),
	)

	// Configure menu.
	menu := newAppMenu(s.ui)
	menu.quitButton.ClickedEvent.AddHandler(func(args any) {
		s.Application.exit()
	})
	s.ui.Container.AddChild(menu.container)

	const screenshotWidth = 180 // side of the screenshot square image.
	const maxCellWidth = 200
	const cellSpacing = 5 // minimum space between cells (horizontally and vertically)

	s.numcols = (s.winw - 2*cellSpacing) / maxCellWidth

	colstretch := make([]bool, s.numcols)
	for i := range colstretch {
		colstretch[i] = true
	}

	bc := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(s.numcols),
			widget.GridLayoutOpts.Stretch(colstretch, nil),
			widget.GridLayoutOpts.Spacing(10, 10),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Top: 10, Bottom: 10,
				Left: 10, Right: 10,
			}),
		)),
	)

	for i := range s.roms {
		img := mustDecodeImage(bytes.NewReader(s.roms[i].Image))

		const frameThickness = 2
		img = fitImage(img, screenshotWidth)
		img = frameImage(img, frameThickness, colornames.Black)

		selected := i == s.selidx

		click := func() {
			s.onClickedROM(s.roms[i].Path)
		}

		cell := createROMCell(img, s.roms[i].Name, screenshotWidth, selected, click)
		bc.AddChild(cell)
	}

	scrollable := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(2, 0),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}),
		)),
	)

	s.sc = widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(bc),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff}),
			Mask: image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff}),
		}),
	)

	scrollable.AddChild(s.sc)

	pageSizeFunc := func() int {
		return int(math.Round(float64(s.sc.ViewRect().Dy())/float64(bc.GetWidget().Rect.Dy())*1000) / 3)
	}

	slider := widget.NewSlider(
		widget.SliderOpts.Orientation(widget.DirectionVertical),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.PageSizeFunc(pageSizeFunc),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			s.sc.ScrollTop = float64(args.Slider.Current) / 1000
		}),
		widget.SliderOpts.Images(
			&widget.SliderTrackImage{
				Idle:  image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
				Hover: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
			},
			&widget.ButtonImage{
				Idle:    image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
				Hover:   image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
				Pressed: image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
			},
		),
	)
	// Set the slider's position if the scrollContainer is scrolled by other means than the slider
	s.sc.GetWidget().ScrolledEvent.AddHandler(func(args any) {
		if a, ok := args.(*widget.WidgetScrolledEventArgs); ok {
			slider.Current -= int(math.Round(a.Y * float64(pageSizeFunc())))
		}
	})

	scrollable.AddChild(slider)

	s.ui.Container.AddChild(scrollable)
}

func (s *romList) Update() {
	if w, h := ebiten.WindowSize(); w != s.winw || h != s.winh {
		s.winw = w
		s.winh = h
		fmt.Println("updated window size", w, h)
		s.initUI()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		s.up()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		s.down()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		s.left()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		s.right()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		s.onClickedROM(s.roms[s.selidx].Path)
	}

	s.ui.Update()
}

func (s *romList) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Lightcoral)
	s.ui.Draw(screen)
}

/*
  0   1   2   3
  4   5   6   7
  8   9   10  11
*/

func (s *romList) up() {
	fmt.Println("up", s.selidx, "numcols", s.numcols)
	if s.selidx < s.numcols {
		return
	}
	s.selidx -= s.numcols
	s.initUI()
}

func (s *romList) down() {
	fmt.Println("down", s.selidx, "numcols", s.numcols)
	if s.selidx+s.numcols >= len(s.roms) {
		return
	}

	s.selidx += s.numcols
	s.initUI()
}

func (s *romList) left() {
	fmt.Println("left", s.selidx, "numcols", s.numcols)
	if s.selidx == 0 {
		return
	}
	s.selidx--
	s.initUI()
}

func (s *romList) right() {
	fmt.Println("right", s.selidx, "numcols", s.numcols)
	if s.selidx == len(s.roms)-1 {
		return
	}
	s.selidx++
	s.initUI()
}

func createROMCell(img *ebiten.Image, name string, side int, selected bool, click func()) *widget.Container {
	const padding = 10

	cellbg := image.NewNineSliceColor(colornames.Darkgrey)
	cellhoverbg := image.NewNineSliceColor(colornames.Lightgrey)
	cellpressedbg := image.NewNineSliceColor(colornames.Black)
	selectedbg := image.NewNineSliceColor(colornames.Blue)
	if selected {
		cellbg = selectedbg
		cellpressedbg = selectedbg
		cellhoverbg = selectedbg
	}

	var cell *widget.Container
	cell = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(cellbg),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{Top: padding}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
				cell.SetBackgroundImage(cellhoverbg)
			}),
			widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
				cell.SetBackgroundImage(cellbg)
			}),
			widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
				if args.Button == ebiten.MouseButtonLeft {
					cell.SetBackgroundImage(cellpressedbg)
				}
			}),
			widget.WidgetOpts.MouseButtonClickedHandler(func(args *widget.WidgetMouseButtonClickedEventArgs) {
				click()
			}),
		),
	)

	cell.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(img),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(
				widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
					Stretch:  false,
				},
			),
		),
	))

	cell.AddChild(widget.NewLabel(
		widget.LabelOpts.Text(name, loadFont(14), &widget.LabelColor{Idle: colornames.Black}),
		widget.LabelOpts.TextOpts(
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionEnd),
			widget.TextOpts.MaxWidth(float64(side-2*padding)),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
					Stretch:  false,
				}),
			),
		),
	))

	return cell
}
