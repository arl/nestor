package mappers

import (
	"fmt"

	"nestor/emu"
	"nestor/ines"
)

var NROM = emu.MapperDesc{
	Name: "NROM",
	Load: loadMapper000,
}

func loadMapper000(rom *ines.Rom, hw *emu.NESHardware) error {
	// CPU memory space mapping.
	//

	// PRG-RAM (Family basic only but we just always provide it).
	PRGRAM := make([]uint8, 0x2000)
	hw.CPU.Bus.MapMemorySlice(0x6000, 0x7FFF, PRGRAM, false)

	switch len(rom.PRGROM) {
	case 0x4000:
		hw.CPU.Bus.MapMemorySlice(0x8000, 0xBFFF, rom.PRGROM, true)
		hw.CPU.Bus.MapMemorySlice(0xC000, 0xFFFF, rom.PRGROM, true) // mirror
	case 0x8000:
		hw.CPU.Bus.MapMemorySlice(0x8000, 0xFFFF, rom.PRGROM, true)
	default:
		return fmt.Errorf("unexpected PRGROM size: 0x%x", len(rom.CHRROM))
	}

	// PPU memory space mapping.
	//

	// Copy CHRROM to Pattern Tables.
	switch len(rom.CHRROM) {
	case 0x0000:
		// Some roms have no CHR-ROM, allocate mem anyway.
		// XXX: not sure about that
		CHRROM := make([]uint8, 0x2000)
		hw.PPU.PatternTables.Data = CHRROM
	case 0x2000:
		hw.PPU.PatternTables.Data = rom.CHRROM
	default:
		return fmt.Errorf("unexpected CHRROM size: 0x%x", len(rom.CHRROM))

	}

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
	return nil
}
