package mappers

import (
	"nestor/hw/snapshot"
	"nestor/ines"
)

var MMC1 = mapperDesc{
	Name:        "MMC1",
	Load:        loadMMC1,
	PRGBankSize: 0x4000,
	CHRBankSize: 0x1000,
}

type mmc1 struct {
	*base

	prevCycle int64

	serial  shiftReg // shift register
	counter uint8    // count of bits shifted

	// CTRL reg bits
	//
	// 4bit0
	// -----
	// CPPMM
	// |||||
	// |||++- Nametable arrangement: (0: one-screen, lower bank; 1: one-screen, upper bank;
	// |||               2: horizontal arrangement ("vertical mirroring", PPU A10);
	// |||               3: vertical arrangement ("horizontal mirroring", PPU A11) )
	// |++--- PRG-ROM bank mode (0, 1: switch 32 KB at $8000, ignoring low bit of bank number;
	// |                         2: fix first bank at $8000 and switch 16 KB bank at $C000;
	// |                         3: fix last bank at $C000 and switch 16 KB bank at $8000)
	// +----- CHR-ROM bank mode (0: switch 8 KB at a time; 1: switch two separate 4 KB banks)
	nt      uint8
	prgmode uint8
	chrmode uint8

	// CHR bank 0 bits
	//
	// 4bit0
	// -----
	// CCCCC
	// |||||
	// +++++- Select 4 KB or 8 KB CHR bank at PPU $0000 (low bit ignored in 8 KB mode)
	chrbank0 uint32

	// CHR bank 1 bits
	//
	// 4bit0
	// -----
	// CCCCC
	// |||||
	// +++++- Select 4 KB CHR bank at PPU $1000 (ignored in 8 KB mode)
	chrbank1 uint32
	lastchr  uint16

	// PRG reg bits
	//
	// 4bit0
	// -----
	// RPPPP
	// |||||
	// |++++- Select 16 KB PRG-ROM bank (low bit ignored in 32 KB mode)
	// +----- MMC1B and later: PRG-RAM chip enable (0: enabled; 1: disabled; ignored on MMC1A)
	//        MMC1A: Bit 3 bypasses fixed bank logic in 16K mode (0: fixed bank affects A17-A14;
	//        1: fixed bank affects A16-A14 and bit 3 directly controls A17)
	prgbank     uint32
	disableWRAM bool // TODO: unused for now
}

func loadMMC1(b *base) (Mapper, error) {
	mmc1 := &mmc1{base: b}

	b.init(mmc1.WritePRGROM)

	// PPU mapping.
	b.setNTMirroring(ines.OnlyAScreen)

	// Mapper initialization.
	// On powerup: bits 2,3 of $8000 are set (this ensures the $8000 is bank 0,
	// and $C000 is the last bank - needed for SEROM/SHROM/SH1ROM which do no
	// support banking)
	mmc1.writeREG(0x8000, 0x0C)
	mmc1.writeREG(0xA000, 0)
	mmc1.writeREG(0xC000, 0)
	mmc1.writeREG(0xE000, 0) // TODO: WRAM Disable: enabled by default for MMC1B
	mmc1.disableWRAM = true  // TODO: always enabled on MMC1A
	mmc1.lastchr = 0xA000
	mmc1.remap()
	return mmc1, nil
}

type shiftReg uint8

func (sr shiftReg) push(val uint8) shiftReg {
	sr >>= 1
	sr |= shiftReg((val << 4) & 0x10)
	return sr
}

func (m *mmc1) WritePRGROM(addr uint16, val uint8) {
	curCycle := m.cpu.CurrentCycle()
	// Ignore consecutive cycle writes
	resetbit := u8tob(val & 0x80)
	if resetbit || curCycle-m.prevCycle >= 2 {
		if resetbit {
			// if the resetbit is set.
			//	- ignore databit
			//	- reset shift register (so that the next write is the "first" write)
			//	- bits 2,3 of control reg are set (16k PRG mode, $8000 swappable)
			//	- other bits of $8000 (and other regs) are unchanged
			m.serial = 0
			m.counter = 0
			m.prgmode = 0b11
			m.remap()
		} else {
			m.serial = m.serial.push(val)
			m.counter++
			if m.counter == 5 {
				m.writeREG(addr, uint8(m.serial))
				m.remap()
				m.serial = 0
				m.counter = 0
			}
		}
	}
	m.prevCycle = m.cpu.CurrentCycle()
}

func (m *mmc1) writeREG(addr uint16, val uint8) {
	switch (addr & 0x6000) >> 13 {
	case 0: // 0X8000
		m.writeCTRL(val)
	case 1: // 0xA000
		m.writeCHR0(val)
		m.lastchr = addr
	case 2: // 0xC000
		m.writeCHR1(val)
		m.lastchr = addr
	case 3: // 0xE000
		m.writePRG(val)
	}
}

func (m *mmc1) writeCTRL(val uint8) {
	m.chrmode = (val & 0x10) >> 4
	m.prgmode = (val & 0x0C) >> 2

	prevNT := m.nt
	m.nt = val & 0x03
	if prevNT != m.nt {
		switch m.nt {
		case 0:
			m.setNTMirroring(ines.OnlyAScreen)
		case 1:
			m.setNTMirroring(ines.OnlyBScreen)
		case 2:
			m.setNTMirroring(ines.VertMirroring)
		case 3:
			m.setNTMirroring(ines.HorzMirroring)
		}
	}

	modMapper.DebugZ("Write CTRL reg").String("mapper", m.desc.Name).
		Uint8("val", val).
		Uint8("prgmode", m.prgmode).
		Uint8("chrmode", m.chrmode).
		End()
}

