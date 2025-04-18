package emu

import (
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/sync/errgroup"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/hw/hwio"
	"nestor/ines"
	"nestor/tests"
)

func TestNestest(t *testing.T) {
	if !testing.Verbose() {
		log.Disable()
	}

	romPath := filepath.Join(tests.RomsPath(t), "other", "nestest.nes")
	rom, err := ines.ReadRom(romPath)
	if err != nil {
		t.Fatal(err)
	}

	flog, err := os.CreateTemp("", "nestor.nestest.*.log")
	if err != nil {
		t.Fatal(err)
	}
	println("nestest log:", flog.Name())

	nes, err := powerUp(rom)
	if err != nil {
		t.Fatal(err)
	}

	// nestest.nes rom has an 'automation' mode. To enable it,
	// PC must be set to C000 (instead of C004 for graphic mode).
	nes.CPU.PC = 0xC000
	nes.CPU.SetTraceOutput(flog)

	// TODO: remove once openbus is implemented
	nes.APU.Square1.Duty.Value = 0x40
	nes.APU.Square1.Sweep.Value = 0x40
	nes.APU.Square1.Timer.Value = 0x40
	nes.APU.Square1.Length.Value = 0x40

	nes.APU.Square2.Duty.Value = 0x40
	nes.APU.Square2.Sweep.Value = 0x40
	nes.APU.Square2.Timer.Value = 0x40
	nes.APU.Square2.Length.Value = 0x40

	// Configure a headless testing output.
	cfg := TestingOutputConfig{
		Height: hw.NTSCHeight,
		Width:  hw.NTSCWidth,
	}
	e := Emulator{
		NES: nes,
		out: newTestingOutput(cfg),
	}
	e.Run()

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

func TestBlarggRoms(t *testing.T) {
	// Various tests from blargg's test roms. They're easily to automate since they
	// write to a specific memory location to signal the test status.
	if !testing.Verbose() {
		log.Disable()
	}

	romsDir := filepath.Join(tests.RomsPath(t))

	tests := []string{
		"apu_reset/4015_cleared.nes",
		"apu_reset/irq_flag_cleared.nes",
		"apu_reset/len_ctrs_enabled.nes",
		"apu_reset/4017_timing.nes",
		"apu_reset/4017_written.nes",
		"apu_reset/works_immediately.nes",

		"apu_test/rom_singles/1-len_ctr.nes",
		"apu_test/rom_singles/2-len_table.nes",
		"apu_test/rom_singles/3-irq_flag.nes",
		"apu_test/rom_singles/4-jitter.nes",
		"apu_test/rom_singles/5-len_timing.nes",
		"apu_test/rom_singles/6-irq_flag_timing.nes",
		"apu_test/rom_singles/7-dmc_basics.nes",
		"apu_test/rom_singles/8-dmc_rates.nes",

		"cpu_dummy_writes/cpu_dummy_writes_oam.nes",
		"cpu_dummy_writes/cpu_dummy_writes_ppumem.nes",

		"cpu_interrupts_v2/rom_singles/1-cli_latency.nes",
		"cpu_interrupts_v2/rom_singles/2-nmi_and_brk.nes",
		"cpu_interrupts_v2/rom_singles/3-nmi_and_irq.nes",
		"cpu_interrupts_v2/rom_singles/4-irq_and_dma.nes",
		"cpu_interrupts_v2/rom_singles/5-branch_delays_irq.nes",

		"cpu_exec_space/test_cpu_exec_space_ppuio.nes",

		"cpu_reset/ram_after_reset.nes",
		"cpu_reset/registers.nes",

		"instr_misc/rom_singles/01-abs_x_wrap.nes",
		"instr_misc/rom_singles/02-branch_wrap.nes",
		"instr_misc/rom_singles/03-dummy_reads.nes",
		"instr_misc/rom_singles/04-dummy_reads_apu.nes",

		"instr_test-v5/rom_singles/01-basics.nes",
		"instr_test-v5/rom_singles/02-implied.nes",
		"instr_test-v5/rom_singles/03-immediate.nes",
		"instr_test-v5/rom_singles/04-zero_page.nes",
		"instr_test-v5/rom_singles/05-zp_xy.nes",
		"instr_test-v5/rom_singles/06-absolute.nes",
		"instr_test-v5/rom_singles/07-abs_xy.nes",
		"instr_test-v5/rom_singles/08-ind_x.nes",
		"instr_test-v5/rom_singles/09-ind_y.nes",
		"instr_test-v5/rom_singles/10-branches.nes",
		"instr_test-v5/rom_singles/11-stack.nes",
		"instr_test-v5/rom_singles/12-jmp_jsr.nes",
		"instr_test-v5/rom_singles/13-rts.nes",
		"instr_test-v5/rom_singles/14-rti.nes",
		"instr_test-v5/rom_singles/15-brk.nes",
		"instr_test-v5/rom_singles/16-special.nes",

		"instr_timing/rom_singles/1-instr_timing.nes",
		"instr_timing/rom_singles/2-branch_timing.nes",

		"oam_read/oam_read.nes",

		"oam_stress/oam_stress.nes",

		"ppu_open_bus/ppu_open_bus.nes",

		"sprdma_and_dmc_dma/sprdma_and_dmc_dma.nes",
		"sprdma_and_dmc_dma/sprdma_and_dmc_dma_512.nes",
	}

	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())

	for _, romName := range tests {
		// Ensure tests run on all platforms.
		romPath := filepath.Join(romsDir, filepath.FromSlash(romName))
		g.Go(func() error {
			t.Run(romName, runBlarggTestRom(romPath))
			return nil
		})
	}
	g.Wait()
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

		magic := "\xde\xb0\x61"
		magicset := 0
		var result uint8

		// Configure a headless testing output.
		out := newTestingOutput(TestingOutputConfig{
			Height: hw.NTSCHeight,
			Width:  hw.NTSCWidth,
		})

		// When reset is required, it needs to be pressed 100ms later, so we
		// start a frame counter to keep track of the time.
		framesBeforeReset := -1 // not requested

		for {
			vbuf := out.BeginFrame()
			nes.RunOneFrame(vbuf)
			out.EndFrame(vbuf)

			data := readString(nes.CPU.Bus, 0x6001, 3)
			if magicset == 0 {
				if data == magic {
					magicset = 1
				}
				// Wait for the magic bytes to appear
				continue
			}

			// Once magic bytes have been written, they must not be overwritten.
			if data != magic {
				t.Fatalf("corrupted memory")
			}
			result = nes.CPU.Bus.Peek8(0x6000)
			if result <= 0x7F {
				break
			}
			if result == 0x80 {
				t.Logf("%s still running...", t.Name())
			}

			// Handle reset request.
			switch {
			case framesBeforeReset == 0:
				t.Log("pressing RESET...")
				nes.Reset(true)
				framesBeforeReset = -1
			case framesBeforeReset > 0:
				framesBeforeReset--
			case result == 0x81:
				framesBeforeReset = 30 // in 20 frames >= 100ms
			}
		}
		if result != 0x00 {
			txt := readString(nes.CPU.Bus, 0x6004, -1)
			t.Fatalf("test failed:\ncode 0x%02x\ntext %s", result, txt)
		}
	}
}

