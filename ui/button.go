package ui

import (
	"image"
	"image/color"
	"math"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

type SmallSquareIconButton struct {
	*widget.Clickable
	tooltip component.Tooltip
	style   SquareIconButtonStyle
	tipArea component.TipArea
}

func NewSmallSquareIconButton(th *material.Theme, icon []byte, tooltip string) SmallSquareIconButton {
	btn := SmallSquareIconButton{
		tooltip:   component.PlatformTooltip(th, tooltip),
		Clickable: new(widget.Clickable),
	}
	btn.style = SquareIconButtonStyle{
		Background:  th.Palette.ContrastBg,
		Color:       th.Palette.ContrastFg,
		Icon:        mustNewIcon(icon),
		Size:        20,
		Inset:       layout.UniformInset(4),
		Button:      btn.Clickable,
		Description: tooltip,
	}

	return btn
}

func (b *SmallSquareIconButton) Layout(gtx layout.Context, enabled bool) layout.Dimensions {
	// We don't use gtx.Disabled when drawing the tooltip or else it would not
	// disappear correctly.
	return b.tipArea.Layout(gtx, b.tooltip, func(gtx C) D {
		if !enabled {
			gtx = gtx.Disabled()
		}
		return b.style.Layout(gtx)
	})
}

func mustNewIcon(data []byte) *widget.Icon {
	icon, err := widget.NewIcon(data)
	if err != nil {
		panic(err)
	}
	return icon
}

/* code copied from the Gio project material.IconButtonStyle
with the following license:
// SPDX-License-Identifier: Unlicense OR MIT
*/

type SquareIconButtonStyle struct {
	Background color.NRGBA
	// Color is the icon color.
	Color color.NRGBA
	Icon  *widget.Icon
	// Size is the icon size.
	Size        unit.Dp
	Inset       layout.Inset
	Button      *widget.Clickable
	Description string

	Theme *material.Theme
}

func (b SquareIconButtonStyle) Layout(gtx layout.Context) layout.Dimensions {
	m := op.Record(gtx.Ops)
	dims := b.Button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		if d := b.Description; d != "" {
			semantic.DescriptionOp(b.Description).Add(gtx.Ops)
		}
		return layout.Background{}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				// rr := (gtx.Constraints.Min.X + gtx.Constraints.Min.Y) / 1
				// defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, rr).Push(gtx.Ops).Pop()
				background := b.Background
				switch {
				case !gtx.Enabled():
					background = Disabled(b.Background)
				case b.Button.Hovered() || gtx.Focused(b.Button):
					background = Hovered(b.Background)
				}
				paint.Fill(gtx.Ops, background)
				for _, c := range b.Button.History() {
					drawInk(gtx, c)
				}
				return layout.Dimensions{Size: gtx.Constraints.Min}
			},
			func(gtx layout.Context) layout.Dimensions {
				return b.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					size := gtx.Dp(b.Size)
					if b.Icon != nil {
						gtx.Constraints.Min = image.Point{X: size}
						b.Icon.Layout(gtx, b.Color)
					}
					return layout.Dimensions{
						Size: image.Point{X: size, Y: size},
					}
				})
			},
		)
	})
	c := m.Stop()
	// bounds := image.Rectangle{Max: dims.Size}
	// defer clip.Ellipse(bounds).Push(gtx.Ops).Pop()
	c.Add(gtx.Ops)

	return dims
}

// Disabled blends color towards the luminance and multiplies alpha.
// Blending towards luminance will desaturate the color.
// Multiplying alpha blends the color together more with the background.
func Disabled(c color.NRGBA) (d color.NRGBA) {
	const r = 80 // blend ratio
	lum := approxLuminance(c)
	d = mix(c, color.NRGBA{A: c.A, R: lum, G: lum, B: lum}, r)
	d = MulAlpha(d, 128+32)
	return
}

// Hovered blends dark colors towards white, and light colors towards
// black. It is approximate because it operates in non-linear sRGB space.
func Hovered(c color.NRGBA) (h color.NRGBA) {
	if c.A == 0 {
		// Provide a reasonable default for transparent widgets.
		return color.NRGBA{A: 0x44, R: 0x88, G: 0x88, B: 0x88}
	}
	const ratio = 0x20
	m := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: c.A}
	if approxLuminance(c) > 128 {
		m = color.NRGBA{A: c.A}
	}
	return mix(m, c, ratio)
}

