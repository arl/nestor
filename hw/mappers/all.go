package mappers

import (
	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
)

var modMapper = log.NewModule("mapper")

var All = map[uint16]hw.MapperDesc{
	0:  NROM,
	2:  UxROM,
	3:  CNROM,
	66: GxROM,
}

func copyCHRROM(ppu *hw.PPU, rom *ines.Rom, bank uint32) {
	// Copy CHRROM bank to PPU memory.
	// CHRROM is 8KB in size
	start := bank * 0x2000
	end := start + 0x2000
	copy(ppu.PatternTables.Data, rom.CHRROM[start:end])
}
