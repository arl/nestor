package emu

import (
	"fmt"

	"nestor/hw"
	"nestor/hw/mappers"
	"nestor/ines"
)

type NES struct {
	CPU   *hw.CPU
	PPU   *hw.PPU
	APU   *hw.APU
	Rom   *ines.Rom
	Mixer *hw.AudioMixer
}

func powerUp(rom *ines.Rom) (*NES, error) {
	audioMixer := hw.NewAudioMixer()
	ppu := hw.NewPPU()
	cpu := hw.NewCPU(ppu)
	apu := hw.NewAPU(cpu, audioMixer)

	cpu.APU = apu
	cpu.InitBus()

	// Load mapper.
	mapper, ok := mappers.All[rom.Mapper()]
	if !ok {
		return nil, fmt.Errorf("unsupported mapper %03d", rom.Mapper())
	}
	if err := mapper.Load(rom, cpu, ppu); err != nil {
		return nil, fmt.Errorf("error while loading mapper %03d (%s): %s", rom.Mapper(), mapper.Name, err)
	}

	nes := &NES{
		CPU:   cpu,
		PPU:   ppu,
		APU:   apu,
		Rom:   rom,
		Mixer: audioMixer,
	}
	nes.Reset(false)
	return nes, nil
}

func (nes *NES) Reset(soft bool) {
	nes.PPU.Reset()
	nes.CPU.Reset(soft)
	nes.APU.Reset(soft)
	nes.Mixer.Reset()
}

func (nes *NES) RunOneFrame(frame hw.Frame) {
	nes.PPU.SetFrameBuffer(frame.Video)
	nes.CPU.Run(29781)
}
