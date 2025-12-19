// Package menu provides a reusable menu bar component for ebitenui.
package menu

import (
	goimage "image"
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/exp/constraints"
)

// Style defines the visual appearance of the menu.
type Style struct {
	Font            *text.Face
	BackgroundColor color.Color
	MenuBackground  color.Color

	TextColorIdle     color.Color
	TextColorDisabled color.Color
	TextColorHover    color.Color
	TextColorPressed  color.Color

	ButtonHoverColor   color.Color
	ButtonPressedColor color.Color
}

// DefaultStyle returns a sensible default style.
func DefaultStyle(face *text.Face) Style {
	return Style{
		Font:               face,
		BackgroundColor:    color.Black,
		MenuBackground:     color.RGBA{R: 50, G: 50, B: 50, A: 255},
		TextColorIdle:      color.White,
		TextColorDisabled:  color.Gray{Y: 128},
		TextColorHover:     color.White,
		TextColorPressed:   color.Black,
		ButtonHoverColor:   color.Gray{Y: 64},
		ButtonPressedColor: color.White,
	}
}

// Action is a constraint for menu action types (typically int or custom action enums).
type Action interface {
	constraints.Integer
}

// Item represents a menu entry that can be clicked.
type Item[A Action] struct {
	Label    string
	ID       string    // Optional string ID for items without actions (e.g., submenu openers)
	Action   A         // Action to trigger; 0 means no action
	Disabled bool      // Start disabled
	SubMenu  []Item[A] // If non-empty, clicking shows a submenu
}

// Menu represents a top-level menu with items.
type Menu[A Action] struct {
	Label string
	Items []Item[A]
}

// Definition describes the complete menu bar structure.
type Definition[A Action] struct {
	Menus []Menu[A]
}

// Bar is a menu bar widget built from a Definition.
type Bar[A Action] struct {
	Container *widget.Container
	ui        *ebitenui.UI
	style     Style
	onAction  func(A)

	// Store buttons for later access (e.g., to enable/disable)
	actionButtons map[A]*widget.Button
	idButtons     map[string]*widget.Button
}

// New creates a new menu bar from the given definition.
func New[A Action](ui *ebitenui.UI, def Definition[A], style Style, onAction func(A)) *Bar[A] {
	bar := &Bar[A]{
		ui:            ui,
		style:         style,
		onAction:      onAction,
		actionButtons: make(map[A]*widget.Button),
		idButtons:     make(map[string]*widget.Button),
	}

	bar.Container = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(style.BackgroundColor)),
		widget.ContainerOpts.Layout(
			widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			),
		),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true}),
		),
	)

	for _, menu := range def.Menus {
		bar.addMenu(menu)
	}

	return bar
}

func (b *Bar[A]) addMenu(menu Menu[A]) {
	btn := b.newMenuButton(menu.Label)
	btn.ClickedEvent.AddHandler(event.WrapHandler(func(args *widget.ButtonClickedEventArgs) {
		b.openMenu(args.Button.GetWidget(), menu.Items)
	}))
	b.Container.AddChild(btn)
}

func (b *Bar[A]) openMenu(opener *widget.Widget, items []Item[A]) {
	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(b.style.MenuBackground)),
		widget.ContainerOpts.Layout(
			widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Spacing(4),
				widget.RowLayoutOpts.Padding(&widget.Insets{Top: 4, Bottom: 4}),
			),
		),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
	)

	for _, item := range items {
		entry := b.newMenuEntry(item)
		c.AddChild(entry)
	}

	w, h := c.PreferredSize()
	x := opener.Rect.Min.X
	y := opener.Rect.Max.Y

	window := widget.NewWindow(
		widget.WindowOpts.Contents(c),
		widget.WindowOpts.CloseMode(widget.CLICK_OUT),
		widget.WindowOpts.Location(goimage.Rect(x, y, x+w, y+h)),
	)

	b.ui.AddWindow(window)
}

func (b *Bar[A]) openSubMenu(opener *widget.Widget, items []Item[A]) {
	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(b.style.MenuBackground)),
		widget.ContainerOpts.Layout(
			widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Spacing(4),
				widget.RowLayoutOpts.Padding(&widget.Insets{Top: 4, Bottom: 4}),
			),
		),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
	)

	for _, item := range items {
		entry := b.newMenuEntry(item)
		c.AddChild(entry)
	}

	w, h := c.PreferredSize()
	x := opener.Rect.Max.X
	y := opener.Rect.Min.Y

	window := widget.NewWindow(
		widget.WindowOpts.Contents(c),
		widget.WindowOpts.CloseMode(widget.CLICK),
		widget.WindowOpts.Location(goimage.Rect(x, y, x+w, y+h)),
	)

	b.ui.AddWindow(window)
}

func (b *Bar[A]) newMenuButton(label string) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    image.NewNineSliceColor(color.Transparent),
			Hover:   image.NewNineSliceColor(b.style.ButtonHoverColor),
			Pressed: image.NewNineSliceColor(b.style.ButtonPressedColor),
		}),
		widget.ButtonOpts.Text(label, b.style.Font, &widget.ButtonTextColor{
			Idle:     b.style.TextColorIdle,
			Disabled: b.style.TextColorDisabled,
			Hover:    b.style.TextColorHover,
			Pressed:  b.style.TextColorPressed,
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Top:    4,
			Left:   4,
			Right:  32,
			Bottom: 4,
		}),
	)
}

func (b *Bar[A]) newMenuEntry(item Item[A]) *widget.Button {
	btn := widget.NewButton(
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    image.NewNineSliceColor(color.Transparent),
			Hover:   image.NewNineSliceColor(b.style.ButtonHoverColor),
			Pressed: image.NewNineSliceColor(b.style.ButtonPressedColor),
		}),
		widget.ButtonOpts.Text(item.Label, b.style.Font, &widget.ButtonTextColor{
			Idle:     b.style.TextColorIdle,
			Disabled: b.style.TextColorDisabled,
			Hover:    b.style.TextColorHover,
			Pressed:  b.style.TextColorPressed,
		}),
		widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 16, Right: 64}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	if item.Disabled {
		btn.GetWidget().Disabled = true
	}

	// Store button reference by action or ID
	if item.Action != 0 {
		b.actionButtons[item.Action] = btn
	}
	if item.ID != "" {
		b.idButtons[item.ID] = btn
	}

	if len(item.SubMenu) > 0 {
		btn.ClickedEvent.AddHandler(event.WrapHandler(func(args *widget.ButtonClickedEventArgs) {
			b.openSubMenu(args.Button.GetWidget(), item.SubMenu)
		}))
	} else if item.Action != 0 {
		action := item.Action
		btn.ClickedEvent.AddHandler(func(args any) {
			if b.onAction != nil {
				b.onAction(action)
			}
		})
	}

	return btn
}

// GetButton returns the button widget for the given action, or nil if not found.
func (b *Bar[A]) GetButton(action A) *widget.Button {
	return b.actionButtons[action]
}

// GetButtonByID returns the button widget for the given string ID, or nil if not found.
func (b *Bar[A]) GetButtonByID(id string) *widget.Button {
	return b.idButtons[id]
}

// SetDisabled enables or disables the button for the given action.
func (b *Bar[A]) SetDisabled(action A, disabled bool) {
	if btn := b.actionButtons[action]; btn != nil {
		btn.GetWidget().Disabled = disabled
	}
}

// SetDisabledByID enables or disables the button for the given string ID.
func (b *Bar[A]) SetDisabledByID(id string, disabled bool) {
	if btn := b.idButtons[id]; btn != nil {
		btn.GetWidget().Disabled = disabled
	}
}
