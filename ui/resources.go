package ui

import (
	"image/color"

	"nestor/assets"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/constantutil"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	backgroundColor = 0x131a22 // rgb(19, 26, 34)

	textIdleColor     = 0xdff4ff // rgb(223, 244, 255)
	textDisabledColor = 0x5a7a91 // rgb(90, 122, 145)

	labelIdleColor     = textIdleColor
	labelDisabledColor = textDisabledColor

	buttonTextIdleColor     = textIdleColor
	buttonTextDisabledColor = labelDisabledColor

	listSelectedBackground         = 0x4b687a // rgb(75, 104, 122)
	listDisabledSelectedBackground = 0x2a3944 // rgb(42, 57, 68)

	listFocusedBackground = 0x2a3944 // rgb(42, 57, 68)

	headerColor = textIdleColor

	textInputCaretColor         = 0xe7c34b // rgb(231, 195, 75)
	textInputDisabledCaretColor = 0x766326 // rgb(118, 99, 38)

	toolTipColor = backgroundColor

	separatorColor = listDisabledSelectedBackground

	sliderBg       = 0x646464 // rgb(100, 100, 100)
	sliderHandleBg = 0xff6464 // rgb(255, 100, 100)
)

type uiResources struct {
	fonts *fonts

	background *image.NineSlice

	separatorColor color.Color

	text        *textResources
	button      *buttonResources
	label       *labelResources
	checkbox    *checkboxResources
	comboButton *comboButtonResources
	list        *listResources
	slider      *sliderResources
	progressBar *progressBarResources
	panel       *panelResources
	tabBook     *tabBookResources
	header      *headerResources
	textInput   *textInputResources
	textArea    *textAreaResources
	toolTip     *toolTipResources
}

type textResources struct {
	idleColor     color.Color
	disabledColor color.Color
	face          *text.Face
	titleFace     *text.Face
	bigTitleFace  *text.Face
	smallFace     *text.Face
}

type buttonResources struct {
	image   *widget.ButtonImage
	text    *widget.ButtonTextColor
	face    *text.Face
	padding *widget.Insets
}

type checkboxResources struct {
	image   *widget.CheckboxImage
	spacing int
}

type labelResources struct {
	text *widget.LabelColor
	face *text.Face
}

type comboButtonResources struct {
	image   *widget.ButtonImage
	text    *widget.ButtonTextColor
	face    *text.Face
	graphic *widget.GraphicImage
	padding *widget.Insets
}

type listResources struct {
	image        *widget.ScrollContainerImage
	track        *widget.SliderTrackImage
	trackPadding *widget.Insets
	handle       *widget.ButtonImage
	handleSize   *int
	face         *text.Face
	entry        *widget.ListEntryColor
	entryPadding *widget.Insets
}

type sliderResources struct {
	trackImage *widget.SliderTrackImage
	handle     *widget.ButtonImage
	handleSize *int
}

type progressBarResources struct {
	trackImage *widget.ProgressBarImage
	fillImage  *widget.ProgressBarImage
}

type panelResources struct {
	image    *image.NineSlice
	titleBar *image.NineSlice
	padding  *widget.Insets
}

type tabBookResources struct {
	buttonFace    *text.Face
	buttonText    *widget.ButtonTextColor
	buttonPadding *widget.Insets
}

type headerResources struct {
	background *image.NineSlice
	padding    *widget.Insets
	face       *text.Face
	color      color.Color
}

type textInputResources struct {
	image   *widget.TextInputImage
	padding *widget.Insets
	face    *text.Face
	color   *widget.TextInputColor
}

type textAreaResources struct {
	image        *widget.ScrollContainerImage
	track        *widget.SliderTrackImage
	trackPadding *widget.Insets
	handle       *widget.ButtonImage
	handleSize   *int
	face         *text.Face
	entryPadding *widget.Insets
}

