package ui

import (
	"fmt"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

type romList struct {
	*Application

	rrw      *recentRomsWidget
	selected int

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

	state.rrw = newRecentROMsWidget(app.screenw, app.screenh, 0, func(path string) {
		state.onClickedROM(path)
	})

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
	// Create a root container for the UI.
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

	s.rrw.recreateUI(s.winw, s.winh, s.selected)
	s.ui.Container.AddChild(s.rrw.container)
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
		s.onClickedROM(roms[s.selected].Path)
	}

	s.ui.Update()
}

func (s *romList) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Lightcoral)
	s.ui.Draw(screen)
}

func (s *romList) up() {
	fmt.Println("up", s.selected, "numcols", s.rrw.numcols, "numrows", s.rrw.numrows)
	if s.selected < s.rrw.numcols {
		return
	}
	s.selected -= s.rrw.numcols
	s.initUI()
}

/*
  0   1   2   3
  4   5   6   7
  8   9   10  11
*/

func (s *romList) down() {
	fmt.Println("down", s.selected, "numcols", s.rrw.numcols, "numrows", s.rrw.numrows)
	tot := s.rrw.numrows * s.rrw.numcols
	if s.selected >= tot-s.rrw.numcols {
		return
	}

	s.selected += s.rrw.numcols
	s.initUI()
}

func (s *romList) left() {
	fmt.Println("left", s.selected, "numcols", s.rrw.numcols, "numrows", s.rrw.numrows)
	if s.selected%s.rrw.numcols == 0 {
		return
	}
	s.selected--
	s.initUI()
}

func (s *romList) right() {
	fmt.Println("right", s.selected, "numcols", s.rrw.numcols, "numrows", s.rrw.numrows)
	if s.selected%s.rrw.numcols == s.rrw.numcols-1 {
		return
	}
	s.selected++
	s.initUI()
}
