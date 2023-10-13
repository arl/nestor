package main

import (
	"fmt"
	"nestor/ines"
)

func loadMapper000(cart *ines.Rom, nes *NES) {
	// TODO: load PRGRAM if present in cartribge.

	switch len(cart.PRGROM) {
	case 0x4000:
		nes.CPU.bus.MapSlice(0x8000, 0xBFFF, cart.PRGROM)
		nes.CPU.bus.MapSlice(0xC000, 0xFFFF, cart.PRGROM) // mirror
	case 0x8000:
		nes.CPU.bus.MapSlice(0x8000, 0xFFFF, cart.PRGROM)
	default:
		panic(fmt.Sprintf("unexpected CHRROM size: 0x%x", len(cart.CHRROM)))
	}
}
