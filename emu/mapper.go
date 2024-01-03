package emu

import (
	"nestor/ines"
)

type MapperDesc struct {
	Name string
	Load func(rom *ines.Rom, hw *NESHardware) error
}
