package emu

import (
	"fmt"
	"image"

	"nestor/hw"
	"nestor/hw/mappers"
	"nestor/ines"
)

type NES struct {
	CPU *hw.CPU
	PPU *hw.PPU
	Rom *ines.Rom

	Frames chan image.RGBA
}

func powerUp(rom *ines.Rom) (*NES, error) {
	ppu := hw.NewPPU()

	cpu := hw.NewCPU(ppu)
	cpu.InitBus()
	ppu.CPU = cpu

	// Load mapper, applying cartridge memory
	// and hardware based on mapper.
	mapper, ok := mappers.All[rom.Mapper()]
	if !ok {
		return nil, fmt.Errorf("unsupported mapper %03d", rom.Mapper())
	}
	if err := mapper.Load(rom, cpu, ppu); err != nil {
		return nil, fmt.Errorf("error while loading mapper %03d (%s): %s", rom.Mapper(), mapper.Name, err)
	}

	nes := &NES{
		CPU: cpu,
		PPU: ppu,
		Rom: rom,
	}
	nes.Reset()
	return nes, nil
}

func (nes *NES) Reset() {
	nes.PPU.Reset()
	nes.CPU.Reset()
}

func (nes *NES) RunOneFrame(frame hw.Frame) bool {
	nes.PPU.SetFrameBuffer(frame.Video)
	return nes.CPU.Run(29781)
}
