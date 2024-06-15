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

	Frames chan image.RGBA

	// TODO: remove when react-debugger is at feature parity.
	Debugger hw.Debugger
}

const NoDebugger = ""

// PowerUp initializes a NES with the given ROM, and optionally attach a
// remote debugger at the given address (leave empty to disable).
func PowerUp(rom *ines.Rom, dbgAddr string) (*NES, error) {
	// nes.Rom = rom
	ppu := hw.NewPPU()
	ppu.InitBus()

	cpu := hw.NewCPU(ppu)
	cpu.InitBus()

	ppu.CPU = cpu

	var idbg hw.Debugger
	if dbgAddr == "" {
		idbg = hw.NopDebugger{}
		cpu.SetDebugger(idbg)
	} else {
		dbg, err := debugger.NewDebugger(cpu, dbgAddr)
		if err != nil {
			return nil, err
		}
		idbg = dbg
	}

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
		Debugger: idbg,
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
		nes.RunOneFrame()
		nes.Debugger.FrameEnd()
		out.EndFrame(screen)
	}
}

func (nes *NES) RunOneFrame() {
	nes.CPU.Run(29781)
}
