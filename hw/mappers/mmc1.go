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
	case 2:
		m.writeCHR1(val)
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
	modMapper.DebugZ("Write CHR0 reg").String("mapper", m.desc.Name).Uint8("val", val).End()
	m.chrbank0 = uint32(val & 0b11111) // TODO: Adjust mask if CHRROM is larger
}

func (m *mmc1) writeCHR1(val uint8) {
	modMapper.DebugZ("Write CHR1 reg").String("mapper", m.desc.Name).Uint8("val", val).End()
	m.chrbank1 = uint32(val & 0b11111) // TODO: Adjust mask if CHRROM is larger
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
		m.selectCHRROMPage8KB(int(m.chrbank0))
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
	mmc1.remap()
	return nil
}
