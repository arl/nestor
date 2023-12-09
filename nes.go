package main

import (
	"fmt"
	"io"

	"nestor/cpu"
	"nestor/emu"
	"nestor/emu/hwio"
	"nestor/emu/mappers"
	"nestor/ines"
	"nestor/ppu"
)

type NES struct {
	Hw emu.Hardware
}

func (nes *NES) PowerUp(rom *ines.Rom) error {
	ppubus := hwio.NewTable("ppu")
	nes.Hw.PPU = ppu.New(ppubus)

	cpubus := hwio.NewTable("cpu")
	nes.Hw.CPU = cpu.NewCPU(cpubus, nes.Hw.PPU)
	nes.Hw.CPU.InitBus()

	// PPU VRAM (name tables) and mirror.
	vram := make([]byte, 0x1000)
	ppubus.MapMemorySlice(0x2000, 0x2FFF, vram, false)

	// Map cartridge memory and hardware based on mapper.
	return mapCartridge(rom, &nes.Hw)
}

func mapCartridge(rom *ines.Rom, hw *emu.Hardware) error {
	mapper, ok := mappers.All[rom.Mapper()]
	if !ok {
		return fmt.Errorf("unsupported mapper %03d", rom.Mapper())
	}

	if err := mapper.Load(rom, hw); err != nil {
		return fmt.Errorf("failed to load mapper %03d (%s): %s", rom.Mapper(), mapper.Name, err)
	}
	return nil
}

func (nes *NES) Reset() {
	nes.Hw.CPU.Reset()
	nes.Hw.PPU.Reset()
}

func (nes *NES) Run() {
	for {
		nes.Hw.CPU.Run(29692) // random
	}
}

func (nes *NES) RunDisasm(out io.Writer, nestest bool) {
	d := cpu.NewDisasm(nes.Hw.CPU, out, nestest)
	for {
		d.Run(29692) // random
	}
}

type ticker struct{}

func (tt *ticker) Tick() {}
