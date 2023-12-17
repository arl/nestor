package hw

import (
	"nestor/emu/hwio"
	log "nestor/emu/logger"
)

const (
	NumScanlines = 262 // Number of scanlines per frame.
	NumCycles    = 341 // Number of PPU cycles per scanline.
)

type PPU struct {
	Bus *hwio.Table // PPU bus
	// CPU  *cpu.CPU
	Regs Regs // PPU registers

	Cycle    int // Current cycle/pixel in scanline
	Scanline int // Current scanline being drawn

	//	$0000-$0FFF	$1000	Pattern table 0
	//	$1000-$1FFF	$1000	Pattern table 1
	PatternTables hwio.Mem `hwio:"offset=0x0000,size=0x2000,wcb"`

	// $2000-$23FF	$0400	Nametable 0
	// $2400-$27FF	$0400	Nametable 1
	// $2800-$2BFF	$0400	Nametable 2
	// $2C00-$2FFF	$0400	Nametable 3
	// $3000-$3EFF	$0F00	Mirrors of $2000-$2EFF
	NameTables hwio.Mem `hwio:"offset=0x2000,size=0x1000,vsize=0x1F00,wcb"`

	// $3F00-$3F1F	$0020	Palette RAM indexes
	// $3F20-$3FFF	$00E0	Mirrors of $3F00-$3F1F
	Palettes hwio.Mem `hwio:"offset=0x3F00,size=0x20,vsize=0x100,wcb"`
}

func NewPPU() *PPU {
	return &PPU{
		Bus: hwio.NewTable("ppu"),
	}
}

func (p *PPU) InitBus() {
	hwio.MustInitRegs(p)
	p.Bus.MapBank(0x0000, p, 0)
	p.Reset()
}

func (p *PPU) Reset() {
	// TODO
}

func (p *PPU) Tick() {
	switch {
	// Pre-render line
	case p.Scanline == 261:
		if p.Cycle == 1 {
			// Clear vblank, sprite0Hit and spriteOverflow
			const mask = 1<<vblank | 1<<sprite0Hit | 1<<spriteOverflow
			p.Regs.PPUSTATUS.ClearBits(mask)

			if p.Regs.PPUCTRL.GetBit(nmi) {
				panic("CONTINUER ICI")
			}
		}

	// Visible scanlines
	case p.Scanline >= 0 && p.Scanline <= 239:
		switch {
		// Idle
		case p.Cycle == 0:
			break
		// Fetch data
		case p.Cycle > 0 && p.Cycle <= 256:
			break
		case p.Cycle > 256 && p.Cycle <= 320:
			break
		case p.Cycle > 321 && p.Cycle <= 336:
			break
		case p.Cycle > 337 && p.Cycle <= 340:
			break
		}

	// Post-render scanline
	case p.Scanline == 240:
		break

	// VBlank start (set nmi)
	case p.Scanline == 241:
		if p.Cycle == 1 {
			p.Regs.PPUSTATUS.Value |= 1 << vblank
			// if
		}
		// { status.vBlank = true; if (ctrl.nmi) CPU::set_nmi(); }

	// VBlank
	case p.Scanline >= 242 && p.Scanline <= 260:
	}

	p.Cycle++
	if p.Cycle >= NumCycles {
		p.Cycle = 0
		p.Scanline++
		if p.Scanline >= NumScanlines {
			p.Scanline = 0
		}
	}
}

func (p *PPU) WritePATTERNTABLES(addr uint16, n int) {
	log.ModPPU.DebugZ("Write to PATTERNTABLES").
		Hex8("val", p.PatternTables.Data[addr]).
		Hex16("addr", addr).
		End()
}

func (p *PPU) WriteNAMETABLES(addr uint16, n int) {
	memaddr := addr & 0x0FFF
	ntnum := memaddr / 0x0400
	log.ModPPU.DebugZ("Write to NAMETABLES").
		Uint16("num", ntnum).
		Hex8("val", p.NameTables.Data[memaddr]).
		Hex16("addr", addr).
		End()
}

func (p *PPU) WritePALETTES(addr uint16, n int) {
	memaddr := addr & 0x01F
	log.ModPPU.DebugZ("Write to PALETTES").
		Hex8("val", p.Palettes.Data[memaddr]).
		Hex16("addr", addr).
		End()
}
