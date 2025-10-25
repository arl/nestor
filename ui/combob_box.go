package ui

import (
	goimage "image"

	"github.com/ebitenui/ebitenui/widget"
)

type Combobox struct {
	*widget.ListComboButton
	items   []string
	indices []any

	previdx int
}

func newCombobox(items []string, selidx int, layoutData any, selectHandler func(item int)) *Combobox {
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

	combo := &Combobox{
		items:   items,
		indices: indices,
		previdx: selidx,
	}

	handler := func(args *widget.ListComboButtonEntrySelectedEventArgs) {
		idx := args.Entry.(int)
		if selectHandler == nil ||
			// avoid spurious events.
			idx == combo.previdx {
			return
		}

		selectHandler(idx)
		combo.previdx = idx
	}

	combo.ListComboButton = widget.NewListComboButton(
		widget.ListComboButtonOpts.Entries(indices),
		widget.ListComboButtonOpts.MaxContentHeight(150),
		widget.ListComboButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(layoutData)),
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
		widget.ListComboButtonOpts.EntrySelectedHandler(handler),
	)

	combo.ListComboButton.SetSelectedEntry(selidx)
	return combo
}