type toolTipResources struct {
	background *image.NineSlice
	padding    *widget.Insets
	face       *text.Face
	color      color.Color
}

func newUIResources() *uiResources {
	background := image.NewNineSliceColor(hex2color(backgroundColor))

	fonts := loadFonts()
	button := newButtonResources(fonts)
	checkbox := newCheckboxResources()
	comboButton := newComboButtonResources(fonts)
	list := newListResources(fonts)
	slider := newSliderResources()
	progressBar := newProgressBarResources()
	panel := newPanelResources()
	tabBook := newTabBookResources(fonts)
	header := newHeaderResources(fonts)
	textInput := newTextInputResources(fonts)
	textArea := newTextAreaResources(fonts)
	toolTip := newToolTipResources(fonts)

	return &uiResources{
		fonts:          fonts,
		button:         button,
		label:          newLabelResources(fonts),
		checkbox:       checkbox,
		comboButton:    comboButton,
		list:           list,
		slider:         slider,
		panel:          panel,
		tabBook:        tabBook,
		header:         header,
		textInput:      textInput,
		toolTip:        toolTip,
		textArea:       textArea,
		progressBar:    progressBar,
		background:     background,
		separatorColor: hex2color(separatorColor),
		text: &textResources{
			idleColor:     hex2color(textIdleColor),
			disabledColor: hex2color(textDisabledColor),
			face:          fonts.face,
			titleFace:     fonts.titleFace,
			bigTitleFace:  fonts.bigTitleFace,
			smallFace:     fonts.toolTipFace,
		},
	}
}

func newButtonResources(fonts *fonts) *buttonResources {
	const (
		buttonIdleColor     = 0xaaaab4 // rgb(170, 170, 180)
		buttonIdleBorder    = 0x5a5a5a // rgb(90, 90, 90)
		buttonHoverColor    = 0x828296 // rgb(130, 130, 150)
		buttonHoverBorder   = 0x464646 // rgb(70, 70, 70)
		buttonPressedColor  = 0x828296 // rgb(130, 130, 150)
		buttonPressedBorder = 0x464646 // rgb(70, 70, 70)
	)

	i := &widget.ButtonImage{
		Idle: image.NewBorderedNineSliceColor(
			hex2color(buttonIdleColor),
			hex2color(buttonIdleBorder),
			3),
		Hover: image.NewBorderedNineSliceColor(
			hex2color(buttonIdleColor),
			hex2color(buttonIdleBorder),
			3,
		),
		Pressed: image.NewAdvancedNineSliceColor(
			hex2color(buttonPressedColor),
			image.NewBorder(3, 2, 2, 2, hex2color(buttonPressedBorder)),
		),
	}

	return &buttonResources{
		image: i,
		text: &widget.ButtonTextColor{
			Idle:     hex2color(buttonTextDisabledColor),
			Disabled: hex2color(buttonTextDisabledColor),
		},
		face: fonts.face,
		padding: &widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		},
	}
}

