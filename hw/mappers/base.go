package mappers

import (
	"fmt"

	"nestor/hw"
	"nestor/ines"
)

type base struct {
	desc MapperDesc

	rom *ines.Rom
	cpu *hw.CPU
	ppu *hw.PPU
}

func ispow2(n int) bool {
	return n&(n-1) == 0
}

func newbase(desc MapperDesc, rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) (*base, error) {
	if !ispow2(len(rom.PRGROM)) {
		return nil, fmt.Errorf("only support PRGROM with power of 2 size, got %d", len(rom.PRGROM))
	}

	return &base{desc: desc, rom: rom, cpu: cpu, ppu: ppu}, nil
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

func (b *base) setNametableMirroring(m ines.NTMirroring) {
	// Unmap all nametables
	b.ppu.Bus.Unmap(0x2000, 0x3EFF)

	A := b.ppu.Nametables[:0x400]
	B := b.ppu.Nametables[0x400:0x800]

	var nt1, nt2, nt3, nt4 []byte

	switch m {
	case ines.HorzMirroring:
		nt1, nt2 = A, A
		nt3, nt4 = B, B
	case ines.VertMirroring:
		nt1, nt2 = A, B
		nt3, nt4 = A, B
	case ines.OnlyAScreen:
		nt1, nt2 = A, A
		nt3, nt4 = A, A
	case ines.OnlyBScreen:
		nt1, nt2 = B, B
		nt3, nt4 = B, B
	default:
		panic(fmt.Sprintf("unsupported mirroring %d", m))
	}

	// Map nametables
	b.ppu.Bus.MapMemorySlice(0x2000, 0x23FF, nt1, false)
	b.ppu.Bus.MapMemorySlice(0x2400, 0x27FF, nt2, false)
	b.ppu.Bus.MapMemorySlice(0x2800, 0x2BFF, nt3, false)
	b.ppu.Bus.MapMemorySlice(0x2C00, 0x2FFF, nt4, false)

	// Mirrors
	b.ppu.Bus.MapMemorySlice(0x3000, 0x33FF, nt1, false)
	b.ppu.Bus.MapMemorySlice(0x3400, 0x37FF, nt2, false)
	b.ppu.Bus.MapMemorySlice(0x3800, 0x3BFF, nt3, false)
	b.ppu.Bus.MapMemorySlice(0x3C00, 0x3EFF, nt4, false)
}
