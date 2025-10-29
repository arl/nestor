package ui

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"

	"nestor/config"
	"nestor/hw/input"
)

type configState struct {
	*app
}

func newConfigState(app *app) *configState {
	state := &configState{
		app: app,
	}

	state.createUI()
	return state
}

func (s *configState) enter(args ...any) {
	if len(args) == 0 {
		return
	}

	btn := args[0].(input.PaddleButton)
	idxpreset := args[1].(int)
	code := args[2].(*input.Code)
	if code == nil {
		// do nothing
		return
	}

	s.app.cfg.Input.Presets[idxpreset].AssignCode(btn, *code)

	modUI.InfoZ("Config state entered with args").String("args", fmt.Sprintf("%+v", args)).End()
}

func (s *configState) exit() {
	if err := config.Save(&s.app.cfg); err != nil {
		modUI.ErrorZ("failed to save config").Error("err", err).End()
		return
	}

	modUI.InfoZ("config saved").String("path", config.Path()).End()
}

func (s *configState) update() {
	s.ui.Update()
}

func (s *configState) draw(screen *ebiten.Image) {
	screen.Fill(colornames.Lightcoral)
	s.ui.Draw(screen)
}

func (s *configState) createUI() {
	// root of the whole config UI.
	root := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.TrackHover(false)),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true, false}),
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Top:    20,
				Bottom: 20,
			}),
			widget.GridLayoutOpts.Spacing(0, 20))),
		widget.ContainerOpts.BackgroundImage(res.background))

	// container for page list and page content.
	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:  25,
				Right: 25,
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(20, 0))))

	pages := []any{
		s.inputConfigPage(),
		s.videoConfigPage(),
	}

	pageContainer := newPageContainer()

	pageList := widget.NewList(
		widget.ListOpts.Entries(pages),
		widget.ListOpts.EntryLabelFunc(func(e any) string {
			return e.(*page).title
		}),
		widget.ListOpts.ScrollContainerImage(res.list.image),
		widget.ListOpts.SliderParams(&widget.SliderParams{
			TrackImage:    res.list.track,
			HandleImage:   res.list.handle,
			MinHandleSize: res.list.handleSize,
			TrackPadding:  res.list.trackPadding,
		}),
		widget.ListOpts.EntryColor(res.list.entry),
		widget.ListOpts.EntryFontFace(res.list.face),
		widget.ListOpts.EntryTextPadding(res.list.entryPadding),
		widget.ListOpts.HideHorizontalSlider(),
		widget.ListOpts.HideVerticalSlider(),

		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			pageContainer.setPage(args.Entry.(*page))
		}))
	pageList.SetSelectedEntry(pages[0])

	container.AddChild(pageList)
	container.AddChild(pageContainer.widget)

	// TODO: add a reset config button
	footer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	footer.AddChild(widget.NewButton(
		widget.ButtonOpts.Text("Back to main menu", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			s.app.setState("rom_list")
		}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	root.AddChild(container)
	root.AddChild(footer)

	s.ui.Container = root
}

type page struct {
	title   string
	content widget.PreferredSizeLocateableWidget
}
