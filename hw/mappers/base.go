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

	PRGROM   [0x8000]byte // $8000-$FFFF
	PRGRAM   []byte
	PRGNVRAM []byte

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
		desc:     desc,
		rom:      rom,
		cpu:      cpu,
		ppu:      ppu,
		PRGRAM:   make([]byte, rom.PRGRAMSize()),
		PRGNVRAM: make([]byte, rom.PRGNVRAMSize()),
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
		// panic(fmt.Sprintf("PRGRAM not implemented, rom has $%XB", b.rom.PRGRAMSize()))
		b.cpu.Bus.MapMem(0x6000, &hwio.Mem{
			Name:  "PRGRAM",
			VSize: 0x2000,
			Data:  make([]byte, b.rom.PRGRAMSize()),
		})
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

func (b *base) BatteryPackedRAM() []byte {
	if b.rom.HasBattery() {
		return b.PRGNVRAM
	}
	return nil
}

func (b *base) SetBatteryPackedRAM(data []byte) error {
	if !b.rom.HasBattery() {
		modMapper.WarnZ("rom doesn't support battery packed RAM").End()
		return nil
	}

	if len(data) != len(b.PRGNVRAM) {
		return fmt.Errorf("invalid battery packed RAM size: %d", len(data))
	}
	copy(b.PRGNVRAM, data)
	return nil
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

func mirrorcopy(dst, src []byte) int {
	n, m := len(dst), len(src)
	if m == 0 || n == 0 {
		return 0
	}
	// Hot path: same size
	if m == n {
		return copy(dst, src)
	}
	copy(dst, src)

	// double-filled region each iteration
	for size := m; size < n; size <<= 1 {
		copy(dst[size:], dst[:size])
	}
	return n
}

// select what 32KB PRG ROM bank to use.
func (b *base) selectPRGPage32KB(bank int) {
	mirrorcopy(b.PRGROM[:], b.rom.PRGROM[32*KB*(bank):])
}

// select what 16KB PRG ROM bank to use into which PRG 16KB page.
func (b *base) selectPRGPage16KB(page uint32, bank int) {
	if len(b.rom.PRGROM) == 0 {
		return
	}
	if bank < 0 {
		// TODO: should probably not be checked here and should not panic.
		if len(b.rom.PRGROM)%(16*KB) != 0 {
			panic("PRGROM not multiple of 16KB")
		}
		bank += len(b.rom.PRGROM) / (16 * KB)
	}

	offbus := 16 * KB * page
	endbus := 16 * KB * (page + 1)
	offrom := 16 * KB * (bank)
	copy(b.PRGROM[offbus:endbus], b.rom.PRGROM[offrom:])

	modMapper.DebugZ("Select 16 kB PRG page").
		Hex16("bus.start", uint16(0x8000+offbus)).
		Hex16("bus.end", uint16(-1+0x8000+endbus)).
		Hex16("rom.start", uint16(16*KB*(bank))).
		Int("bank", bank).End()
}

// select what 8KB PRG ROM bank to use.
func (b *base) selectCHRROMPage8KB(bank int) {
	if len(b.rom.CHRROM) == 0 {
		return
	}
	if bank < 0 {
		bank += len(b.rom.CHRROM) / (8 * KB)
	}

	offbus, endbus := 0, 8*KB
	offrom := 8 * KB * bank
	copy(b.CHRROM[offbus:endbus], b.rom.CHRROM[offrom:])

	modMapper.DebugZ("Select 8 kB CHR page").
		Hex16("bus.start", uint16(offbus)).
		Hex16("bus.end", uint16(-1+endbus)).
		Hex16("rom.start", uint16(offrom)).
		Int("bank", bank).End()
}

// select what 4KB PRG ROM bank to use into which PRG 4KB page.
func (b *base) selectCHRROMPage4KB(page uint32, bank int) {
	if len(b.rom.CHRROM) == 0 {
		return
	}
	if bank < 0 {
		bank += len(b.rom.CHRROM) / (4 * KB)
	}

	offbus := 4 * KB * page
	endbus := 4 * KB * (page + 1)
	offrom := min(4*KB*bank, len(b.rom.CHRROM)-1)
	copy(b.CHRROM[offbus:endbus], b.rom.CHRROM[offrom:])
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
