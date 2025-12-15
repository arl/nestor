package mappers

import (
	"nestor/hw/snapshot"
)

var BNROM = mapperDesc{
	Name:        "Mapper34",
	Load:        loadBNROM,
	PRGBankSize: 0x8000,
	CHRBankSize: 0x2000, // 8KB CHR-RAM
}

type mapper34 struct {
	*base
	prgbank uint32
}

func loadBNROM(b *base) (Mapper, error) {
	m34 := &mapper34{
		base:    b,
		prgbank: 0,
	}

	// BNROM: PRG bank selected by writes to PRG-ROM space ($8000-$FFFF)
	b.registers.Reset()
	b.registers.SetRange(0x8000, 0x10000)

	b.init(m34.WritePRGROM)

	// BNROM uses fixed mirroring specified by the ROM header
	b.setNTMirroring(b.rom.Mirroring())
	// BNROM uses CHR-RAM (8KB), no bank switching
	b.selectCHRROMPage8KB(0)
	b.selectPRGPage32KB(0)

	return m34, nil
}

func (m *mapper34) WritePRGROM(addr uint16, val uint8) {
	modMapper.DebugZ("WritePRGROM").
		Hex16("addr", addr).
		Hex8("val", val).
		End()

	// BNROM: PRG bank selected by writes to PRG-ROM space ($8000-$FFFF). Ignore
	// writes in PRG-RAM space.
	if addr < 0x8000 {
		return
	}

	// PRG bank select (32KB banks)
	//
	// Mask is conservative for typical 32KB banking;
	// selection wraps via mapper logic/ROM size.
	prevprg := m.prgbank
	m.prgbank = uint32(val & 0x3)
	if prevprg != m.prgbank {
		m.selectPRGPage32KB(int(m.prgbank))
	}
}

func (m *mapper34) State() *snapshot.MapperState {
	state := &snapshot.Mapper34State{
		BaseState: m.base.state(),
		PRGBank:   m.prgbank,
		CHRBank:   0, // BNROM doesn't use CHR banking
	}

	return encodeState(m.rom.Number(), state)
}

func (m *mapper34) SetState(ms *snapshot.MapperState) {
	s := decodeState[snapshot.Mapper34State](ms)

	m.base.setState(s.BaseState)
	m.prgbank = s.PRGBank

	m.selectPRGPage32KB(int(m.prgbank))
}
