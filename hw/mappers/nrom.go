package mappers

import "nestor/hw/hwio"

var NROM = MapperDesc{
	Name:         "NROM",
	Load:         loadNROM,
	CHRROMbanksz: 0x2000,
}

type nrom struct {
	*base
	/* CPU */
	PRGRAM hwio.Mem    `hwio:"offset=0x6000,size=0x2000"`
	PRGROM hwio.Device `hwio:"offset=0x8000,size=0x8000,rcb"`

	/* PPU */
	PatternTables hwio.Mem `hwio:"bank=1,offset=0x0000,size=0x2000"`
}

func (m *nrom) ReadPRGROM(addr uint16) uint8 {
	addr &= 0x7FFF                        // max PRGROM size is 32KB
	addr &= uint16(len(m.rom.PRGROM) - 1) // PRGROM mirrors
	return m.rom.PRGROM[addr]
}

func loadNROM(b *base) error {
	nrom := &nrom{base: b}
	hwio.MustInitRegs(nrom)

	// CPU mapping.
	b.cpu.Bus.MapBank(0x0000, nrom, 0)

	// PPU mapping.
	b.setNTMirroring(b.rom.Mirroring())
	b.ppu.Bus.MapBank(0x0000, nrom, 1)
	b.copyCHRROM(nrom.PatternTables.Data, 0)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
