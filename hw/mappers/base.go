package mappers

import (
	"strconv"

	"nestor/hw"
	"nestor/ines"
)

type base struct {
	desc MapperDesc

	rom *ines.Rom
	cpu *hw.CPU
	ppu *hw.PPU
}

func newbase(desc MapperDesc, rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) *base {
	return &base{
		desc: desc,
		rom:  rom,
		cpu:  cpu,
		ppu:  ppu,
	}
}

func (b *base) load() error {
	return b.desc.Load(b)
}

func copyCHRROM(ppu *hw.PPU, rom *ines.Rom, bank uint32) {
	// Copy CHRROM bank to PPU memory.
	// CHRROM is 8KB in size
	start := bank * 0x2000
	end := start + 0x2000
	copy(ppu.PatternTables.Data, rom.CHRROM[start:end])
}

func (b *base) setNametableMirroring() {
	A := b.ppu.Nametables[:0x400]
	B := b.ppu.Nametables[0x400:0x800]

	// NameTables
	switch m := b.rom.Mirroring(); m {
	case ines.HorzMirroring:
		// A A B B
		b.ppu.Bus.MapMemorySlice(0x2000, 0x23FF, A, false)
		b.ppu.Bus.MapMemorySlice(0x2400, 0x27FF, A, false)
		b.ppu.Bus.MapMemorySlice(0x2800, 0x2BFF, B, false)
		b.ppu.Bus.MapMemorySlice(0x2C00, 0x2FFF, B, false)

		// nametables mirrors
		b.ppu.Bus.MapMemorySlice(0x3000, 0x33FF, A, false)
		b.ppu.Bus.MapMemorySlice(0x3400, 0x37FF, A, false)
		b.ppu.Bus.MapMemorySlice(0x3800, 0x3BFF, B, false)
		b.ppu.Bus.MapMemorySlice(0x3C00, 0x3EFF, B, false)

	case ines.VertMirroring:
		// A B A B
		b.ppu.Bus.MapMemorySlice(0x2000, 0x23FF, A, false)
		b.ppu.Bus.MapMemorySlice(0x2400, 0x27FF, B, false)
		b.ppu.Bus.MapMemorySlice(0x2800, 0x2BFF, A, false)
		b.ppu.Bus.MapMemorySlice(0x2C00, 0x2FFF, B, false)

		// nametables mirrors
		b.ppu.Bus.MapMemorySlice(0x3000, 0x33FF, A, false)
		b.ppu.Bus.MapMemorySlice(0x3400, 0x37FF, B, false)
		b.ppu.Bus.MapMemorySlice(0x3800, 0x3BFF, A, false)
		b.ppu.Bus.MapMemorySlice(0x3C00, 0x3EFF, B, false)
	default:
		panic("unsupported mirroring: " + strconv.Itoa(int(m)))
	}
}
