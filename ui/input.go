package ui

import (
	"fmt"
	"math"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"nestor/hw"
)

type inputConfigDialog struct{}

func openInputConfigDialog(cfg *hw.InputConfig) {
	var icd inputConfigDialog

	// Create a new GTK dialog
	dlg := mustT(gtk.DialogNew())
	dlg.SetTitle("NES Controller Configuration")
	dlg.SetModal(true)
	dlg.SetDefaultSize(800, 400) // Adjusted window size to accommodate all widgets

	// Create a Notebook for tabs
	notebook := mustT(gtk.NotebookNew())

	for i := range cfg.Paddles {
		tabLabel := fmt.Sprintf("Controller %d", i+1)
		content := icd.createControllerTab(&cfg.Paddles[i])
		notebook.AppendPage(content, mustT(gtk.LabelNew(tabLabel)))
	}

	// Add the notebook to the dialog's content area
	contentArea := mustT(dlg.GetContentArea())
	contentArea.Add(notebook)

	// Show all widgets in the dialog
	dlg.ShowAll()
	dlg.Run()
	dlg.Destroy()
}

// createControllerTab creates the content for each controller tab.
func (inputConfigDialog) createControllerTab(padcfg *hw.PaddleConfig) *gtk.Widget {
	hbox := mustT(gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5))

	// Create a mouse-enabled DrawingArea widget.
	drawingArea := mustT(gtk.DrawingAreaNew())
	drawingArea.SetSizeRequest(400, 300)
	drawingArea.SetEvents(int(gdk.BUTTON_PRESS_MASK))

	treeView, listStore := createPropertyList()

	// Wrap the treeview into a frame.
	frame := mustT(gtk.FrameNew("Controller Mappings"))
	frame.SetShadowType(gtk.SHADOW_ETCHED_IN)
	frame.SetBorderWidth(5)
	frame.Add(treeView)
	frame.SetSizeRequest(200, -1) // Set a minimum width for the property list

	hbox.PackStart(drawingArea, true, true, 0)
	hbox.PackStart(frame, false, false, 0)

	cc := &controllerConfig{
		drawingArea: drawingArea,
		treeView:    treeView,
		listStore:   listStore,
		padcfg:      padcfg,
	}

	// Initialize property list.
	cc.updatePropertyList()

	drawingArea.Connect("draw", cc.drawController)
	drawingArea.Connect("button-press-event", cc.onClick)

	return &hbox.Container.Widget
}

type controllerConfig struct {
	drawingArea *gtk.DrawingArea
	treeView    *gtk.TreeView
	listStore   *gtk.ListStore
	padcfg      *hw.PaddleConfig
}

// createPropertyList creates the TreeView and ListStore for the property list
func createPropertyList() (*gtk.TreeView, *gtk.ListStore) {
	listStore := mustT(gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING))
	treeView := mustT(gtk.TreeViewNewWithModel(listStore))

	// Create the Button Name column
	buttonNameColumn := mustT(gtk.TreeViewColumnNewWithAttribute("Button", mustT(gtk.CellRendererTextNew()), "text", 0))
	buttonNameColumn.SetResizable(true)
	treeView.AppendColumn(buttonNameColumn)

	// Create the Assigned Key column
	assignedKeyColumn := mustT(gtk.TreeViewColumnNewWithAttribute("Assigned Key", mustT(gtk.CellRendererTextNew()), "text", 1))
	assignedKeyColumn.SetResizable(true)
	treeView.AppendColumn(assignedKeyColumn)

	// Set grid lines to make it look like a table
	treeView.SetGridLines(gtk.TREE_VIEW_GRID_LINES_BOTH)

	return treeView, listStore
}

func (cc *controllerConfig) updatePropertyList() {
	cc.listStore.Clear()

	for btn := hw.PadA; btn <= hw.PadRight; btn++ {
		iter := cc.listStore.Append()
		must(cc.listStore.Set(iter, []int{0, 1}, []any{btn.String(), cc.padcfg.GetMapping(btn)}))
	}
}

// boundingBox is an axis-aligned bounding box.
type boundingBox struct {
	xmin, ymin, xmax, ymax float64
}

// contains reports whether the point (x,y) is inside bb.
func (bb boundingBox) contains(x, y float64) bool {
	return x >= bb.xmin && x <= bb.xmax && y >= bb.ymin && y <= bb.ymax
}

// controllerGeometry holds the computed dimensions and positions
type controllerGeometry struct {
	scaleX, scaleY, scale float64
	marginX, marginY      float64
	left, upy             float64
	xselect, xstart       float64
	xb, xa                float64
	yselect               float64
	side                  float64
	vertCenter            float64
	margin                float64
	section               float64

	buttonBoxes [hw.PadButtonCount]boundingBox
}

