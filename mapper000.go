package main

import (
	"fmt"

	"nestor/emu/hwio"
	"nestor/ines"
)

func loadMapper000(cart *ines.Rom, cpubus *hwio.MemMap) error {
	// TODO: load PRGRAM if present in cartridge.

	switch len(cart.PRGROM) {
	case 0x4000:
		cpubus.MapSlice(0x8000, 0xBFFF, cart.PRGROM)
		cpubus.MapSlice(0xC000, 0xFFFF, cart.PRGROM) // mirror
	case 0x8000:
		cpubus.MapSlice(0x8000, 0xFFFF, cart.PRGROM)
	default:
		return fmt.Errorf("unexpected CHRROM size: 0x%x", len(cart.CHRROM))
	}
	return nil
}
