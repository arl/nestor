package ui

import (
	"fmt"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
)

type romList struct {
	*Application

	rrw *recentRomsWidget

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

	state.rrw = newRecentROMsWidget(app.screenw, app.screenh, func(path string) {
		state.onClickedROM(path)
	})

	state.initUI()

	return state
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

	fmt.Println("window size:", s.winw, s.winh)

	s.ui.Container.AddChild(s.rrw.container)
	s.rrw.initUI(s.winw, s.winh)
}

func (s *romList) onClickedROM(path string) {
	modUI.InfoZ("selected ROM").String("path", path).End()
	if err := s.Application.runRom(path); err == nil {
		s.Application.setState("running")
	} else {
		modUI.ErrorZ("failed to run ROM").String("path", path).Error("err", err).End()
	}
}

func (s *romList) Update() {
	if w, h := ebiten.WindowSize(); w != s.winw || h != s.winh {
		s.winw = w
		s.winh = h
		fmt.Println("window resized:", s.winw, s.winh)
		s.ui.Container.RequestRelayout()
	}

	s.ui.Update()
}

func (s *romList) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Lightcoral)
	s.ui.Draw(screen)
}
