package mappers

import (
	"nestor/hw/snapshot"
	"nestor/ines"
)

var Mapper34 = mapperDesc{
	Name:        "Mapper34",
	Load:        loadMapper34,
	PRGBankSize: 0x8000,
	CHRBankSize: 0x1000,
}

type mapper34 struct {
	*base

	isNINA  bool
	prgbank uint32
	chrbank uint32
}

func loadMapper34(b *base) (Mapper, error) {
	isNINA := b.rom.SubMapper() == 1
	if !isNINA && b.rom.SubMapper() != 2 {
		if len(b.rom.CHRROM) > 8*1024 {
			isNINA = true
		}
	}

	m34 := &mapper34{
		base:   b,
		isNINA: isNINA,
	}
	b.init(m34.WritePRGROM)

	b.setNTMirroring(ines.VertMirroring)
	if isNINA {
		b.selectCHRROMPage4KB(0, 0)
		b.selectCHRROMPage4KB(1, 1)
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

	prevprg := m.prgbank
	m.prgbank = uint32(val & 0x3)
	if prevprg != m.prgbank {
		m.selectPRGPage32KB(int(m.prgbank))
	}

	if m.isNINA {
		prevchr := m.chrbank
		m.chrbank = uint32((val >> 2) & 0xF)
		if prevchr != m.chrbank {
			m.selectCHRROMPage4KB(0, int(m.chrbank))
			m.selectCHRROMPage4KB(1, int(m.chrbank)+1)
		}
	}
}

func (m *mapper34) State() *snapshot.MapperState {
	state := &snapshot.Mapper34State{
		BaseState: m.base.state(),
		IsNINA:    m.isNINA,
		PRGBank:   m.prgbank,
		CHRBank:   m.chrbank,
	}

	return encodeState(m.rom.Number(), state)
}

func (m *mapper34) SetState(ms *snapshot.MapperState) {
	s := decodeState[snapshot.Mapper34State](ms)

	m.base.setState(s.BaseState)
	m.isNINA = s.IsNINA
	m.prgbank = s.PRGBank
	m.chrbank = s.CHRBank

	m.selectPRGPage32KB(int(m.prgbank))
	if m.isNINA {
		m.selectCHRROMPage4KB(0, int(m.chrbank))
		m.selectCHRROMPage4KB(1, int(m.chrbank)+1)
	}
}
