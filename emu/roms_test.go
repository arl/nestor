package emu

import (
	"bytes"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/google/go-cmp/cmp"

	"nestor/emu/hwio"
	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
	"nestor/tests"
)

func TestNestest(t *testing.T) {
	log.SetOutput(io.Discard)

	romPath := filepath.Join(tests.RomsPath(t), "other", "nestest.nes")
	rom, err := ines.ReadRom(romPath)
	if err != nil {
		t.Fatal(err)
	}

	nes, err := powerUp(rom)
	if err != nil {
		t.Fatal(err)
	}

	flog, err := os.CreateTemp("", "nestor.nestest.*.log")
	if err != nil {
		t.Fatal(err)
	}

	// nestest.nes rom has an 'automation' mode. To enable it,
	// PC must be set to C000 (instead of C004 for graphic mode).
	nes.CPU.PC = 0xC000
	nes.CPU.SetTraceOutput(flog)

	// Configure a headless testing output.
	outcfg := TestingOutputConfig{
		Height: hw.OutputNTSC().Height,
		Width:  hw.OutputNTSC().Width,
	}
	nes.SetOutput(newTestingOutput(outcfg))
	nes.Run()

	result := nes.CPU.Read16(0x02)
	if result != 0 {
		t.Fatalf("nestest CPU tests failed with result 0x%04x (check nestest.txt)", result)
	}

	flog.Close()

	want := filepath.Join("testdata", "nestest.mesen.log")
	CompareFileWithGolden(t, flog.Name(), want, false)
	if t.Failed() {
		t.Log("execution trace saved to", flog.Name())
	}
}

type testdir struct {
	dir   string
	files []string
}

func tdir(dir string, items ...any) *testdir {
	td := testdir{
		dir: dir,
	}
	for _, item := range items {
		switch v := item.(type) {
		case string:
			td.files = append(td.files, filepath.Join(dir, v))
		case *testdir:
			for _, v := range v.files {
				td.files = append(td.files, filepath.Join(dir, v))
			}
		}
	}
	return &td
}

func (td *testdir) list(fn func(string)) {
	for _, f := range td.files {
		fn(f)
	}
}

