package ui

import (
	"fmt"
	goimage "image"

	"github.com/ebitenui/ebitenui/widget"
)

type Combobox struct {
	Widget  *widget.ListComboButton
	items   []string
	indices []any
}

func newCombobox[T comparable](items []string, layoutData any, onSelect func(item T)) *Combobox {
	indices := make([]any, 0, len(items))
	for i := range items {
		indices = append(indices, i)
	}

	fbtn := func(e any) string {
		idx := e.(int)
		return items[idx]
	}
	flist := func(e any) string {
		idx := e.(int)
		return items[idx]
	}
	cbox := widget.NewListComboButton(
		widget.ListComboButtonOpts.Entries(indices),
		widget.ListComboButtonOpts.MaxContentHeight(150),
		widget.ListComboButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(layoutData),
		),
		widget.ListComboButtonOpts.ButtonParams(&widget.ButtonParams{
			Image:       res.button.image,
			TextPadding: res.comboButton.padding,
			TextColor:   res.button.text,
			TextFace:    res.comboButton.face,
			MinSize:     &goimage.Point{200, 0},
		}),
		widget.ListComboButtonOpts.ListParams(&widget.ListParams{
			ScrollContainerImage: res.list.image,
			Slider: &widget.SliderParams{
				TrackImage:    res.slider.trackImage,
				HandleImage:   res.slider.handle,
				MinHandleSize: ptrTo(5),
				TrackPadding:  widget.NewInsetsSimple(2),
			},
			EntryFace:        res.fonts.small,
			EntryColor:       res.list.entry,
			EntryTextPadding: widget.NewInsetsSimple(5),
			MinSize:          &goimage.Point{200, 0},
		}),

		widget.ListComboButtonOpts.EntryLabelFunc(fbtn, flist),
		widget.ListComboButtonOpts.EntrySelectedHandler(func(args *widget.ListComboButtonEntrySelectedEventArgs) {
			fmt.Println("in EntrySelectedHandler")
			if onSelect != nil {
				return
			}
			onSelect(args.Entry.(T))
		}))

	cbox.SetSelectedEntry(0)
	return &Combobox{Widget: cbox, items: items, indices: indices}
}
