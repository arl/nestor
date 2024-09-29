package mappers

import (
	"fmt"

	"nestor/emu/hwio"
	"nestor/hw"
	"nestor/ines"
)

var UxROM = hw.MapperDesc{
	Name: "UxROM",
	Load: loadUxROM,
}

type uxrom struct {
	rom *ines.Rom

	// switchable PRGROM bank
	PRGROM hwio.Manual `hwio:"offset=0x0000,size=0x8000"`
	cur    uint        // current bank
}

func (m *uxrom) ReadPRGROM(addr uint16, _ bool) uint8 {
	var bank uint32
	if addr >= 0xC000 {
		bank = uint32(m.rom.PRGROMSlots() - 1)
	} else {
		bank = uint32(m.cur)
	}
	romaddr := (bank * 0x4000) + uint32(addr&0x3FFF)
	val := m.rom.PRGROM[romaddr]
	modMapper.DebugZ("PRGROM read").
		Hex16("addr", addr).
		Hex32("romaddr", romaddr).
		Uint("bank", m.cur).
		Hex8("val", val).
		End()
	return val
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
	m.cur = uint(val & 0b111)
	modMapper.DebugZ("PRGROM bank switch").
		Uint("prev", prev).
		Uint("new", m.cur).
		End()
}

func loadUxROM(rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) error {
	// CPU memory space mapping.

	// PRG-RAM (always provide it, though it should only be for 'Family basic').
	PRGRAM := make([]uint8, 0x2000)

	cpu.Bus.MapMemorySlice(0x6000, 0x7FFF, PRGRAM, false)
	modMapper.DebugZ("map PRGRAM to cpu").
		String("range", "0x6000-0x7FFF").
		String("src", "slice").
		End()

	uxrom := &uxrom{
		rom: rom,
		cur: 0,
	}
	hwio.MustInitRegs(uxrom)
	cpu.Bus.MapBank(0x8000, uxrom, 0)

	// PPU memory space mapping.
	// Nametables.
	switch rom.Mirroring() {
	case ines.HorzMirroring: // A A B B
		ppu.SetMirroring(hw.HorzMirroring)
	case ines.VertMirroring: // A B A B
		ppu.SetMirroring(hw.VertMirroring)
	default:
		return fmt.Errorf("unexpected mirroring: %d", rom.Mirroring())
	}

	// continuer ici avec le log
	// Copy CHRROM to Pattern Tables.
	switch len(rom.CHRROM) {
	case 0x0000:
		// Some roms have no CHR-ROM, allocate mem anyway.
		// XXX: not sure about that
		// copy(ppu.PatternTables.Data, make([]uint8, 0x2000))
		// modMapper.DebugZ("map CHR-ROM to pattern tables").
		// 	String("range", "0x0000-0x2000").
		// 	String("src", "rom.CHR-ROM").
		// 	End()
		//panic("not sure about that, check comment")

	case 0x2000:
		copy(ppu.PatternTables.Data, rom.CHRROM)
		modMapper.DebugZ("map CHR-ROM to pattern tables").
			String("range", "0x0000-0x2000").
			String("src", "rom.CHR-ROM").
			End()
	default:
		return fmt.Errorf("unexpected CHRROM size: 0x%x", len(rom.CHRROM))
	}

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
	return nil
}
