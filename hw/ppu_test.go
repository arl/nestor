package hw

import (
	"testing"
)

func TestPPUScroll(t *testing.T) {
	ppu := NewPPU()
	ppu.InitBus()
	cpu := NewCPU(ppu)
	cpu.InitBus()
	ppu.CPU = cpu
	ppu.CreateScreen()

	ppu.vramTmp = 0xffff

	// Write to PPUCTRL
	cpu.Write8(0x2000, 0)
	if got := ppu.vramTmp.nametable(); got != 0b00 {
		t.Errorf("t.nametable = 0b%08b, want 0b00", got)
	}

	// Read from PPUSTATUS
	_ = cpu.Read8(0x2002)
	if ppu.writeLatch {
		t.Errorf("writeLatch = %t, want false", ppu.writeLatch)
	}

	// First write to PPUSCROLL
	cpu.Write8(0x2005, 0b01111_101)
	if got := ppu.vramTmp.coarsex(); got != 0b01111 {
		t.Errorf("t.coarsex = 0b%08b, want 0b01111", got)
	}
	if ppu.bg.finex != 0b101 {
		t.Errorf("finex = 0b%08b, want 0b101", ppu.bg.finex)
	}
	if !ppu.writeLatch {
		t.Errorf("writeLatch = %t, want true", ppu.writeLatch)
	}

	// Second write to PPUSCROLL
	cpu.Write8(0x2005, 0b01_011_110)
	if got := ppu.vramTmp.coarsey(); got != 0b01011 {
		t.Errorf("t.coarsey = 0b%08b, want 0b01011", got)
	}
	if got := ppu.vramTmp.finey(); got != 0b110 {
		t.Errorf("t.finey = 0b%08b, want 0b110", got)
	}
	if ppu.writeLatch {
		t.Errorf("writeLatch = %t, want false", ppu.writeLatch)
	}

	// First write to PPUADDR
	cpu.Write8(0x2006, 0b00_111101)
	if got := ppu.vramTmp.high(); got != 0b111101 {
		t.Errorf("t.high = %08b, want 0b111101", got)
	}
	// Bit 14 (15th bit) of t gets set to zero
	if ppu.vramTmp.val() != 0b0111101_01101111 {
		t.Errorf("t.val = %015b, want 0b0111101_01101111", ppu.vramTmp.val())
	}

	// Second write to PPUADDR
	cpu.Write8(0x2006, 0b11110000)
	if got := ppu.vramTmp.low(); got != 0b11110000 {
		t.Errorf("t.low = %08b, want 0b11110000", got)
	}
	if ppu.vramTmp.val() != 0b0111101_11110000 {
		t.Errorf("t.val = %015b, want 0b0111101_11110000", ppu.vramTmp.val())
	}
	// After t is updated, contents of t copied into v
	if ppu.vramTmp.val() != ppu.vramAddr.val() {
		t.Errorf("v != t")
	}
}
