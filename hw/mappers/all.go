// Package mappers provides the interface and implementations for NES mappers.
package mappers

import (
	"fmt"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/hw/snapshot"
	"nestor/ines"
)

var modMapper = log.NewModule("mapper")

func Load(rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) (Mapper, error) {
	desc, ok := All[rom.Number()]
	if !ok {
		return nil, fmt.Errorf("unsupported mapper %d", rom.Number())
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
	SetState(*snapshot.MapperState)
	State() *snapshot.MapperState
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
