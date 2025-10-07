package ui

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
)

type romList struct {
	*Application

	winw, winh int
	ui         *ebitenui.UI
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
	s.ui = &ebitenui.UI{}

	// Create a root container for the UI.
	s.ui.Container = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(&widget.Insets{}),
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
		)))

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
		s.initUI()
	}
	s.ui.Update()
}

func (s *romList) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Lightcoral)
	s.ui.Draw(screen)
}
