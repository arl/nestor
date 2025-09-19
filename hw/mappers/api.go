package mappers

import (
	"fmt"

	"nestor/hw"
	"nestor/hw/snapshot"
	"nestor/ines"
)

type Mapper interface {
	BatteryPackedRAM() []byte
	SetBatteryPackedRAM(data []byte) error
	SetState(*snapshot.MapperState)
	State() *snapshot.MapperState
}

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
