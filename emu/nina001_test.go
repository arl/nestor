package emu

import (
	"encoding/binary"
	"os"
	"testing"

	"nestor/ines"
)

func TestNINA001_BankRegistersInPRGRAMSwitchPRG(t *testing.T) {
	romPath := writeTempROM(t, buildMapper34NINA001ROM(t))

	rom, err := ines.ReadROM(romPath)
	if err != nil {
		t.Fatalf("ReadROM(%q) failed: %v", romPath, err)
	}
	nes, err := powerUp(rom)
	if err != nil {
		t.Fatalf("powerUp failed: %v", err)
	}

	// Prevent false positives if the program never runs.
	nes.CPU.Bus.Write8(0x6000, 0xFF)

	out := NewOutput(nil, OutputConfig{
		Height: NTSCHeight,
		Width:  NTSCWidth,
	})

	// Program should quickly:
	// - write $7FFF to switch to PRG bank 2
	// - read $9000 (bank-dependent) into $6004
	// - write $6000 = 0 to signal completion
	const maxFrames = 10
	for range maxFrames {
		f := out.BeginFrame()
		nes.RunOneFrame(&f)
		out.EndFrame(&f)
		if nes.CPU.Bus.Peek8(0x6000) == 0x00 {
			break
		}
	}

	if got := nes.CPU.Bus.Peek8(0x6000); got != 0x00 {
		t.Fatalf("program did not complete: $6000=%02x", got)
	}
	// If bank switching works, $9000 comes from PRG bank 2 and the program stores it at $6004.
	if got := nes.CPU.Bus.Peek8(0x6004); got != 0x02 {
		t.Fatalf("bank switch via NINA-001 PRG-RAM register failed: $6004=%02x, want 02", got)
	}
}

func writeTempROM(t *testing.T, data []byte) string {
	t.Helper()
	f, err := os.CreateTemp("", "nestor.nina001.*.nes")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		t.Fatalf("Write failed: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	return f.Name()
}

func buildMapper34NINA001ROM(t *testing.T) []byte {
	t.Helper()

	const (
		mapper    = 34
		submapper = 1 // NINA-001

		prgBanks32KB = 4  // enough to switch to bank 2
		chrSize8KB   = 4  // 32KB CHR-ROM
		prgRAMExp    = 7  // 8KB PRG-RAM (NES2.0 exponent, 128 * 2^(exp-1))
		bankSize     = 32 * 1024
	)

	// NES 2.0 header (16 bytes).
	hdr := make([]byte, 16)
	copy(hdr[:4], []byte(ines.Magic))
	hdr[4] = uint8(prgBanks32KB * 2) // PRG size in 16KB units
	hdr[5] = uint8(chrSize8KB)       // CHR size in 8KB units
	hdr[6] = uint8((mapper & 0x0F) << 4)
	hdr[7] = uint8(mapper&0xF0) | 0x08 // NES2.0 marker
	hdr[8] = uint8((submapper << 4) | uint8((mapper>>8)&0x0F))
	hdr[10] = prgRAMExp // PRG-RAM size (volatile)

	prg := make([]byte, prgBanks32KB*bankSize)
	for bank := 0; bank < prgBanks32KB; bank++ {
		fill(prg[bank*bankSize:(bank+1)*bankSize], 0xEA) // NOP
		installNINA001TestProgram(prg[bank*bankSize:(bank+1)*bankSize])
		// Bank-dependent value at $9000.
		prg[bank*bankSize+0x1000] = byte(bank)
	}

	chr := make([]byte, chrSize8KB*0x2000)
	fill(chr, 0)

	return append(append(hdr, prg...), chr...)
}

func installNINA001TestProgram(bank []byte) {
	// Simple program at $8000 (bank offset 0x0000) that switches to bank 2 and
	// stores a bank-dependent byte at $9000 into $6004, then signals completion.
	//
	// The program must be identical in every 32KB bank because NINA-001 bank
	// switching replaces the entire $8000-$FFFF mapping while code is executing.
	prog := []byte{
		0x78,       // SEI
		0xD8,       // CLD
		0xA9, 0x80, // LDA #$80
		0x8D, 0x00, 0x60, // STA $6000 (running)
		0xA9, 0x02, // LDA #$02 (select PRG bank 2)
		0x8D, 0xFF, 0x7F, // STA $7FFF (NINA-001 PRG bank register in PRG-RAM space)
		0xAD, 0x00, 0x90, // LDA $9000 (bank-dependent value)
		0x8D, 0x04, 0x60, // STA $6004
		0xA9, 0x00, // LDA #$00
		0x8D, 0x00, 0x60, // STA $6000 (done)
		0x4C, 0x17, 0x80, // JMP $8017 (spin)
	}
	copy(bank[0x0000:], prog)

	// Vectors (NMI, RESET, IRQ/BRK) at $FFFA-$FFFF (bank offsets 0x7FFA..0x7FFF).
	binary.LittleEndian.PutUint16(bank[0x7FFA:], 0x8000)
	binary.LittleEndian.PutUint16(bank[0x7FFC:], 0x8000)
	binary.LittleEndian.PutUint16(bank[0x7FFE:], 0x8000)
}

func fill(b []byte, v byte) {
	for i := range b {
		b[i] = v
	}
}