// approxLuminance is a fast approximate version of RGBA.Luminance.
func approxLuminance(c color.NRGBA) byte {
	const (
		r = 13933 // 0.2126 * 256 * 256
		g = 46871 // 0.7152 * 256 * 256
		b = 4732  // 0.0722 * 256 * 256
		t = r + g + b
	)
	return byte((r*int(c.R) + g*int(c.G) + b*int(c.B)) / t)
}

// mix mixes c1 and c2 weighted by (1 - a/256) and a/256 respectively.
func mix(c1, c2 color.NRGBA, a uint8) color.NRGBA {
	ai := int(a)
	return color.NRGBA{
		R: byte((int(c1.R)*ai + int(c2.R)*(256-ai)) / 256),
		G: byte((int(c1.G)*ai + int(c2.G)*(256-ai)) / 256),
		B: byte((int(c1.B)*ai + int(c2.B)*(256-ai)) / 256),
		A: byte((int(c1.A)*ai + int(c2.A)*(256-ai)) / 256),
	}
}

// MulAlpha applies the alpha to the color.
func MulAlpha(c color.NRGBA, alpha uint8) color.NRGBA {
	c.A = uint8(uint32(c.A) * uint32(alpha) / 0xFF)
	return c
}

func drawInk(gtx layout.Context, c widget.Press) {
	// duration is the number of seconds for the
	// completed animation: expand while fading in, then
	// out.
	const (
		expandDuration = float32(0.5)
		fadeDuration   = float32(0.9)
	)

	now := gtx.Now

	t := float32(now.Sub(c.Start).Seconds())

	end := c.End
	if end.IsZero() {
		// If the press hasn't ended, don't fade-out.
		end = now
	}

	endt := float32(end.Sub(c.Start).Seconds())

	// Compute the fade-in/out position in [0;1].
	var alphat float32
	{
		var haste float32
		if c.Cancelled {
			// If the press was cancelled before the inkwell
			// was fully faded in, fast forward the animation
			// to match the fade-out.
			if h := 0.5 - endt/fadeDuration; h > 0 {
				haste = h
			}
		}
		// Fade in.
		half1 := t/fadeDuration + haste
		if half1 > 0.5 {
			half1 = 0.5
		}

		// Fade out.
		half2 := float32(now.Sub(end).Seconds())
		half2 /= fadeDuration
		half2 += haste
		if half2 > 0.5 {
			// Too old.
			return
		}

		alphat = half1 + half2
	}

	// Compute the expand position in [0;1].
	sizet := t
	if c.Cancelled {
		// Freeze expansion of cancelled presses.
		sizet = endt
	}
	sizet /= expandDuration

	// Animate only ended presses, and presses that are fading in.
	if !c.End.IsZero() || sizet <= 1.0 {
		gtx.Execute(op.InvalidateCmd{})
	}

	if sizet > 1.0 {
		sizet = 1.0
	}

	if alphat > .5 {
		// Start fadeout after half the animation.
		alphat = 1.0 - alphat
	}
	// Twice the speed to attain fully faded in at 0.5.
	t2 := alphat * 2
	// Beziér ease-in curve.
	alphaBezier := t2 * t2 * (3.0 - 2.0*t2)
	sizeBezier := sizet * sizet * (3.0 - 2.0*sizet)
	size := gtx.Constraints.Min.X
	if h := gtx.Constraints.Min.Y; h > size {
		size = h
	}
	// Cover the entire constraints min rectangle and
	// apply curve values to size and color.
	size = int(float32(size) * 2 * float32(math.Sqrt(2)) * sizeBezier)
	alpha := 0.7 * alphaBezier
	const col = 0.8
	ba, bc := byte(alpha*0xff), byte(col*0xff)
	rgba := MulAlpha(color.NRGBA{A: 0xff, R: bc, G: bc, B: bc}, ba)
	ink := paint.ColorOp{Color: rgba}
	ink.Add(gtx.Ops)
	rr := size / 2
	defer op.Offset(c.Position.Add(image.Point{
		X: -rr,
		Y: -rr,
	})).Push(gtx.Ops).Pop()
	// defer clip.UniformRRect(image.Rectangle{Max: image.Pt(size, size)}, rr).Push(gtx.Ops).Pop()
	paint.PaintOp{}.Add(gtx.Ops)
}
