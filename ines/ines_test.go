package ines

import (
	"path/filepath"
	"testing"

	"nestor/tests"
)

func TestRomOpen(t *testing.T) {
	romsDir := filepath.Join(tests.RomsPath(t), "instr_test-v5", "rom_singles")

	paths := []string{
		"01-basics.nes",
		"02-implied.nes",
		"03-immediate.nes",
		"04-zero_page.nes",
		"05-zp_xy.nes",
		"06-absolute.nes",
		"07-abs_xy.nes",
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

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			rom, err := ReadRom(filepath.Join(romsDir, path))
			if err != nil {
				t.Fatal(err)
			}
			_ = rom
		})
	}
}
