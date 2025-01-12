package emu

import (
	"bytes"
	"path/filepath"
	"testing"

	"nestor/hw/hwio"
	"nestor/ines"
	"nestor/tests"
)

func checkedRead8(tb testing.TB, bus *hwio.Table, addr uint16, want uint8) {
	tb.Helper()

	if v := bus.Read8(addr); v != want {
		tb.Errorf("%s[0x%04X] should be 0x%02X, got 0x%02X", bus.Name, addr, want, v)
	}
}

func TestMapper000(t *testing.T) {
	// Check that mapper 000 correctly maps ROM to the hardware.
	romPath := filepath.Join(tests.RomsPath(t), "other", "nestest.nes")
	rom, err := ines.ReadRom(romPath)
	if err != nil {
		t.Fatal(err)
	}
	nes, err := powerUp(rom)
	if err != nil {
		t.Fatal(err)
	}

	/* CPU mapping */

	// Check that PRGROM is mapped to CPU memory.
	checkedRead8(t, nes.CPU.Bus, 0x8000, rom.PRGROM[0x0000])

	switch len(rom.PRGROM) {
	case 0x4000:
		// Check the 64 first bytes are mirrored.
		for i := uint16(0); i < 64; i++ {
			checkedRead8(t, nes.CPU.Bus, 0x8000+i, rom.PRGROM[0x0000+i])
		}
	case 0x8000:
		checkedRead8(t, nes.CPU.Bus, 0x8000, rom.PRGROM[0x4000])
	}

	/* PPU mapping */

	// Check that pattern tables is mapped to CHRROM.
	idx := bytes.IndexFunc(rom.CHRROM, func(r rune) bool { return r != 0 })
	if idx == -1 {
		panic("rom not adapted to this test")
	}
	checkedRead8(t, nes.PPU.Bus, uint16(idx), rom.CHRROM[idx])

	// Write into nametable, read from mirror.
	nes.PPU.Bus.Write8(0x2EFF, 0x27)
	checkedRead8(t, nes.PPU.Bus, 0x2EFF, 0x27)

	// Write into palette ram, read from mirror.
	nes.PPU.Bus.Write8(0x3F1F, 0x23)
	checkedRead8(t, nes.PPU.Bus, 0x3F1F, 0x23)
}

func TestPPURegisterMapping(t *testing.T) {
	// Check that PPU registers are correctly mapped
	// and mirrored into the CPU memory space.

	romPath := filepath.Join(tests.RomsPath(t), "other", "nestest.nes")
	rom, err := ines.ReadRom(romPath)
	if err != nil {
		t.Fatal(err)
	}
	nes, err := powerUp(rom)
	if err != nil {
		t.Fatal(err)
	}

	nes.CPU.Bus.Write8(0x2000, 0x27)
	if nes.PPU.PPUCTRL != 0x27 {
		t.Errorf("PPUCTRL should be 0x27, got 0x%02X", nes.PPU.PPUCTRL)
	}

	nes.CPU.Bus.Write8(0x3001, 0x18)
	if nes.PPU.PPUMASK != 0x18 {
		t.Errorf("PPUMASK should be 0x18, got 0x%02X", nes.PPU.PPUMASK)
	}

	nes.CPU.Bus.Write8(0x3F02, 0xF5) // PPUSTATUS is readonly
	if nes.PPU.PPUSTATUS != 0x00 {
		t.Errorf("PPUSTATUS should be 0x00, got 0x%02X", nes.PPU.PPUSTATUS)
	}
}
