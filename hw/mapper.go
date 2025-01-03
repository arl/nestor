package hw

import (
	"strconv"

	"nestor/ines"
)

type MapperDesc struct {
	Name string
	Load func(*ines.Rom, *CPU, *PPU) error
}

func SetNTMirroring(ppu *PPU, m ines.Mirroring) {
	A := ppu.Nametables[:0x400]
	B := ppu.Nametables[0x400:0x800]

	// NameTables
	switch m {
	case ines.HorzMirroring:
		// A A B B
		ppu.Bus.MapMemorySlice(0x2000, 0x23FF, A, false)
		ppu.Bus.MapMemorySlice(0x2400, 0x27FF, A, false)
		ppu.Bus.MapMemorySlice(0x2800, 0x2BFF, B, false)
		ppu.Bus.MapMemorySlice(0x2C00, 0x2FFF, B, false)

		// nametables mirrors
		ppu.Bus.MapMemorySlice(0x3000, 0x33FF, A, false)
		ppu.Bus.MapMemorySlice(0x3400, 0x37FF, A, false)
		ppu.Bus.MapMemorySlice(0x3800, 0x3BFF, B, false)
		ppu.Bus.MapMemorySlice(0x3C00, 0x3EFF, B, false)

	case ines.VertMirroring:
		// A B A B
		ppu.Bus.MapMemorySlice(0x2000, 0x23FF, A, false)
		ppu.Bus.MapMemorySlice(0x2400, 0x27FF, B, false)
		ppu.Bus.MapMemorySlice(0x2800, 0x2BFF, A, false)
		ppu.Bus.MapMemorySlice(0x2C00, 0x2FFF, B, false)

		// nametables mirrors
		ppu.Bus.MapMemorySlice(0x3000, 0x33FF, A, false)
		ppu.Bus.MapMemorySlice(0x3400, 0x37FF, B, false)
		ppu.Bus.MapMemorySlice(0x3800, 0x3BFF, A, false)
		ppu.Bus.MapMemorySlice(0x3C00, 0x3EFF, B, false)
	default:
		panic("unsupported mirroring: " + strconv.Itoa(int(m)))
	}
}
