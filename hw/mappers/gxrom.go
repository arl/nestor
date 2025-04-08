package mappers

var GxROM = MapperDesc{
	Name:         "GxROM",
	Load:         loadGxROM,
	PRGROMbanksz: 0x8000,
	CHRROMbanksz: 0x2000,
}

type gxrom struct {
	*base

	chrbank uint32
	prgbank uint32
}

func (m *gxrom) WritePRGROM(addr uint16, val uint8) {
	// 7  bit  0
	// ---- ----
	// xxPP xxCC
	//   ||   ||
	//   ||   ++- Select 8 KB CHR ROM bank for PPU $0000-$1FFF
	//   ++------ Select 32 KB PRG ROM bank for CPU $8000-$FFFF
	prevchr := m.chrbank
	m.chrbank = uint32(val & 0x3)
	if prevchr != m.chrbank {
		m.selectCHRROMPage8KB(int(m.chrbank))
	}

	prevprg := m.prgbank
	m.prgbank = uint32((val >> 4) & 0x3)
	if prevprg != m.prgbank {
		m.selectPRGPage32KB(int(m.prgbank))
	}
}

func loadGxROM(b *base) error {
	gxrom := &gxrom{base: b}
	b.init(gxrom.WritePRGROM)

	b.setNTMirroring(b.rom.Mirroring())
	b.selectCHRROMPage8KB(0)
	b.selectPRGPage32KB(0)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
