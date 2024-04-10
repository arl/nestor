package emu

import (
	"fmt"
	"image"
	"io"

	"nestor/emu/debugger"
	"nestor/hw"
	"nestor/hw/mappers"
	"nestor/ines"
)

type NES struct {
	CPU *hw.CPU
	PPU *hw.PPU
	Rom *ines.Rom

	Debugger hw.Debugger
	screenCh chan *image.RGBA
}

func (nes *NES) PowerUp(rom *ines.Rom) error {
	nes.Rom = rom
	nes.PPU = hw.NewPPU()
	nes.PPU.InitBus()

	nes.CPU = hw.NewCPU(nes.PPU)
	nes.CPU.InitBus()

	nes.Debugger = debugger.NewDebugger(nes.CPU)

	nes.PPU.CPU = nes.CPU

	// Load mapper, and ap cartridge memory and hardware based on mapper.
	mapper, ok := mappers.All[rom.Mapper()]
	if !ok {
		return fmt.Errorf("unsupported mapper %03d", rom.Mapper())
	}
	if err := mapper.Load(rom, nes.CPU, nes.PPU); err != nil {
		return fmt.Errorf("error while loading mapper %03d (%s): %s", rom.Mapper(), mapper.Name, err)
	}

	nes.PPU.CreateScreen()
	nes.Reset()
	return nil
}

func (nes *NES) Reset() {
	nes.PPU.Reset()
	nes.CPU.Reset()
}

func (nes *NES) FrameEvents() <-chan *image.RGBA {
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
			nes.screenCh <- nes.PPU.Output()
		}
		nes.Debugger.FrameEnd()
	}
}

func (nes *NES) RunOneFrame() {
	nes.CPU.Run(29781)
	nes.CPU.Clock -= 29781
}

func (nes *NES) RunDisasm(out io.Writer) {
	d := hw.NewDisasm(nes.CPU, out)
	for {
		d.Run(29781)
		nes.CPU.Clock -= 29781
	}
}
