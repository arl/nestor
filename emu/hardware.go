package emu

import (
	"nestor/hw"
	"nestor/ines"
)

type NESHardware struct {
	CPU *hw.CPU
	PPU *hw.PPU
	Rom *ines.Rom
}
