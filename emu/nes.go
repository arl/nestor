package emu

import (
	"bytes"

	"github.com/tinylib/msgp/msgp"

	"nestor/hw"
	"nestor/hw/apu"
	"nestor/hw/hwdefs"
	"nestor/hw/mappers"
	"nestor/hw/snapshot"
	"nestor/ines"
)

type NES struct {
	CPU   *hw.CPU
	PPU   *hw.PPU
	APU   *apu.APU
	Rom   *ines.Rom
	Mixer *apu.Mixer

	isRunAheadFrame bool
}

func powerUp(rom *ines.Rom) (*NES, error) {
	var nes NES
	nes.Mixer = apu.NewMixer(&nes)
	nes.PPU = hw.NewPPU()
	nes.CPU = hw.NewCPU(nes.PPU)
	nes.APU = apu.New(nes.CPU, nes.Mixer)

	nes.CPU.APU = nes.APU
	nes.CPU.InitBus()

	if err := mappers.Load(rom, nes.CPU, nes.PPU); err != nil {
		return nil, err
	}

	nes.Reset(hwdefs.HardReset)
	return &nes, nil
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

func (nes *NES) IsRunAheadFrame() bool {
	return nes.isRunAheadFrame
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
		Mixer:   nes.Mixer.State(),
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
	nes.Mixer.SetState(state.Mixer)
	return nil
}
