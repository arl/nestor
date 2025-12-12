package ui

import (
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

// accepts 0xrrggbb or 0xaarrggbb
func hex2color(val uint32) color.Color {
	alpha := uint8(0xFF)
	if val > 0xffffff {
		alpha = uint8(val >> 24)
		val &= 0xffffff
	}

	return color.NRGBA{
		R: uint8(val & 0xff0000 >> 16),
		G: uint8(val & 0xff00 >> 8),
		B: uint8(val & 0xff),
		A: alpha,
	}
}

func ninesliceFromHex(val uint32) *image.NineSlice {
	return image.NewNineSliceColor(hex2color(val))
}

type pageContainer struct {
	root      *widget.Container
	titleText *widget.Text
	content   *widget.Container
	sc        *widget.ScrollContainer
	slider    *widget.Slider
}

func newPageContainer() *pageContainer {
	pc := &pageContainer{}

	pc.content = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(10),
		)),
	)

	pc.sc = widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(pc.content),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: res.panel.image,
			Mask: res.panel.image,
		}),
	)

	pageSizeFunc := func() int {
		viewH := pc.sc.ViewRect().Dy()
		contentH := pc.content.GetWidget().Rect.Dy()
		if contentH <= viewH {
			return 1000
		}
		return int(float64(viewH) / float64(contentH) * 1000)
	}

	pc.slider = widget.NewSlider(
		widget.SliderOpts.Orientation(widget.DirectionVertical),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.PageSizeFunc(pageSizeFunc),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			pc.sc.ScrollTop = float64(args.Slider.Current) / 1000
		}),
		widget.SliderOpts.Images(res.slider.trackImage, res.slider.handle),
	)

	pc.sc.GetWidget().ScrolledEvent.AddHandler(func(args any) {
		if a, ok := args.(*widget.WidgetScrolledEventArgs); ok {
			pc.slider.Current -= int(a.Y * float64(pageSizeFunc()))
		}
	})

	pc.titleText = widget.NewText(
		widget.TextOpts.Text("", res.text.titleFace, res.text.idleColor),
	)

	scrollRow := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{
			HorizontalPosition: widget.GridLayoutPositionStart,
			VerticalPosition:   widget.GridLayoutPositionStart,
			MaxWidth:           0,
			MaxHeight:          0,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(4, 0),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{true}),
		)),
	)
	scrollRow.AddChild(pc.sc)
	scrollRow.AddChild(pc.slider)

	pc.root = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Padding(res.panel.padding),
			widget.GridLayoutOpts.Spacing(0, 15),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
		)),
	)
	pc.root.AddChild(pc.titleText)
	pc.root.AddChild(scrollRow)

	return pc
}

func (p *pageContainer) setPage(pg *page) {
	p.titleText.Label = pg.title
	p.content.RemoveChildren()
	p.content.AddChild(pg.content)
	p.sc.ScrollTop = 0
	p.slider.Current = 0
}

func newPageContentContainer() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(10),
		)))
}

func newSeparator(ld any) widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:    10,
				Bottom: 10,
			}))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(ld)))

	c.AddChild(widget.NewGraphic(
		widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch:   true,
			MaxHeight: 2,
		})),
		widget.GraphicOpts.ImageNineSlice(image.NewNineSliceColor(res.separatorColor)),
	))

	return c
}

func newCheckbox2States(label string, initial bool, onchanged func(bool)) *widget.Checkbox {
	return widget.NewCheckbox(
		widget.CheckboxOpts.Image(res.checkbox.image),
		widget.CheckboxOpts.InitialState(bool2state(initial)),
		widget.CheckboxOpts.Text(label, res.label.face, res.label.text),
		widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
			onchanged(state2bool(args.State))
		}),
	)
}

func bool2state(b bool) widget.WidgetState {
	if b {
		return widget.WidgetChecked
	}
	return widget.WidgetUnchecked
}

func state2bool(s widget.WidgetState) bool {
	return s == widget.WidgetChecked
}
