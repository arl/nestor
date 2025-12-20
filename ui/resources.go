package ui

import (
	"image/color"
	_ "image/png"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"nestor/assets"
)

const (
	backgroundColor = 0x131a22   // rgba(19, 26, 34, 1)
	panelColor      = 0x192a3b   // rgba(25, 42, 59, 1)
	menuBackground  = 0x5e5e99e6 // rgba(94, 94, 153, 0.9)

	widgetDisabledColor = 0x5a7a91 // rgba(90, 122, 145, 1)
	widgetIdleColor     = backgroundColor
	widgetMaskColor     = 0xbb296d //  rgba(187, 41, 109, 1)

	textIdleColor     = 0xdff4ff // rgba(242, 223, 255, 1)
	textDisabledColor = widgetDisabledColor

	textErrorColor = widgetMaskColor

	labelIdleColor     = textIdleColor
	labelDisabledColor = textDisabledColor

	buttonTextIdleColor     = textIdleColor
	buttonTextDisabledColor = labelDisabledColor

	listSelectedBackground         = 0x4b687a // rgba(75, 104, 122, 1)
	listDisabledSelectedBackground = 0x2a3944 // rgba(42, 57, 68, 1)
	listFocusedBackground          = 0x2a3944 // rgba(42, 57, 68, 1)

	headerColor = textIdleColor

	textInputCaretColor         = 0xe7c34b // rgba(231, 195, 75, 1)
	textInputDisabledCaretColor = 0x766326 // rgba(118, 99, 38, 1)

	toolTipColor = backgroundColor

	separatorColor = listDisabledSelectedBackground

	sliderBg       = 0x646464 // rgba(100, 100, 100, 1)
	sliderHandleBg = 0xff6464 // rgba(255, 100, 100, 1)
)

type uiResources struct {
	fonts      *fonts
	background *image.NineSlice

	separatorColor color.Color

	base   *baseResources
	images *imageResources

	text        *textResources
	button      *buttonResources
	label       *labelResources
	checkbox    *checkboxResources
	comboButton *comboButtonResources
	list        *listResources
	slider      *sliderResources
	panel       *panelResources
	tabBook     *tabBookResources
	header      *headerResources
	textInput   *textInputResources
	textArea    *textAreaResources
}

var res *uiResources

func initResources() {
	fonts := loadFonts()

	res = &uiResources{}
	res.images = newImageResources()
	res.separatorColor = hex2color(separatorColor)
	res.base = newBaseResources()
	res.background = ninesliceFromHex(backgroundColor)
	res.fonts = fonts
	res.label = newLabelResources(fonts)
	res.text = newTextResources(fonts)
	res.button = newButtonResources(fonts)
	res.checkbox = newCheckboxResources()
	res.slider = newSliderResources()
	res.list = newListResources(fonts)
	res.comboButton = newComboButtonResources(fonts)
	res.panel = newPanelResources()
	res.tabBook = newTabBookResources(fonts)
	res.header = newHeaderResources(fonts)
	res.textInput = newTextInputResources(fonts)
	res.textArea = newTextAreaResources(fonts)
}

type imageResources struct {
	paddle    *image.NineSlice
	paddleimg *ebiten.Image
}

func newImageResources() *imageResources {
	paddleimg, _, err := ebitenutil.NewImageFromFileSystem(assets.FS, "graphics/nes-paddle.png")
	if err != nil {
		panic(err)
	}
	return &imageResources{
		paddle:    image.NewFixedNineSlice(paddleimg),
		paddleimg: paddleimg,
	}
}

type baseResources struct {
	scrollimg *widget.ScrollContainerImage
}

func newBaseResources() *baseResources {
	return &baseResources{
		scrollimg: &widget.ScrollContainerImage{
			Idle:     ninesliceFromHex(widgetIdleColor),
			Disabled: ninesliceFromHex(widgetDisabledColor),
			Mask:     ninesliceFromHex(widgetMaskColor),
		},
	}
}

type textResources struct {
	idleColor     color.Color
	disabledColor color.Color
	face          *text.Face
	titleFace     *text.Face
	bigTitleFace  *text.Face
	smallFace     *text.Face
}

func newTextResources(fonts *fonts) *textResources {
	return &textResources{
		idleColor:     hex2color(textIdleColor),
		disabledColor: hex2color(textDisabledColor),
		face:          fonts.face,
		titleFace:     fonts.titleFace,
		bigTitleFace:  fonts.bigTitleFace,
		smallFace:     fonts.toolTipFace,
	}
}

