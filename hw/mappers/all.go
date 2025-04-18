package mappers

import (
	"fmt"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
)

var modMapper = log.NewModule("mapper")

func Load(rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) error {
	desc, ok := All[rom.Mapper()]
	if !ok {
		return fmt.Errorf("unsupported mapper %d", rom.Mapper())
	}
	base, err := newbase(desc, rom, cpu, ppu)
	if err != nil {
		return fmt.Errorf("mapper initialization failed: %w", err)
	}
	if err := desc.Load(base); err != nil {
		return fmt.Errorf("failed to load mapper %s: %w", desc.Name, err)
	}
	return nil
}

type ErrUnsuppportedPRGROMSize int

func (e ErrUnsuppportedPRGROMSize) Error() string {
	return fmt.Sprintf("unsupported PRGROM size: %d bytes", int(e))
}

type MapperDesc struct {
	Name            string
	Load            func(*base) error
	PRGROMbanksz    uint32
	CHRROMbanksz    uint32
	PRGRAMbanksz    uint32
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
