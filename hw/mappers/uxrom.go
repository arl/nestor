package mappers

var UxROM = MapperDesc{
	Name:         "UxROM",
	Load:         loadUxROM,
	PRGROMbanksz: 0x4000,
	CHRROMbanksz: 0x2000,
}

type uxrom struct {
	*base

	prgbank      uint32
	bankmask     uint8
	busConflicts bool
}

func (m *uxrom) WritePRGROM(addr uint16, val uint8) {
		val &= m.cpu.Bus.Peek8(addr)
	if m.busConflicts {
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
		m.selectPRGPage16KB(0, int(m.prgbank))
	}
}

func loadUxROM(b *base) error {
	uxrom := &uxrom{
		base:         b,
		busConflicts: b.rom.SubMapper() == 2,
		bankmask:     uint8(len(b.rom.PRGROM)>>14) - 1,
	}
	b.init(uxrom.WritePRGROM)

	b.setNTMirroring(b.rom.Mirroring())
	b.selectCHRPage8KB(0)
	b.selectPRGPage16KB(0, 0)
	b.selectPRGPage16KB(1, -1)
	return nil
}
