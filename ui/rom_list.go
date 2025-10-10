package ui

import (
	"bytes"
	"fmt"

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
	numrows int

	winw, winh int
	ui         *ebitenui.UI
}

func newRomListState(app *Application) *romList {
	state := &romList{
		Application: app,
		winw:        app.screenw,
		winh:        app.screenh,
		ui:          &ebitenui.UI{},
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
				// This is what contains our grid (1 column, 2 rows):
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

	roms := loadRecentROMs()

	for i := range roms {
		// rowidx := i / s.numcols
		// if rowidx == s.numrows {
		// 	break
		// }

		img := mustDecodeImage(bytes.NewReader(roms[i].Image))

		const frameThickness = 3
		img = fitImage(img, screenshotWidth)
		img = frameImage(img, 3, colornames.Black)

		selected := i == s.selidx
		cell := createROMCell(img, roms[i].Name, screenshotWidth, selected, func() {
			s.onClickedROM(roms[i].Path)
		})
		bc.AddChild(cell)
	}

	s.ui.Container.AddChild(bc)
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
		// TODO: continuer ici
		roms := loadRecentROMs()
		s.onClickedROM(roms[s.selidx].Path)
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
	fmt.Println("up", s.selidx, "numcols", s.numcols, "numrows", s.numrows)
	if s.selidx < s.numcols {
		return
	}
	s.selidx -= s.numcols
	s.initUI()
}

func (s *romList) down() {
	fmt.Println("down", s.selidx, "numcols", s.numcols, "numrows", s.numrows)
	tot := s.numrows * s.numcols
	if s.selidx >= tot-s.numcols {
		return
	}

	s.selidx += s.numcols
	s.initUI()
}

func (s *romList) left() {
	fmt.Println("left", s.selidx, "numcols", s.numcols, "numrows", s.numrows)
	if s.selidx%s.numcols == 0 {
		return
	}
	s.selidx--
	s.initUI()
}

func (s *romList) right() {
	fmt.Println("right", s.selidx, "numcols", s.numcols, "numrows", s.numrows)
	if s.selidx%s.numcols == s.numcols-1 {
		return
	}
	s.selidx++
	s.initUI()
}

func createROMCell(img *ebiten.Image, name string, side int, selected bool, click func()) *widget.Container {
	const padding = 10

	screenshot := widget.NewGraphic(
		widget.GraphicOpts.Image(img),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(
				widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
					Stretch:  false,
				},
			),
		),
	)

	cellbg := image.NewNineSliceColor(colornames.Darkgrey)
	cellhoverbg := image.NewNineSliceColor(colornames.Lightgrey)
	cellpressedbg := image.NewNineSliceColor(colornames.Black)

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
	cell.AddChild(screenshot)

	label := widget.NewLabel(
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
	)
	cell.AddChild(label)

	return cell
}
