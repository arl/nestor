package ui

import (
	"sort"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type configPage struct {
	*app
}

func newConfig(app *app) *configPage {
	state := &configPage{
		app: app,
	}

	state.createUI()
	return state
}

var bgcolor = image.NewNineSliceColor(colornames.Blueviolet)

func (s *configPage) createUI() {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.TrackHover(false)),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			// It is using a GridLayout with a single column
			widget.GridLayoutOpts.Columns(1),
			// It uses the Stretch parameter to define how the rows will be layed out.
			// - a fixed sized header
			// - a content row that stretches to fill all remaining space
			// - a fixed sized footer
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true, false}),
			// Padding defines how much space to put around the outside of the grid.
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Top:    20,
				Bottom: 20,
			}),
			// Spacing defines how much space to put between each column and row
			widget.GridLayoutOpts.Spacing(0, 20))),
		widget.ContainerOpts.BackgroundImage(bgcolor))

	rootContainer.AddChild(headerContainer())

	rootContainer.AddChild(configContainer())

	footerContainer := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(
		widget.RowLayoutOpts.Padding(&widget.Insets{
			Left:  25,
			Right: 25,
		}),
	)))
	rootContainer.AddChild(footerContainer)

	footerContainer.AddChild(widget.NewText(
		widget.TextOpts.Text("github.com/ebitenui/ebitenui", res.text.smallFace, res.text.disabledColor)))
}

func (s *configPage) draw(screen *ebiten.Image) {
	screen.Fill(colornames.Lightcoral)
	s.ui.Draw(screen)
}

func configContainer() widget.PreferredSizeLocateableWidget {

	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(&widget.Insets{
				Left:  25,
				Right: 25,
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(20, 0),
		)))

	pages := []interface{}{

		inputPage(),
		// buttonPage(res),
		// checkboxPage(res),
		// listPage(res),
		// comboButtonPage(res),
		// tabBookPage(res),
		// gridLayoutPage(res),
		// rowLayoutPage(res),
		// sliderPage(res),
		// toolTipPage(res),
		// dragAndDropPage(res),
		// textInputPage(res),
		// radioGroupPage(res),
		// windowPage(res, ui),
		// anchorLayoutPage(res),
		// textAreaPage(res),
		// progressBarPage(res),
	}

	collator := collate.New(language.English)
	sort.Slice(pages, func(a int, b int) bool {
		p1 := pages[a].(*page)
		p2 := pages[b].(*page)
		return collator.CompareString(p1.title, p2.title) < 0
	})

	pageContainer := newPageContainer()

	pageList := widget.NewList(
		widget.ListOpts.Entries(pages),
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
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

		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			pageContainer.setPage(args.Entry.(*page))
		}))
	container.AddChild(pageList)

	container.AddChild(pageContainer.widget)

	pageList.SetSelectedEntry(pages[0])

	return container
}
