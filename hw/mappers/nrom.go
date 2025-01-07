package mappers

import (
	"nestor/hw/hwio"
)

var NROM = MapperDesc{
	Name:           "NROM",
	Load:           loadNROM,
	PRGROMpagesize: 0x4000,
	CHRROMpagesize: 0x2000,
}

type nrom struct {
	PRGRAM hwio.Mem `hwio:"offset=0x6000,size=0x2000"`
	PRGROM hwio.Mem `hwio:"offset=0x8000,vsize=0x8000,readonly"`
}

func loadNROM(b *base) error {
	nrom := &nrom{}
	hwio.MustInitRegs(nrom)

	// CPU mapping.

	// Dimension the PRGROM based on the length of the cartridge PRGROM.
	// PRGROM mirrors are taken care of by hwio.Mem 'vsize'.
	nrom.PRGROM.Data = make([]byte, len(b.rom.PRGROM))
	copy(nrom.PRGROM.Data, b.rom.PRGROM)

	b.cpu.Bus.MapBank(0x0000, nrom, 0)

	// PPU mapping.
	b.setNametableMirroring()
	copy(b.ppu.PatternTables.Data, b.rom.CHRROM)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