func readString(t *hwio.Table, addr uint16, maxlen int) string {
	var sb strings.Builder
	if maxlen == -1 {
		maxlen = math.MaxUint16
	}
	for i := 0; ; i++ {
		val := t.Peek8(addr + uint16(i))
		if val == 0 || i >= maxlen {
			break
		}
		sb.WriteByte(val)
	}

	return sb.String()
}

func TestSprite0Hit(t *testing.T) {
	if !testing.Verbose() {
		log.Disable()
	}

	outdir := filepath.Join("testdata", t.Name())
	os.Mkdir(outdir, 0755)

	roms := []string{
		"01.basics.nes",
		"02.alignment.nes",
		"03.corners.nes",
		"04.flip.nes",
		"05.left_clip.nes",
		"06.right_edge.nes",
		"07.screen_bottom.nes",
		"08.double_height.nes",
		"09.timing_basics.nes",
		"10.timing_order.nes",
		"11.edge_timing.nes",
	}

	const frameidx = 70
	for _, romName := range roms {
		t.Run(romName, func(t *testing.T) {
			romPath := filepath.Join(tests.RomsPath(t), "sprite_hit_tests_2005.10.05", romName)
			runAndCompareFrame(t, romPath, outdir, romName, frameidx)
		})
	}
}

func TestSpriteOverflow(t *testing.T) {
	if !testing.Verbose() {
		log.Disable()
	}

	outdir := filepath.Join("testdata", t.Name())
	os.Mkdir(outdir, 0755)

	roms := []string{
		"1.Basics.nes",
		// "2.Details.nes", // failed #9
		// "3.Timing.nes",
		// "4.Obscure.nes",
		"5.Emulator.nes",
	}

	const frameidx = 25
	for _, romName := range roms {
		t.Run(romName, func(t *testing.T) {
			romPath := filepath.Join(tests.RomsPath(t), "sprite_overflow_tests", romName)
			runAndCompareFrame(t, romPath, outdir, romName, frameidx)
		})
	}
}

