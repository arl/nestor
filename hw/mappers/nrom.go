package mappers

var NROM = MapperDesc{
	Name:         "NROM",
	Load:         loadNROM,
	CHRROMbanksz: 0x2000,
}

type nrom struct{ *base }

func loadNROM(b *base) (Mapper, error) {
	nrom := &nrom{base: b}
	b.init(nil)

	b.setNTMirroring(b.rom.Mirroring())
	b.selectCHRROMPage8KB(0)
	switch len(b.rom.PRGROM) {
	case 16 * KB:
		b.selectPRGPage16KB(0, 0)
		b.selectPRGPage16KB(1, 0) // mirror
	case 32 * KB:
		b.selectPRGPage32KB(0)
	default:
		return nil, ErrUnsuppportedPRGROMSize(len(b.rom.PRGROM))
	}

	// TODO: handle ROMS with CHRRAM
	return nrom, nil
}
