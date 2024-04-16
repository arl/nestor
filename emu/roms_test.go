package emu

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"nestor/emu/hwio"
	"nestor/hw"
	"nestor/ines"
)

func TestNestest(t *testing.T) {
	nes := new(NES)
	romPath := filepath.Join("..", "testdata", "nes-test-roms", "other", "nestest.nes")
	rom, err := ines.ReadRom(romPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := nes.PowerUp(rom); err != nil {
		t.Fatal(err)
	}

	flog, err := os.CreateTemp("", "nestor.nestest.*.log")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(flog.Name())
	t.Cleanup(func() {
		flog.Close()
		t.Logf("log saved to %s", flog.Name())
	})

	// This rom has an 'automation' mode. To enable it, PC must be set to C000.
	// We do that by overwriting the rom location of the reset vector.
	// binary.LittleEndian.PutUint16(cartridge.PRGROM[0x3FFC:], 0xC000)
	nes.CPU.PC = 0xC000
	nes.CPU.P = hw.P(0x24)
	nes.CPU.Clock = 7
	nes.PPU.Cycle = 21
	disasm := hw.NewDisasm(nes.CPU, flog)
	disasm.Run(26560)

	r1, r2 := nes.CPU.Read8(0x02), nes.CPU.Read8(0x03)
	if r1 != 0x00 || r2 != 0x00 {
		t.Fatalf("nestest failed: 0x%02x%02x", r1, r2)
	}
}

func TestInstructionsV5(t *testing.T) {
	dir := filepath.Join("..", "testdata", "nes-test-roms", "instr_test-v5", "rom_singles")
	files := []string{
		"01-basics.nes",
		"02-implied.nes",
		// "03-immediate.nes", uses unofficial  0xAB (LXA)
		"04-zero_page.nes",
		"05-zp_xy.nes",
		"06-absolute.nes",
		// "07-abs_xy.nes", uses unofficial 0x9C (SHY)
		"08-ind_x.nes",
		"09-ind_y.nes",
		"10-branches.nes",
		"11-stack.nes",
		"12-jmp_jsr.nes",
		"13-rts.nes",
		"14-rti.nes",
		"15-brk.nes",
		"16-special.nes",
	}

	log.SetOutput(io.Discard)
	for _, path := range files {
		t.Run(path, runTestRom(filepath.Join(dir, path)))
	}
}

func TestInterruptsV2(t *testing.T) {
	t.Skip("all failing for now")
	dir := filepath.Join("testdata", "nes-test-roms", "cpu_interrupts_v2", "rom_singles")
	files := []string{
		"1-cli_latency.nes", // APU should generate IRQ when $4017 = $00
		"2-nmi_and_brk.nes",
		"3-nmi_and_irq.nes",
		"4-irq_and_dma.nes",
		"5-branch_delays_irq.nes",
	}

	log.SetOutput(io.Discard)
	for _, path := range files {
		t.Run(path, runTestRom(filepath.Join(dir, path)))
	}
}

func runTestRom(path string) func(t *testing.T) {
	// All text output is written starting at $6004, with a zero-byte terminator
	// at the end. As more text is written, the terminator is moved forward, so
	// an emulator can print the current text at any time.

	// The test status is written to $6000. $80 means the test is running, $81
	// means the test needs the reset button pressed, but delayed by at least
	// 100 msec from now. $00-$7F means the test has completed and given that
	// result code.

	// To allow an emulator to know when one of these tests is running and the
	// data at $6000+ is valid, as opposed to some other NES program, $DE $B0
	// $G1 is written to $6001-$6003.
	return func(t *testing.T) {
		rom, err := ines.ReadRom(path)
		if err != nil {
			t.Fatal(err)
		}
		nes := new(NES)
		if err := nes.PowerUp(rom); err != nil {
			t.Fatal(err)
		}
		nes.Reset()

		magic := []byte{0xde, 0xb0, 0x61}
		magicset := 0
		var result uint8

		for {
			nes.RunOneFrame()
			data := nes.CPU.Bus.FetchPointer(0x6001)
			if magicset == 0 {
				if bytes.Equal(data[:3], magic) {
					magicset = 1
				}
				// Wait for the magic bytes to appear
				continue
			}

			// Once magic bytes have been written, they must not be overwritten.
			if !bytes.Equal(data[:3], magic) {
				t.Fatalf("corrupted memory")
			}
			result = nes.CPU.Read8(0x6000)
			if result <= 0x7F {
				break
			}
			if result == 0x80 {
				t.Log("test still running...")
			}
			if result == 0x81 {
				t.Log("needs reset button pressed in the next 100ms")
				panic("not implemented")
			}
		}
		if result != 0x00 {
			txt := memToString(nes.CPU.Bus, 0x6004)
			t.Fatalf("test failed:\ncode 0x%02x\ntext %s", result, txt)
		}
	}
}

func memToString(t *hwio.Table, addr uint16) string {
	data := t.FetchPointer(addr)
	i := 0
	for data[i] != 0 {
		i++
	}
	return unsafe.String(&data[0], i)
}

func TestNametableMirroring(t *testing.T) {
	romPath := filepath.Join("..", "testdata", "nes-test-roms", "other", "snow.nes")
	rom, err := ines.ReadRom(romPath)
	if err != nil {
		t.Fatal(err)
	}
	if rom.Mirroring() != ines.HorzMirroring {
		t.Errorf("incorrect nt mirroring")
	}
	nes := new(NES)
	if err := nes.PowerUp(rom); err != nil {
		t.Fatal(err)
	}
	nes.Reset()

	nes.PPU.Bus.Write8(0x2000, 'A')
	nes.PPU.Bus.Write8(0x2800, 'B')

	addrs := []uint16{
		0x2000, // A
		0x2400, // A
		0x2800, // B
		0x2C00, // B
		0x3000, // A
		0x3400, // A
		0x3800, // B
		0x3C00, // B
	}
	var nts []byte
	for _, a := range addrs {
		nts = append(nts, nes.PPU.Bus.Read8(a))
	}

	if string(nts) != "AABBAABB" {
		t.Errorf("mirrors = %s", nts)
	}
}
