package mappers

import (
	"nestor/hw/hwio"
)

var GxROM = MapperDesc{
	Name:         "GxROM",
	Load:         loadGxROM,
	PRGROMbanksz: 0x8000,
	CHRROMbanksz: 0x2000,
}

type gxrom struct {
	*base
	PRGRAM hwio.Mem `hwio:"offset=0x6000,size=0x2000"`
	PRGROM hwio.Device

	chrbank uint32
	prgbank uint32
}

func (m *gxrom) ReadPRGROM(addr uint16) uint8 {
	addr &= uint16(m.desc.PRGROMbanksz - 1) // limit to max PRGROM size
	romaddr := (m.prgbank * m.desc.PRGROMbanksz) + uint32(addr)
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
	prevchr := m.chrbank
	m.chrbank = uint32(val & 0x3)
	if prevchr != m.chrbank {
		copyCHRROM(m.ppu, m.rom, m.chrbank)
		modMapper.DebugZ("CHRROM bank switch").String("mapper", m.desc.Name).Uint32("prev", prevchr).Uint32("new", m.chrbank).End()
	}

	prevprg := m.prgbank
	m.prgbank = uint32((val >> 4) & 0x3)
	if prevprg != m.prgbank {
		modMapper.DebugZ("PRGROM bank switch").String("mapper", m.desc.Name).Uint32("prev", prevprg).Uint32("new", m.prgbank).End()
	}
}

func loadGxROM(b *base) error {
	gxrom := &gxrom{base: b}
	hwio.MustInitRegs(gxrom)

	// CPU mapping.
	gxrom.PRGROM = hwio.Device{
		Name:    "PRGROM",
		Size:    0x8000,
		ReadCb:  gxrom.ReadPRGROM,
		PeekCb:  gxrom.ReadPRGROM,
		WriteCb: gxrom.WritePRGROM,
	}
	b.cpu.Bus.MapDevice(0x8000, &gxrom.PRGROM)
	b.cpu.Bus.MapBank(0x0000, gxrom, 0)

	// PPU mapping.
	b.setNametableMirroring(b.rom.Mirroring())
	copy(b.ppu.PatternTables.Data, b.rom.CHRROM)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
