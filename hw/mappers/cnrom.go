package mappers

import (
	"nestor/hw/hwio"
)

var CNROM = MapperDesc{
	Name: "CNROM",
	Load: loadCNROM,
}

type cnrom struct {
	*base

	// switchable CHRROM bank
	PRGROM hwio.Device `hwio:"offset=0x8000,size=0x8000,rcb,pcb=ReadPRGROM,wcb"`
	cur    uint32      // current CHRROM bank
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
	prev := m.cur
	m.cur = uint32(val & 0b11)
	if prev != m.cur {
		copyCHRROM(m.ppu, m.rom, m.cur)
		modMapper.InfoZ("CHRROM bank switch").
			Uint32("prev", prev).
			Uint32("new", m.cur).
			End()
	}
}

// TODO: bus conflicts
func loadCNROM(b *base) error {
	cnrom := &cnrom{
		base: b,
	}
	hwio.MustInitRegs(cnrom)

	// Map CNROM banks onto CPU address space.
	b.cpu.Bus.MapBank(0x0000, cnrom, 0)

	// PPU mapping.
	b.setNametableMirroring()
	copyCHRROM(b.ppu, b.rom, 0)
	// copy(ppu.PatternTables.Data, rom.CHRROM)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