func TestDMCDMADuringRead(t *testing.T) {
	if !testing.Verbose() {
		log.Disable()
	}

	outdir := filepath.Join("testdata", t.Name())
	os.Mkdir(outdir, 0755)

	roms := []string{
		"dma_2007_read.nes",
		"dma_2007_write.nes",
		"dma_4016_read.nes",
		// "double_2007_read.nes",
		"read_write_2007.nes",
	}

	const frameidx = 400
	for _, romName := range roms {
		t.Run(romName, func(t *testing.T) {
			romPath := filepath.Join(tests.RomsPath(t), "dmc_dma_during_read4", romName)
			runAndCompareFrame(t, romPath, outdir, romName, frameidx)
		})
	}
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
		nts = append(nts, nes.PPU.Bus.Read8(a))
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
			runAndCompareFrame(t, romPath, outdir, romName, frameidx)
		})
	}
}

func TestBlarggAPUtests(t *testing.T) {
	const frameidx = 25

	outdir := filepath.Join("testdata", t.Name())
	os.Mkdir(outdir, 0755)

	roms := []string{
		"01.len_ctr.nes",
		"02.len_table.nes",
		"03.irq_flag.nes",
		"04.clock_jitter.nes",
		"05.len_timing_mode0.nes",
		"06.len_timing_mode1.nes",
		"07.irq_flag_timing.nes",
		"08.irq_timing.nes",
		"09.reset_timing.nes",
		"10.len_halt_timing.nes",
		"11.len_reload_timing.nes",
	}
	for _, romName := range roms {
		t.Run(romName, func(t *testing.T) {
			romPath := filepath.Join(tests.RomsPath(t), "blargg_apu_2005.07.30", romName)
			runAndCompareFrame(t, romPath, outdir, romName, frameidx)
		})
	}
}

func TestTimingVBlankNMI(t *testing.T) {
	if !testing.Verbose() {
		log.Disable()
	}

	const frameidx = 200

	outdir := filepath.Join("testdata", t.Name())
	os.Mkdir(outdir, 0755)

	roms := []string{
		"1.frame_basics.nes",
		"2.vbl_timing.nes",
		"3.even_odd_frames.nes",
		"4.vbl_clear_timing.nes",
		"5.nmi_suppression.nes",
		"6.nmi_disable.nes",
		"7.nmi_timing.nes",
	}
	for _, romName := range roms {
		t.Run(romName, func(t *testing.T) {
			romPath := filepath.Join(tests.RomsPath(t), "vbl_nmi_timing", romName)
			runAndCompareFrame(t, romPath, outdir, romName, frameidx)
		})
	}
}

func TestPPUReadBuffer(t *testing.T) {
	const frameidx = 1300

	romPath := filepath.Join(tests.RomsPath(t), "ppu_read_buffer", "test_ppu_read_buffer.nes")
	runAndCompareFrame(t, romPath, "testdata", t.Name(), frameidx)
}

func TestUxROMSubmappers(t *testing.T) {
	const frameidx = 60

	outdir := filepath.Join("testdata", t.Name())
	os.Mkdir(outdir, 0755)

	roms := []string{
		"2_test_0.nes",
		"2_test_1.nes",
		"2_test_2.nes",
	}
	for _, romName := range roms {
		t.Run(romName, func(t *testing.T) {
			romPath := filepath.Join("..", "tests", "mapper2", romName)
			runAndCompareFrame(t, romPath, outdir, romName, frameidx)
		})
	}
}

func TestAxROMSubmappers(t *testing.T) {
	const frameidx = 60

	outdir := filepath.Join("testdata", t.Name())
	os.Mkdir(outdir, 0755)

	roms := []string{
		"7_test_0.nes",
		"7_test_1.nes",
		"7_test_2.nes",
	}
	for _, romName := range roms {
		t.Run(romName, func(t *testing.T) {
			romPath := filepath.Join("..", "tests", "mapper7", romName)
			runAndCompareFrame(t, romPath, outdir, romName, frameidx)
		})
	}
}

func TestCNROMSubmappers(t *testing.T) {
	const frameidx = 60

	outdir := filepath.Join("testdata", t.Name())
	os.Mkdir(outdir, 0755)

	roms := []string{
		"3_test_0.nes",
		"3_test_1.nes",
		"3_test_2.nes",
	}
	for _, romName := range roms {
		t.Run(romName, func(t *testing.T) {
			romPath := filepath.Join("..", "tests", "mapper3", romName)
			runAndCompareFrame(t, romPath, outdir, romName, frameidx)
		})
	}
}

func runAndCompareFrame(t *testing.T, romPath, frameDir, framePath string, frame int64) {
	t.Parallel()

	rom, err := ines.ReadRom(romPath)
	if err != nil {
		t.Fatal(err)
	}

	nes, err := powerUp(rom)
	if err != nil {
		t.Fatal(err)
	}

	out := newTestingOutput(
		TestingOutputConfig{
			Height:        hw.NTSCHeight,
			Width:         hw.NTSCWidth,
			SaveFrameDir:  frameDir,
			SaveFrameFile: framePath,
			SaveFrameNum:  frame,
		},
	)

	e := Emulator{NES: nes, out: out}
	e.Run()

	out.CompareFrame(t)
}
