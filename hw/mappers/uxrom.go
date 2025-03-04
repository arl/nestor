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
	old := val
	if m.busConflicts {
		old = m.cpu.Bus.Peek8(addr)
		val &= old
	}
	modMapper.DebugZ("WritePRGROM").
		Hex16("addr", addr).
		Hex8("old", old).
		Hex8("val", val).
		Hex8("bank", val&m.bankmask).
		Bool("conflicts", m.busConflicts).
		End()

	// 7  bit  0
	// ---- ----
	// xxxx pPPP
	//      ||||
	//      ++++- Select 16 KB PRG ROM bank for CPU $8000-$BFFF
	//            (UNROM uses bits 2-0; UOROM uses bits 3-0)
	m.prgbank = uint32(val & m.bankmask)
	m.selectPRGPage16KB(0, int(m.prgbank))
}

func loadUxROM(b *base) error {
	uxrom := &uxrom{
		base:         b,
		busConflicts: b.rom.SubMapper() == 2,
		bankmask:     uint8(len(b.rom.PRGROM)>>14) - 1,
	}
	b.init(uxrom.WritePRGROM)

	b.setNTMirroring(b.rom.Mirroring())
	b.selectCHRROMPage8KB(0)
	b.selectPRGPage16KB(0, 0)
	b.selectPRGPage16KB(1, -1)
	return nil
}
