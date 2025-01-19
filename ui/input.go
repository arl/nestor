package ui

import (
	_ "embed"
	"fmt"
	"math"
	"strconv"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"

	"nestor/hw/input"
)

//go:embed input.glade
var inputUI string

type controlCfgDialog struct {
	*gtk.Dialog

	cfg    *input.Config
	curpad int // currently visible paddle

	drawArea  *gtk.DrawingArea
	listStore *gtk.ListStore
	plugcheck *gtk.CheckButton
	bboxes    [input.PadButtonCount]aabbox

	devices map[string]int // allows to give each joystick a number without using the GUID

	drawScale float64
}

func showControllerConfig(cfg *input.Config) {
	builder := mustT(gtk.BuilderNewFromString(inputUI))
	dlg := controlCfgDialog{
		cfg:       cfg,
		curpad:    0,
		drawScale: 3.6,
		devices:   map[string]int{"": 0},
		Dialog:    build[gtk.Dialog](builder, "input_dialog"),
		plugcheck: build[gtk.CheckButton](builder, "plugged_chk"),
		drawArea:  build[gtk.DrawingArea](builder, "paddle_drawing"),
		listStore: mustT(gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)),
	}
	radioPad1 := build[gtk.RadioButton](builder, "paddle1_radio")
	radioPad2 := build[gtk.RadioButton](builder, "paddle2_radio")
	treeView := build[gtk.TreeView](builder, "treeview")
	presets := build[gtk.ComboBoxText](builder, "presets_combo")

	treeView.SetModel(dlg.listStore)
	typecell := mustT(gtk.CellRendererTextNew())
	namecell := mustT(gtk.CellRendererTextNew())
	devcell := mustT(gtk.CellRendererTextNew())
	typecell.SetProperty("weight", pango.WEIGHT_LIGHT)
	namecell.SetProperty("weight", pango.WEIGHT_NORMAL)
	devcell.SetProperty("weight", pango.WEIGHT_NORMAL)
	typecol := mustT(gtk.TreeViewColumnNewWithAttribute("Type", typecell, "text", 0))
	namecol := mustT(gtk.TreeViewColumnNewWithAttribute("Name", namecell, "text", 1))
	devcol := mustT(gtk.TreeViewColumnNewWithAttribute("Device", devcell, "text", 2))
	treeView.AppendColumn(typecol)
	treeView.AppendColumn(namecol)
	treeView.AppendColumn(devcol)

	dlg.drawArea.Connect("draw", dlg.onDrawPaddle)
	presets.Connect("changed", dlg.onPresetChanged)
	dlg.drawArea.Connect("button-press-event", dlg.onClick)
	dlg.plugcheck.Connect("toggled", func(cb *gtk.CheckButton) {
		dlg.cfg.Paddles[dlg.curpad].Plugged = cb.GetActive()
	})
	radioPad1.Connect("clicked", func() {
		dlg.curpad = 0
		presets.SetActive(int(cfg.Paddles[dlg.curpad].PaddlePreset))
	})
	radioPad2.Connect("clicked", func() {
		dlg.curpad = 1
		presets.SetActive(int(cfg.Paddles[dlg.curpad].PaddlePreset))
	})

	presets.SetActive(int(cfg.Paddles[0].PaddlePreset))

	dlg.ShowAll()
	dlg.Run()
	dlg.Destroy()
}

func (dlg *controlCfgDialog) onPresetChanged(presets *gtk.ComboBoxText) {
	dlg.cfg.Paddles[dlg.curpad].PaddlePreset = uint(presets.GetActive())
	dlg.cfg.Paddles[dlg.curpad].Preset = &dlg.cfg.Presets[dlg.cfg.Paddles[dlg.curpad].PaddlePreset]
	dlg.updatePaddleCfg()
}

func (dlg *controlCfgDialog) updatePaddleCfg() {
	dlg.plugcheck.SetActive(dlg.cfg.Paddles[dlg.curpad].Plugged)
	dlg.updatePropertyList()
}

func (dlg *controlCfgDialog) updatePropertyList() {
	dlg.listStore.Clear()

	for btn := input.PadA; btn <= input.PadRight; btn++ {
		iter := dlg.listStore.Append()
		mapping := dlg.cfg.Paddles[dlg.curpad].Preset.Buttons[btn]

		typ := mapping.Type.String()
		name := mapping.Name()
		dev, ok := dlg.devices[mapping.CtrlGUID]
		if !ok {
			dev = len(dlg.devices)
			dlg.devices[mapping.CtrlGUID] = dev
		}
		devstr := ""
		if dev > 0 {
			devstr = strconv.Itoa(dev)
		}

		must(dlg.listStore.Set(iter, []int{0, 1, 2}, []any{typ, name, devstr}))
	}
}

