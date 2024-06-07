package emu

import (
	"path/filepath"
	"testing"

	"nestor/emu/hwio"
	"nestor/ines"
	"nestor/tests"
)

func checkedRead8(tb testing.TB, bus *hwio.Table, addr uint16, want uint8) {
	tb.Helper()

	if v := bus.Read8(addr); v != want {
		tb.Errorf("%s[0x%04X] should be 0x%02X, got 0x%02X", bus.Name, addr, want, v)
	}
}

func firstNonZero(b []byte) int {
	for i, v := range b {
		if v != 0 {
			return i
		}
	}
	return -1
}

func TestMapper000(t *testing.T) {
	// Check that mapper 000 correctly maps ROM to the hardware.

	nes := new(NES)
	romPath := filepath.Join(tests.RomsPath(t), "other", "nestest.nes")
	cartridge, err := ines.ReadRom(romPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := nes.PowerUp(cartridge); err != nil {
		t.Fatal(err)
	}

	/* CPU mapping */

	// Check that PRGROM is mapped to CPU memory.
	checkedRead8(t, nes.CPU.Bus, 0x8000, cartridge.PRGROM[0x0000])

	switch len(cartridge.PRGROM) {
	case 0x4000:
		// Check the 64 first bytes are mirrored.
		for i := uint16(0); i < 64; i++ {
			checkedRead8(t, nes.CPU.Bus, 0x8000+i, cartridge.PRGROM[0x0000+i])
		}
	case 0x8000:
		checkedRead8(t, nes.CPU.Bus, 0x8000, cartridge.PRGROM[0x4000])
	}

	/* PPU mapping */

	// Check that pattern tables is mapped to CHRROM.
	idx := firstNonZero(cartridge.CHRROM)
	if idx == -1 {
		panic("rom not adapted to this test")
	}
	checkedRead8(t, nes.PPU.Bus, uint16(idx), cartridge.CHRROM[idx])

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

	nes := new(NES)
	romPath := filepath.Join(tests.RomsPath(t), "other", "nestest.nes")
	cartridge, err := ines.ReadRom(romPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := nes.PowerUp(cartridge); err != nil {
		t.Fatal(err)
	}

	nes.CPU.Bus.Write8(0x2000, 0x27)
	if nes.PPU.PPUCTRL.Value != 0x27 {
		t.Errorf("PPUCTRL should be 0x27, got 0x%02X", nes.PPU.PPUCTRL.Value)
	}

	nes.CPU.Bus.Write8(0x3456, 0x18)
	if nes.PPU.PPUADDR.Value != 0x18 {
		t.Errorf("PPUADDR should be 0x18, got 0x%02X", nes.PPU.PPUADDR.Value)
	}

	nes.CPU.Bus.Write8(0x3FFF, 0xF5)
	if nes.PPU.PPUDATA.Value != 0xF5 {
		t.Errorf("PPUDATA should be 0xF5, got 0x%02X", nes.PPU.PPUDATA.Value)
	}
}
