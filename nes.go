package main

import "nestor/ines"

type NES struct {
	cart *ines.Rom
}

func startNES(rom *ines.Rom) (*NES, error) {
	nes := &NES{
		cart: rom,
	}
	return nes, nil
}
