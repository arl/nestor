package emu

import "nestor/hw"

type NESHardware struct {
	CPU *hw.CPU
	PPU *hw.PPU
}

func (hw *NESHardware) Reset() {
	hw.CPU.Reset()
	hw.PPU.Reset()
}
