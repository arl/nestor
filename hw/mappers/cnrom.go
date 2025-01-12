package mappers

import (
	"nestor/hw/hwio"
)

var CNROM = MapperDesc{
	Name:            "CNROM",
	Load:            loadCNROM,
	PRGROMbanksz:    0x8000,
	CHRROMbanksz:    0x2000,
	HasBusConflicts: func(b *base) bool { return b.rom.SubMapper() == 2 },
}

type cnrom struct {
	*base

	PatternTables hwio.Mem `hwio:"offset=0x0000,size=0x2000,readonly"`

	// switchable CHRROM bank
	PRGROM  hwio.Device
	chrbank uint32
}

func (m *cnrom) ReadPRGROM(addr uint16) uint8 {
	addr -= 0x8000
	addr &= uint16(len(m.rom.PRGROM) - 1)
	return m.rom.PRGROM[addr]
}

func (m *cnrom) WritePRGROM(addr uint16, val uint8) {
	if m.hasBusConflicts {
		val &= m.ReadPRGROM(addr)
	}

	// 7  bit  0
	// ---- ----
	// cccc ccCC
	// |||| ||||
	// ++++-++++- Select 8 KB CHR ROM bank for PPU $0000-$1FFF
	// CNROM only uses lowest 2 bits
	prev := m.chrbank
	m.chrbank = uint32(val & 0b11)
	if prev != m.chrbank {
		copyCHRROM(m.PatternTables.Data, m.rom, m.chrbank)
		modMapper.InfoZ("CHRROM bank switch").String("mapper", m.desc.Name).Uint32("prev", prev).Uint32("new", m.chrbank).End()
	}
}

func loadCNROM(b *base) error {
	cnrom := &cnrom{base: b}
	hwio.MustInitRegs(cnrom)
	b.ppu.Bus.MapBank(0x0000, cnrom, 0)

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
	b.setNTMirroring(b.rom.Mirroring())
	copyCHRROM(cnrom.PatternTables.Data, b.rom, 0)

	return nil

	// TODO: load and map CHR-RAM if present in cartridge.
}
