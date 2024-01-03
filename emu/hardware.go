package emu

import "nestor/hw"

type NESHardware struct {
	CPU *hw.CPU
	PPU *hw.PPU
}
