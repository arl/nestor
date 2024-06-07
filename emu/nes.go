package emu

import (
	"fmt"
	"image"

	"nestor/emu/debugger"
	"nestor/hw"
	"nestor/hw/mappers"
	"nestor/ines"
)

type NES struct {
	CPU *hw.CPU
	PPU *hw.PPU
	Rom *ines.Rom

	Frames   chan image.RGBA
	Debugger hw.Debugger
}

func (nes *NES) PowerUp(rom *ines.Rom) error {
	nes.Rom = rom
	nes.PPU = hw.NewPPU()
	nes.PPU.InitBus()

	nes.CPU = hw.NewCPU(nes.PPU)
	nes.CPU.InitBus()

	nes.Debugger = debugger.NewDebugger(nes.CPU)

	nes.PPU.CPU = nes.CPU

	// Load mapper, applying cartridge memory and hardware based on mapper.
	mapper, ok := mappers.All[rom.Mapper()]
	if !ok {
		return fmt.Errorf("unsupported mapper %03d", rom.Mapper())
	}
	if err := mapper.Load(rom, nes.CPU, nes.PPU); err != nil {
		return fmt.Errorf("error while loading mapper %03d (%s): %s", rom.Mapper(), mapper.Name, err)
	}

	nes.Reset()
	return nil
}

func (nes *NES) Reset() {
	nes.PPU.Reset()
	nes.CPU.Reset()
}

func (nes *NES) Run(out *hw.Output) {
	for {
		screen := out.BeginFrame()
		nes.PPU.SetFrameBuffer(screen)
		nes.RunOneFrame()
		nes.Debugger.FrameEnd()
		out.EndFrame(screen)
	}
}

func (nes *NES) RunOneFrame() {
	nes.CPU.Run(29781)
}
