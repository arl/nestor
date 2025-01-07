package mappers

import (
	"fmt"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
)

var modMapper = log.NewModule("mapper")

var ErrUnsupportedMapper = fmt.Errorf("unsupported mapper")

func Load(rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) error {
	desc, ok := All[rom.Mapper()]
	if !ok {
		return ErrUnsupportedMapper
	}
	base := newbase(desc, rom, cpu, ppu)
	return base.load()
}

type MapperDesc struct {
	Name           string
	Load           func(*base) error
	PRGROMpagesize uint32
	CHRROMpagesize uint32
}

var All = map[uint16]MapperDesc{
	0:  NROM,
	2:  UxROM,
	3:  CNROM,
	66: GxROM,
}
