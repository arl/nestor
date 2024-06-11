package debugger

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"nestor/hw"
	"nestor/ui"
)

type listing struct {
	th *material.Theme
	// list widget.List
	list SelectList
	pos  layout.Position

	cpu *hw.CPU
	dbg *gioDebugger

	cs      callStack
	pc      uint16
	resetPC uint16 // pc at reset vector

	selectedPC uint16

	lblStyle material.LabelStyle
}

func newListing(dbg *gioDebugger, theme *material.Theme, cpu *hw.CPU) listing {
	lblStyle := material.LabelStyle{
		Color:          theme.Palette.Fg,
		SelectionColor: ui.MulAlpha(theme.Palette.ContrastBg, 0x60),
		TextSize:       theme.TextSize * 14.0 / 16.0,
		Shaper:         theme.Shaper,
		MaxLines:       1,
		Alignment:      text.Start,
		WrapPolicy:     text.WrapWords,
	}
	lblStyle.Font.Typeface = "Go Mono"
	lblStyle.Font.Weight = font.Light

	return listing{
		list: SelectList{
			List: widget.List{
				List: layout.List{
					Axis: layout.Vertical,
				},
			},
			ItemHeight: unit.Dp(24),
		},
		lblStyle: lblStyle,
		th:       theme,
		cpu:      cpu,
		dbg:      dbg,
	}
}

func (l *listing) update(cs callStack, stat debuggerState) {
	l.pc = stat.pc
	l.cs = cs
}

func (l *listing) Layout(gtx C, stat status) D {
	nlines := 1000

	drawOp := func(gtx C, i int) D {
		// look for the first pc of current stack frame
		var bottom uint16
		if len(l.cs) == 0 {
			// We're in the bottom stack frame
			bottom = l.dbg.resetPC
		} else {
			bottom = l.cs[0].src
		}

		// iterate on all pcs, looking for current position
		curpc := bottom
		pcidx := 0
		for pcidx < i {

			// continuer here

			op := l.cpu.Disasm(curpc)
			curpc += uint16(len(op.Bytes))

			if l.pc == curpc {
				l.list.Selected = i
			}

			pcidx++
		}
		if pcidx != i {
			panic("inconsistent pc chain")
		}

		op := l.cpu.Disasm(curpc)

		defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()

		switch {
		case l.list.Selected == i:
			paint.Fill(gtx.Ops, color.NRGBA{R: 0xFF, G: 0xF0, B: 0xF0, A: 0xFF})
		case l.list.Hovered == i:
			paint.Fill(gtx.Ops, color.NRGBA{R: 0xF0, G: 0xFF, B: 0xF0, A: 0xFF})
		}

		inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
		return inset.Layout(gtx, func(gtx C) D {
			dims := layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					l.lblStyle.Text = fmt.Sprintf("%04X", op.PC)
					gtx.Constraints = layout.Exact(image.Pt(45, 25))
					return l.lblStyle.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					var sb strings.Builder
					for _, b := range op.Bytes {
						fmt.Fprintf(&sb, "%02X ", b)
					}

					l.lblStyle.Text = sb.String()
					gtx.Constraints = layout.Exact(image.Pt(80, 25))
					return l.lblStyle.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					l.lblStyle.Text = op.Opcode
					gtx.Constraints = layout.Exact(image.Pt(45, 25))
					return l.lblStyle.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					l.lblStyle.Text = op.Oper
					gtx.Constraints = layout.Exact(image.Pt(45, 25))
					return l.lblStyle.Layout(gtx)
				}),
			)

			l.pos = l.list.Position
			return dims
		})
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return material.H6(l.th, "Disassembly").Layout(gtx)
		}),
		layout.Flexed(1, func(gtx C) D {
			// Do not show source when the debugger is running.
			if stat == running {
				return layout.Dimensions{}
			}

			return l.list.Layout(l.th, gtx, nlines, drawOp)
		}),
	)
}

