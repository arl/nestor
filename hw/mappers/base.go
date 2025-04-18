package mappers

import (
	"fmt"

	"nestor/hw"
	"nestor/hw/hwio"
	"nestor/ines"
)

type base struct {
	rom *ines.Rom

	cpu *hw.CPU

	// TODO: should PRGRAM this always be there?
	PRGRAM hwio.Mem `hwio:"offset=0x6000,size=0x2000"`

	PRGROM [0x8000]byte // $8000-$FFFF

	ppu        *hw.PPU
	CHRROM     [0x2000]byte
	nametables [0x800]byte

	desc MapperDesc

	// set by base.init
	registers hwio.Bitset
	writeReg  func(addr uint16, value uint8) // optional
}

func newbase(desc MapperDesc, rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) (*base, error) {
	if !ispow2(len(rom.PRGROM)) {
		return nil, fmt.Errorf("only support PRGROM with power of 2 size, got %d", len(rom.PRGROM))
	}

	b := &base{
		desc: desc,
		rom:  rom,
		cpu:  cpu,
		ppu:  ppu,
	}

	start := uint(0x8000)
	end := uint(0x10000)
	if desc.RegisterStart != 0 {
		start = uint(desc.RegisterStart)
	}
	if desc.RegisterEnd != 0 {
		end = uint(desc.RegisterEnd)
	}
	b.registers.SetRange(uint(start), uint(end))
	return b, nil
}

func (b *base) init(writeReg func(uint16, uint8)) {
	// CPU mapping.
	hwio.MustInitRegs(b)
	b.cpu.Bus.MapBank(0x0000, b, 0)

	if b.rom.PRGRAMSize() > 0 {
		// panic("PRGRAM not implemented")
	}

	b.writeReg = writeReg
	b.cpu.Bus.MapMem(0x8000, &hwio.Mem{
		Name:    "PRGROM",
		Data:    b.PRGROM[:],
		VSize:   0x8000,
		Flags:   hwio.MemFlagReadOnlyNoLog,
		WriteCb: b.write,
	})

	// Handle CHR RAM if CHRROM is empty.
	chrFlag := hwio.MemFlagReadOnly
	if len(b.rom.CHRROM) == 0 {
		chrFlag = hwio.MemFlagReadWrite // 8 KB CHR RAM
	}

	b.ppu.Bus.MapMem(0x0000, &hwio.Mem{
		Name:  "CHRROM",
		Data:  b.CHRROM[:],
		VSize: 0x2000,
		Flags: chrFlag,
	})
}

func (b *base) write(addr uint16, value uint8) {
	// is this a register write?
	if b.registers.Test(uint(addr)) {
		if b.writeReg != nil {
			b.writeReg(addr, value)
		}
	}
}

const KB = 1 << 10

// select what 32KB PRG ROM bank to use.
func (b *base) selectPRGPage32KB(bank int) {
	// TODO: what if instead of copying we were using
	// table.MapMemorySlice. in this case we would avoid a copy, as well as
	// define if the memory is read-only or read-write.
	copy(b.PRGROM[:], b.rom.PRGROM[32*KB*(bank):])
}

// select what 16KB PRG ROM bank to use into which PRG 16KB page.
func (b *base) selectPRGPage16KB(page uint32, bank int) {
	if bank < 0 {
		// TODO: should probably not be checked here and should not panic.
		if len(b.rom.PRGROM)%(16*KB) != 0 {
			panic("PRGROM not multiple of 16KB")
		}
		bank += len(b.rom.PRGROM) / (16 * KB)
	}

	start := 16 * KB * page
	end := 16 * KB * (page + 1)
	copy(b.PRGROM[start:end], b.rom.PRGROM[16*KB*(bank):])

	modMapper.DebugZ("Select 16 kB PRG page").
		Hex16("bus.start", uint16(0x8000+start)).
		Hex16("bus.end", uint16(-1+0x8000+end)).
		Hex16("rom.start", uint16(16*KB*(bank))).
		Int("bank", bank).End()
}

// TODO: remove and use selectCHRROM... instead
func (b *base) copyCHRROM(dest []byte, bank uint32) {
	// Copy CHRROM bank to PPU memory.
	// CHRROM is 8KB in size (when present).
	start := min(uint32(len(b.rom.CHRROM)-1), bank*b.desc.CHRROMbanksz)
	end := min(uint32(len(b.rom.CHRROM)), start+b.desc.CHRROMbanksz)
	copy(dest, b.rom.CHRROM[start:end])
}

// select what 8KB PRG ROM bank to use.
func (b *base) selectCHRROMPage8KB(bank int) {
	if bank < 0 {
		bank += len(b.rom.CHRROM) / (8 * KB)
	}

	// b:bus r:rom
	bstart, bend := 0, 8*KB
	rstart := 8 * KB * bank
	copy(b.CHRROM[bstart:bend], b.rom.CHRROM[rstart:])

	modMapper.DebugZ("Select 8 kB CHR page").
		Hex16("bus.start", uint16(bstart)).
		Hex16("bus.end", uint16(-1+bend)).
		Hex16("rom.start", uint16(rstart)).
		Int("bank", bank).End()
}

// select what 4KB PRG ROM bank to use into which PRG 4KB page.
func (b *base) selectCHRROMPage4KB(page uint32, bank int) {
	if bank < 0 {
		bank += len(b.rom.CHRROM) / (4 * KB)
	}

	if len(b.rom.CHRROM) != 0 {
		romoff := min(4*KB*bank, len(b.rom.CHRROM)-1)
		copy(b.CHRROM[4*KB*page:], b.rom.CHRROM[romoff:])
	}
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

func ispow2(n int) bool  { return n&(n-1) == 0 }
func u8tob(v uint8) bool { return v != 0 }
