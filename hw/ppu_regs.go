package hw

import (
	"nestor/emu/hwio"
	log "nestor/emu/logger"
)

// CPU-exposed memory-mapped PPU registers.
// This bank is mapped at 0x2000-0x3FFF, with mirrors.
type Regs struct {
	PPUCTRL   hwio.Reg8 `hwio:"offset=0x0,writeonly,wcb"`
	PPUMASK   hwio.Reg8 `hwio:"offset=0x1,writeonly,wcb"`
	PPUSTATUS hwio.Reg8 `hwio:"offset=0x2,readonly,rcb"`
	// OAMADDR   hwio.Reg8 `hwio:"offset=0x3,writeonly,wcb"`
	// OAMDATA   hwio.Reg8 `hwio:"offset=0x4"`
	PPUSCROLL hwio.Reg8 `hwio:"offset=0x5,writeonly,wcb"`
	PPUADDR   hwio.Reg8 `hwio:"offset=0x6,writeonly,wcb"`
	PPUDATA   hwio.Reg8 `hwio:"offset=0x7,rcb,wcb,"`

	// OAMDMA hwio.Reg8 `hwio:"bank:1,writeonly,wcb"`
}

const (
	// PPUCTRL bits

	// Generate an NMI at the start of the
	// vertical blanking interval (0: off; 1: on)
	nmi = 7

	// PPU master/slave select
	// (0: read backdrop from EXT pins; 1: output color on EXT pins)
	ppuMasterSlave = 6

	// Sprite size (0: 8x8 pixels; 1: 8x16 pixels – see PPU OAM#Byte 1)
	spriteSize = 5

	// Background pattern table address (0: $0000; 1: $1000)
	backgroundAddr = 4

	// Sprite pattern table address for 8x8 sprites
	// (0: $0000; 1: $1000; ignored in 8x16 mode)
	spriteAddr = 3

	// VRAM address increment per CPU read/write of PPUDATA
	// (0: add 1, going across; 1: add 32, going down)
	vramIncr = 2

	// Base nametable address
	// (0 = $2000; 1 = $2400; 2 = $2800; 3 = $2C00)
	baseNTmask = 0b11
)

func (r *Regs) WritePPUCTRL(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUCTRL").Hex8("val", val).End()
}

const (
	// PPUSTATUS bits

	// Vertical blank has started (0: not in vblank; 1: in vblank).
	// Set at dot 1 of line 241 (the line *after* the post-render
	// line); cleared after reading $2002 and at dot 1 of the
	// pre-render line.
	vblank = 7

	// Sprite 0 Hit.  Set when a nonzero pixel of sprite 0 overlaps
	// a nonzero background pixel; cleared at dot 1 of the pre-render
	// line.  Used for raster timing.
	sprite0Hit = 6

	// Sprite overflow. The intent was for this flag to be set
	// whenever more than eight sprites appear on a scanline, but a
	// hardware bug causes the actual behavior to be more complicated
	// and generate false positives as well as false negatives; see
	// PPU sprite evaluation. This flag is set during sprite
	// evaluation and cleared at dot 1 (the second dot) of the
	// pre-render line.
	spriteOverflow = 5

	// Returns stale PPU bus contents.
	openbusMask = 0b11111
)

func (r *Regs) ReadPPUSTATUS(val uint8) uint8 {
	if val != 0 {
		log.ModPPU.DebugZ("Read from PPUSTATUS").Hex8("val", val).End()
	}
	return val
}

func (r *Regs) WritePPUMASK(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUMASK").Hex8("val", val).End()
}

func (r *Regs) WritePPUSCROLL(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUSCROLL").Hex8("val", val).End()
}

func (r *Regs) WritePPUADDR(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUADDR").Hex8("val", val).End()
}

func (r *Regs) ReadPPUDATA(val uint8) uint8 {
	log.ModPPU.DebugZ("Read from PPUDATA").Hex8("val", val).End()
	return val
}

func (r *Regs) WritePPUDATA(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUDATA").Hex8("val", val).End()
}
