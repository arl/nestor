package mappers

import (
	"nestor/hw/snapshot"
)

var Mapper34 = mapperDesc{
	Name:        "Mapper34",
	Load:        loadMapper34,
	PRGBankSize: 0x8000,
	CHRBankSize: 0x1000,
	// Mapper 34 (NINA-001) uses bank registers at $7FFD-$7FFF (PRG-RAM space).
	// BNROM variants use $8000-$FFFF (PRG-ROM space). Start at $7FFD and ignore
	// the PRG-RAM registers when not in NINA mode.
	RegisterStart: 0x7FFD,
}

type mapper34 struct {
	*base

	isNINA  bool
	prgbank uint32
	chrbank0 uint32 // PPU $0000-$0FFF (4KB), NINA-001
	chrbank1 uint32 // PPU $1000-$1FFF (4KB), NINA-001
}

func loadMapper34(b *base) (Mapper, error) {
	isNINA := b.rom.SubMapper() == 1
	if !isNINA && b.rom.SubMapper() != 2 {
		if len(b.rom.CHRROM) > 8*1024 {
			isNINA = true
		}
	}

	m34 := &mapper34{
		base:     b,
		isNINA:   isNINA,
		chrbank0: 0,
		chrbank1: 1,
	}
	b.init(m34.WritePRGROM)

	// BNROM and NINA-001 use fixed mirroring specified by the ROM header.
	b.setNTMirroring(b.rom.Mirroring())
	if isNINA {
		b.selectCHRROMPage4KB(0, int(m34.chrbank0))
		b.selectCHRROMPage4KB(1, int(m34.chrbank1))
	} else {
		b.selectCHRROMPage8KB(0)
	}
	b.selectPRGPage32KB(0)

	return m34, nil
}

func (m *mapper34) WritePRGROM(addr uint16, val uint8) {
	modMapper.DebugZ("WritePRGROM").
		Hex16("addr", addr).
		Hex8("val", val).
		Bool("nina", m.isNINA).
		End()

	// NINA-001: bank registers are in PRG-RAM space at $7FFD-$7FFF.
	// BNROM: PRG bank selected by writes to PRG-ROM space ($8000-$FFFF).
	if m.isNINA {
		switch addr {
		case 0x7FFD:
			chr := uint32(val & 0x0F)
			if m.chrbank0 != chr {
				m.chrbank0 = chr
				m.selectCHRROMPage4KB(0, int(m.chrbank0))
			}
			return
		case 0x7FFE:
			chr := uint32(val & 0x0F)
			if m.chrbank1 != chr {
				m.chrbank1 = chr
				m.selectCHRROMPage4KB(1, int(m.chrbank1))
			}
			return
		case 0x7FFF:
			// PRG bank (32KB) is selected here.
			// fallthrough to PRG select below
		default:
			// Some boards mirror PRG bank selection into $8000-$FFFF as well.
			if addr < 0x8000 {
				return
			}
		}
	}

	// PRG bank select (BNROM and NINA-001 @ $7FFF).
	// Mask is conservative for typical 32KB banking; selection wraps via mapper logic/ROM size.
	prevprg := m.prgbank
	m.prgbank = uint32(val & 0x3)
	if prevprg != m.prgbank {
		m.selectPRGPage32KB(int(m.prgbank))
	}
}

func (m *mapper34) State() *snapshot.MapperState {
	// Pack the two 4KB CHR banks into one field to preserve snapshot format.
	chrPacked := m.chrbank0 | (m.chrbank1 << 16)
	state := &snapshot.Mapper34State{
		BaseState: m.base.state(),
		IsNINA:    m.isNINA,
		PRGBank:   m.prgbank,
		CHRBank:   chrPacked,
	}

	return encodeState(m.rom.Number(), state)
}

func (m *mapper34) SetState(ms *snapshot.MapperState) {
	s := decodeState[snapshot.Mapper34State](ms)

	m.base.setState(s.BaseState)
	m.isNINA = s.IsNINA
	m.prgbank = s.PRGBank
	m.chrbank0 = s.CHRBank & 0xFFFF
	m.chrbank1 = (s.CHRBank >> 16) & 0xFFFF

	m.selectPRGPage32KB(int(m.prgbank))
	if m.isNINA {
		m.selectCHRROMPage4KB(0, int(m.chrbank0))
		m.selectCHRROMPage4KB(1, int(m.chrbank1))
	}
}
