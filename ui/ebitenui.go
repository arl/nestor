package ui

import (
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

type MouseClickArgs = widget.WidgetMouseButtonClickedEventArgs

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

func nineSliceBorderFromHex(width int, borderColor, bgColor uint32) *image.NineSlice {
	return image.NewBorderedNineSliceColor(
		hex2color(borderColor),
		hex2color(bgColor),
		width,
	)
}

type pageContainer struct {
	widget    widget.PreferredSizeLocateableWidget
	titleText *widget.Text
	flipBook  *widget.FlipBook
}

func newPageContainer() *pageContainer {
	c := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.TrackHover(false)),
		widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(res.panel.padding),
			widget.RowLayoutOpts.Spacing(15))),
	)

	titleText := widget.NewText(
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.TextOpts.Text("", res.text.titleFace, res.text.idleColor))
	c.AddChild(titleText)

	flipBook := widget.NewFlipBook(
		widget.FlipBookOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		}))),
	)
	c.AddChild(flipBook)

	return &pageContainer{
		widget:    c,
		titleText: titleText,
		flipBook:  flipBook,
	}
}

func (p *pageContainer) setPage(page *page) {
	p.titleText.Label = page.title
	p.flipBook.SetPage(page.content)
	p.flipBook.RequestRelayout()
}

func newPageContentContainer() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
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
