// Package mappers provides the interface and implementations for NES mappers.
package mappers

import (
	"fmt"

	"github.com/tinylib/msgp/msgp"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/hw/snapshot"
	"nestor/ines"
)

var modMapper = log.NewModule("mapper")

func Load(rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) (Mapper, error) {
	desc, ok := All[rom.Mapper()]
	if !ok {
		return nil, fmt.Errorf("unsupported mapper %d", rom.Mapper())
	}
	base, err := newbase(desc, rom, cpu, ppu)
	if err != nil {
		return nil, fmt.Errorf("mapper initialization failed: %w", err)
	}
	mapper, err := desc.Load(base)
	if err != nil {
		return nil, fmt.Errorf("failed to load mapper %s: %w", desc.Name, err)
	}
	return mapper, nil
}

type ErrUnsuppportedPRGROMSize int

func (e ErrUnsuppportedPRGROMSize) Error() string {
	return fmt.Sprintf("unsupported PRGROM size: %d bytes", int(e))
}

type Mapper interface {
	BatteryPackedRAM() []byte
	SetBatteryPackedRAM(data []byte) error
}

func EncodeState(m Mapper) snapshot.MapperState {
	var (
		mshler msgp.Marshaler
		number uint16
	)

	switch m := m.(type) {
	case *nrom:
		mshler = m.state()
		number = m.rom.Mapper()
	case *mmc1:
		mshler = m.state()
		number = m.rom.Mapper()
	case *uxrom:
		mshler = m.state()
		number = m.rom.Mapper()
	case *cnrom:
		mshler = m.state()
		number = m.rom.Mapper()
	case *axrom:
		mshler = m.state()
		number = m.rom.Mapper()
	case *gxrom:
		mshler = m.state()
		number = m.rom.Mapper()
	default:
		panic("unknown mapper type")
	}

	data, err := mshler.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}

	return snapshot.MapperState{
		Num:  number,
		Data: data,
	}
}

func DecodeState(m Mapper, s snapshot.MapperState) {
	switch s.Num {
	case 0:
		var state snapshot.NROMState
		if _, err := state.UnmarshalMsg(s.Data); err != nil {
			panic(err)
		}
		m.(*nrom).setState(&state)
	case 1:
		var state snapshot.MMC1State
		if _, err := state.UnmarshalMsg(s.Data); err != nil {
			panic(err)
		}
		m.(*mmc1).setState(&state)
	case 2:
		var state snapshot.UxROMState
		if _, err := state.UnmarshalMsg(s.Data); err != nil {
			panic(err)
		}
		m.(*uxrom).setState(&state)
	case 3:
		var state snapshot.CNROMState
		if _, err := state.UnmarshalMsg(s.Data); err != nil {
			panic(err)
		}
		m.(*cnrom).setState(&state)
	case 7:
		var state snapshot.AxROMState
		if _, err := state.UnmarshalMsg(s.Data); err != nil {
			panic(err)
		}
		m.(*axrom).setState(&state)
	case 66:
		var state snapshot.GxROMState
		if _, err := state.UnmarshalMsg(s.Data); err != nil {
			panic(err)
		}
		m.(*gxrom).setState(&state)
	default:
		panic("unknown mapper type")
	}
}

func SetState(m Mapper, state msgp.Decodable) {
	switch m := m.(type) {
	case *nrom:
		m.setState(state.(*snapshot.NROMState))
	case *mmc1:
		m.setState(state.(*snapshot.MMC1State))
	case *uxrom:
		m.setState(state.(*snapshot.UxROMState))
	case *cnrom:
		m.setState(state.(*snapshot.CNROMState))
	case *axrom:
		m.setState(state.(*snapshot.AxROMState))
	case *gxrom:
		m.setState(state.(*snapshot.GxROMState))
	default:
		panic("unknown mapper type")
	}
}

type MapperDesc struct {
	Name            string
	Load            func(*base) (Mapper, error)
	PRGBankSize     uint32
	CHRBankSize     uint32
	HasBusConflicts func(*base) bool

	RegisterStart uint16 // defaults to 0x8000 if not set
	RegisterEnd   uint16 // defaults to 0xFFFF if not set
}

var All = map[uint16]MapperDesc{
	0:  NROM,
	1:  MMC1,
	2:  UxROM,
	3:  CNROM,
	7:  AxROM,
	66: GxROM,
}
