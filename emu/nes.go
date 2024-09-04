package emu

import (
	"fmt"
	"image"
	"io"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/hw/mappers"
	"nestor/ines"
)

type NES struct {
	CPU *hw.CPU
	PPU *hw.PPU
	Rom *ines.Rom

	Frames chan image.RGBA
	Out    Output
}

func PowerUp(rom *ines.Rom) (*NES, error) {
	ppu := hw.NewPPU()
	ppu.InitBus()

	cpu := hw.NewCPU(ppu)
	cpu.InitBus()
	// TODO: gtk3
	// dbg := debugger.NewDebugger(cpu)
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
		CPU: cpu,
		PPU: ppu,
		Rom: rom,
		// TODO: gtk3
		// Debugger: dbg,
	}
	nes.Reset()
	return nes, nil
}

func (nes *NES) Reset() {
	nes.PPU.Reset()
	nes.CPU.Reset()
}

type Output interface {
	io.Closer

	BeginFrame() []byte
	EndFrame([]byte)
	Poll() bool
}

func (nes *NES) SetOutput(out Output) {
	nes.Out = out
}

// Run run the emulator loop until the CPU halts
// or the output window is closed.
func (nes *NES) Run() {
	for nes.Out.Poll() {
		vbuf := nes.Out.BeginFrame()
		halted := !nes.RunOneFrame(vbuf)
		// TODO: gtk3
		// nes.Debugger.FrameEnd()
		nes.Out.EndFrame(vbuf)

		if halted {
			break
		}
	}
	log.ModEmu.InfoZ("Emulation stopped").End()
	if err := nes.Out.Close(); err != nil {
		log.ModEmu.WarnZ("Error closing emulator window").Error("error", err).End()
	}
}

func (nes *NES) RunOneFrame(vbuf []byte) bool {
	nes.PPU.SetFrameBuffer(vbuf)
	return nes.CPU.Run(29781)
}
