// Package input provides the input handling for NES controllers.
package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type EbitenInput struct {
	keys [2][8]ebiten.Key

	scratch [256]ebiten.Key
	cfg     Config
}

func NewEbitenInput(cfg Config) *EbitenInput {
	return &EbitenInput{cfg: cfg}
}

func (ei *EbitenInput) state(idx int) uint8 {
	padcfg := ei.cfg.Paddles[idx]
	if !padcfg.Plugged {
		// TODO: check this
		return 0
	}

	preset := ei.cfg.Paddles[idx].Preset

	keys := inpututil.AppendPressedKeys(ei.scratch[:0])
	buttons := preset.ToButtons()

	state := uint8(0)
	for i, code := range buttons {
		pressed := uint8(0)
		switch code.Type {
		case Keyboard:
			for _, k := range keys {
				if k == code.Scancode {
					pressed = 1
					break
				}
			}
			// case ButtonCtrl:
			// 	if ctrl := Gamectrls.getByGUID(code.CtrlGUID); ctrl != nil {
			// 		pressed = ctrl.Button(code.CtrlButton)
			// 	}
			// case AxisCtrl:
			// 	if ctrl := Gamectrls.getByGUID(code.CtrlGUID); ctrl != nil {
			// 		if ctrl.Axis(code.CtrlAxis) >= JoyAxisThreshold {
			// 			pressed = 1
			// 		}
			// 	}
		}
		state |= pressed << i
	}
	return state
}

func (ei *EbitenInput) LoadState() (uint8, uint8) {
	return ei.state(0), ei.state(1)
}
