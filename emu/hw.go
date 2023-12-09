package emu

import (
	"nestor/cpu"
	"nestor/ppu"
)

type Hardware struct {
	CPU *cpu.CPU
	PPU *ppu.PPU
}
