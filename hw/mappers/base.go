package mappers

import (
	"fmt"
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

func ispow2(n int) bool {
	return n&(n-1) == 0
}

func newbase(desc MapperDesc, rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) (*base, error) {
	if !ispow2(len(rom.PRGROM)) {
		return nil, fmt.Errorf("only support PRGROM with size that is power of 2, got %d", len(rom.PRGROM))
	}
	return &base{
		desc: desc,
		rom:  rom,
		cpu:  cpu,
		ppu:  ppu,
	}, nil
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

func (b *base) mapNametable(idx, ntidx uint16) {
	ntoff := uint32(ntidx) * 0x400
	fmt.Printf("mapping from 0x%04x to 0x%04x\n", 0x2000+(idx*0x400), 0x2000+(idx+1)*0x400-1)
	b.ppu.Bus.MapMemorySlice(0x2000+(idx*0x400), 0x2000+(idx+1)*0x400-1, b.ppu.Nametables[ntoff:ntoff+0x400], false)

	fmt.Printf("mapping from 0x%04x to 0x%04x\n", 0x3000+(idx*0x400), min(0x3EFF, 0x3000+(idx+1)*0x400-1))
	b.ppu.Bus.MapMemorySlice(0x3000+(idx*0x400), min(0x3EFF, 0x3000+(idx+1)*0x400-1), b.ppu.Nametables[ntoff:ntoff+0x400], false)
}

func (b *base) setNametables(ntidx0, ntidx1, ntidx2, ntidx3 uint16) {
	b.mapNametable(0, ntidx0)
	b.mapNametable(1, ntidx1)
	b.mapNametable(2, ntidx2)
	b.mapNametable(3, ntidx3)
}

func (b *base) setNametableMirroring(m ines.NTMirroring) {
	b.ppu.Bus.Unmap(0x2000, 0x3EFF)
	A := b.ppu.Nametables[:0x400]
	B := b.ppu.Nametables[0x400:0x800]

	switch m {
	case ines.HorzMirroring:
		// b.setNametables(0, 0, 1, 1)
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
		// b.setNametables(0, 1, 0, 1)
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
	case ines.OnlyAScreen:
		// b.setNametables(0, 0, 0, 0)
		// A A A A
		b.ppu.Bus.MapMemorySlice(0x2000, 0x23FF, A, false)
		b.ppu.Bus.MapMemorySlice(0x2400, 0x27FF, A, false)
		b.ppu.Bus.MapMemorySlice(0x2800, 0x2BFF, A, false)
		b.ppu.Bus.MapMemorySlice(0x2C00, 0x2FFF, A, false)

		// nametables mirrors
		b.ppu.Bus.MapMemorySlice(0x3000, 0x33FF, A, false)
		b.ppu.Bus.MapMemorySlice(0x3400, 0x37FF, A, false)
		b.ppu.Bus.MapMemorySlice(0x3800, 0x3BFF, A, false)
		b.ppu.Bus.MapMemorySlice(0x3C00, 0x3EFF, A, false)
	case ines.OnlyBScreen:
		// b.setNametables(1, 1, 1, 1)
		// A A A A
		b.ppu.Bus.MapMemorySlice(0x2000, 0x23FF, B, false)
		b.ppu.Bus.MapMemorySlice(0x2400, 0x27FF, B, false)
		b.ppu.Bus.MapMemorySlice(0x2800, 0x2BFF, B, false)
		b.ppu.Bus.MapMemorySlice(0x2C00, 0x2FFF, B, false)

		// nametables mirrors
		b.ppu.Bus.MapMemorySlice(0x3000, 0x33FF, B, false)
		b.ppu.Bus.MapMemorySlice(0x3400, 0x37FF, B, false)
		b.ppu.Bus.MapMemorySlice(0x3800, 0x3BFF, B, false)
		b.ppu.Bus.MapMemorySlice(0x3C00, 0x3EFF, B, false)
	default:
		panic("unsupported mirroring: " + strconv.Itoa(int(m)))
	}
}
