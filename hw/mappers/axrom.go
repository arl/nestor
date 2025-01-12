package mappers

import (
	"nestor/hw/hwio"
	"nestor/ines"
)

var AxROM = MapperDesc{
	Name:            "AxROM",
	Load:            loadAxROM,
	PRGROMbanksz:    0x8000,
	HasBusConflicts: func(b *base) bool { return b.rom.SubMapper() == 2 },
}

type axrom struct {
	*base

	PatternTables hwio.Mem `hwio:"offset=0x0000,size=0x2000"`
	PRGRAM        hwio.Mem `hwio:"offset=0x6000,size=0x2000"`

	// switchable PRGROM bank
	PRGROM  hwio.Device
	prgbank uint32
	ntm     ines.NTMirroring
}

func (m *axrom) ReadPRGROM(addr uint16) uint8 {
	addr &= uint16(m.desc.PRGROMbanksz - 1) // limit to max PRGROM size
	romaddr := (m.prgbank * m.desc.PRGROMbanksz) + uint32(addr)
	return m.rom.PRGROM[romaddr]
}

func (m *axrom) WritePRGROM(addr uint16, val uint8) {
	if m.hasBusConflicts {
		val &= m.ReadPRGROM(addr)
	}

	// 7  bit  0
	// ---- ----
	// xxxM xPPP
	//    |  |||
	//    |  +++- Select 32 KB PRG ROM bank for CPU $8000-$FFFF
	//    +------ Select 1 KB VRAM page for all 4 nametables
	prev := m.prgbank
	m.prgbank = uint32(val & 0x7)
	if prev != m.prgbank {
		modMapper.DebugZ("PRGROM bank switch").String("mapper", m.desc.Name).Uint32("prev", prev).Uint32("new", m.prgbank).End()
	}

	prevntm := m.ntm
	if val&0x10 == 0x10 {
		m.ntm = ines.OnlyBScreen
	} else {
		m.ntm = ines.OnlyAScreen
	}
	if prevntm != m.ntm {
		m.setNTMirroring(m.ntm)
		modMapper.DebugZ("select NT mirroring").String("mapper", m.desc.Name).Stringer("prev", prevntm).Stringer("new", m.ntm).End()
	}
}

func loadAxROM(b *base) error {
	axrom := &axrom{base: b}
	hwio.MustInitRegs(axrom)
	b.ppu.Bus.MapBank(0x0000, axrom, 0)

	// CPU mapping.
	axrom.PRGROM = hwio.Device{
		Name:    "PRGROM",
		Size:    0x8000,
		ReadCb:  axrom.ReadPRGROM,
		PeekCb:  axrom.ReadPRGROM,
		WriteCb: axrom.WritePRGROM,
	}
	b.cpu.Bus.MapDevice(0x8000, &axrom.PRGROM)

	// PPU mapping.
	copy(axrom.PatternTables.Data, b.rom.CHRROM)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
