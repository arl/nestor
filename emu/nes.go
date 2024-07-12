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

func PowerUp(rom *ines.Rom) (*NES, error) {
	// nes.Rom = rom
	ppu := hw.NewPPU()
	ppu.InitBus()

	cpu := hw.NewCPU(ppu)
	cpu.InitBus()
	dbg := debugger.NewDebugger(cpu)
	ppu.CPU = cpu

	// Load mapper, applying cartridge memory and hardware based on mapper.
	mapper, ok := mappers.All[rom.Mapper()]
	if !ok {
		return nil, fmt.Errorf("unsupported mapper %03d", rom.Mapper())
	}
	if err := mapper.Load(rom, cpu, ppu); err != nil {
		return nil, fmt.Errorf("error while loading mapper %03d (%s): %s", rom.Mapper(), mapper.Name, err)
	}

	nes := &NES{
		CPU:      cpu,
		PPU:      ppu,
		Rom:      rom,
		Debugger: dbg,
	}
	nes.Reset()
	return nes, nil
}

func (nes *NES) Reset() {
	nes.PPU.Reset()
	nes.CPU.Reset()
}

func (nes *NES) Run(out *hw.Output) {
	for {
		screen := out.BeginFrame()
		nes.PPU.SetFrameBuffer(screen)
		if !nes.RunOneFrame() {
			out.Close()
			break
		}
		nes.Debugger.FrameEnd()
		out.EndFrame(screen)
	}
}

func (nes *NES) RunOneFrame() bool {
	return nes.CPU.Run(29781)
}
