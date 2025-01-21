package emu

import (
	"nestor/hw"
	"nestor/hw/hwdefs"
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

	if err := mappers.Load(rom, cpu, ppu); err != nil {
		return nil, err
	}

	nes := &NES{
		CPU:   cpu,
		PPU:   ppu,
		APU:   apu,
		Rom:   rom,
		Mixer: audioMixer,
	}
	nes.Reset(hwdefs.HardReset)
	return nes, nil
}

func (nes *NES) Reset(soft bool) {
	nes.PPU.Reset()
	nes.APU.Reset(soft)
	nes.CPU.Reset(soft)
	nes.Mixer.Reset()
}

func (nes *NES) RunOneFrame(frame hw.Frame) {
	nes.PPU.SetFrameBuffer(frame.Video)
	nes.CPU.Run(29781)
	nes.APU.EndFrame()
}
