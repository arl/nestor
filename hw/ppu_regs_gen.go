//go:build ignore

package hw

//go:generate bitfield -out ppu_regs.go

// 'Loopy' register
type loopy struct {
	coarsex   uint8  `bitfield:"5"` // Coarse X
	coarsey   uint8  `bitfield:"5"` // Coarse Y
	nametable uint8  `bitfield:"2"` // Nametable
	finey     uint16 `bitfield:"3"` // Fine Y

	low  uint8 `bitfield:"8,union=lohi"`
	high uint8 `bitfield:"7,union=lohi"`

	addr uint16 `bitfield:"14,union=addr"`
	val  uint16 `bitfield:"15,union=val"`
}

// ppuctrl register ($2000)
type ppuctrl struct {
	// Nametable selection mask
	// (0 = $2000; 1 = $2400; 2 = $2800; 3 = $2C00)
	nametable uint8 `bitfield:"2"`

	// VRAM address increment per CPU read/write of PPUDATA
	// (0: +1 i.e. horizontal; 1: +32 i.e. vertical)
	incr bool `bitfield:"1"`

	// Sprite pattern table address for 8x8 sprites
	// (0: $0000; 1: $1000; ignored in 8x16 mode)
	spriteTable bool `bitfield:"1"`

	// Background pattern table address (0: $0000; 1: $1000)
	bgTable uint16 `bitfield:"1"`

	// Sprite size (0: 8x8 pixels; 1: 8x16 pixels – see byte 1 of OAM)
	spriteSize bool `bitfield:"1"`

	// PPU master/slave select
	// (0: read backdrop from EXT pins; 1: output color on EXT pins)
	slave bool `bitfield:"1"`

	// Generate an NMI at the start of the
	// vertical blanking interval (0: off; 1: on)
	nmi bool `bitfield:"1"`

	val uint8 `bitfield:"8,union=val"`
}

// ppumask register ($2001)
type ppumask struct {
	// Grayscale. (0: normal color, 1: produce a greyscale display)
	gray bool `bitfield:"1"`

	// Show background in leftmost 8 pixels of screen
	bgLeft bool `bitfield:"1"`

	// Show sprites in leftmost 8 pixels of screen
	spriteLeft bool `bitfield:"1"`

	// Show background
	bg bool `bitfield:"1"`

	// Show sprites
	sprites bool `bitfield:"1"`

	// Emphasize red, green or blue.
	red   bool `bitfield:"1"`
	green bool `bitfield:"1"`
	blue  bool `bitfield:"1"`

	val uint8 `bitfield:"8,union=val"`
}

// ppustatus register ($2002)
type ppustatus struct {
	openBus uint8 `bitfield:"5"`

	// The intent was for this flag to be set whenever more than eight sprites
	// appear on a scanline, but a hardware bug causes the actual behavior to be
	// more complicated and generate false positives as well as false negatives;
	// This flag is set during sprite evaluation and cleared at dot 1 (the
	// second dot) of the pre-render line.
	spriteOverflow bool `bitfield:"1"`

	// Set when a nonzero pixel of sprite 0 overlaps a nonzero background pixel;
	// cleared at dot 1 of the pre-render line. Used for raster timing.
	spriteHit bool `bitfield:"1"`

	// Indicates whether vertical blank has started
	//  (0: not in vblank; 1: in vblank).
	//
	// Set at dot 1 of line 241 (the line *after* the post-render line); cleared
	// after reading $2002 and at dot 1 of the pre-render line.
	vblank bool `bitfield:"1"`

	val uint8 `bitfield:"8,union=val"`
}
