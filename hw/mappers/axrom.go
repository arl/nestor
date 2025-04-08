package mappers

import (
	"nestor/ines"
)

var AxROM = MapperDesc{
	Name:         "AxROM",
	Load:         loadAxROM,
	PRGROMbanksz: 0x8000,
}

type axrom struct {
	*base

	ntm          ines.NTMirroring
	prgbank      uint32
	busConflicts bool
}

func (m *axrom) WritePRGROM(addr uint16, val uint8) {
	old := val
	if m.busConflicts {
		old = m.cpu.Bus.Peek8(addr)
		val &= old
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
		m.selectPRGPage32KB(int(m.prgbank))
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
	axrom := &axrom{
		base:         b,
		busConflicts: b.rom.SubMapper() == 2,
	}
	b.init(axrom.WritePRGROM)

	b.selectCHRROMPage8KB(0)
	b.selectPRGPage32KB(0)
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}
