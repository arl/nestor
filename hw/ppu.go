package hw

import (
	"unsafe"

	"nestor/emu/hwio"
	"nestor/emu/log"
)

const (
	NumScanlines = 262 // Number of scanlines per frame.
	NumCycles    = 341 // Number of PPU cycles per scanline.
)

// Throwaway frame buffer for the first PPU cycles, before an actual one for the
// actual output.
var tmpFramebuf = make([]uint32, 256*240)

type PPU struct {
	Bus *hwio.Table
	CPU *CPU

	Cycle    int // Current cycle/pixel in scanline
	Scanline int // Current scanline being drawn

	//	$0000-$0FFF	$1000	Pattern table 0
	//	$1000-$1FFF	$1000	Pattern table 1
	PatternTables hwio.Mem `hwio:"offset=0x0000,size=0x2000,wcb"`

	// Nametables mapping depends on rom/mapper.
	//  $2000-$23FF	$0400	Nametable 0
	//  $2400-$27FF	$0400	Nametable 1
	//  $2800-$2BFF	$0400	Nametable 2
	//  $2C00-$2FFF	$0400	Nametable 3
	//  $3000-$3EFF	$0F00	Mirrors of $2000-$2EFF
	Nametables [0x800]byte

	// $3F00-$3F1F	$0020	Palette RAM indexes
	// $3F20-$3FFF	$00E0	Mirrors of $3F00-$3F1F
	Palettes hwio.Mem `hwio:"offset=0x3F00,size=0x20,vsize=0x100,wcb"`

	// CPU-exposed memory-mapped PPU registers
	// mapped from $2000 to $2007, mirrored up to $3fff
	PPUCTRL   hwio.Reg8 `hwio:"bank=1,offset=0x0,writeonly,wcb"`
	PPUMASK   hwio.Reg8 `hwio:"bank=1,offset=0x1,writeonly,wcb"`
	PPUSTATUS hwio.Reg8 `hwio:"bank=1,offset=0x2,readonly,rcb"`
	// OAMADDR   hwio.Reg8 `hwio:"bank=1,offset=0x3,writeonly,wcb"`
	// OAMDATA   hwio.Reg8 `hwio:"bank=1,offset=0x4"`
	PPUSCROLL hwio.Reg8 `hwio:"bank=1,offset=0x5,writeonly,wcb"`
	PPUADDR   hwio.Reg8 `hwio:"bank=1,offset=0x6,writeonly,wcb"`
	PPUDATA   hwio.Reg8 `hwio:"bank=1,offset=0x7,rcb,wcb"`

	// OAMDMA hwio.Reg8 `hwio:"bank=2,writeonly,wcb"`

	framebuf []uint32 // RGBA framebuffer
	oddFrame bool

	// VRAM read/write
	vramAddr   loopy
	vramTmp    loopy
	writeLatch bool

	ppuDataRbuf uint8 // only used for PPUDATA reads
	busAddr     uint16

	bg bgregs
}

func NewPPU() *PPU {
	return &PPU{
		Bus:      hwio.NewTable("ppu"),
		framebuf: tmpFramebuf,
	}
}

func (p *PPU) SetFrameBuffer(framebuf []byte) {
	p.framebuf = unsafe.Slice((*uint32)(unsafe.Pointer(&framebuf[0])), len(framebuf)/4)
}

func (p *PPU) InitBus() {
	hwio.MustInitRegs(p)
	p.Bus.MapBank(0x0000, p, 0)
}

func (p *PPU) Reset() {
	p.Scanline = 0
	p.Cycle = 0
	p.writeLatch = false
	p.vramAddr = 0
	p.PPUCTRL.Value = 0
	p.PPUMASK.Value = 0
	p.PPUSTATUS.Value = 0
	p.oddFrame = false
	for i := range p.Nametables {
		p.Nametables[i] = 0xff
	}
}

func (p *PPU) Tick() {
	switch {
	case p.Scanline < 240:
		p.doScanline(renderMode)
	case p.Scanline == 240:
		p.doScanline(postRender)
	case p.Scanline == 241:
		p.doScanline(vblankNMI)
	case p.Scanline == 261:
		p.doScanline(preRender)
	}
	p.Cycle++
	if p.Cycle >= NumCycles {
		p.Cycle %= NumCycles
		p.Scanline++
		if p.Scanline >= NumScanlines {
			p.Scanline = 0
			p.oddFrame = !p.oddFrame
		}
	}
}

type scanlineMode int

const (
	preRender scanlineMode = iota
	renderMode
	postRender
	vblankNMI
)

