package ui

import (
	"slices"
	"strconv"

	"github.com/ebitenui/ebitenui/utilities/mobile"
	"github.com/ebitenui/ebitenui/widget"

	"github.com/arl/nestor/ui/shader"
)

func (s *configState) videoConfigPage() *page {
	c := newPageContentContainer()
	c.SetBackgroundImage(res.panel.image)

	monitorBlock := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(20))))

	onMonitorChange := func(args *widget.TextInputChangedEventArgs) {
		var idxmon int
		idxmon, err := strconv.Atoi(args.InputText)
		if err != nil {
			modUI.ErrorZ("monitor: unexpected input").String("txt", args.InputText).End()
			return
		}
		if idxmon == int(s.app.cfg.Video.Monitor) {
			return
		}

		prev := s.app.cfg.Video.Monitor
		s.app.cfg.Video.Monitor = uint(idxmon)
		modUI.InfoZ("changed monitor").
			Uint("from", prev).
			Uint("to", s.app.cfg.Video.Monitor).
			End()

		s.app.savecfg()
	}

	monitorInput := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),

		widget.TextInputOpts.MobileInputMode(mobile.NUMERIC),
		widget.TextInputOpts.Image(res.textInput.image),
		widget.TextInputOpts.Face(res.textInput.face),
		widget.TextInputOpts.Color(res.textInput.color),
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),
		widget.TextInputOpts.SubmitHandler(onMonitorChange),
		widget.TextInputOpts.ChangedHandler(onMonitorChange),
		widget.TextInputOpts.Validation(func(txt string) (bool, *string) {
			if _, err := strconv.Atoi(txt); err != nil {
				return false, nil
			}
			return true, nil
		}),
	)
	monitorInput.SetText(strconv.Itoa(int(s.app.cfg.Video.Monitor)))

	monitorBlock.AddChild(
		widget.NewLabel(
			widget.LabelOpts.Text("Monitor", res.label.face, res.label.text),
			widget.LabelOpts.LabelPadding(&widget.Insets{Top: 5, Left: 10}),
			widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionEnd,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				})))),
		monitorInput)

	// shader selection.
	shaderNames := shader.Names()
	idxshader := slices.Index(shaderNames, s.app.cfg.Video.Shader)
	if idxshader < 0 {
		idxshader = shader.DefaultIndex()
	}

	onShaderChanged := func(idxpreset int) {
		from := s.app.cfg.Video.Shader
		to := shaderNames[idxpreset]
		modUI.InfoZ("changed shader").
			String("from", from).
			String("to", to).
			End()

		s.app.cfg.Video.Shader = to
		s.app.savecfg()
		s.createUI()
	}

	shaderBlock := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal))))

	shaderBlock.AddChild(
		widget.NewLabel(
			widget.LabelOpts.Text("Shader", res.label.face, res.label.text),
			widget.LabelOpts.LabelPadding(&widget.Insets{Top: 5, Left: 10}),
			widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionEnd,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				})))),
		newCombobox(shaderNames, idxshader,
			widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				MaxWidth:           200,
			},
			onShaderChanged))

	c.AddChild(
		newCheckbox2States("Enable V-Sync", s.app.cfg.Video.VSync, func(enabled bool) {
			s.app.cfg.Video.VSync = enabled
			modUI.InfoZ("toggled 'enable vsync'").
				Bool("from", !enabled).
				Bool("to", enabled).
				End()

			s.app.savecfg()
		}),
		newCheckbox2States("Start in fullscreen", s.app.cfg.Video.StartFullscreen, func(enabled bool) {
			s.app.cfg.Video.StartFullscreen = enabled
			modUI.InfoZ("toggled start-fullscreen").
				Bool("from", !enabled).
				Bool("to", enabled).
				End()
			s.app.savecfg()
		}),
		monitorBlock,
		shaderBlock,
	)
	return &page{title: "Video", content: c}
}
