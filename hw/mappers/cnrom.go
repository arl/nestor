package mappers

import (
	"nestor/hw/hwio"
)

var CNROM = MapperDesc{
	Name:           "CNROM",
	Load:           loadCNROM,
	PRGROMpagesize: 0x8000,
	CHRROMpagesize: 0x2000,
}

type cnrom struct {
	*base

	// switchable CHRROM bank
	PRGROM  hwio.Device
	chrPage uint32
}

func (m *cnrom) ReadPRGROM(addr uint16) uint8 {
	addr &= 0x7FFF                        // max PRGROM size is 32KB
	addr &= uint16(len(m.rom.PRGROM) - 1) // PRGROM mirrors
	return m.rom.PRGROM[addr]
}

func (m *cnrom) WritePRGROM(addr uint16, val uint8) {
	// Switch bank.

	// 7  bit  0
	// ---- ----
	// cccc ccCC
	// |||| ||||
	// ++++-++++- Select 8 KB CHR ROM bank for PPU $0000-$1FFF
	// CNROM only uses loweest 2 bits
	prev := m.chrPage
	m.chrPage = uint32(val & 0b11)
	if prev != m.chrPage {
		copyCHRROM(m.ppu, m.rom, m.chrPage)
		modMapper.InfoZ("CHRROM bank switch").
			Uint32("prev", prev).
			Uint32("new", m.chrPage).
			End()
	}
}

// TODO: bus conflicts
func loadCNROM(b *base) error {
	cnrom := &cnrom{base: b}
	hwio.MustInitRegs(cnrom)

	// CPU mapping.
	cnrom.PRGROM = hwio.Device{
		Name:    "PRGROM",
		Size:    0x8000,
		ReadCb:  cnrom.ReadPRGROM,
		PeekCb:  cnrom.ReadPRGROM,
		WriteCb: cnrom.WritePRGROM,
	}
	b.cpu.Bus.MapDevice(0x8000, &cnrom.PRGROM)

	// PPU mapping.
	b.setNametableMirroring()
	copyCHRROM(b.ppu, b.rom, 0)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
