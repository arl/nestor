package debugger

import (
	"fmt"
	"image"
	"image/color"
	"slices"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

type stackFrameFlag uint8

const (
	sffNone stackFrameFlag = iota
	sffNMI
	sffIRQ
)

type stackFrame struct {
	src    uint16
	target uint16
	ret    uint16
	flag   stackFrameFlag
}

type callStack []stackFrame

func (cs *callStack) push(src, dst, ret uint16, flag stackFrameFlag) {
	*cs = append(*cs, stackFrame{
		src:    src,
		target: dst,
		ret:    ret,
		flag:   flag,
	})
}

func (cs *callStack) len() int {
	return len(*cs)
}

func (cs *callStack) pop() {
	if cs.len() == 0 {
		return
	}
	*cs = (*cs)[:cs.len()-1]
}

func (cs *callStack) reset() {
	*cs = (*cs)[:0]
}

type frameInfo [2]string

func (cs *callStack) build(pc uint16) []frameInfo {
	nfos := make([]frameInfo, 0, cs.len()+1)
	var curf *stackFrame
	for i, f := range *cs {
		if i > 0 {
			curf = &((*cs)[i-1])
		}
		src := fmt.Sprintf("$%04X", f.src)
		nfos = slices.Insert(nfos, 0, frameInfo{
			cs.entryPoint(curf),
			src,
		})
	}

	// Current frame
	curf = nil
	if cs.len() > 0 {
		curf = &((*cs)[cs.len()-1])
	}

	return slices.Insert(nfos, 0, frameInfo{
		cs.entryPoint(curf),
		fmt.Sprintf("$%04X", pc),
	})
}

func (callStack) entryPoint(f *stackFrame) string {
	if f == nil {
		return "[bottom of stack]"
	}

	str := fmt.Sprintf("%04X", f.target)
	switch f.flag {
	case sffNMI:
		return "[nmi] $" + str
	case sffIRQ:
		return "[irq] $" + str
	default:
		return str
	}
}

type callstackViewer struct {
	theme  *material.Theme
	grid   component.GridState
	frames []frameInfo
}

var callstackHeadings = []string{"Function", "PC"}

func (v *callstackViewer) update(cs callStack, pc uint16) {
	v.frames = cs.build(pc)
}

func (v *callstackViewer) Layout(gtx C) D {
	// Configure width based on available space and a minimum size.
	minSize := gtx.Dp(unit.Dp(100))
	border := widget.Border{
		Color: color.NRGBA{A: 255},
		Width: unit.Dp(1),
	}

	inset := layout.UniformInset(unit.Dp(2))

	// Configure a label styled to be a heading.
	headingLabel := material.Body1(v.theme, "")
	headingLabel.Font.Weight = font.Bold
	headingLabel.Alignment = text.Middle
	headingLabel.MaxLines = 1
	headingLabel.TextSize = unit.Sp(11)

	// Configure a label styled to be a data element.
	dataLabel := material.Body1(v.theme, "")
	dataLabel.Font.Typeface = "Go Mono"
	dataLabel.MaxLines = 1
	dataLabel.Alignment = text.End
	dataLabel.TextSize = unit.Sp(12)

	// Measure the height of a heading row.
	orig := gtx.Constraints
	gtx.Constraints.Min = image.Point{}
	macro := op.Record(gtx.Ops)
	dims := inset.Layout(gtx, headingLabel.Layout)
	_ = macro.Stop()
	gtx.Constraints = orig

	const numCols = 2
	return component.Table(v.theme, &v.grid).Layout(gtx, len(v.frames), numCols,
		func(axis layout.Axis, index, constraint int) int {
			widthUnit := max(int(float32(constraint)/3), minSize)
			switch axis {
			case layout.Horizontal:
				switch index {
				case 0, 1:
					return int(widthUnit)
				case 2, 3:
					return int(widthUnit / 2)
				default:
					return 0
				}
			default:
				return dims.Size.Y
			}
		},
		func(gtx C, col int) D {
			return border.Layout(gtx, func(gtx C) D {
				return inset.Layout(gtx, func(gtx C) D {
					headingLabel.Text = callstackHeadings[col]
					return headingLabel.Layout(gtx)
				})
			})
		},
		func(gtx C, row, col int) D {
			return inset.Layout(gtx, func(gtx C) D {
				dataLabel.Text = v.frames[row][col]
				return dataLabel.Layout(gtx)
			})
		},
	)
}
