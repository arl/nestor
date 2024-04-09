package hw

import (
	"nestor/ines"
)

type MapperDesc struct {
	Name string
	Load func(*ines.Rom, *CPU, *PPU) error
}
