package mappers

import (
	"nestor/hw/hwio"
)

var NROM = MapperDesc{
	Name: "NROM",
	Load: loadNROM,
}

type nrom struct {
	*base
	PRGRAM hwio.Mem `hwio:"offset=0x6000,size=0x2000"`
	PRGROM hwio.Device
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
	nrom.PRGROM = hwio.Device{
		Name:    "PRGROM",
		Size:    0x8000,
		ReadCb:  nrom.ReadPRGROM,
		PeekCb:  nrom.ReadPRGROM,
		WriteCb: nil, // no bank-switching
	}
	b.cpu.Bus.MapDevice(0x8000, &nrom.PRGROM)

	b.cpu.Bus.MapBank(0x0000, nrom, 0)

	// PPU mapping.
	b.setNTMirroring(b.rom.Mirroring())
	copy(b.ppu.PatternTables.Data, b.rom.CHRROM)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
