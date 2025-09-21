package emu

import (
	"bytes"
	"fmt"

	"github.com/tinylib/msgp/msgp"

	"nestor/hw"
	"nestor/hw/apu"
	"nestor/hw/hwdefs"
	"nestor/hw/mappers"
	"nestor/hw/snapshot"
	"nestor/ines"
)

type NES struct {
	CPU    *hw.CPU
	PPU    *hw.PPU
	APU    *apu.APU
	Rom    *ines.Rom
	Mixer  *apu.Mixer
	Mapper mappers.Mapper

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

	mapper, err := mappers.Load(rom, nes.CPU, nes.PPU)
	if err != nil {
		return nil, fmt.Errorf("error loading rom: %w", err)
	}

	nes.Mapper = mapper
	nes.Reset(hwdefs.HardReset)
	return &nes, nil
}

func (nes *NES) Reset(soft bool) {
	nes.PPU.Reset()
	nes.APU.Reset(soft)
	nes.CPU.Reset(soft)
	nes.Mixer.Reset()
}

func (nes *NES) RunOneFrame(frame *Frame) {
	nes.PPU.BeginFrame(frame.Video)
	nes.CPU.Run(29781)
	nes.APU.EndFrame(&frame.Audio)
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
		PPU:     nes.PPU.State(),
		APU:     nes.APU.State(),
		Mixer:   nes.Mixer.State(),
		Mapper:  nes.Mapper.State(),
	}
	// TODO: move RAM state in CPU state.
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

	nes.Mapper.SetState(state.Mapper)
	nes.CPU.SetState(state.CPU)
	nes.PPU.SetState(state.PPU)
	nes.APU.SetState(state.APU)
	nes.Mixer.SetState(state.Mixer)
	copy(nes.CPU.RAM.Data, state.RAM[:])

	return nil
}
