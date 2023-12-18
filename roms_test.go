package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"nestor/emu/hwio"
	log "nestor/emu/logger"
	"nestor/hw"
	"nestor/ines"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestNestest(t *testing.T) {
	nes := new(NES)
	cartridge, err := ines.ReadRom("testdata/nes-test-roms/other/nestest.nes")

	// This rom has an 'automation' mode. To enable it, PC must be set to C000.
	// We do that by overwriting the rom location of the reset vector.
	binary.LittleEndian.PutUint16(cartridge.PRGROM[0x3FFC:], 0xC000)
	tcheck(t, err)
	tcheck(t, nes.PowerUp(cartridge))

	flog, err := os.CreateTemp("", "nestor.nestest.*.log")
	tcheck(t, err)
	t.Log(flog.Name())

	t.Cleanup(func() {
		flog.Close()
		t.Logf("log saved to %s", flog.Name())
		content, err := os.ReadFile(flog.Name())
		tcheck(t, err)

		want, err := os.ReadFile("testdata/nes-test-roms/other/nestest.log")
		tcheck(t, err)

		// TODO(arl) diff should be nil once APU is implemented.
		// With the APU not yet implemented, we have a few mismatched lines at
		// the end of the log. For now, we check this doesn't drift.
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(want), string(content), true)
		if dmp.DiffLevenshtein(diffs) != 8 {
			t.Fatalf("investigate why the Levenshtein disatnce distance has changed")
		}
	})

	nes.Hw.CPU.Reset()
	nes.Hw.CPU.P = hw.P(0x24)
	disasm := hw.NewDisasm(nes.Hw.CPU, flog, true)
	disasm.Run(26560)
}

func TestInstructionsV5(t *testing.T) {
	dir := filepath.Join("testdata", "nes-test-roms", "instr_test-v5", "rom_singles")
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
	t.Skip("need APU")
	dir := filepath.Join("testdata", "nes-test-roms", "cpu_interrupts_v2", "rom_singles")
	files := []string{
		"1-cli_latency.nes", // APU should generate IRQ when $4017 = $00
	}

	log.SetOutput(io.Discard)
	for _, path := range files {
		t.Run(path, runTestRom(filepath.Join(dir, path)))
	}
}

func runTestRom(path string) func(t *testing.T) {
	// All text output is written starting at $6004, with a zero-byte
	// terminator at the end. As more text is written, the terminator is moved
	// forward, so an emulator can print the current text at any time.

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

		var nes NES
		checkf(nes.PowerUp(rom), "error during power up")
		nes.Reset()

		magic := []byte{0xde, 0xb0, 0x61}
		var result uint8

		for {
			nes.RunOneFrame()
			data := nes.Hw.CPU.Bus.FetchPointer(0x6001)
			if !bytes.Equal(data[:3], magic) {
				t.Fatalf("corrupted memory")
			}
			result = nes.Hw.CPU.Read8(0x6000)
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
			txt := memToString(nes.Hw.CPU.Bus, 0x6004)
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
