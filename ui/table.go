package ui

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type Table struct {
	Cols int

	ColBorder unit.Dp
	RowBorder unit.Dp
}

func (t Table) Layout(gtx C, nrows int, cell func(gtx C, i, j int) D) D {
	lightgrey := color.NRGBA{R: 235, G: 235, B: 235, A: 255}
	darkgrey := color.NRGBA{R: 128, G: 128, B: 128, A: 255}

	colBorder := func(gtx C) D {
		paint.FillShape(gtx.Ops, darkgrey, clip.Rect(image.Rect(0, 0, int(t.ColBorder), gtx.Constraints.Max.Y)).Op())
		return layout.Spacer{Width: t.ColBorder}.Layout(gtx)
	}

	macro := op.Record(gtx.Ops)

	colsizes := make([]D, t.Cols)
	cellsizes := make([]D, nrows)

	columns := make([]layout.FlexChild, 0, t.Cols)
	cells := make([]layout.FlexChild, nrows)

	columns = append(columns, layout.Rigid(colBorder))
	for i := 0; i < t.Cols; i++ {
		i := i
		columns = append(columns, layout.Rigid(func(gtx C) D {
			for j := 0; j < nrows; j++ {
				j := j
				cells[j] = layout.Rigid(func(gtx C) D {
					cellsizes[j] = cell(gtx, i, j)
					return cellsizes[j]
				})
			}
			colsizes[i] = layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEnd, Alignment: layout.End}.Layout(gtx, cells...)
			return colsizes[i]
		}))
		columns = append(columns, layout.Rigid(colBorder))
	}

	dims := layout.Flex{Spacing: layout.SpaceEnd, Alignment: layout.Start}.Layout(gtx, columns...)
	call := macro.Stop()

	var tot D
	for _, sz := range colsizes {
		tot.Size.X += sz.Size.X
	}
	tot.Size.X += int(t.ColBorder) * (t.Cols + 1)
	for _, sz := range cellsizes {
		tot.Size.Y += sz.Size.Y
	}

	defer clip.Rect{Max: tot.Size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: lightgrey}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	call.Add(gtx.Ops)

	// TODO(arl) use offset to do that, reusing the same op multiple time?
	y := 0
	for _, sz := range cellsizes {
		y += sz.Size.Y
		paint.FillShape(gtx.Ops, darkgrey, clip.Rect(image.Rect(0, y-int(t.RowBorder), gtx.Constraints.Max.X, y)).Op())
	}
	return dims
}
