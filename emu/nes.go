package emu

import (
	"bytes"

	"github.com/tinylib/msgp/msgp"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/hw/hwdefs"
	"nestor/hw/mappers"
	"nestor/hw/snapshot"
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

	buf, err := nes.SaveSnapshot()
	if err != nil {
		log.ModEmu.FatalZ("failed to save snapshot").Error("err", err).End()
		return
	}

	if err := nes.LoadSnapshot(buf); err != nil {
		log.ModEmu.FatalZ("failed to load snapshot").Error("err", err).End()
		return
	}
}

const SaveStateVersion = 1

func (nes *NES) SaveSnapshot() ([]byte, error) {
	buf := bytes.Buffer{}
	mw := msgp.NewWriter(&buf)

	state := snapshot.NES{
		Version: SaveStateVersion,
		CPU:     nes.CPU.State(),
		DMA:     nes.CPU.DMA.State(),
		PPU:     nes.PPU.State(),
		APU:     nes.APU.State(),
	}
	copy(state.RAM[:], nes.CPU.RAM.Data)

	if err := state.EncodeMsg(mw); err != nil {
		return nil, err
	}

	mw.Flush()
	return buf.Bytes(), nil
}

func (nes *NES) LoadSnapshot(buf []byte) error {
	r := msgp.NewReader(bytes.NewReader(buf))
	var state snapshot.NES
	if err := state.DecodeMsg(r); err != nil {
		return err
	}

	copy(nes.CPU.RAM.Data, state.RAM[:])

	nes.CPU.SetState(state.CPU)
	nes.CPU.DMA.SetState(state.DMA)
	nes.PPU.SetState(state.PPU)
	nes.APU.SetState(state.APU)
	return nil
}