type FocusBorderStyle struct {
	Focused     bool
	BorderWidth unit.Dp
	Color       color.NRGBA
}

func FocusBorder(th *material.Theme, focused bool) FocusBorderStyle {
	return FocusBorderStyle{
		Focused:     focused,
		BorderWidth: unit.Dp(2),
		Color:       th.ContrastBg,
	}
}

func (focus FocusBorderStyle) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	inset := layout.UniformInset(focus.BorderWidth)
	if !focus.Focused {
		return inset.Layout(gtx, w)
	}

	return widget.Border{
		Color: focus.Color,
		Width: focus.BorderWidth,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return inset.Layout(gtx, w)
	})
}

type SelectList struct {
	widget.List

	Selected int
	Hovered  int

	ItemHeight unit.Dp

	focused bool
}

func (list *SelectList) Layout(th *material.Theme, gtx layout.Context, length int, element layout.ListElement) layout.Dimensions {
	return FocusBorder(th, list.focused).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		gtx.Constraints = layout.Exact(size)
		defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()

		changed := false
		grabbed := false

		itemHeight := gtx.Metric.Dp(list.ItemHeight)

		pointerClicked := false
		pointerHovered := false
		pointerPosition := f32.Point{}

		event.Op(gtx.Ops, &list.List)

		for {
			ev, ok := gtx.Event(
				key.Filter{
					Focus: list,
					Name: key.NameUpArrow + "|" + key.NameDownArrow + "|" +
						key.NameHome + "|" + key.NameEnd + "|" +
						key.NamePageUp + "|" + key.NamePageDown,
				},
				pointer.Filter{
					Target: &list.List,
					Kinds:  pointer.Press | pointer.Move,
				},
			)
			if !ok {
				break
			}

			switch ev := ev.(type) {
			case key.Event:
				if ev.State == key.Press {
					offset := 0
					switch ev.Name {
					case key.NameHome:
						offset = -length
					case key.NameEnd:
						offset = length
					case key.NameUpArrow:
						offset = -1
					case key.NameDownArrow:
						offset = 1
					case key.NamePageUp:
						offset = -list.List.Position.Count
					case key.NamePageDown:
						offset = list.List.Position.Count
					}

					if offset != 0 {
						target := list.Selected + offset
						if target < 0 {
							target = 0
						}
						if target >= length {
							target = length - 1
						}
						if list.Selected != target {
							list.Selected = target
							changed = true
						}
					}
				}
			case key.FocusEvent:
				if list.focused != ev.Focus {
					list.focused = ev.Focus
					gtx.Execute(op.InvalidateCmd{})
				}
			case pointer.Event:
				switch ev.Kind {
				case pointer.Press:
					if !list.focused && !grabbed {
						grabbed = true
						gtx.Execute(key.FocusCmd{Tag: list})
					}
					// TODO: find the item
					pointerClicked = true
					pointerPosition = ev.Position
				case pointer.Move:
					pointerHovered = true
					pointerPosition = ev.Position
				case pointer.Cancel:
					list.Hovered = -1
				}
			}
		}

		if pointerClicked || pointerHovered {
			clientClickY := list.Position.First*itemHeight + list.Position.Offset + int(pointerPosition.Y)
			target := clientClickY / itemHeight
			if 0 <= target && target <= length {
				if pointerClicked && list.Selected != target {
					list.Selected = target
				}
				if pointerHovered && list.Hovered != target {
					list.Hovered = target
				}
			}
		}

		if changed {
			pos := &list.List.Position
			switch {
			case list.Selected < pos.First+1:
				list.List.Position = layout.Position{First: list.Selected - 1}
			case pos.First+pos.Count-1 <= list.Selected:
				list.List.Position = layout.Position{First: list.Selected - pos.Count + 2}
			}
		}

		return material.List(th, &list.List).Layout(gtx, length,
			func(gtx layout.Context, index int) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Point{
					X: gtx.Constraints.Max.X,
					Y: itemHeight,
				})
				return element(gtx, index)
			})
	})
}