func Test_testdir(t *testing.T) {
	var got []string
	tdir("a",
		tdir("b",
			"b1",
			"b2",
			"b3",
		),
		tdir("c",
			"c1",
			"c2",
			tdir("c2",
				"c2a",
				"c2b",
			),
		),
	).list(func(path string) { got = append(got, path) })

	want := []string{
		filepath.Join("a", "b", "b1"),
		filepath.Join("a", "b", "b2"),
		filepath.Join("a", "b", "b3"),
		filepath.Join("a", "c", "c1"),
		filepath.Join("a", "c", "c2"),
		filepath.Join("a", "c", "c2", "c2a"),
		filepath.Join("a", "c", "c2", "c2b"),
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBlarggRoms(t *testing.T) {
	// Various tests from blargg's test roms. They're easily to automate since they
	// write to a specific memory location to signal the test status.
	if !testing.Verbose() {
		log.SetOutput(io.Discard)
	}

	romsDir := filepath.Join(tests.RomsPath(t))

	tdir(romsDir,
		tdir("instr_test-v5",
			tdir("rom_singles",
				"01-basics.nes",
				"02-implied.nes",
				// "03-immediate.nes", // uses unofficial  0xAB (LXA)
				"04-zero_page.nes",
				"05-zp_xy.nes",
				"06-absolute.nes",
				// "07-abs_xy.nes",// uses unofficial 0x9C (SHY)
				"08-ind_x.nes",
				"09-ind_y.nes",
				"10-branches.nes",
				"11-stack.nes",
				"12-jmp_jsr.nes",
				"13-rts.nes",
				"14-rti.nes",
				"15-brk.nes",
				"16-special.nes",
			),
		),
		tdir("instr_misc",
			tdir("rom_singles",
				"01-abs_x_wrap.nes",
				"02-branch_wrap.nes",
				// "03-dummy_reads.nes",
				// "04-dummy_reads_apu.nes",
			),
		),
		tdir("cpu_dummy_writes",
			// "cpu_dummy_writes_ppumem.nes",
			"cpu_dummy_writes_oam.nes",
		),
		tdir("cpu_interrupts_v2",
			tdir("rom_singles"), // all failing for now
			//"1-cli_latency.nes", // APU should generate IRQ when $4017 = $00
			//"2-nmi_and_brk.nes",
			//"3-nmi_and_irq.nes",
			//"4-irq_and_dma.nes",
			//"5-branch_delays_irq.nes",

		),
		tdir("oam_read", "oam_read.nes"),
		// tdir("oam_stress", "oam_stress.nes"),
	).list(func(path string) {
		t.Run(filepath.Base(path), runBlarggTestRom(path))
	})
}

func runBlarggTestRom(path string) func(t *testing.T) {
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
		nes, err := powerUp(rom)
		if err != nil {
			t.Fatal(err)
		}

		magic := []byte{0xde, 0xb0, 0x61}
		magicset := 0
		var result uint8

		// Configure a headless testing output.
		out := newTestingOutput(TestingOutputConfig{
			Height: hw.OutputNTSC().Height,
			Width:  hw.OutputNTSC().Width,
		})

		for {
			vbuf := out.BeginFrame()
			nes.RunOneFrame(vbuf)
			out.EndFrame(vbuf)

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
	romPath := filepath.Join(tests.RomsPath(t), "other", "snow.nes")
	rom, err := ines.ReadRom(romPath)
	if err != nil {
		t.Fatal(err)
	}
	if rom.Mirroring() != ines.HorzMirroring {
		t.Errorf("incorrect nt mirroring")
	}
	nes, err := powerUp(rom)
	if err != nil {
		t.Fatal(err)
	}

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
		nts = append(nts, nes.PPU.Bus.Read8(a, false))
	}

	if string(nts) != "AABBAABB" {
		t.Errorf("mirrors = %s", nts)
	}
}

func TestBlarggPPUtests(t *testing.T) {
	const frameidx = 25

	outdir := filepath.Join("testdata", t.Name())
	os.Mkdir(outdir, 0755)

	roms := []string{
		"palette_ram.nes",
		"power_up_palette.nes",
		"vram_access.nes",
		"sprite_ram.nes",
		"vbl_clear_time.nes",
	}
	for _, romName := range roms {
		t.Run(romName, func(t *testing.T) {
			romPath := filepath.Join(tests.RomsPath(t), "blargg_ppu_tests_2005.09.15b", romName)
			runTestRomAndCompareFrame(t, romPath, outdir, romName, frameidx)
		})
	}
}

func TestTimingVBlankNMI(t *testing.T) {
	const frameidx = 200

	outdir := filepath.Join("testdata", t.Name())
	os.Mkdir(outdir, 0755)

	roms := []string{
		"1.frame_basics.nes", // onlt this passes for now
		// "2.vbl_timing.nes",
		// "3.even_odd_frames.nes",
		"4.vbl_clear_timing.nes",
		// "5.nmi_suppression.nes",
		"6.nmi_disable.nes",
		"7.nmi_timing.nes",
	}
	for _, romName := range roms {
		t.Run(romName, func(t *testing.T) {
			romPath := filepath.Join(tests.RomsPath(t), "vbl_nmi_timing", romName)
			runTestRomAndCompareFrame(t, romPath, outdir, romName, frameidx)
		})
	}
}

func runTestRomAndCompareFrame(t *testing.T, romPath, frameDir, framePath string, frame int64) {
	rom, err := ines.ReadRom(romPath)
	if err != nil {
		t.Fatal(err)
	}
	nes, err := powerUp(rom)
	if err != nil {
		t.Fatal(err)
	}

	filepath.SplitList(romPath)

	out := newTestingOutput(
		TestingOutputConfig{
			Height:        hw.OutputNTSC().Height,
			Width:         hw.OutputNTSC().Width,
			SaveFrameDir:  frameDir,
			SaveFrameFile: framePath,
			SaveFrameNum:  frame,
		})
	nes.SetOutput(out)
	nes.Run()

	out.CompareFrame(t)
}
