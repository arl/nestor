package ui

import "github.com/ebitenui/ebitenui/widget"

type StaticList struct {
	List    *widget.List
	items   []string
	indices []any
}

func newStaticList[T comparable](items []string, onSelect func(item T)) *StaticList {
	indices := make([]any, 0, len(items))
	for i := range items {
		indices = append(indices, i)
	}

	list := widget.NewList(
		widget.ListOpts.Entries(indices),
		widget.ListOpts.EntryLabelFunc(func(e any) string {
			idx := e.(int)
			return items[idx]
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
			if onSelect != nil {
				return
			}
			onSelect(args.Entry.(T))
		}))

	return &StaticList{List: list, items: items, indices: indices}
}
