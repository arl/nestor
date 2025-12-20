package ui

import (
	"bytes"
	"math"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"

	"nestor/config"
)

type mainState struct {
	*app

	selidx  int
	numcols int
	roms    []config.RecentROM

	sc     *widget.ScrollContainer
	slider *widget.Slider
	grid   *widget.Container
	cells  []*widget.Container

	menu *appMenu
}

func newMainState(app *app) *mainState {
	state := &mainState{app: app}
	return state
}

func (s *mainState) enter(_ any) {
	s.roms = config.LoadRecentROMs()
}

func (s *mainState) exit() {}

func (s *mainState) buildMenu() *widget.Container {
	s.menu = newAppMenu(&s.ui, s.actions, menuOptions{
		getROMName:       s.app.romName,
		settingsDisabled: false,
	})
	return s.menu.Container
}

func (s *mainState) startROM() {
	if len(s.roms) == 0 {
		return
	}
	recentrom := s.roms[s.selidx]
	if err := s.runRom(recentrom.Path, recentrom.SaveState); err == nil {
		s.setState("running", nil)
	} else {
		modUI.ErrorZ("failed to run ROM").String("path", recentrom.Path).Error("err", err).End()
	}
}

func (s *mainState) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		s.up()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		s.down()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		s.left()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		s.right()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		s.startROM()
	}

	s.ui.Update()
	s.updateSliderVisibility()
}

func (s *mainState) updateSliderVisibility() {
	if s.sc == nil || s.slider == nil || s.grid == nil {
		return
	}
	viewH := s.sc.ViewRect().Dy()
	contentH := s.grid.GetWidget().Rect.Dy()
	if contentH == 0 || viewH == 0 {
		return
	}
	if contentH <= viewH {
		s.slider.GetWidget().Visibility = widget.Visibility_Hide
	} else {
		s.slider.GetWidget().Visibility = widget.Visibility_Show
	}
}

func (s *mainState) draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

func (s *mainState) up() {
	if s.selidx < s.numcols {
		return
	}
	s.selectCell(s.selidx-s.numcols, true)
}

func (s *mainState) down() {
	if s.selidx+s.numcols >= len(s.roms) {
		return
	}
	s.selectCell(s.selidx+s.numcols, false)
}

func (s *mainState) left() {
	if s.selidx == 0 {
		return
	}
	s.selectCell(s.selidx-1, true)
}

func (s *mainState) right() {
	if s.selidx == len(s.roms)-1 {
		return
	}
	s.selectCell(s.selidx+1, false)
}

func (s *mainState) selectCell(idx int, alignTop bool) {
	if idx < 0 || idx >= len(s.cells) {
		return
	}
	s.cells[s.selidx].SetBackgroundImage(cellbg)
	s.selidx = idx
	s.cells[s.selidx].SetBackgroundImage(selectedbg)

	cellrect := s.cells[s.selidx].GetWidget().Rect
	viewrect := s.sc.ViewRect()
	if !cellrect.In(viewrect) {
		contentRect := s.sc.ContentRect()
		if alignTop {
			s.sc.ScrollTop = float64(cellrect.Min.Y-contentRect.Min.Y) / float64(contentRect.Dy()-viewrect.Dy())
		} else {
			s.sc.ScrollTop = float64(cellrect.Max.Y-contentRect.Min.Y-viewrect.Dy()) / float64(contentRect.Dy()-viewrect.Dy())
		}
		s.slider.Current = int(s.sc.ScrollTop * 1000)
	}
}

func (s *mainState) createUI() {
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

	s.ui.Container.AddChild(s.buildMenu())

	const screenshotWidth = 180 // side of the screenshot square image.
	const maxCellWidth = 200
	const cellSpacing = 5 // minimum space between cells (horizontally and vertically)

	s.numcols = (s.displayWidth - 2*cellSpacing) / maxCellWidth

	colstretch := make([]bool, s.numcols)
	for i := range colstretch {
		colstretch[i] = true
	}

	s.grid = widget.NewContainer(
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

	s.cells = s.cells[:0]
	for i := range s.roms {
		img := mustDecodeImage(bytes.NewReader(s.roms[i].Image))

		const frameThickness = 2
		img = fitImage(img, screenshotWidth)
		img = frameImage(img, frameThickness, colornames.Black)

		cell := s.createROMCell(i, img, screenshotWidth)
		s.cells = append(s.cells, cell)
		s.grid.AddChild(cell)
	}

	scrollable := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.background),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(2, 0),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}),
		)),
	)

	s.sc = widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(s.grid),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: res.background,
			Mask: res.background,
		}),
	)

	scrollable.AddChild(s.sc)

	pageSizeFunc := func() int {
		return int(math.Round(float64(s.sc.ViewRect().Dy())/float64(s.grid.GetWidget().Rect.Dy())*1000) / 3)
	}

	s.slider = widget.NewSlider(
		widget.SliderOpts.Orientation(widget.DirectionVertical),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.PageSizeFunc(pageSizeFunc),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			s.sc.ScrollTop = float64(args.Slider.Current) / 1000
		}),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
	)

	// Set the slider's position if the scrollContainer is scrolled by other
	// means than the slider.
	s.sc.GetWidget().ScrolledEvent.AddHandler(func(args any) {
		if a, ok := args.(*widget.WidgetScrolledEventArgs); ok {
			s.slider.Current -= int(math.Round(a.Y * float64(pageSizeFunc())))
		}
	})

	scrollable.AddChild(s.slider)
	s.selectCell(s.selidx, true)
	s.ui.Container.AddChild(scrollable)
}

var (
	cellbg        = image.NewNineSliceColor(colornames.Darkgrey)
	cellhoverbg   = image.NewNineSliceColor(colornames.Lightgrey)
	cellpressedbg = image.NewNineSliceColor(colornames.Black)
	selectedbg    = image.NewNineSliceColor(colornames.Blue)
)

func (s *mainState) createROMCell(idx int, img *ebiten.Image, side int) *widget.Container {
	const padding = 10

	var cell *widget.Container
	cell = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{Top: padding}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
				if idx == s.selidx {
					cell.SetBackgroundImage(selectedbg)
				} else {
					cell.SetBackgroundImage(cellhoverbg)
				}
			}),
			widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
				if idx == s.selidx {
					cell.SetBackgroundImage(selectedbg)
				} else {
					cell.SetBackgroundImage(cellbg)
				}
			}),
			widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
				if args.Button == ebiten.MouseButtonLeft {
					if idx == s.selidx {
						cell.SetBackgroundImage(selectedbg)
					} else {
						cell.SetBackgroundImage(cellpressedbg)
					}
				}
			}),
			widget.WidgetOpts.MouseButtonClickedHandler(func(args *widget.WidgetMouseButtonClickedEventArgs) {
				if args.Button == ebiten.MouseButtonLeft {
					s.selectCell(idx, true)
					s.startROM()
				}
			}),
		),
	)

	if idx == s.selidx {
		cell.SetBackgroundImage(selectedbg)
	} else {
		cell.SetBackgroundImage(cellbg)
	}

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
		widget.LabelOpts.Text(s.roms[idx].Name, res.fonts.face, &widget.LabelColor{Idle: colornames.Black}),
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
