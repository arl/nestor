package mappers

import (
	"nestor/hw/hwio"
)

var UxROM = MapperDesc{
	Name: "UxROM",
	Load: loadUxROM,
}

type uxrom struct {
	*base

	PRGRAM hwio.Mem `hwio:"offset=0x6000,size=0x2000"`

	// switchable PRGROM bank
	PRGROM hwio.Device `hwio:"offset=0x8000,size=0x8000,rcb,pcb=ReadPRGROM,wcb"`
	cur    uint32      // current bank
}

func (m *uxrom) ReadPRGROM(addr uint16) uint8 {
	banknum := m.cur
	if addr >= 0xC000 {
		banknum = uint32(m.rom.PRGROMSlots() - 1)
	}
	romaddr := (banknum * 0x4000) + uint32(addr&0x3FFF)
	return m.rom.PRGROM[romaddr]
}

func (m *uxrom) WritePRGROM(addr uint16, val uint8) {
	// Switch bank.

	// 7  bit  0
	// ---- ----
	// xxxx pPPP
	//      ||||
	//      ++++- Select 16 KB PRG ROM bank for CPU $8000-$BFFF
	//            (UNROM uses bits 2-0; UOROM uses bits 3-0)
	prev := m.cur
	m.cur = uint32(val & 0b111)
	if prev != m.cur {
		modMapper.DebugZ("PRGROM bank switch").
			Uint32("prev", prev).
			Uint32("new", m.cur).
			End()
	}
}

func loadUxROM(b *base) error {
	uxrom := &uxrom{base: b}
	hwio.MustInitRegs(uxrom)

	// CPU mapping.
	b.cpu.Bus.MapBank(0x0000, uxrom, 0)

	// PPU mapping.
	b.setNametableMirroring()
	copy(b.ppu.PatternTables.Data, b.rom.CHRROM)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
