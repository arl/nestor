package mappers

var CNROM = MapperDesc{
	Name:         "CNROM",
	Load:         loadCNROM,
	PRGROMbanksz: 0x8000,
	CHRROMbanksz: 0x2000,
}

type cnrom struct {
	*base

	chrbank uint32

	busConflicts bool
}

func (m *cnrom) WritePRGROM(addr uint16, val uint8) {
	old := val
	if m.busConflicts {
		old = m.cpu.Bus.Peek8(addr)
		val &= old
	}
	modMapper.DebugZ("WritePRGROM").
		Hex16("addr", addr).
		Hex8("old", old).
		Hex8("val", val).
		Bool("conflicts", m.busConflicts).
		End()

	// 7  bit  0
	// ---- ----
	// cccc ccCC
	// |||| ||||
	// ++++-++++- Select 8 KB CHR ROM bank for PPU $0000-$1FFF
	// CNROM only uses lowest 2 bits
	prev := m.chrbank
	m.chrbank = uint32(val) & 0b11
	if prev != m.chrbank {
		m.selectCHRROMPage8KB(int(m.chrbank))
	}
}

func loadCNROM(b *base) error {
	cnrom := &cnrom{
		base:         b,
		busConflicts: b.rom.SubMapper() == 2,
	}
	b.init(cnrom.WritePRGROM)

	// PPU mapping.
	b.setNTMirroring(b.rom.Mirroring())
	b.selectCHRROMPage8KB(0)
	b.selectPRGPage32KB(0)

	return nil

	// TODO: load and map CHR-RAM if present in cartridge.
}