// computeGeometry computes geometry based on the drawing area size.
func (cc *controllerConfig) computeGeometry(da *gtk.DrawingArea) controllerGeometry {
	// Get the size of the drawing area
	allocation := da.GetAllocation()
	width := float64(allocation.GetWidth())
	height := float64(allocation.GetHeight())

	// TODO: just define once
	const baseWidth = 600.0
	const baseHeight = 300.0
	const controllerAspectRatio = baseWidth / baseHeight

	// Calculate the maximum width and height that maintain the aspect ratio
	drawingWidth := width
	drawingHeight := width / controllerAspectRatio

	if drawingHeight > height {
		drawingHeight = height
		drawingWidth = height * controllerAspectRatio
	}

	// Calculate the margins to center the controller
	marginX := (width - drawingWidth) / 2
	marginY := (height - drawingHeight) / 2

	// Calculate the scaling factor
	scaleX := drawingWidth / baseWidth
	scaleY := drawingHeight / baseHeight
	scale := math.Min(scaleX, scaleY)

	// Compute constants
	margin := baseWidth / 20
	side := baseHeight / 8
	vertCenter := baseHeight / 2
	left := baseWidth / 10
	section := 3 * side
	upy := vertCenter - section/2

	xselect := left + section + margin
	yselect := vertCenter - side/2
	xstart := xselect + side*3

	xb := xstart + section + side
	xa := xb + 1.5*side

	// Compute paddle buttons bounding boxes.
	buttonBoxes := [hw.PadButtonCount]boundingBox{
		hw.PadUp: {
			xmin: left + side, ymin: upy,
			xmax: left + 2*side, ymax: upy + side,
		},
		hw.PadDown: {
			xmin: left + side, ymin: upy + 2*side,
			xmax: left + 2*side, ymax: upy + 3*side,
		},
		hw.PadLeft: {
			xmin: left, ymin: upy + side,
			xmax: left + side, ymax: upy + 2*side,
		},
		hw.PadRight: {
			xmin: left + 2*side, ymin: upy + side,
			xmax: left + 3*side, ymax: upy + 2*side,
		},
		hw.PadSelect: {
			xmin: xselect, ymin: yselect,
			xmax: xselect + 2*side, ymax: yselect + side,
		},
		hw.PadStart: {
			xmin: xstart, ymin: yselect,
			xmax: xstart + 2*side, ymax: yselect + side,
		},
		hw.PadB: {
			xmin: xb - side*0.60, ymin: vertCenter - side*0.60,
			xmax: xb + side*0.60, ymax: vertCenter + side*0.60,
		},
		hw.PadA: {
			xmin: xa - side*0.60, ymin: vertCenter - side*0.60,
			xmax: xa + side*0.60, ymax: vertCenter + side*0.60,
		},
	}

	return controllerGeometry{
		scaleX:      scaleX,
		scaleY:      scaleY,
		scale:       scale,
		marginX:     marginX,
		marginY:     marginY,
		left:        left,
		upy:         upy,
		xselect:     xselect,
		xstart:      xstart,
		xb:          xb,
		xa:          xa,
		yselect:     yselect,
		side:        side,
		vertCenter:  vertCenter,
		margin:      margin,
		section:     section,
		buttonBoxes: buttonBoxes,
	}
}

