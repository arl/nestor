package main

import (
	"fmt"
	"io"

	"nestor/cpu"
	"nestor/emu/hwio"
	"nestor/ines"
	"nestor/ppu"
)

type NES struct {
	CPU *cpu.CPU
	PPU *ppu.PPU
}

func (nes *NES) PowerUp(rom *ines.Rom) error {
	ppubus := hwio.NewTable("ppu")
	nes.PPU = ppu.New(ppubus)

	cpubus := hwio.NewTable("cpu")
	nes.CPU = cpu.NewCPU(cpubus, nes.PPU)
	nes.CPU.InitBus()

	// PPU VRAM (name tables) and mirror.
	vram := make([]byte, 0x1000)
	ppubus.MapMemorySlice(0x2000, 0x2FFF, vram, false)

	if rom.Mapper() != 0 {
		// Only handle mapper 000 (NROM) for now.
		return fmt.Errorf("unsupported mapper: %d", rom.Mapper())
	}

	if err := loadMapper000(rom, nes); err != nil {
		return fmt.Errorf("failed to load mapper %03d: %s", rom.Mapper(), err)
	}
	return nil
}

// Reset forwards the reset signal to all hardware.
func (nes *NES) Reset() {
	nes.CPU.Reset()
	nes.PPU.Reset()
}

func (nes *NES) Run() {
	for {
		nes.CPU.Run(29692) // random
	}
}

func (nes *NES) RunDisasm(out io.Writer, nestest bool) {
	d := cpu.NewDisasm(nes.CPU, out, nestest)
	for {
		d.Run(29692) // random
	}
}

type ticker struct{}

func (tt *ticker) Tick() {}
