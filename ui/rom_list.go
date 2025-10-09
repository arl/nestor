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

	const screenshotSize = 200 // side of the screenshot square image.
	const cellSpacing = 5      // minimum space between cells (horizontally and vertically)

	s.numcols = s.winw / (screenshotSize + cellSpacing)
	s.numrows = s.winh / (screenshotSize + cellSpacing) // max rows to display
	if s.numrows == 0 {
		s.numrows = 1
	}

	vertSpacing := (s.winh - (s.numrows * screenshotSize)) / (s.numcols + 1)

	colstretch := make([]bool, s.numcols)
	for i := range colstretch {
		colstretch[i] = true
	}
	rowstrech := make([]bool, s.numrows)
	for i := range rowstrech {
		rowstrech[i] = true
	}

	bc := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(s.numcols),
			widget.GridLayoutOpts.Stretch(colstretch, rowstrech),
			widget.GridLayoutOpts.Spacing(10, vertSpacing),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Top: 10, Bottom: 10,
				Left: 10, Right: 10,
			}),
		)),
	)

	roms := loadRecentROMs()

	for i := range roms {
		rowidx := i / s.numcols
		if rowidx == s.numrows {
			break
		}

		img := mustDecodeImage(bytes.NewReader(roms[i].Image))
		selected := i == s.selidx
		cell := createROMCell(img, roms[i].Name, screenshotSize, selected, func() {
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

func (s *romList) up() {
	fmt.Println("up", s.selidx, "numcols", s.numcols, "numrows", s.numrows)
	if s.selidx < s.numcols {
		return
	}
	s.selidx -= s.numcols
	s.initUI()
}

/*
  0   1   2   3
  4   5   6   7
  8   9   10  11
*/

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
	const frameThickness = 2
	const padding = 20
	const labelHeight = 40

	imageSide := side - (labelHeight + padding*2)

	img = resizeImage(img, float64(imageSide), float64(imageSide))
	btnimg := &widget.ButtonImage{
		Idle:    image.NewFixedNineSlice(img),
		Pressed: image.NewFixedNineSlice(frameImage(img, frameThickness, colornames.Black)),
		Hover:   image.NewFixedNineSlice(frameImage(img, frameThickness, colornames.Darkgray)),
	}

	btn := widget.NewButton(
		widget.ButtonOpts.Image(btnimg),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			click()
		}),
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(
			widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding:            &widget.Insets{Top: padding},
			},
		)),
	)

	bgcol := colornames.Lightcyan
	if selected {
		bgcol = colornames.Yellow
	}
	cell := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(bgcol)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	cell.AddChild(btn)

	label := widget.NewLabel(
		widget.LabelOpts.Text(name, loadFont(14), &widget.LabelColor{Idle: colornames.Black}),
		widget.LabelOpts.TextOpts(
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionEnd),
			widget.TextOpts.MaxWidth(float64(side-2*padding)),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionEnd,
					Padding: &widget.Insets{
						Bottom: padding,
						Left:   padding,
					},
					StretchHorizontal: true,
				}),
			),
		),
	)
	cell.AddChild(label)

	return cell
}
