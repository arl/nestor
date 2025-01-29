package mappers

import "nestor/hw/hwio"

var UxROM = MapperDesc{
	Name:            "UxROM",
	Load:            loadUxROM,
	PRGROMbanksz:    0x4000,
	CHRROMbanksz:    0x2000,
	HasBusConflicts: func(b *base) bool { return b.rom.SubMapper() == 2 },
}

type uxrom struct {
	*base

	/* CPU */
	PRGRAM          hwio.Mem    `hwio:"offset=0x6000,size=0x2000"`
	PRGROM          hwio.Device `hwio:"offset=0x8000,size=0x8000,rcb,wcb"`
	prgbank         uint32
	bankmask        uint8
	hasBusConflicts bool

	/* PPU */
	PatternTables hwio.Mem `hwio:"bank=1,offset=0x0000,size=0x2000"`
}

func (m *uxrom) ReadPRGROM(addr uint16) uint8 {
	banknum := m.prgbank
	if addr >= 0xC000 {
		banknum = uint32(len(m.rom.PRGROM))/m.desc.PRGROMbanksz - 1
	}
	romaddr := (banknum * m.desc.PRGROMbanksz) + uint32(addr)&(m.desc.PRGROMbanksz-1)
	return m.rom.PRGROM[romaddr]
}

func (m *uxrom) WritePRGROM(addr uint16, val uint8) {
	if m.hasBusConflicts {
		val &= m.ReadPRGROM(addr)
	}

	// 7  bit  0
	// ---- ----
	// xxxx pPPP
	//      ||||
	//      ++++- Select 16 KB PRG ROM bank for CPU $8000-$BFFF
	//            (UNROM uses bits 2-0; UOROM uses bits 3-0)
	prev := m.prgbank
	m.prgbank = uint32(val & m.bankmask)
	if prev != m.prgbank {
		modMapper.DebugZ("PRGROM bank switch").String("mapper", m.desc.Name).Uint32("prev", prev).Uint32("new", m.prgbank).End()
	}
}

func loadUxROM(b *base) error {
	uxrom := &uxrom{
		base:            b,
		hasBusConflicts: b.rom.SubMapper() == 2,
		bankmask:        uint8(len(b.rom.PRGROM)>>14) - 1,
	}
	hwio.MustInitRegs(uxrom)

	// CPU mapping.
	b.cpu.Bus.MapBank(0x0000, uxrom, 0)

	// PPU mapping.
	b.setNTMirroring(b.rom.Mirroring())
	b.ppu.Bus.MapBank(0x0000, uxrom, 1)
	b.copyCHRROM(uxrom.PatternTables.Data, 0)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
