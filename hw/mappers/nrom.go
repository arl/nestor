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
	PRGROM hwio.Mem
}

func loadNROM(b *base) error {
	nrom := &nrom{}
	hwio.MustInitRegs(nrom)

	// CPU mapping.
	nrom.PRGROM = hwio.Mem{
		Name:  "PRGROM",
		Data:  b.rom.PRGROM,
		VSize: 0x8000,
		Flags: hwio.MemFlag8ReadOnly,
	}
	b.cpu.Bus.MapMem(0x8000, &nrom.PRGROM)
	b.cpu.Bus.MapBank(0x0000, nrom, 0)

	// PPU mapping.
	b.setNametableMirroring()
	copy(b.ppu.PatternTables.Data, b.rom.CHRROM)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