func newCheckboxResources() *checkboxResources {
	f1 := must(assets.FS.Open("graphics/checkbox-idle.png"))
	idle, _, _ := ebitenutil.NewImageFromReader(f1)

	f2 := must(assets.FS.Open("graphics/checkbox-checked.png"))
	checked, _, _ := ebitenutil.NewImageFromReader(f2)

	f3 := must(assets.FS.Open("graphics/checkbox-greyed.png"))
	greyed, _, _ := ebitenutil.NewImageFromReader(f3)

	f4 := must(assets.FS.Open("graphics/checkbox-hover.png"))
	idle_hovered, _, _ := ebitenutil.NewImageFromReader(f4)

	f5 := must(assets.FS.Open("graphics/checkbox-checked-hover.png"))
	checked_hovered, _, _ := ebitenutil.NewImageFromReader(f5)

	f6 := must(assets.FS.Open("graphics/checkbox-greyed-hover.png"))
	greyed_hovered, _, _ := ebitenutil.NewImageFromReader(f6)

	f7 := must(assets.FS.Open("graphics/checkbox-disabled.png"))
	idle_disabled, _, _ := ebitenutil.NewImageFromReader(f7)

	f8 := must(assets.FS.Open("graphics/checkbox-checked-disabled.png"))
	checked_disabled, _, _ := ebitenutil.NewImageFromReader(f8)

	f9 := must(assets.FS.Open("graphics/checkbox-greyed-disabled.png"))
	greyed_disabled, _, _ := ebitenutil.NewImageFromReader(f9)

	return &checkboxResources{
		spacing: 10,
		image: &widget.CheckboxImage{
			Unchecked:         image.NewFixedNineSlice(idle),
			Checked:           image.NewFixedNineSlice(checked),
			Greyed:            image.NewFixedNineSlice(greyed),
			UncheckedHovered:  image.NewFixedNineSlice(idle_hovered),
			CheckedHovered:    image.NewFixedNineSlice(checked_hovered),
			GreyedHovered:     image.NewFixedNineSlice(greyed_hovered),
			UncheckedDisabled: image.NewFixedNineSlice(idle_disabled),
			CheckedDisabled:   image.NewFixedNineSlice(checked_disabled),
			GreyedDisabled:    image.NewFixedNineSlice(greyed_disabled),
		},
	}
}

func newLabelResources(fonts *fonts) *labelResources {
	return &labelResources{
		face: fonts.face,
		text: &widget.LabelColor{
			Idle:     hex2color(labelIdleColor),
			Disabled: hex2color(labelDisabledColor),
		},
	}
}

func newComboButtonResources(fonts *fonts) *comboButtonResources {
	idle := must(loadImageNineSlice("graphics/combo-button-idle.png", 12, 0))
	hover := must(loadImageNineSlice("graphics/combo-button-hover.png", 12, 0))
	pressed := must(loadImageNineSlice("graphics/combo-button-pressed.png", 12, 0))
	disabled := must(loadImageNineSlice("graphics/combo-button-disabled.png", 12, 0))

	i := &widget.ButtonImage{
		Idle:     idle,
		Hover:    hover,
		Pressed:  pressed,
		Disabled: disabled,
	}

	arrowDown := must(loadGraphicImages("graphics/arrow-down-idle.png", "graphics/arrow-down-disabled.png"))

	return &comboButtonResources{
		image:   i,
		face:    fonts.face,
		graphic: arrowDown,
		padding: &widget.Insets{
			Left:  30,
			Right: 30,
		},
		text: &widget.ButtonTextColor{
			Idle:     hex2color(buttonTextIdleColor),
			Disabled: hex2color(buttonTextDisabledColor),
		},
	}
}

