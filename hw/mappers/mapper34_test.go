package mappers

import (
	"testing"

	"nestor/hw"
	"nestor/ines"
)

func TestMapper34_NINA001_PRGRAMRegistersSwitchBanks(t *testing.T) {
	rom := mustDecodeTestROM(t, testROMSpec{
		mapper:    34,
		submapper: 1, // NINA-001
		prg32KB:   4, // banks 0..3
		chr4KB:    16,
		prgRAM8KB: true, // ensure $6000-$7FFF is mapped
	})

	ppu := hw.NewPPU()
	cpu := hw.NewCPU(ppu)
	cpu.InitBus()

	m, err := Load(rom, cpu, ppu)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	m34, ok := m.(*mapper34)
	if !ok {
		t.Fatalf("expected *mapper34, got %T", m)
	}
	if !m34.isNINA {
		t.Fatalf("expected NINA mode")
	}

	// Initial mapping.
	if got := m34.base.PRGROM[0]; got != 0 {
		t.Fatalf("initial PRG mapping mismatch: got %d, want 0", got)
	}
	if got := m34.base.CHRROM[0]; got != 0 {
		t.Fatalf("initial CHR0 mapping mismatch: got %d, want 0", got)
	}
	if got := m34.base.CHRROM[0x1000]; got != 1 {
		t.Fatalf("initial CHR1 mapping mismatch: got %d, want 1", got)
	}

	// NINA-001 bank selection registers are in PRG-RAM space at $7FFD-$7FFF.
	// These writes must reach the mapper even though they target $6000-$7FFF.
	cpu.Bus.Write8(0x7FFF, 0x02) // PRG 32KB bank = 2
	if got := m34.base.PRGROM[0]; got != 2 {
		t.Fatalf("PRG bank switch via $7FFF failed: got %d, want 2", got)
	}

	cpu.Bus.Write8(0x7FFD, 0x03) // CHR bank for $0000-$0FFF = 3
	cpu.Bus.Write8(0x7FFE, 0x05) // CHR bank for $1000-$1FFF = 5
	if got := m34.base.CHRROM[0]; got != 3 {
		t.Fatalf("CHR0 bank switch via $7FFD failed: got %d, want 3", got)
	}
	if got := m34.base.CHRROM[0x1000]; got != 5 {
		t.Fatalf("CHR1 bank switch via $7FFE failed: got %d, want 5", got)
	}
}

type testROMSpec struct {
	mapper    uint16
	submapper uint8
	prg32KB   int
	chr4KB    int
	prgRAM8KB bool
}

func mustDecodeTestROM(t *testing.T, s testROMSpec) *ines.Rom {
	t.Helper()

	if s.prg32KB <= 0 {
		t.Fatalf("invalid prg32KB=%d", s.prg32KB)
	}
	if s.chr4KB <= 0 || s.chr4KB%2 != 0 {
		t.Fatalf("invalid chr4KB=%d (must be even, CHR header counts 8KB units)", s.chr4KB)
	}

	// NES 2.0 header (16 bytes).
	hdr := make([]byte, 16)
	copy(hdr[:4], []byte(ines.Magic))
	hdr[4] = uint8(s.prg32KB * 2)     // PRG size in 16KB units
	hdr[5] = uint8(s.chr4KB / 2)      // CHR size in 8KB units
	hdr[6] = uint8((s.mapper & 0x0F) << 4) // mapper low nibble
	hdr[7] = uint8(s.mapper&0xF0) | 0x08   // mapper high nibble + NES2.0 marker
	hdr[8] = uint8((s.submapper << 4) | uint8((s.mapper>>8)&0x0F))

	// PRG-RAM size encoding (NES2.0): low nibble is exponent for size in 128B units.
	// We want 8KB when requested: 128 * 2^(7-1) = 8192.
	if s.prgRAM8KB {
		hdr[10] = 0x07
	}

	prg := make([]byte, s.prg32KB*0x8000)
	for bank := 0; bank < s.prg32KB; bank++ {
		fill(prg[bank*0x8000:(bank+1)*0x8000], byte(bank))
	}

	chr := make([]byte, s.chr4KB*0x1000)
	for bank := 0; bank < s.chr4KB; bank++ {
		fill(chr[bank*0x1000:(bank+1)*0x1000], byte(bank))
	}

	buf := append(append(hdr, prg...), chr...)
	rom, err := ines.Decode(buf)
	if err != nil {
		t.Fatalf("ines.Decode() failed: %v", err)
	}
	return rom
}

func fill(b []byte, v byte) {
	for i := range b {
		b[i] = v
	}
}