func (m *mmc1) writeCHR0(val uint8) {
	bank := val & 0b11111
	m.chrbank0 = uint32(bank) // TODO: Adjust mask if CHRROM is larger
	modMapper.DebugZ("Write CHR0 reg").Hex8("val", val).Hex8("bank", bank).End()
}

func (m *mmc1) writeCHR1(val uint8) {
	bank := val & 0b11111
	m.chrbank1 = uint32(bank) // TODO: Adjust mask if CHRROM is larger
	modMapper.DebugZ("Write CHR1 reg").Hex8("val", val).Hex8("bank", bank).End()
}

func (m *mmc1) writePRG(val uint8) {
	modMapper.DebugZ("Write PRG reg").String("mapper", m.desc.Name).Uint8("val", val).End()

	// $E000-FFFF:  [...W PPPP]
	// W = WRAM Disable (0=enabled, 1=disabled)
	// P = PRG Reg
	m.disableWRAM = u8tob(val & 0b1_0000)
	m.prgbank = uint32(val & 0b1111)
	if m.disableWRAM {
		panic("disable WRAM not implemented")
	}
}

func (m *mmc1) remap() {
	extrareg := m.chrbank0
	if m.lastchr == 0xC000 && m.chrmode != 0 {
		extrareg = m.chrbank1
	}
	var prgbankSelect uint32
	if len(m.rom.PRGROM) == 0x80000 {
		// 512kb carts use bit 7 of $A000/$C000 to select page
		// This is used for SUROM (Dragon Warrior 3/4, Dragon Quest 4)
		prgbankSelect = extrareg & 0x10
	}

	const _forceWramOn = false // TODO: read from ROM header

	readonly := m.disableWRAM && !_forceWramOn

	switch totalram := m.rom.PRGRAMSize() + m.rom.PRGNVRAMSize(); {
	case totalram > 0x4000: // SxROM, 32kb of save ram.
		ram := m.PRGNVRAM
		if !m.rom.HasBattery() {
			ram = m.PRGRAM
		}
		bank := (extrareg >> 2) & 0x03
		m.cpu.Bus.Unmap(0x6000, 0x7FFF)
		m.cpu.Bus.MapMemorySlice(0x6000, 0x7FFF, ram[0x2000*bank:], readonly)

	case totalram > 0x2000:

		// TODO: test persistency

		// SOROM, half of the 16kb ram is battery backed.
		ram := m.PRGNVRAM
		if ((extrareg >> 3) & 0x01) != 0 {
			ram = m.PRGRAM
		}
		m.cpu.Bus.Unmap(0x6000, 0x7FFF)
		m.cpu.Bus.MapMemorySlice(0x6000, 0x7FFF, ram, readonly)

	case totalram == 0x2000:

		// TODO: test persistency

		m.cpu.Bus.Unmap(0x6000, 0x7FFF)
		ram := m.PRGNVRAM
		if m.rom.PRGRAMSize() == 0x2000 {
			ram = m.PRGRAM
		}
		m.cpu.Bus.MapMemorySlice(0x6000, 0x7FFF, ram, readonly)

	case totalram == 0:
		// Do not map any RAM

	default:
		panic("not supported")
	}

	// TODO: handle submapper 5

	switch m.prgmode {
	case 0, 1:
		// ignore low bit of bank number
		m.selectPRGPage32KB(int((m.prgbank & 0xFE) | prgbankSelect))
	case 2:
		m.selectPRGPage16KB(0, int(0|prgbankSelect))
		m.selectPRGPage16KB(1, int(m.prgbank|prgbankSelect))
	case 3:
		m.selectPRGPage16KB(0, int(m.prgbank|prgbankSelect))
		m.selectPRGPage16KB(1, int(0x0F|prgbankSelect))
	}

	switch m.chrmode {
	case 0:
		bank := int(m.chrbank0 & 0x1E)
		m.selectCHRROMPage4KB(0, bank)
		m.selectCHRROMPage4KB(1, bank+1)
	case 1:
		m.selectCHRROMPage4KB(0, int(m.chrbank0))
		m.selectCHRROMPage4KB(1, int(m.chrbank1))
	}
}

func (m *mmc1) State() *snapshot.MapperState {
	state := &snapshot.MMC1State{
		BaseState:   m.base.state(),
		PrevCycle:   m.prevCycle,
		Serial:      uint8(m.serial),
		Counter:     m.counter,
		NT:          m.nt,
		PRGMode:     m.prgmode,
		CHRMode:     m.chrmode,
		CHRBank0:    m.chrbank0,
		CHRBank1:    m.chrbank1,
		LastCHR:     m.lastchr,
		PRGBank:     m.prgbank,
		DisableWRAM: m.disableWRAM,
	}

	return encodeState(m.rom.Number(), state)
}

func (m *mmc1) SetState(ms *snapshot.MapperState) {
	s := decodeState[snapshot.MMC1State](ms)

	m.base.setState(s.BaseState)
	m.prevCycle = s.PrevCycle
	m.serial = shiftReg(s.Serial)
	m.counter = s.Counter
	m.nt = s.NT
	m.prgmode = s.PRGMode
	m.chrmode = s.CHRMode
	m.chrbank0 = s.CHRBank0
	m.chrbank1 = s.CHRBank1
	m.lastchr = s.LastCHR
	m.prgbank = s.PRGBank
	m.disableWRAM = s.DisableWRAM

	// Remap based on restored state
	m.remap()
}
