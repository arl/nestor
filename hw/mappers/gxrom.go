package mappers

import (
	"nestor/hw/hwio"
)

var GxROM = MapperDesc{
	Name: "GxROM",
	Load: loadGxROM,
}

type gxrom struct {
	*base

	PRGRAM hwio.Mem `hwio:"offset=0x6000,size=0x2000"`

	// 32KB switchable PRGROM bank
	// 8KB switchable CHRROM bank
	PRGROM hwio.Device `hwio:"offset=0x8000,size=0x8000,rcb,wcb"`
	curchr uint32
	curprg uint32
}

func (m *gxrom) ReadPRGROM(addr uint16) uint8 {
	banknum := m.curprg
	addr &= 0x7FFF // max PRGROM size is 32KB
	romaddr := (banknum * 0x8000) + uint32(addr)
	return m.rom.PRGROM[romaddr]
}

func (m *gxrom) WritePRGROM(addr uint16, val uint8) {
	// Switch bank.

	// 7  bit  0
	// ---- ----
	// xxPP xxCC
	//   ||   ||
	//   ||   ++- Select 8 KB CHR ROM bank for PPU $0000-$1FFF
	//   ++------ Select 32 KB PRG ROM bank for CPU $8000-$FFFF

	prevchr := m.curchr
	m.curchr = uint32(val & 0x3)
	if prevchr != m.curchr {
		copyCHRROM(m.ppu, m.rom, m.curchr)
		modMapper.DebugZ("CHRROM bank switch").
			Uint32("prev", prevchr).
			Uint32("new", m.curchr).
			End()
	}

	prevprg := m.curprg
	m.curprg = uint32((val >> 4) & 0x3)
	if prevprg != m.curprg {
		modMapper.DebugZ("PRGROM bank switch").
			Uint32("prev", prevprg).
			Uint32("new", m.curprg).
			End()
	}
}

func loadGxROM(b *base) error {
	gxrom := &gxrom{
		base: b,
	}

	hwio.MustInitRegs(gxrom)

	// CPU mapping.
	b.cpu.Bus.MapBank(0x0000, gxrom, 0)

	// PPU mapping.
	b.setNametableMirroring()
	copy(b.ppu.PatternTables.Data, b.rom.CHRROM)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
