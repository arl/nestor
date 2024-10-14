package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"nestor/hw"
)

func showControllerConfig(cfg *hw.InputConfig) {
	dlg := mustT(gtk.DialogNew())
	dlg.SetTitle("NES Controller Configuration")
	dlg.SetModal(true)
	dlg.SetDefaultSize(550, 250)

	tabs := mustT(gtk.NotebookNew())

	for i := range cfg.Paddles {
		tabLabel := fmt.Sprintf("Controller %d", i+1)
		content := createControllerTab(&cfg.Paddles[i])
		tabs.AppendPage(content, mustT(gtk.LabelNew(tabLabel)))
	}

	area := mustT(dlg.GetContentArea())
	area.Add(tabs)

	dlg.ShowAll()
	dlg.Run()
	dlg.Destroy()
}

// createControllerTab creates the content for each controller tab.
func createControllerTab(padcfg *hw.PaddleConfig) *gtk.Widget {
	area := mustT(gtk.DrawingAreaNew())
	area.SetEvents(int(gdk.BUTTON_PRESS_MASK))

	const margin = 10
	area.SetMarginStart(margin)
	area.SetMarginEnd(margin)
	area.SetMarginTop(margin)
	area.SetMarginBottom(margin)

	treeView, listStore := createPropertyList()

	// Checkbox for the plugged state.
	plugcheck := mustT(gtk.CheckButtonNewWithLabel("Plugged"))
	plugcheck.SetActive(padcfg.Plugged)
	plugcheck.Connect("toggled", func(cb *gtk.CheckButton) {
		padcfg.Plugged = cb.GetActive()
	})

	// Wrap treeview and checkbox into a vbox and into a frame.
	vbox := mustT(gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 5))
	vbox.PackStart(treeView, false, true, 0)
	vbox.PackStart(plugcheck, false, false, 0)

	frame := mustT(gtk.FrameNew(""))
	frame.SetShadowType(gtk.SHADOW_ETCHED_IN)
	frame.SetBorderWidth(5)
	frame.Add(vbox)
	frame.SetSizeRequest(200, -1)

	// Pack the drawing area and frame into a hbox.
	hbox := mustT(gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5))
	hbox.PackStart(area, true, true, 0)
	hbox.PackStart(frame, false, false, 0)

	cc := &controllerConfig{
		drawingArea: area,
		treeView:    treeView,
		listStore:   listStore,
		padcfg:      padcfg,
	}

	// Initialize property list.
	cc.updatePropertyList()

	area.Connect("draw", cc.onDraw)
	area.Connect("button-press-event", cc.onClick)

	return &hbox.Container.Widget
}

type controllerConfig struct {
	drawingArea *gtk.DrawingArea
	treeView    *gtk.TreeView
	listStore   *gtk.ListStore
	padcfg      *hw.PaddleConfig

	bboxes [hw.PadButtonCount]aabb

	scale float64
}

// createPropertyList creates the TreeView and ListStore for the property list
func createPropertyList() (*gtk.TreeView, *gtk.ListStore) {
	listStore := mustT(gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING))
	treeView := mustT(gtk.TreeViewNewWithModel(listStore))

	// Create the Button Name column
	col := mustT(gtk.TreeViewColumnNewWithAttribute("Button",
		mustT(gtk.CellRendererTextNew()), "text", 0))
	col.SetResizable(true)
	treeView.AppendColumn(col)

	// Create the Assigned Key column
	col = mustT(gtk.TreeViewColumnNewWithAttribute("Assigned Key",
		mustT(gtk.CellRendererTextNew()), "text", 1))
	col.SetResizable(true)
	treeView.AppendColumn(col)

	// Set grid lines to make it look like a table
	treeView.SetGridLines(gtk.TREE_VIEW_GRID_LINES_BOTH)

	return treeView, listStore
}

func (cc *controllerConfig) updatePropertyList() {
	cc.listStore.Clear()

	for btn := hw.PadA; btn <= hw.PadRight; btn++ {
		iter := cc.listStore.Append()
		mapping := cc.padcfg.GetMapping(btn)
		must(cc.listStore.Set(iter, []int{0, 1}, []any{btn.String(), mapping}))
	}
}

func (cc *controllerConfig) computeScale() {
	const (
		xmax = 100.0
		ymax = 42.0
	)

	alloc := cc.drawingArea.GetAllocation()
	w := float64(alloc.GetWidth())
	h := float64(alloc.GetHeight())

	// Compute scale to maintain aspect ratio
	if w/h >= xmax/ymax {
		cc.scale = h / ymax
	} else {
		cc.scale = w / xmax
	}
}

