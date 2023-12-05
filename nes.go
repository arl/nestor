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
	ppubus := hwio.NewTable("ppu")
	nes.PPU = ppu.New(ppubus)

	cpubus := hwio.NewTable("cpu")
	nes.CPU = cpu.NewCPU(cpubus, nes.PPU)

	// RAM is 0x800 bytes, mirrored.
	ram := make([]byte, 0x0800)
	cpubus.MapMemorySlice(0x0000, 0x07FF, ram, false)
	cpubus.MapMemorySlice(0x0800, 0x0FFF, ram, false)
	cpubus.MapMemorySlice(0x1000, 0x17FF, ram, false)
	cpubus.MapMemorySlice(0x1800, 0x1FFF, ram, false)

	// Map PPU registers and mirrors.
	for i := uint16(0x2000); i < 0x4000; i += 8 {
		cpubus.MapBank(i, nes.PPU, 0)
	}
	// PPU VRAM (name tables) and mirror.
	vram := make([]byte, 0x1000)
	ppubus.MapMemorySlice(0x2000, 0x2FFF, vram, false)
	// CONTINUER ICI: we can't do that since the result slice is not a power of 2.
	// What we should do is copy ALL of ndsemu hwio and use hwio.Mem with the size and vsize struct tags. See examples in:

	// geometry.go:92: GxFifo  hwio.Mem `hwio:"bank=0,offset=0x0,size=4,vsize=0x40,rw8=off,rw16=off,wcb"`
	// geometry.go:93: GxCmd   hwio.Mem `hwio:"bank=0,offset=0x40,size=4,vsize=0x190,rw8=off,rw16=off,wcb"`

	ppubus.MapMemorySlice(0x3000, 0x3EFF, vram[:0xF00], false)

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
