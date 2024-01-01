package main

import (
	"fmt"
	"image"
	"io"

	"nestor/emu"
	"nestor/emu/mappers"
	"nestor/hw"
	"nestor/ines"
)

type NES struct {
	Hw emu.NESHardware

	screenCh chan *image.RGBA
}

func (nes *NES) PowerUp(rom *ines.Rom) error {
	nes.Hw.PPU = hw.NewPPU()
	nes.Hw.PPU.InitBus()

	nes.Hw.CPU = hw.NewCPU(nes.Hw.PPU)
	nes.Hw.CPU.InitBus()

	nes.Hw.PPU.CPU = nes.Hw.CPU

	// Map cartridge memory and hardware based on mapper.
	err := mapCartridge(rom, &nes.Hw)
	if err != nil {
		return fmt.Errorf("mapper failed to map cartridge: %s", err)
	}

	nes.Reset()
	return nil
}

func mapCartridge(rom *ines.Rom, hw *emu.NESHardware) error {
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
	nes.Hw.PPU.Reset()
	nes.Hw.CPU.Reset()
}

func (nes *NES) AttachScreen() <-chan *image.RGBA {
	if nes.screenCh != nil {
		panic("screen already attached")
	}
	nes.screenCh = make(chan *image.RGBA)
	return nes.screenCh
}

func (nes *NES) Run() {
	for {
		nes.RunOneFrame()
		if nes.screenCh != nil {
			nes.screenCh <- nes.Hw.PPU.Output()
		}
	}
}

func (nes *NES) RunOneFrame() {
	nes.Hw.CPU.Run(29781)
	nes.Hw.CPU.Clock -= 29781
}

func (nes *NES) RunDisasm(out io.Writer) {
	d := hw.NewDisasm(nes.Hw.CPU, out)
	for {
		d.Run(29781)
		nes.Hw.CPU.Clock -= 29781
	}
}