// onDraw handles the drawing event
func (cc *controllerConfig) onDraw(da *gtk.DrawingArea, cr *cairo.Context) {
	// Set the transformation matrix
	cc.computeScale()
	cr.Scale(cc.scale, cc.scale)

	// Controller body.
	cr.SetSourceRGB(0.8, 0.8, 0.8) // Light grey
	cr.Rectangle(0, 0, 100, 42)
	drawRoundedRectangle(cr, 0, 0, 100, 42, 2, allCorners)
	cr.Fill()

	// Internal panel.
	cr.SetSourceRGB(0.3, 0.3, 0.3) // Dark grey
	drawRoundedRectangle(cr, 3, 6, 94, 32, 1.5)
	cr.Fill()

	// Directional pad panel.
	cr.SetSourceRGB(0.1, 0.1, 0.1) // Nearly black
	cr.Rectangle(7, 21, 18, 6)
	cr.Rectangle(13, 15, 6, 18)
	cr.Fill()

	cr.SetSourceRGB(0.2, 0.2, 0.2)
	drawArrow(cr, 14, 16, 4, 4, arrowUp)
	cr.Fill()
	drawArrow(cr, 14, 28, 4, 4, arrowDown)
	cr.Fill()
	drawArrow(cr, 8, 22, 4, 4, arrowLeft)
	cr.Fill()
	drawArrow(cr, 20, 22, 4, 4, arrowRight)
	cr.Fill()
	cr.Arc(16, 24, 2, 0, 2*math.Pi)
	cr.Fill()

	cc.bboxes[hw.PadUp] = aabb{13, 15, 20, 21}
	cc.bboxes[hw.PadDown] = aabb{13, 27, 20, 33}
	cc.bboxes[hw.PadLeft] = aabb{7, 21, 13, 27}
	cc.bboxes[hw.PadRight] = aabb{20, 21, 27, 27}

	// Central H lines.
	cr.SetSourceRGB(0.5, 0.5, 0.5)
	drawRoundedRectangle(cr, 31, 6, 28, 5, 1.5, bottomLeft, bottomRight)
	cr.Fill()
	drawRoundedRectangle(cr, 31, 12, 28, 5, 1.5)
	cr.Fill()
	drawRoundedRectangle(cr, 31, 18, 28, 5, 1.5)
	cr.Fill()
	drawRoundedRectangle(cr, 31, 35, 28, 3, 1.5, topLeft, topRight)
	cr.Fill()

	// Select/Start
	//
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
	drawRoundedRectangle(cr, 31, 24, 28, 9, 1.5)
	cr.Fill()

	// Select/start buttons.
	cr.SetSourceRGB(0.1, 0.1, 0.1)
	drawRoundedRectangle(cr, 34, 27.5, 8, 3, 1.5)
	cr.Fill()
	drawRoundedRectangle(cr, 48, 27.5, 8, 3, 1.5)
	cr.Fill()

	cc.bboxes[hw.PadSelect] = aabb{34, 27.5, 42, 30.5}
	cc.bboxes[hw.PadStart] = aabb{48, 27.5, 56, 30.5}

	// B/A

	// B/A panels.
	cr.SetSourceRGB(0.9, 0.9, 0.9)
	drawRoundedRectangle(cr, 65, 24, 10, 10, 1.5)
	cr.Fill()
	drawRoundedRectangle(cr, 77, 24, 10, 10, 1.5)
	cr.Fill()

	cc.bboxes[hw.PadB] = aabb{65, 24, 75, 34}
	cc.bboxes[hw.PadA] = aabb{77, 24, 87, 34}

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

func (cc *controllerConfig) onClick(da *gtk.DrawingArea, event *gdk.Event) {
	cc.computeScale()

	evbtn := gdk.EventButtonNewFromEvent(event)
	x, y := evbtn.MotionVal()

	// Account for scaling.
	x /= cc.scale
	y /= cc.scale

	for i, bbox := range cc.bboxes {
		if bbox.contains(x, y) {
			btn := hw.PaddleButton(i)
			code, err := hw.AskForKeybding(btn.String())
			if err != nil {
				gtk.MessageDialogNew(nil, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Error: %s", err).Run()
				return
			}
			code = strings.TrimSpace(code)
			if code != "" {
				cc.padcfg.SetMapping(btn, code)
				cc.updatePropertyList()
			}
			return
		}
	}
}

type arrowDir int

const (
	arrowUp arrowDir = iota
	arrowDown
	arrowLeft
	arrowRight
)

func drawArrow(cr *cairo.Context, x, y, length, width float64, dir arrowDir) {
	cr.NewPath()

	// Dimensions for the arrow
	shaftWidth := width * 0.6  // Width of the arrow shaft
	headLength := length * 0.4 // Length of the arrow head

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

	// Draw shaft
	cr.Rectangle(shaftx, shafty, shaftw, shafth)
	cr.Fill()

	// Draw head
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

func drawRoundedRectangle(cr *cairo.Context, x, y, width, height, radius float64, corners ...corner) {
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

type aabb struct {
	xmin, ymin, xmax, ymax float64
}

func (bb aabb) contains(x, y float64) bool {
	return x >= bb.xmin && x <= bb.xmax && y >= bb.ymin && y <= bb.ymax
}
