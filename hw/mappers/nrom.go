package mappers

import (
	"fmt"

	"nestor/hw"
	"nestor/ines"
)

var NROM = hw.MapperDesc{
	Name: "NROM",
	Load: loadNROM,
}

func loadNROM(rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) error {
	// CPU memory space mapping.
	//

	// PRG-RAM (always provide it, though it should only be for 'Family basic').
	PRGRAM := make([]uint8, 0x2000)
	cpu.Bus.MapMemorySlice(0x6000, 0x7FFF, PRGRAM, false)
	modMapper.DebugZ("map PRG-RAM to cpu").
		String("range", "0x6000-0x7FFF").
		String("src", "slice").
		End()

	switch len(rom.PRGROM) {
	case 0x4000:
		cpu.Bus.MapMemorySlice(0x8000, 0xBFFF, rom.PRGROM, true)
		modMapper.DebugZ("map PRG-ROM to cpu").
			String("range", "0x8000-0xBFFF").
			String("src", "rom.PRG-ROM").
			End()

		cpu.Bus.MapMemorySlice(0xC000, 0xFFFF, rom.PRGROM, true) // mirror
		modMapper.DebugZ("map PRG-ROM to cpu").
			String("range", "0xC000-0xFFFF").
			String("src", "rom.PRG-ROM").
			End()
	case 0x8000:
		modMapper.DebugZ("map PRG-ROM to cpu").
			String("range", "0x8000-0xFFFF").
			String("src", "rom.PRG-ROM").
			End()
		cpu.Bus.MapMemorySlice(0x8000, 0xFFFF, rom.PRGROM, true)
	default:
		return fmt.Errorf("unexpected PRGROM size: 0x%x", len(rom.CHRROM))
	}

	// PPU memory space mapping.
	// NameTables
	switch rom.Mirroring() {
	case ines.HorzMirroring: // A A B B
		// 4 nametables
		ppu.Bus.MapMemorySlice(0x2000, 0x23FF, ppu.Nametables[:0x400], false)
		ppu.Bus.MapMemorySlice(0x2400, 0x27FF, ppu.Nametables[:0x400], false)
		ppu.Bus.MapMemorySlice(0x2800, 0x2BFF, ppu.Nametables[0x400:0x800], false)
		ppu.Bus.MapMemorySlice(0x2C00, 0x2FFF, ppu.Nametables[0x400:0x800], false)

		// mirrors of the nametable area
		ppu.Bus.MapMemorySlice(0x3000, 0x33FF, ppu.Nametables[:0x400], false)
		ppu.Bus.MapMemorySlice(0x3400, 0x37FF, ppu.Nametables[:0x400], false)
		ppu.Bus.MapMemorySlice(0x3800, 0x3BFF, ppu.Nametables[0x400:0x800], false)
		ppu.Bus.MapMemorySlice(0x3C00, 0x3EFF, ppu.Nametables[0x400:0x800], false)

	case ines.VertMirroring: // A B A B
		// 4 nametables
		ppu.Bus.MapMemorySlice(0x2000, 0x23FF, ppu.Nametables[:0x400], false)
		ppu.Bus.MapMemorySlice(0x2400, 0x27FF, ppu.Nametables[0x400:0x800], false)
		ppu.Bus.MapMemorySlice(0x2800, 0x2BFF, ppu.Nametables[:0x400], false)
		ppu.Bus.MapMemorySlice(0x2C00, 0x2FFF, ppu.Nametables[0x400:0x800], false)

		// mirrors of the nametable area
		ppu.Bus.MapMemorySlice(0x3000, 0x33FF, ppu.Nametables[:0x400], false)
		ppu.Bus.MapMemorySlice(0x3400, 0x37FF, ppu.Nametables[0x400:0x800], false)
		ppu.Bus.MapMemorySlice(0x3800, 0x3BFF, ppu.Nametables[:0x400], false)
		ppu.Bus.MapMemorySlice(0x3C00, 0x3EFF, ppu.Nametables[0x400:0x800], false)
	}

	// continuer ici avec le log
	// Copy CHRROM to Pattern Tables.
	switch len(rom.CHRROM) {
	case 0x0000:
		// Some roms have no CHR-ROM, allocate mem anyway.
		// XXX: not sure about that
		copy(ppu.PatternTables.Data, make([]uint8, 0x2000))
		modMapper.DebugZ("map CHR-ROM to pattern tables").
			String("range", "0x0000-0x2000").
			String("src", "rom.CHR-ROM").
			End()
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