func (dlg *controlCfgDialog) captureInput(btn input.PaddleButton) {
	dlg.SetSensitive(false)

	// The input capture window is SDL, not gtk, we to run it in a different
	// goroutine to not block gtk event loop.
	go func() {
		text := fmt.Sprintf("%s (Paddle %d)", btn, dlg.curpad+1)
		code, err := input.Capture(text)

		glib.IdleAdd(func() {
			defer dlg.SetSensitive(true)

			if err != nil {
				gtk.MessageDialogNew(nil, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Error: %s", err).Run()
				return
			}

			if code.Type == input.ControlNotSet {
				return
			}

			dlg.cfg.Paddles[dlg.curpad].Preset.Buttons[btn] = code
			dlg.updatePropertyList()
		})
	}()
}

func (dlg *controlCfgDialog) onClick(da *gtk.DrawingArea, event *gdk.Event) {
	x, y := gdk.EventButtonNewFromEvent(event).MotionVal()
	x /= dlg.drawScale
	y /= dlg.drawScale

	for i, bbox := range dlg.bboxes {
		if bbox.contains(x, y) {
			dlg.captureInput(input.PaddleButton(i))
			return
		}
	}
}

func (dlg *controlCfgDialog) onDrawPaddle(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.Scale(dlg.drawScale, dlg.drawScale)

	// Paddle body.
	cr.SetSourceRGB(0.8, 0.8, 0.8)
	cr.Rectangle(0, 0, 100, 42)
	roundedRect(cr, 0, 0, 100, 42, 2, allCorners)
	cr.Fill()

	// Internal panel.
	cr.SetSourceRGB(0.3, 0.3, 0.3)
	roundedRect(cr, 3, 6, 94, 32, 1.5)
	cr.Fill()

	// Dpad panel.
	cr.SetSourceRGB(0.1, 0.1, 0.1)
	cr.Rectangle(7, 21, 18, 6)
	cr.Rectangle(13, 15, 6, 18)
	cr.Fill()

	cr.SetSourceRGB(0.2, 0.2, 0.2)
	arrow(cr, 14, 16, 4, 4, arrowUp)
	cr.Fill()
	arrow(cr, 14, 28, 4, 4, arrowDown)
	cr.Fill()
	arrow(cr, 8, 22, 4, 4, arrowLeft)
	cr.Fill()
	arrow(cr, 20, 22, 4, 4, arrowRight)
	cr.Fill()
	cr.Arc(16, 24, 2, 0, 2*math.Pi)
	cr.Fill()

	dlg.bboxes[input.PadUp] = aabbox{13, 15, 20, 21}
	dlg.bboxes[input.PadDown] = aabbox{13, 27, 20, 33}
	dlg.bboxes[input.PadLeft] = aabbox{7, 21, 13, 27}
	dlg.bboxes[input.PadRight] = aabbox{20, 21, 27, 27}

	// Central H lines.
	cr.SetSourceRGB(0.5, 0.5, 0.5)
	roundedRect(cr, 31, 6, 28, 5, 1.5, bottomLeft, bottomRight)
	cr.Fill()
	roundedRect(cr, 31, 12, 28, 5, 1.5)
	cr.Fill()
	roundedRect(cr, 31, 18, 28, 5, 1.5)
	cr.Fill()
	roundedRect(cr, 31, 35, 28, 3, 1.5, topLeft, topRight)
	cr.Fill()

	cr.SelectFontFace("Sans", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_BOLD)
	cr.SetFontSize(2.4)

	cr.SetSourceRGB(0.9, 0, 0)
	cr.MoveTo(33, 21.5)
	cr.ShowText("SELECT")

	cr.SetSourceRGB(0.9, 0, 0)
	cr.MoveTo(47, 21.5)
	cr.ShowText("START")

	// Select/start panel.
	cr.SetSourceRGB(0.9, 0.9, 0.9)
	roundedRect(cr, 31, 24, 28, 9, 1.5)
	cr.Fill()

	// Select/start buttons.
	cr.SetSourceRGB(0.1, 0.1, 0.1)
	roundedRect(cr, 34, 27.5, 8, 3, 1.5)
	cr.Fill()
	roundedRect(cr, 48, 27.5, 8, 3, 1.5)
	cr.Fill()

	dlg.bboxes[input.PadSelect] = aabbox{34, 27.5, 42, 30.5}
	dlg.bboxes[input.PadStart] = aabbox{48, 27.5, 56, 30.5}

	// B/A panels.
	cr.SetSourceRGB(0.9, 0.9, 0.9)
	roundedRect(cr, 65, 24, 10, 10, 1.5)
	cr.Fill()
	roundedRect(cr, 77, 24, 10, 10, 1.5)
	cr.Fill()

	dlg.bboxes[input.PadB] = aabbox{65, 24, 75, 34}
	dlg.bboxes[input.PadA] = aabbox{77, 24, 87, 34}

	// B/A buttons.
	cr.SetSourceRGB(1, 0, 0)
	cr.Arc(70, 29, 4, 0, 2*math.Pi)
	cr.Arc(82, 29, 4, 0, 2*math.Pi)
	cr.Fill()
	cr.SetFontSize(2.6)
	cr.MoveTo(73, 37)
	cr.ShowText("B")
	cr.MoveTo(85, 37)
	cr.ShowText("A")
}

type arrowDir int

const (
	arrowUp arrowDir = iota
	arrowDown
	arrowLeft
	arrowRight
)

func arrow(cr *cairo.Context, x, y, length, width float64, dir arrowDir) {
	cr.NewPath()

	shaftWidth := width * 0.6  // arrow shaft width
	headLength := length * 0.4 // arrow head length

	var (
		shaftx, shafty float64
		shaftw, shafth float64
		headx0, heady0 float64
		headx1, heady1 float64
		headx2, heady2 float64
	)
	switch dir {
	case arrowUp:
		shaftx, shafty = x+(width-shaftWidth)/2, y+headLength
		shaftw, shafth = shaftWidth, length-headLength
		headx0, heady0 = x+width/2, y
		headx1, heady1 = x, y+headLength
		headx2, heady2 = x+width, y+headLength
	case arrowDown:
		shaftx, shafty = x+(width-shaftWidth)/2, y
		shaftw, shafth = shaftWidth, length-headLength
		headx0, heady0 = x+width/2, y+length
		headx1, heady1 = x, y+length-headLength
		headx2, heady2 = x+width, y+length-headLength
	case arrowLeft:
		shaftx, shafty = x+headLength, y+(width-shaftWidth)/2
		shaftw, shafth = length-headLength, shaftWidth
		headx0, heady0 = x, y+width/2
		headx1, heady1 = x+headLength, y
		headx2, heady2 = x+headLength, y+width
	case arrowRight:
		shaftx, shafty = x, y+(width-shaftWidth)/2
		shaftw, shafth = length-headLength, shaftWidth
		headx0, heady0 = x+length, y+width/2
		headx1, heady1 = x+length-headLength, y
		headx2, heady2 = x+length-headLength, y+width
	default:
		panic("unexpected arrow direction")
	}

	cr.Rectangle(shaftx, shafty, shaftw, shafth)
	cr.Fill()
	cr.MoveTo(headx0, heady0)
	cr.LineTo(headx1, heady1)
	cr.LineTo(headx2, heady2)
	cr.ClosePath()
	cr.Fill()
}

type corner byte

const (
	topLeft corner = 1 << iota
	topRight
	bottomLeft
	bottomRight

	allCorners = topLeft | topRight | bottomLeft | bottomRight
)

func roundedRect(cr *cairo.Context, x, y, width, height, radius float64, corners ...corner) {
	cr.NewPath()

	c := allCorners
	if len(corners) > 0 {
		c = 0
		for _, corner := range corners {
			c |= corner
		}
	}

	// Start from the top-left corner
	if c&topLeft != 0 {
		cr.MoveTo(x+radius, y)
	} else {
		cr.MoveTo(x, y)
	}

	// Top edge
	if c&topRight != 0 {
		cr.LineTo(x+width-radius, y)
		cr.Arc(x+width-radius, y+radius, radius, -math.Pi/2, 0)
	} else {
		cr.LineTo(x+width, y)
	}

	// Right edge
	if c&bottomRight != 0 {
		cr.LineTo(x+width, y+height-radius)
		cr.Arc(x+width-radius, y+height-radius, radius, 0, math.Pi/2)
	} else {
		cr.LineTo(x+width, y+height)
	}

	// Bottom edge
	if c&bottomLeft != 0 {
		cr.LineTo(x+radius, y+height)
		cr.Arc(x+radius, y+height-radius, radius, math.Pi/2, math.Pi)
	} else {
		cr.LineTo(x, y+height)
	}

	// Left edge
	if c&topLeft != 0 {
		cr.LineTo(x, y+radius)
		cr.Arc(x+radius, y+radius, radius, math.Pi, 3*math.Pi/2)
	} else {
		cr.LineTo(x, y)
	}

	cr.ClosePath()
}

type aabbox struct{ xmin, ymin, xmax, ymax float64 }

func (bb aabbox) contains(x, y float64) bool {
	return x >= bb.xmin && x <= bb.xmax && y >= bb.ymin && y <= bb.ymax
}
