package mappers

import (
	"fmt"

	"nestor/hw"
	"nestor/ines"
)

type base struct {
	rom *ines.Rom
	cpu *hw.CPU
	ppu *hw.PPU

	nametables [0x800]byte

	hasBusConflicts bool

	desc MapperDesc
}

func ispow2(n int) bool {
	return n&(n-1) == 0
}

func newbase(desc MapperDesc, rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) (*base, error) {
	if !ispow2(len(rom.PRGROM)) {
		return nil, fmt.Errorf("only support PRGROM with power of 2 size, got %d", len(rom.PRGROM))
	}

	return &base{
		desc:            desc,
		rom:             rom,
		cpu:             cpu,
		ppu:             ppu,
		hasBusConflicts: desc.HasBusConflicts != nil && desc.HasBusConflicts(&base{rom: rom}),
	}, nil
}

func (b *base) load() error {
	return b.desc.Load(b)
}

func copyCHRROM(dest []byte, rom *ines.Rom, bank uint32) {
	// Copy CHRROM bank to PPU memory.
	// CHRROM is 8KB in size (when present).
	start := min(uint32(len(rom.CHRROM)-1), bank*0x2000)
	end := min(uint32(len(rom.CHRROM)), start+0x2000)
	copy(dest, rom.CHRROM[start:end])
}

func (b *base) setNTMirroring(m ines.NTMirroring) {
	A := b.nametables[:0x400]
	B := b.nametables[0x400:0x800]

	switch m {
	case ines.HorzMirroring:
		b.remapNametables(A, A, B, B)
	case ines.VertMirroring:
		b.remapNametables(A, B, A, B)
	case ines.OnlyAScreen:
		b.remapNametables(A, A, A, A)
	case ines.OnlyBScreen:
		b.remapNametables(B, B, B, B)
	default:
		panic(fmt.Sprintf("unsupported mirroring %d", m))
	}
}

func (b *base) remapNametables(nt1, nt2, nt3, nt4 []byte) {
	// Unmap all nametables
	b.ppu.Bus.Unmap(0x2000, 0x3EFF)

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
