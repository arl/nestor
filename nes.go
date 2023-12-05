package main

import (
	"fmt"
	"io"

	"nestor/cpu"
	"nestor/emu/hwio"
	"nestor/ines"
	"nestor/ppu"
)

type NES struct {
	CPU *cpu.CPU
	PPU *ppu.PPU
}

func (nes *NES) PowerUp(rom *ines.Rom) error {
	nes.PPU = ppu.New()

	cpubus := &hwio.MemMap{Name: "cpu"}
	nes.CPU = cpu.NewCPU(cpubus, nes.PPU)

	// RAM is 0x800 bytes, mirrored.
	ram := make([]byte, 0x0800)
	cpubus.MapSlice(0x0000, 0x07FF, ram)
	cpubus.MapSlice(0x0800, 0x0FFF, ram)
	cpubus.MapSlice(0x1000, 0x17FF, ram)
	cpubus.MapSlice(0x1800, 0x1FFF, ram)

	// Map PPU registers and their mirrors.
	for i := uint16(0x2000); i < 0x4000; i += 8 {
		cpubus.MapBank(i, nes.PPU, 0)
	}
	cpubus.Write8(0x2006, 0x23)

	if rom.Mapper() != 0 {
		// Only handle mapper 000 (NROM) for now.
		return fmt.Errorf("unsupported mapper: %d", rom.Mapper())
	}

	if err := loadMapper000(rom, nes); err != nil {
		return fmt.Errorf("failed to load mapper %03d: %s", rom.Mapper(), err)
	}
	return nil
}

// Reset forwards the reset signal to all hardware.
func (nes *NES) Reset() {
	nes.CPU.Reset()
	nes.PPU.Reset()
}

func (nes *NES) Run() {
	for {
		nes.CPU.Run(29692) // random
	}
}

func (nes *NES) RunDisasm(out io.Writer, nestest bool) {
	d := cpu.NewDisasm(nes.CPU, out, nestest)
	for {
		d.Run(29692) // random
	}
}

type ticker struct{}

func (tt *ticker) Tick() {}
