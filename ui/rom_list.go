package ui

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type romList struct {
	*Application

	winw, winh int
	ui         *ebitenui.UI
	root       *widget.Container
	rrw        *recentRomsWidget
}

func newRomListState(app *Application) *romList {
	winw, _ := ebiten.WindowSize()
	state := &romList{
		Application: app,
		winw:        winw,
	}

	state.initUI()

	return state
}

// use a grid layout (look at the ebitenui demo example (grid layout))
func (s *romList) initUI() {
	rrw := newRecentROMsWidget(s.winw, s.winh, func(path string) {
		s.onClickedROM(path)
	})

	// Create a root container for the UI.
	s.root = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:  25,
				Right: 25,
			}),
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(20, 0),
		)))
	s.root.AddChild(rrw.container)
	s.ui = &ebitenui.UI{Container: s.root}
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
		s.initUI()
	}
	s.ui.Update()
}

func (s *romList) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}