func (p *PPU) doScanline(sm scanlineMode) {
	switch sm {
	case vblankNMI:
		if p.Cycle == 1 {
			ppustatus := ppustatus(p.PPUSTATUS.Value)
			ppustatus.setVblank(true)
			p.PPUSTATUS.Value = uint8(ppustatus)
			if ppuctrl(p.PPUCTRL.Value).nmi() {
				p.CPU.setNMIflag()
				log.ModPPU.DebugZ("Set NMI flag").End()
			}
		}
	case postRender:
		// nothing to do
		if p.Cycle == 1 {
			// At the start of vblank, the bus address is set back to
			// VideoRamAddr.
			p.busAddr = p.vramAddr.addr()
		}

	case preRender, renderMode:
		if p.Cycle == 1 {
			if sm == preRender {
				ppustatus := ppustatus(p.PPUSTATUS.Value)
				ppustatus.setSpriteHit(false)
				ppustatus.setSpriteOverflow(false)
				p.PPUSTATUS.Value = uint8(ppustatus)
			}
		}

		switch {
		case p.Cycle >= 2 && p.Cycle <= 255,
			p.Cycle >= 322 && p.Cycle <= 337:
			p.renderPixel()
			switch p.Cycle & 0b111 {

			// nametable
			case 1:
				p.bg.addrLatch = p.ntAddr()
				p.refillShifters()
			case 2:
				p.bg.nt = p.Read8(p.bg.addrLatch)

			// attribute table
			case 3:
				p.bg.addrLatch = p.atAddr()
			case 4:
				p.bg.at = p.Read8(p.bg.addrLatch)
				if p.vramAddr.coarsey()&2 != 0 {
					p.bg.at >>= 4
				}
				if p.vramAddr.coarsex()&2 != 0 {
					p.bg.at >>= 2
				}

			// low background byte
			case 5:
				p.bg.addrLatch = p.bgAddr()
			case 6:
				p.bg.bglo = p.Read8(p.bg.addrLatch)

			// high background byte
			case 7:
				p.bg.addrLatch += 8
			case 0:
				p.bg.bghi = p.Read8(p.bg.addrLatch)
				p.horzScroll()
			}
		case p.Cycle == 256:
			p.renderPixel()
			p.bg.bghi = p.Read8(p.bg.addrLatch)
			p.vertScroll()
		case p.Cycle == 257:
			p.renderPixel()
			p.refillShifters()
			p.horzUpdate()
		case p.Cycle >= 280 && p.Cycle <= 304:
			if sm == preRender {
				p.vertUpdate()
			}

		// shifters aren't refilled
		case p.Cycle == 1:
			p.bg.addrLatch = p.ntAddr()
			if sm == preRender {
				ppustatus := ppustatus(p.PPUSTATUS.Value)
				ppustatus.setVblank(false)
				p.PPUSTATUS.Value = uint8(ppustatus)
			}
		case p.Cycle == 321:
			fallthrough
		case p.Cycle == 339:
			p.bg.addrLatch = p.ntAddr()

		// 'garbage' fetches
		case p.Cycle == 338:
			p.bg.nt = p.Read8(p.bg.addrLatch)
		case p.Cycle == 340:
			p.bg.nt = p.Read8(p.bg.addrLatch)
			if sm == preRender && p.renderingEnabled() && p.oddFrame {
				p.Cycle++
			}
		}
	}
}

// background registers
type bgregs struct {
	// temporary address latch storing the address for next cycle,
	// since fetches takes 2 cycles.
	addrLatch uint16

	// 3-bit 'fine x scroll' register.
	finex uint8

	// latches for background rendering.
	nt, at, bglo, bghi uint8

	// shift registers.
	bgShiftlo, bgShifthi uint16
}

func (p *PPU) ntAddr() uint16 {
	return 0x2000 | p.vramAddr.val()&0xfff
}

func (p *PPU) atAddr() uint16 {
	return 0x23C0 | p.vramAddr.nametable()<<10 | p.vramAddr.coarsey()/4<<3 | p.vramAddr.coarsex()/4
}

func (p *PPU) bgAddr() uint16 {
	ppuctrl := ppuctrl(p.PPUCTRL.Value)
	return ppuctrl.bgTable()*0x1000 + (uint16(p.bg.nt) * 16) + p.vramAddr.finey()
}

func (p *PPU) refillShifters() {
	p.bg.bgShiftlo = (p.bg.bgShiftlo & 0xFF00) | uint16(p.bg.bglo)
	p.bg.bgShifthi = (p.bg.bgShifthi & 0xFF00) | uint16(p.bg.bghi)
}

func (p *PPU) horzScroll() {
	if !p.renderingEnabled() {
		return
	}
	if p.vramAddr.coarsex() == 31 {
		p.vramAddr.setVal(p.vramAddr.val() ^ 0x41F)
	} else {
		p.vramAddr.setCoarsex(p.vramAddr.coarsex() + 1)
	}
}

