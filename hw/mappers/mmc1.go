package mappers

import (
	"nestor/hw/hwio"
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

	/* CPU */
	PRGRAM  hwio.Mem    `hwio:"offset=0x6000,size=0x2000"`
	PRGROM1 hwio.Device `hwio:"offset=0x8000,size=0x4000,rcb=ReadPRGROM,wcb=WritePRGROM"`
	PRGROM2 hwio.Device `hwio:"offset=0xC000,size=0x4000,rcb=ReadPRGROM,wcb=WritePRGROM"`

	// /* PPU */
	CHRROM1 hwio.Device `hwio:"bank=1,offset=0x0000,size=0x1000,rcb=ReadCHRROM,wcb=WriteCHRROM"`
	CHRROM2 hwio.Device `hwio:"bank=1,offset=0x1000,size=0x1000,rcb=ReadCHRROM,wcb=WriteCHRROM"`

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
	disableWRAM bool
	prgbank     uint32
}

type shiftReg uint8

func (sr shiftReg) push(val uint8) shiftReg {
	sr >>= 1
	sr |= shiftReg((val << 4) & 0x10)
	return sr
}

func (m *mmc1) ReadPRGROM(addr uint16) uint8 {
	var romaddr uint32
	switch m.prgmode {
	case 0, 1:
		addr &= 0x7fff
		bank := m.prgbank & 0b1110 // ignore low bit of bank number
		if bank >= uint32(len(m.rom.PRGROM)/0x8000) {
			bank = uint32(len(m.rom.PRGROM)/0x8000) - 1 // fallback to last bank if out of bounds
		}
		romaddr = uint32(addr) + 0x8000*(bank)
	case 2:
		if addr >= 0xC000 {
			addr &= 0x3fff
			bank := m.prgbank
			if bank >= uint32(len(m.rom.PRGROM)/0x4000) {
				bank = uint32(len(m.rom.PRGROM)/0x4000) - 1 // fallback to last bank if out of bounds
			}
			romaddr = uint32(addr) + 0x4000*(bank)
		} else {
			addr &= 0x3fff
			bank := uint32(0)
			romaddr = uint32(addr) + bank*0x4000
		}
	case 3:
		if addr >= 0xC000 {
			addr &= 0x3fff
			bank := len(m.rom.PRGROM)/0x4000 - 1
			romaddr = uint32(addr) + uint32(bank)*0x4000
		} else {
			addr &= 0x3fff
			bank := m.prgbank
			if bank >= uint32(len(m.rom.PRGROM)/0x4000) {
				bank = uint32(len(m.rom.PRGROM)/0x4000) - 1 // fallback to last bank if out of bounds
			}
			romaddr = uint32(addr) + 0x4000*(bank)
		}
	default:
		panic("invalid PRG mode")
	}

	modMapper.DebugZ("read PRGROM").String("mapper", m.desc.Name).
		Hex8("mode", m.prgmode).
		Hex16("bank", uint16(m.prgbank)).
		Hex16("addr", addr).
		Hex32("romaddr", romaddr).
		End()
	return m.rom.PRGROM[romaddr]
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
		} else {
			m.serial = m.serial.push(val)
			m.counter++
			if m.counter == 5 {
				m.writeReg(addr, uint8(m.serial))
				m.serial = 0
				m.counter = 0
			}
		}
	}
	m.prevCycle = m.cpu.CurrentCycle()
}

func (m *mmc1) writeReg(addr uint16, val uint8) {
	switch (addr & 0x6000) >> 13 {
	case 0:
		m.writeCTRL(val)
	case 1:
		m.writeCHR0(val)
	case 2:
		m.writeCHR1(val)
	case 3:
		m.writePRG(val)
	default:
		panic("invalid reg write")
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

func (m *mmc1) ReadCHRROM(addr uint16) uint8 {
	romaddr := m.chrromAddr(addr)
	modMapper.DebugZ("read CHR").String("mapper", m.desc.Name).
		Hex8("mode", m.chrmode).
		Hex16("bank0", uint16(m.chrbank0)).
		Hex16("bank1", uint16(m.chrbank1)).
		Hex16("addr", addr).
		Hex32("romaddr", romaddr).
		Hex16("CHRROM size", uint16(len(m.rom.CHRROM))).
		End()
	return m.rom.CHRROM[romaddr]
}

func (m *mmc1) WriteCHRROM(addr uint16, val uint8) {
	romaddr := m.chrromAddr(addr)
	modMapper.DebugZ("read CHR").String("mapper", m.desc.Name).
		Hex8("mode", m.chrmode).
		Hex16("bank0", uint16(m.chrbank0)).
		Hex16("bank1", uint16(m.chrbank1)).
		Hex16("addr", addr).
		Hex32("romaddr", romaddr).
		Hex16("CHRROM size", uint16(len(m.rom.CHRROM))).
		End()

	// TODO: handle disableWRAM
	m.rom.CHRROM[romaddr] = val
}

func (m *mmc1) chrromAddr(addr uint16) uint32 {
	addr &= 0x1fff

	//(0: switch 8 KB at a time; 1: switch two separate 4 KB banks)
	switch m.chrmode {
	case 0:
		nbanks := uint32(len(m.rom.CHRROM) / 0x2000)
		bank := m.chrbank0
		if bank >= nbanks {
			bank = 0 // fallback to first bank if out of bounds
		}
		return uint32(addr) + 0x2000*(bank)
	case 1:
		nbanks := uint32(len(m.rom.CHRROM) / 0x1000)

		var bank uint32
		if addr < 0x1000 {
			bank = m.chrbank0
		} else {
			bank = m.chrbank1
			addr -= 0x1000
		}
		if bank >= nbanks {
			bank = 0 // fallback to first bank if out of bounds
		}
		return uint32(addr) + 0x1000*(bank)
	default:
		panic("invalid CHR mode")
	}
}

func loadMMC1(b *base) error {
	mmc1 := &mmc1{base: b}
	hwio.MustInitRegs(mmc1)

	// CPU mapping.
	b.cpu.Bus.MapBank(0x0000, mmc1, 0)

	if b.rom.PRGRAMSize() > 0 {
		// panic("PRGRAM not implemented")
	}

	// Handle CHR RAM if CHRROM is empty.
	if len(b.rom.CHRROM) == 0 {
		b.rom.CHRROM = make([]byte, 0x2000) // 8 KB CHR RAM
	}

	// PPU mapping.
	b.ppu.Bus.MapBank(0x0000, mmc1, 1)
	mmc1.setNTMirroring(ines.OnlyAScreen)

	// Mapper initialization.
	// On powerup: bits 2,3 of $8000 are set (this ensures the $8000 is bank 0, and
	// $C000 is the last bank - needed for SEROM/SHROM/SH1ROM which do no support
	// banking)
	mmc1.writeReg(0x8000, 0x0C)
	mmc1.writeReg(0xA000, 0)
	mmc1.writeReg(0xC000, 0)
	// TODO: WRAM Disable: enabled by default for MMC1B
	mmc1.writeReg(0xE000, 0)

	// TODO: always enabled on MMC1A
	mmc1.disableWRAM = true
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}

func u8tob(v uint8) bool { return v != 0 }