type buttonResources struct {
	image   *widget.ButtonImage
	text    *widget.ButtonTextColor
	face    *text.Face
	padding *widget.Insets
}

func newButtonResources(fonts *fonts) *buttonResources {
	const (
		buttonIdleColor     = 0x2a3944 // rgba(42, 57, 68, 1)
		buttonIdleBorder    = 0x4b687a // rgba(75, 104, 122, 1)
		buttonHoverColor    = 0x4b687a // rgba(75, 104, 122, 1)
		buttonHoverBorder   = 0x5a7a91 // rgba(90, 122, 145, 1)
		buttonPressedColor  = 0x192a3b // rgba(25, 42, 59, 1)
		buttonPressedBorder = 0x4b687a // rgba(75, 104, 122, 1)
	)

	i := &widget.ButtonImage{
		Idle: image.NewBorderedNineSliceColor(
			hex2color(buttonIdleColor),
			hex2color(buttonIdleBorder),
			1),
		Hover: image.NewBorderedNineSliceColor(
			hex2color(buttonHoverColor),
			hex2color(buttonHoverBorder),
			1,
		),
		Pressed: image.NewAdvancedNineSliceColor(
			hex2color(buttonPressedColor),
			image.NewBorder(1, 1, 1, 1, hex2color(buttonPressedBorder)),
		),
	}

	return &buttonResources{
		image: i,
		text: &widget.ButtonTextColor{
			Idle:     hex2color(buttonTextIdleColor),
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

type checkboxResources struct {
	image   *widget.CheckboxImage
	spacing int
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

type labelResources struct {
	text *widget.LabelColor
	face *text.Face
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

type comboButtonResources struct {
	text    *widget.ButtonTextColor
	face    *text.Face
	padding *widget.Insets
}

func newComboButtonResources(fonts *fonts) *comboButtonResources {
	return &comboButtonResources{
		face:    fonts.small,
		padding: res.button.padding,
		text:    res.button.text,
	}
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

func newListResources(fonts *fonts) *listResources {
	return &listResources{
		face: fonts.face,
		entryPadding: &widget.Insets{
			Left:   30,
			Right:  30,
			Top:    2,
			Bottom: 2,
		},
		image: res.base.scrollimg,
		track: res.slider.trackImage,
		trackPadding: &widget.Insets{
			Top:    5,
			Bottom: 24,
		},
		handle:     res.slider.handle,
		handleSize: res.slider.handleSize,
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

type sliderResources struct {
	trackImage *widget.SliderTrackImage
	handle     *widget.ButtonImage
	handleSize *int
}

func newSliderResources() *sliderResources {
	idle := ninesliceFromHex(sliderBg)
	hover := ninesliceFromHex(sliderBg)
	disabled := ninesliceFromHex(sliderBg)

	handleIdle := ninesliceFromHex(sliderHandleBg)
	handleHover := ninesliceFromHex(sliderHandleBg)
	handlePressed := ninesliceFromHex(sliderHandleBg)
	handleDisabled := ninesliceFromHex(sliderHandleBg)

	return &sliderResources{
		handleSize: ptrTo(6),
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

type panelResources struct {
	image   *image.NineSlice
	padding *widget.Insets
}

func newPanelResources() *panelResources {
	return &panelResources{
		image: ninesliceFromHex(panelColor),
		padding: &widget.Insets{
			Left:   30,
			Right:  30,
			Top:    20,
			Bottom: 20,
		},
	}
}

type tabBookResources struct {
	buttonFace    *text.Face
	buttonText    *widget.ButtonTextColor
	buttonPadding *widget.Insets
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

type headerResources struct {
	background *image.NineSlice
	padding    *widget.Insets
	face       *text.Face
	color      color.Color
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

type textInputResources struct {
	image   *widget.TextInputImage
	padding *widget.Insets
	face    *text.Face
	color   *widget.TextInputColor
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

type textAreaResources struct {
	image        *widget.ScrollContainerImage
	track        *widget.SliderTrackImage
	trackPadding *widget.Insets
	handle       *widget.ButtonImage
	handleSize   *int
	face         *text.Face
	entryPadding *widget.Insets
}

func newTextAreaResources(fonts *fonts) *textAreaResources {
	return &textAreaResources{
		face:       fonts.face,
		handleSize: ptrTo(5),
		image:      res.base.scrollimg,
		track:      res.slider.trackImage,
		trackPadding: &widget.Insets{
			Top:    5,
			Bottom: 24,
		},
		handle: res.slider.handle,
		entryPadding: &widget.Insets{
			Left:   30,
			Right:  30,
			Top:    2,
			Bottom: 2,
		},
	}
}
