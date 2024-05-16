package debugger

import (
	"fmt"
	"image"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"nestor/hw"
	"nestor/ui"
)

type listing struct {
	th   *material.Theme
	list widget.List
	pos  layout.Position

	cpu *hw.CPU
	dbg *debugger

	cs      callStack
	pc      uint16
	resetPC uint16 // pc at reset vector

	lblStyle material.LabelStyle
}

func newListing(dbg *debugger, theme *material.Theme, cpu *hw.CPU) listing {
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
		list:     widget.List{List: layout.List{Axis: layout.Vertical}},
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

	drawSource := func(gtx C, i int) D {
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
			op := l.cpu.Disasm(l.pc)
			curpc += uint16(len(op.Bytes))
			pcidx++
		}
		if pcidx != i {
			panic("inconsistent pc chain")
		}

		op := l.cpu.Disasm(curpc)

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

			return material.List(l.th, &l.list).Layout(gtx, nlines, drawSource)
		}),
	)
}
