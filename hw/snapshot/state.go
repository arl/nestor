package snapshot

//go:generate go tool msgp -tests=false -marshal=false

type NES struct {
	Version int
	CPU     *CPU
	RAM     [0x800]uint8
	DMA     *DMA
	PPU     *PPU
}

type CPU struct {
	PC uint16
	SP uint8
	P  uint8
	A  uint8
	X  uint8
	Y  uint8

	Cycles      int64
	MasterClock int64

	IRQFlag    uint8
	RunIRQ     bool
	PrevRunIRQ bool

	NMIFlag     bool
	PrevNeedNMI bool
	PrevNMIFlag bool
	NeedNMI     bool
}

type DMA struct {
	DMCRunning bool
	AbortDMC   bool
	OAMRunning bool
	DummyCycle bool
	NeedHalt   bool
}

type PPU struct {
	Palette [0x20]uint8
	OAMMem  [0x100]uint8

	OAM  [8]Sprite
	OAM2 [8]Sprite

	OpenBus         uint8
	OpenBusDecayBuf [8]uint32

	BusAddr    uint16
	OAMAddr    uint8
	VRAMAddr   uint16
	VRAMTemp   uint16
	WriteLatch bool
	PPUDataBuf uint8

	PPUBgRegs PPUBgRegs

	PPUCTRL   uint8
	PPUMASK   uint8
	PPUSTATUS uint8

	MasterClock uint64
	Cycle       uint32
	Scanline    int
	FrameCount  uint32

	OddFrame      bool
	PreventVBlank bool
}

type Sprite struct {
	ID    uint8
	X     uint8
	Y     uint8
	Tile  uint8
	Attr  uint8
	DataL uint8
	DataH uint8
}

type PPUBgRegs struct {
	AddrLatch uint16
	Finex     uint8
	NT        uint8
	AT        uint8
	BgLo      uint8
	BgHi      uint8

	// shift registers/latches.
	BgShiftLo uint16
	BgShiftHi uint16
	ATShiftLo uint8
	ATShiftHi uint8
	ATLatchLo bool
	ATLatchHi bool
}