func (p *PPU) vertScroll() {
	if !p.renderingEnabled() {
		return
	}
	if finey := p.vramAddr.finey(); finey < 7 {
		p.vramAddr.setFiney(finey + 1)
		return
	}

	p.vramAddr.setFiney(0)
	if p.vramAddr.coarsey() == 31 {
		p.vramAddr.setCoarsey(0)
	} else if p.vramAddr.coarsey() == 29 {
		p.vramAddr.setCoarsey(0)
		p.vramAddr.setNametable(p.vramAddr.nametable() ^ 0b10)
	} else {
		p.vramAddr.setCoarsey(p.vramAddr.coarsey() + 1)
	}
}

func (p *PPU) horzUpdate() {
	if !p.renderingEnabled() {
		return
	}
	p.vramAddr.setVal((p.vramAddr.val() & ^uint16(0x041F)) | (p.vramTmp.val() & 0x041F))
}

func (p *PPU) vertUpdate() {
	if !p.renderingEnabled() {
		return
	}
	p.vramAddr.setVal((p.vramAddr.val() & ^uint16(0x7BE0)) | (p.vramTmp.val() & 0x7BE0))
}

func (p *PPU) renderingEnabled() bool {
	mask := ppumask(p.PPUMASK.Value)
	return mask.bg() || mask.sprites()
}

func (p *PPU) renderPixel() {
	var palette uint8
	var x = p.Cycle - 2

	mask := ppumask(p.PPUMASK.Value)
	if p.Scanline < 240 && p.Cycle >= 0 && p.Cycle < 256 {
		if mask.bg() {
			hibit := uint8(p.bg.bgShifthi>>(15-p.bg.finex)) & 1
			lobit := uint8(p.bg.bgShiftlo>>(15-p.bg.finex)) & 1
			palette = (hibit << 1) | lobit
		}
		paddr := uint16(0x3f00)
		if p.renderingEnabled() {
			paddr += uint16(palette)
		}
		pidx := p.Read8(paddr)
		p.framebuf[p.Scanline*256+x] = nesPalette[pidx]
	}

	// Perform background shifts:
	p.bg.bgShiftlo <<= 1
	p.bg.bgShifthi <<= 1
	// atShiftL = (atShiftL << 1) | atLatchL
	// atShiftH = (atShiftH << 1) | atLatchH
}

func (p *PPU) WritePATTERNTABLES(addr uint16, n int) {
	log.ModPPU.DebugZ("Write to PATTERNTABLES").
		Hex8("val", p.PatternTables.Data[addr]).
		Hex16("addr", addr).
		End()
}

// PPUCTRL: $2000
func (p *PPU) WritePPUCTRL(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUCTRL").Hex8("val", val).End()

	// By toggling the nmi bit (bit 7 of PPUCTRL) during vblank without reading
	// PPUSTATUS, a program can cause /nmi to be pulled low multiple times,
	// causing multiple NMIs to be generated.
	ppuctrl := ppuctrl(val)
	ppustatus := ppustatus(p.PPUSTATUS.Value)
	if !ppuctrl.nmi() {
		p.CPU.clearNMIflag()
	} else if ppustatus.vblank() {
		p.CPU.setNMIflag()
	}

	// Transfer the nametable bits.
	p.vramTmp.setNametable(ppuctrl.nametable())
}

// PPUMASK: $2001
func (p *PPU) WritePPUMASK(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUMASK").Hex8("val", val).End()
}

// PPUSTATUS: $2002
func (p *PPU) ReadPPUSTATUS(val uint8) uint8 {
	p.writeLatch = false

	ppustatus := ppustatus(val)
	ppustatus.setSpriteOverflow(true)
	ppustatus.setSpriteHit(true)
	ppustatus.setVblank(true)

	p.CPU.clearNMIflag()
	// TODO: emulate open bus?
	return uint8(ppustatus)
}

// PPUSCROLL: $2005
func (p *PPU) WritePPUSCROLL(old, val uint8) {
	log.ModPPU.DebugZ("Write to PPUSCROLL").Hex8("val", val).End()

	if !p.writeLatch { // first write
		p.bg.finex = val & 0b111
		p.vramTmp.setCoarsex(uint16(val) >> 3)
	} else { // second write
		p.vramTmp.setFiney(uint16(val))
		p.vramTmp.setCoarsey(uint16(val) >> 3)
	}

	p.writeLatch = !p.writeLatch
}

// PPUADDR: $2006
func (p *PPU) WritePPUADDR(old, val uint8) {
	if !p.writeLatch { //first write
		p.vramTmp.setHigh(val & 0x3f)
	} else { // second write
		p.vramTmp.setLow(val)
		p.vramAddr.setVal(uint16(p.vramTmp))
	}

	p.writeLatch = !p.writeLatch
}

