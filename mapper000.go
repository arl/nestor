package main

import (
	"fmt"

	"nestor/ines"
)

func loadMapper000(rom *ines.Rom, nes *NES) error {
	// TODO: load and map PRG-RAM if present in cartridge.
	switch len(rom.PRGROM) {
	case 0x4000:
		nes.CPU.Bus.MapMemorySlice(0x8000, 0xBFFF, rom.PRGROM, true)
		nes.CPU.Bus.MapMemorySlice(0xC000, 0xFFFF, rom.PRGROM, true) // mirror
	case 0x8000:
		nes.CPU.Bus.MapMemorySlice(0x8000, 0xFFFF, rom.PRGROM, true)
	default:
		return fmt.Errorf("unexpected CHRROM size: 0x%x", len(rom.CHRROM))
	}

	// TODO: load and map CHR-RAM if present in cartridge.
	if len(rom.CHRROM) != 0x2000 {
		return fmt.Errorf("not implemented CHRROM size: 0x%x", len(rom.CHRROM))
	}
	nes.PPU.Bus.MapMemorySlice(0x0000, 0x1FFF, rom.CHRROM, true)

	return nil
}
