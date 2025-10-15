package ui

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
)

type page struct {
	title   string
	content widget.PreferredSizeLocateableWidget
}

func inputPage() *page {
	c := newPageContentContainer()

	bs := []*widget.Button{}
	for i := 0; i < 3; i++ {
		b := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
			widget.ButtonOpts.Image(res.button.image),
			widget.ButtonOpts.Text(fmt.Sprintf("Button %d", i+1), res.button.face, res.button.text),
			widget.ButtonOpts.TextPadding(res.button.padding),
			widget.ButtonOpts.CursorEnteredHandler(func(args *widget.ButtonHoverEventArgs) { fmt.Println("Cursor Entered: " + args.Button.Text().Label) }),
			widget.ButtonOpts.CursorExitedHandler(func(args *widget.ButtonHoverEventArgs) { fmt.Println("Cursor Exited: " + args.Button.Text().Label) }),
		)
		c.AddChild(b)
		bs = append(bs, b)
	}

	c.AddChild(newSeparator(widget.RowLayoutData{Stretch: true}))

	toggles := []*widget.Button{}
	for i := 0; i < 3; i++ {
		b := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
			widget.ButtonOpts.Image(res.button.image),
			widget.ButtonOpts.Text(fmt.Sprintf("Toggle Button %d", i+1), res.button.face, res.button.text),
			widget.ButtonOpts.TextPadding(res.button.padding),
			widget.ButtonOpts.CursorEnteredHandler(func(args *widget.ButtonHoverEventArgs) { fmt.Println("Cursor Entered: " + args.Button.Text().Label) }),
			widget.ButtonOpts.CursorExitedHandler(func(args *widget.ButtonHoverEventArgs) { fmt.Println("Cursor Exited: " + args.Button.Text().Label) }),
		)
		c.AddChild(b)
		bs = append(bs, b)
		toggles = append(toggles, b)
	}
	elements := []widget.RadioGroupElement{}
	for _, cb := range toggles {
		elements = append(elements, cb)
	}
	widget.NewRadioGroup(widget.RadioGroupOpts.Elements(elements...))

	c.AddChild(newSeparator(widget.RowLayoutData{Stretch: true}))

	c.AddChild(newCheckbox("Disabled", func(args *widget.CheckboxChangedEventArgs) {
		for _, b := range bs {
			b.GetWidget().Disabled = args.State == widget.WidgetChecked
		}
	}))

	return &page{
		title:   "Button",
		content: c,
	}
}