func newListResources(fonts *fonts) *listResources {
	idle := must(newImageFromFile("graphics/list-idle.png"))
	disabled := must(newImageFromFile("graphics/list-disabled.png"))
	mask := must(newImageFromFile("graphics/list-mask.png"))
	trackIdle := must(newImageFromFile("graphics/list-track-idle.png"))
	trackDisabled := must(newImageFromFile("graphics/list-track-disabled.png"))
	handleIdle := must(newImageFromFile("graphics/slider-handle-idle.png"))
	handleHover := must(newImageFromFile("graphics/slider-handle-hover.png"))

	return &listResources{
		face: fonts.face,
		entryPadding: &widget.Insets{
			Left:   30,
			Right:  30,
			Top:    2,
			Bottom: 2,
		},
		image: &widget.ScrollContainerImage{
			Idle:     image.NewNineSlice(idle, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(disabled, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Mask:     image.NewNineSlice(mask, [3]int{26, 10, 23}, [3]int{26, 10, 26}),
		},
		track: &widget.SliderTrackImage{
			Idle:     image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Hover:    image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(trackDisabled, [3]int{0, 5, 0}, [3]int{25, 12, 25}),
		},
		trackPadding: &widget.Insets{
			Top:    5,
			Bottom: 24,
		},
		handle: &widget.ButtonImage{
			Idle:     image.NewNineSliceSimple(handleIdle, 0, 5),
			Hover:    image.NewNineSliceSimple(handleHover, 0, 5),
			Pressed:  image.NewNineSliceSimple(handleHover, 0, 5),
			Disabled: image.NewNineSliceSimple(handleIdle, 0, 5),
		},
		handleSize: constantutil.ConstantToPointer(5),
		entry: &widget.ListEntryColor{
			Unselected:                 hex2color(textIdleColor),
			DisabledUnselected:         hex2color(textDisabledColor),
			Selected:                   hex2color(textIdleColor),
			DisabledSelected:           hex2color(textDisabledColor),
			SelectedBackground:         hex2color(listSelectedBackground),
			DisabledSelectedBackground: hex2color(listDisabledSelectedBackground),
			FocusedBackground:          hex2color(listFocusedBackground),
			SelectedFocusedBackground:  hex2color(listSelectedBackground),
		},
	}
}

func newSliderResources() *sliderResources {
	idle := image.NewNineSliceColor(hex2color(sliderBg))
	hover := image.NewNineSliceColor(hex2color(sliderBg))
	disabled := image.NewNineSliceColor(hex2color(sliderBg))

	handleIdle := image.NewNineSliceColor(hex2color(sliderHandleBg))
	handleHover := image.NewNineSliceColor(hex2color(sliderHandleBg))
	handlePressed := image.NewNineSliceColor(hex2color(sliderHandleBg))
	handleDisabled := image.NewNineSliceColor(hex2color(sliderHandleBg))

	return &sliderResources{
		handleSize: constantutil.ConstantToPointer(6),
		handle: &widget.ButtonImage{
			Idle:     handleIdle,
			Hover:    handleHover,
			Pressed:  handlePressed,
			Disabled: handleDisabled,
		},
		trackImage: &widget.SliderTrackImage{
			Idle:     idle,
			Hover:    hover,
			Disabled: disabled,
		},
	}
}

func newProgressBarResources() *progressBarResources {
	idle := must(newImageFromFile("graphics/progressbar-track-idle.png"))
	fillIdle := must(newImageFromFile("graphics/progressbar-fill-idle.png"))
	disabled := must(newImageFromFile("graphics/slider-track-disabled.png"))

	return &progressBarResources{
		trackImage: &widget.ProgressBarImage{
			Idle:     image.NewNineSlice(idle, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
			Hover:    image.NewNineSlice(idle, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
			Disabled: image.NewNineSlice(disabled, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
		},

		fillImage: &widget.ProgressBarImage{
			Idle:     image.NewNineSlice(fillIdle, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
			Hover:    image.NewNineSlice(fillIdle, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
			Disabled: image.NewNineSlice(fillIdle, [3]int{4, 11, 4}, [3]int{2, 2, 2}),
		},
	}
}

func newPanelResources() *panelResources {
	i := must(loadImageNineSlice("graphics/panel-idle.png", 10, 10))
	t := must(loadImageNineSlice("graphics/titlebar-idle.png", 10, 10))
	return &panelResources{
		image:    i,
		titleBar: t,
		padding: &widget.Insets{
			Left:   30,
			Right:  30,
			Top:    20,
			Bottom: 20,
		},
	}
}

func newTabBookResources(fonts *fonts) *tabBookResources {
	return &tabBookResources{
		buttonFace: fonts.face,
		buttonText: &widget.ButtonTextColor{
			Idle:     hex2color(buttonTextIdleColor),
			Disabled: hex2color(buttonTextDisabledColor),
		},
		buttonPadding: &widget.Insets{
			Left:  30,
			Right: 30,
		},
	}
}

func newHeaderResources(fonts *fonts) *headerResources {
	bg := must(loadImageNineSlice("graphics/header.png", 446, 9))

	return &headerResources{
		face:       fonts.bigTitleFace,
		color:      hex2color(headerColor),
		background: bg,
		padding: &widget.Insets{
			Left:   25,
			Right:  25,
			Top:    4,
			Bottom: 4,
		},
	}
}

func newTextInputResources(fonts *fonts) *textInputResources {
	idle := must(newImageFromFile("graphics/text-input-idle.png"))
	disabled := must(newImageFromFile("graphics/text-input-disabled.png"))

	return &textInputResources{
		image: &widget.TextInputImage{
			Idle:     image.NewNineSlice(idle, [3]int{9, 14, 6}, [3]int{9, 14, 6}),
			Disabled: image.NewNineSlice(disabled, [3]int{9, 14, 6}, [3]int{9, 14, 6}),
		},

		padding: &widget.Insets{
			Left:   8,
			Right:  8,
			Top:    4,
			Bottom: 4,
		},

		face: fonts.face,

		color: &widget.TextInputColor{
			Idle:          hex2color(textIdleColor),
			Disabled:      hex2color(textDisabledColor),
			Caret:         hex2color(textInputCaretColor),
			DisabledCaret: hex2color(textInputDisabledCaretColor),
		},
	}
}

func newTextAreaResources(fonts *fonts) *textAreaResources {
	idle := must(newImageFromFile("graphics/list-idle.png"))
	disabled := must(newImageFromFile("graphics/list-disabled.png"))
	mask := must(newImageFromFile("graphics/list-mask.png"))
	trackIdle := must(newImageFromFile("graphics/list-track-idle.png"))
	trackDisabled := must(newImageFromFile("graphics/list-track-disabled.png"))
	handleIdle := must(newImageFromFile("graphics/slider-handle-idle.png"))
	handleHover := must(newImageFromFile("graphics/slider-handle-hover.png"))

	return &textAreaResources{
		face:       fonts.face,
		handleSize: constantutil.ConstantToPointer(5),

		image: &widget.ScrollContainerImage{
			Idle:     image.NewNineSlice(idle, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(disabled, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Mask:     image.NewNineSlice(mask, [3]int{26, 10, 23}, [3]int{26, 10, 26}),
		},

		track: &widget.SliderTrackImage{
			Idle:     image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Hover:    image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(trackDisabled, [3]int{0, 5, 0}, [3]int{25, 12, 25}),
		},

		trackPadding: &widget.Insets{
			Top:    5,
			Bottom: 24,
		},

		handle: &widget.ButtonImage{
			Idle:     image.NewNineSliceSimple(handleIdle, 0, 5),
			Hover:    image.NewNineSliceSimple(handleHover, 0, 5),
			Pressed:  image.NewNineSliceSimple(handleHover, 0, 5),
			Disabled: image.NewNineSliceSimple(handleIdle, 0, 5),
		},

		entryPadding: &widget.Insets{
			Left:   30,
			Right:  30,
			Top:    2,
			Bottom: 2,
		},
	}
}

func newToolTipResources(fonts *fonts) *toolTipResources {
	bg := must(newImageFromFile("graphics/tool-tip.png"))

	return &toolTipResources{
		face:       fonts.toolTipFace,
		color:      hex2color(toolTipColor),
		background: image.NewNineSlice(bg, [3]int{19, 6, 13}, [3]int{19, 5, 13}),

		padding: &widget.Insets{
			Left:   15,
			Right:  15,
			Top:    10,
			Bottom: 10,
		},
	}
}

// accepts 0xrrggbb or 0xrrggbbaa
func hex2color(val uint32) color.Color {
	alpha := uint8(0xFF)
	if val > 0xffffff {
		alpha = uint8(val & 0xff)
		val = val >> 8
	}

	return color.NRGBA{
		R: uint8(val & 0xff0000 >> 16),
		G: uint8(val & 0xff00 >> 8),
		B: uint8(val & 0xff),
		A: alpha,
	}
}