// drawController renders the NES controller on the drawing area
func (cc *controllerConfig) drawController(da *gtk.DrawingArea, cr *cairo.Context) {
	cr.SetLineWidth(2)

	// Compute geometry
	geom := cc.computeGeometry(da)

	// Apply scaling and translation
	cr.Translate(geom.marginX, geom.marginY)
	cr.Scale(geom.scale, geom.scale)

	// Now, draw the controller using the base dimensions
	const baseWidth = 600.0
	const baseHeight = 300.0

	// Draw the controller body with rounded corners
	cr.SetSourceRGB(0.9, 0.9, 0.9) // Light grey body
	radius := 20.0
	cr.NewPath()
	cr.MoveTo(geom.margin+radius, geom.margin)
	cr.LineTo(baseWidth-geom.margin-radius, geom.margin)
	cr.Arc(baseWidth-geom.margin-radius, geom.margin+radius, radius, -math.Pi/2, 0)
	cr.LineTo(baseWidth-geom.margin, baseHeight-geom.margin-radius)
	cr.Arc(baseWidth-geom.margin-radius, baseHeight-geom.margin-radius, radius, 0, math.Pi/2)
	cr.LineTo(geom.margin+radius, baseHeight-geom.margin)
	cr.Arc(geom.margin+radius, baseHeight-geom.margin-radius, radius, math.Pi/2, math.Pi)
	cr.LineTo(geom.margin, geom.margin+radius)
	cr.Arc(geom.margin+radius, geom.margin+radius, radius, math.Pi, 3*math.Pi/2)
	cr.ClosePath()
	cr.FillPreserve()
	cr.SetSourceRGB(0, 0, 0)
	cr.Stroke()

	// Draw a stripe at the top
	cr.SetSourceRGB(0.7, 0.7, 0.7)
	cr.Rectangle(geom.margin, geom.margin, baseWidth-2*geom.margin, baseHeight/6)
	cr.Fill()

	// Draw the D-Pad with rounded edges
	cr.SetSourceRGB(0, 0, 0)
	cr.Save()
	cr.Translate(geom.left+geom.side+geom.side/2, geom.upy+1.5*geom.side)
	cr.MoveTo(-geom.side/2, -1.5*geom.side)
	cr.LineTo(geom.side/2, -1.5*geom.side)
	cr.LineTo(geom.side/2, -geom.side/2)
	cr.LineTo(1.5*geom.side, -geom.side/2)
	cr.LineTo(1.5*geom.side, geom.side/2)
	cr.LineTo(geom.side/2, geom.side/2)
	cr.LineTo(geom.side/2, 1.5*geom.side)
	cr.LineTo(-geom.side/2, 1.5*geom.side)
	cr.LineTo(-geom.side/2, geom.side/2)
	cr.LineTo(-1.5*geom.side, geom.side/2)
	cr.LineTo(-1.5*geom.side, -geom.side/2)
	cr.LineTo(-geom.side/2, -geom.side/2)
	cr.ClosePath()
	cr.Fill()
	cr.Restore()

	// Draw Select/Start with rounded rectangles.
	cr.SetSourceRGB(0.7, 0.7, 0.7)
	drawRoundedRectangle(cr, geom.xselect, geom.yselect, geom.side*2, geom.side, geom.side/4)
	cr.Fill()

	drawRoundedRectangle(cr, geom.xstart, geom.yselect, geom.side*2, geom.side, geom.side/4)
	cr.Fill()

	// Draw A and B as circles.
	cr.SetSourceRGB(1, 0, 0)
	cr.Arc(geom.xb, geom.vertCenter, geom.side*0.60, 0, 2*math.Pi)
	cr.Fill()
	cr.Arc(geom.xa, geom.vertCenter, geom.side*0.60, 0, 2*math.Pi)
	cr.Fill()

	// Draw labels (centered text)
	cr.SetSourceRGB(0, 0, 0)
	cr.SelectFontFace("Sans", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_BOLD)
	cr.SetFontSize(14)

	text := "Select"
	extents := cr.TextExtents(text)
	textX := geom.xselect + (geom.side*2-extents.Width)/2 - extents.XBearing
	textY := geom.yselect + geom.side - extents.Height
	cr.MoveTo(textX, textY)
	cr.ShowText(text)

	text = "Start"
	extents = cr.TextExtents(text)
	textX = geom.xstart + (geom.side*2-extents.Width)/2 - extents.XBearing
	cr.MoveTo(textX, textY)
	cr.ShowText(text)

	text = "B"
	extents = cr.TextExtents(text)
	textX = geom.xb - extents.Width/2 - extents.XBearing
	textY = geom.vertCenter + geom.side*0.3 - extents.Height/2
	cr.MoveTo(textX, textY)
	cr.ShowText(text)

	text = "A"
	extents = cr.TextExtents(text)
	textX = geom.xa - extents.Width/2 - extents.XBearing
	cr.MoveTo(textX, textY)
	cr.ShowText(text)
}

func drawRoundedRectangle(cr *cairo.Context, x, y, width, height, radius float64) {
	cr.NewPath()
	cr.MoveTo(x+radius, y)
	cr.LineTo(x+width-radius, y)
	cr.Arc(x+width-radius, y+radius, radius, -math.Pi/2, 0)
	cr.LineTo(x+width, y+height-radius)
	cr.Arc(x+width-radius, y+height-radius, radius, 0, math.Pi/2)
	cr.LineTo(x+radius, y+height)
	cr.Arc(x+radius, y+height-radius, radius, math.Pi/2, math.Pi)
	cr.LineTo(x, y+radius)
	cr.Arc(x+radius, y+radius, radius, math.Pi, 3*math.Pi/2)
	cr.ClosePath()
}

func (cc *controllerConfig) onClick(da *gtk.DrawingArea, event *gdk.Event) {
	geom := cc.computeGeometry(da)

	buttonEvent := gdk.EventButtonNewFromEvent(event)
	x, y := buttonEvent.MotionVal()

	// Account for scaling and translation
	x = (x - geom.marginX) / geom.scale
	y = (y - geom.marginY) / geom.scale

	// What paddle button does the user wants to map?
	var tomap hw.PaddleButton = 0xff
	for btn, bb := range geom.buttonBoxes {
		if bb.contains(x, y) {
			tomap = hw.PaddleButton(btn)
			break
		}
	}
	if tomap == 0xff {
		return // Clicked outside of any button
	}

	// Configure/map this button to a key
	code, err := hw.ShowMapInputWindow(tomap.String())
	if err != nil {
		gtk.MessageDialogNew(nil, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Error: %s", err).Run()
		return
	}

	cc.padcfg.SetMapping(tomap, code)
	cc.updatePropertyList()
}
