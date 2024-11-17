package mappers

import (
	"nestor/hw"
	"nestor/hw/hwio"
	"nestor/ines"
)

var NROM = hw.MapperDesc{
	Name: "NROM",
	Load: loadNROM,
}

type nrom struct {
	PRGRAM hwio.Mem `hwio:"offset=0x6000,size=0x2000"`
	PRGROM hwio.Mem `hwio:"offset=0x8000,vsize=0x8000,readonly"`
}

func loadNROM(rom *ines.Rom, cpu *hw.CPU, ppu *hw.PPU) error {
	nrom := &nrom{}
	hwio.MustInitRegs(nrom)

	// CPU mapping.

	// Dimension the PRGROM based on the length of the cartridge PRGROM.
	// PRGROM mirrors are taken care of by hwio.Mem 'vsize'.
	nrom.PRGROM.Data = make([]byte, len(rom.PRGROM))
	copy(nrom.PRGROM.Data, rom.PRGROM)

	cpu.Bus.MapBank(0x0000, nrom, 0)

	// PPU mapping.
	hw.SetNTMirroring(ppu, rom.Mirroring())
	copy(ppu.PatternTables.Data, rom.CHRROM)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