// PPUDATA: $2007
func (p *PPU) ReadPPUDATA(_ uint8) uint8 {
	// Reading VRAM is too slow so the actual data
	// will be returned at the next read.
	val := p.ppuDataRbuf
	p.ppuDataRbuf = p.Read8(p.vramAddr.addr())

	if p.busAddr >= 0x3F00 {
		// Reading palette data is immediate.
		// val = p.Read8(p.vramAddr.addr())
		// Still it overwrites the read buffer.
		val = p.readPalette(p.busAddr)
	}

	p.vramIncr()
	log.ModPPU.DebugZ("VRAM read").
		Hex16("addr", p.vramAddr.addr()).
		Hex8("val", val).
		End()
	return val
}

// PPUDATA: $2007
func (p *PPU) WritePPUDATA(old, val uint8) {
	p.Write8(p.vramAddr.addr(), val)
	p.vramIncr()

	log.ModPPU.DebugZ("VRAM write").
		Hex16("addr", p.vramAddr.addr()).
		Hex8("val", val).
		End()
}

func (p *PPU) vramIncr() {
	ppuctrl := ppuctrl(p.PPUCTRL.Value)
	var incr uint16 = 1
	if ppuctrl.incr() {
		incr = 32 // Increment by 32 if increment mode is set.
	}
	p.vramAddr.setAddr(uint16(p.vramAddr.addr()) + incr)
}

func (p *PPU) Read8(addr uint16) uint8 {
	p.busAddr = addr
	return p.Bus.Read8(addr)
}

func (p *PPU) Write8(addr uint16, val uint8) {
	p.busAddr = addr
	p.Bus.Write8(addr, val)
}

var nesPalette = [...]uint32{
	0xFF7C7C7C, 0xFF0000FC, 0xFF0000BC, 0xFF4428BC, 0xFF940084, 0xFFA80020, 0xFFA81000, 0xFF881400,
	0xFF503000, 0xFF007800, 0xFF006800, 0xFF005800, 0xFF004058, 0xFF000000, 0xFF000000, 0xFF000000,
	0xFFBCBCBC, 0xFF0078F8, 0xFF0058F8, 0xFF6844FC, 0xFFD800CC, 0xFFE40058, 0xFFF83800, 0xFFE45C10,
	0xFFAC7C00, 0xFF00B800, 0xFF00A800, 0xFF00A844, 0xFF008888, 0xFF000000, 0xFF000000, 0xFF000000,
	0xFFF8F8F8, 0xFF3CBCFC, 0xFF6888FC, 0xFF9878F8, 0xFFF878F8, 0xFFF85898, 0xFFF87858, 0xFFFCA044,
	0xFFF8B800, 0xFFB8F818, 0xFF58D854, 0xFF58F898, 0xFF00E8D8, 0xFF787878, 0xFF000000, 0xFF000000,
	0xFFFCFCFC, 0xFFA4E4FC, 0xFFB8B8F8, 0xFFD8B8F8, 0xFFF8B8F8, 0xFFF8A4C0, 0xFFF0D0B0, 0xFFFCE0A8,
	0xFFF8D878, 0xFFD8F878, 0xFFB8F8B8, 0xFFB8F8D8, 0xFF00FCFC, 0xFFF8D8F8, 0xFF000000, 0xFF000000,
}

func (p *PPU) WritePALETTES(addr uint16, n int) {
	memaddr := addr & 0x01F
	val := p.Palettes.Data[memaddr]
	p.writePalette(addr, val)
	log.ModPPU.DebugZ("Write to PALETTES").
		Hex8("val", val).
		Hex16("addr", addr).
		End()
}

func (p *PPU) readPalette(addr uint16) uint8 {
	addr &= 0x1F
	if addr == 0x10 || addr == 0x14 || addr == 0x18 || addr == 0x1C {
		addr &^= 0x10
	}
	return p.Palettes.Data[addr]
}

func (p *PPU) writePalette(addr uint16, val uint8) {
	val &= 0x3F
	addr &= 0x1F
	switch addr {
	case 0x00, 0x10:
		p.Palettes.Data[0x00] = val
		p.Palettes.Data[0x10] = val
	case 0x04, 0x14:
		p.Palettes.Data[0x04] = val
		p.Palettes.Data[0x14] = val
	case 0x08, 0x18:
		p.Palettes.Data[0x08] = val
		p.Palettes.Data[0x18] = val
	case 0x0C, 0x1C:
		p.Palettes.Data[0x0C] = val
		p.Palettes.Data[0x1C] = val
	default:
		p.Palettes.Data[addr] = val
	}
}
