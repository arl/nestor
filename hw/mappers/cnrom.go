package mappers

import (
	"nestor/hw"
	"nestor/hw/hwio"
	"nestor/ines"
)

var CNROM = hw.MapperDesc{
	Name: "CNROM",
	Load: loadCNROM,
}

type cnrom struct {
	rom *ines.Rom
	ppu *hw.PPU

	// switchable CHRROM bank
	PRGROM hwio.Device `hwio:"offset=0x8000,size=0x8000,rcb,wcb"`
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

	copyCHRROM(m.ppu, m.rom, m.cur)

	modMapper.InfoZ("CHRROM bank switch").
		Uint32("prev", prev).
		Uint32("new", m.cur).
		End()
}

func copyCHRROM(ppu *hw.PPU, rom *ines.Rom, bank uint32) {
	// Copy CHRROM bank to PPU memory.
	// CHRROM is 8KB in size
	start := bank * 0x2000
	end := start + 0x2000
	copy(ppu.PatternTables.Data, rom.CHRROM[start:end])
}

// TODO: bus conflicts
func loadCNROM(rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) error {
	cnrom := &cnrom{
		rom: rom,
		ppu: ppu,
	}
	hwio.MustInitRegs(cnrom)

	// Map CNROM banks onto CPU address space.
	cpu.Bus.MapBank(0x0000, cnrom, 0)

	// PPU mapping.
	hw.SetNTMirroring(ppu, rom.Mirroring())
	copyCHRROM(ppu, rom, 0)
	// copy(ppu.PatternTables.Data, rom.CHRROM)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
