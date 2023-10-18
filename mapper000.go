package main

import (
	"fmt"
	"nestor/ines"
)

func loadMapper000(cart *ines.Rom, cpubus *cpuBus) {
	// TODO: load PRGRAM if present in cartribge.

	switch len(cart.PRGROM) {
	case 0x4000:
		cpubus.MapSlice(0x8000, 0xBFFF, cart.PRGROM)
		cpubus.MapSlice(0xC000, 0xFFFF, cart.PRGROM) // mirror
	case 0x8000:
		cpubus.MapSlice(0x8000, 0xFFFF, cart.PRGROM)
	default:
		panic(fmt.Sprintf("unexpected CHRROM size: 0x%x", len(cart.CHRROM)))
	}
}
