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
	PRGRAM hwio.Mem `hwio:"offset=0x6000,size=0x2000"`

	PRGROM [0x8000]byte
	CHRROM [0x2000]byte

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

const KB = 1 << 10

// TODO: move to base with PRGROM

// select what 32KB PRG ROM bank to use.
func (m *mmc1) selectPRGPage32KB(bank int) {
	copy(m.PRGROM[:], m.rom.PRGROM[32*KB*(bank):])
}

// select what 16KB PRG ROM bank to use into which PRG 16KB page.
func (m *mmc1) selectPRGPage16KB(page uint32, bank int) {
	if bank < 0 {
		bank += len(m.rom.PRGROM) / (16 * KB)
	}
	copy(m.PRGROM[16*KB*page:], m.rom.PRGROM[16*KB*(bank):])
}

// select what 8KB PRG ROM bank to use.
func (m *mmc1) selectCHRPage8KB(bank int) {
	copy(m.CHRROM[:], m.rom.CHRROM[8*KB*(bank):])
}

// select what 4KB PRG ROM bank to use into which PRG 4KB page.
func (m *mmc1) selectCHRPage4KB(page uint32, bank int) {
	if bank < 0 {
		bank += len(m.rom.CHRROM) / (4 * KB)
	}
	copy(m.CHRROM[4*KB*page:], m.rom.CHRROM[4*KB*(bank):])
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
	default:
		panic("invalid PRG mode")
	}

	switch m.chrmode {
	case 0:
		m.selectCHRPage8KB(int(m.chrbank0))
	case 1:
		m.selectCHRPage4KB(0, int(m.chrbank0))
		m.selectCHRPage4KB(1, int(m.chrbank1))
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

	b.cpu.Bus.MapMem(0x8000, &hwio.Mem{
		Name:    "PRGROM",
		Data:    mmc1.PRGROM[:],
		VSize:   len(mmc1.PRGROM),
		Flags:   hwio.MemFlagReadWrite,
		WriteCb: mmc1.WritePRGROM,
	})

	// PPU mapping.
	mmc1.setNTMirroring(ines.OnlyAScreen)
	// Handle CHR RAM if CHRROM is empty.
	if len(b.rom.CHRROM) == 0 {
		b.rom.CHRROM = make([]byte, 0x2000) // 8 KB CHR RAM
	}

	b.ppu.Bus.MapMem(0x0000, &hwio.Mem{
		Name:  "CHRROM",
		Data:  mmc1.CHRROM[:],
		VSize: len(mmc1.CHRROM),
		Flags: hwio.MemFlag8ReadOnly,
	})

	// Mapper initialization.
	// On powerup: bits 2,3 of $8000 are set (this ensures the $8000 is bank 0, and
	// $C000 is the last bank - needed for SEROM/SHROM/SH1ROM which do no support
	// banking)
	mmc1.writeREG(0x8000, 0x0C)
	mmc1.writeREG(0xA000, 0)
	mmc1.writeREG(0xC000, 0)
	mmc1.writeREG(0xE000, 0) // TODO: WRAM Disable: enabled by default for MMC1B
	mmc1.disableWRAM = true  // TODO: always enabled on MMC1A
	mmc1.remap()
	return nil

	// TODO: load and map PRG-RAM if present in cartridge.
	// TODO: load and map CHR-RAM if present in cartridge.
}

func u8tob(v uint8) bool { return v != 0 }
