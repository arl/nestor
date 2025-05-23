package mappers

import (
	"nestor/ines"
)

var MMC1 = MapperDesc{
	Name: "MMC1",
	Load: loadMMC1,
	// PRGROMbanksz: 0x8000,
	// PRGRAMbanksz: 0x2000,
}

type mmc1 struct {
	*base

	prevCycle int64

	serial  shiftReg // shift register
	counter uint8    // count of bits shifted

	// CTRL reg bits
	chrmode uint8
	prgmode uint8
	ntm     uint8

	// CHR reg 0 bits
	chrbank0 uint32
	chrbank1 uint32
	lastchr  uint16

	// PRG reg bits
	disableWRAM bool // TODO: unused for now
	prgbank     uint32
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
	case 0:
		m.writeCTRL(val)
	case 1:
		m.writeCHR0(val)
		m.lastchr = addr
	case 2:
		m.writeCHR1(val)
		m.lastchr = addr
	case 3:
		m.writePRG(val)
	}
}

func (m *mmc1) writeCTRL(val uint8) {
	m.chrmode = (val & 0x10) >> 4
	m.prgmode = (val & 0x0C) >> 2

	prevNT := m.ntm
	m.ntm = val & 0x03
	if prevNT != m.ntm {
		switch m.ntm {
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

	const _forceWramOn = false // TODO: read from ROM header

	readonly := false
	if m.disableWRAM && !_forceWramOn {
		// no access
		readonly = true
	}

	totalram := m.rom.PRGRAMSize() + m.rom.PRGNVRAMSize()
	switch {
	case totalram > 0x4000:
		// SXROM with 32kB of save+work ram
		panic("MMC1 with 32kB of save+work ram not implemented")

	case totalram > 0x2000:

		// TODO: test persistency
		// SOROM, half of the 16kb ram is battery backed
		ram := m.PRGNVRAM
		if ((extrareg >> 3) & 0x01) != 0 {
			ram = m.PRGRAM
		}
		m.cpu.Bus.Unmap(0x6000, 0x7FFF)
		m.cpu.Bus.MapMemorySlice(0x6000, 0x7FFF, ram, readonly)

	case totalram == 0x2000:
	case totalram == 0:
		// Do not map any RAM

	default:
		panic("not supported")
	}

	switch m.prgmode {
	case 0, 1:
		// ignore low bit of bank number
		m.selectPRGPage32KB(int(m.prgbank & 0xFE))
	case 2:
		m.selectPRGPage16KB(0, 0)
		m.selectPRGPage16KB(1, int(m.prgbank))
	case 3:
		m.selectPRGPage16KB(0, int(m.prgbank))
		m.selectPRGPage16KB(1, -1)
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

func loadMMC1(b *base) error {
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
	return nil
}
