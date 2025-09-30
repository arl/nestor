// Package snapshot provides types and functions for snapshot encoding and
// decoding.
package snapshot

import (
	"github.com/tinylib/msgp/msgp"

	"nestor/hw/hwdefs"
)

//go:generate go tool msgp -tests=false -marshal=false

type NES struct {
	Version int
	CPU     *CPU
	RAM     [0x800]uint8
	PPU     *PPU
	APU     *APU
	Mixer   *APUMixer
	Mapper  *MapperState
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

	Input InputPorts
	DMA   DMA
}

type InputPorts struct {
	State      [2]uint8
	PrevStrobe bool
	Strobe     bool
}

type DMA struct {
	NeedHalt     bool
	DummyCycle   bool
	DMCRunning   bool
	AbortDMC     bool
	OAMRunning   bool
	OAMPage      uint8
	OAMDMARegVal uint8
}

type PPU struct {
	MasterClock uint64

	PaletteRAM         [0x20]uint8
	SpriteRAM          [0x100]uint8
	SecondarySpriteRAM [0x20]uint8
	OpenBusDecayStamp  [8]int32
	MemoryReadBuffer   uint8

	SpriteRAMAddr uint8
	VRAMAddr      uint16
	XScroll       uint8
	TmpVRAMAddr   uint16
	WriteToggle   bool

	HighBitShift  uint16
	LowBitShift   uint16
	PPUCTRL       PPUCTRL
	PPUMASK       PPUMASK
	PPUSTATUS     PPUSTATUS
	PPUBusAddress uint16

	PaletteRAMMask      uint16
	IntensifyColorBits  uint16
	Scanline            int64
	Cycle               int64
	FrameCount          uint32
	CurrentTilePalette  uint8
	PreviousTilePalette uint8
	Tile                PPUTileInfo
	OpenBus             uint8

	SpriteTiles          [64]PPUSpriteInfo
	SpriteIndex          int
	SpriteCount          int
	SpriteAddrH          uint8
	SpriteAddrL          uint8
	Sprite0Added         bool
	Sprite0Visible       bool
	OAMCopybuffer        uint8
	SecondaryOAMAddr     uint8
	SpriteInRange        bool
	RenderingEnabled     bool
	PrevRenderingEnabled bool
	IgnoreVRAMRead       int
	OAMCopyDone          bool

	NeedStateUpdate       bool
	NeedVideoRAMIncrement bool
	PreventVblFlag        bool
	OverflowBugCounter    uint8
	UpdateVRAMAddr        uint16
	UpdateVRAMAddrDelay   uint8
	AllowFullPpuAccess    bool

	Region int
}

type PPUCTRL struct {
	VerticalWrite         bool
	SpritePatternAddr     uint16
	BackgroundPatternAddr uint16
	LargeSprites          bool
	SecondaryPpu          bool // not saved
	NmiOnVerticalBlank    bool
}

type PPUMASK struct {
	Grayscale         bool
	BackgroundMask    bool
	SpriteMask        bool
	BackgroundEnabled bool
	SpritesEnabled    bool
	IntensifyRed      bool
	IntensifyGreen    bool
	IntensifyBlue     bool
}

type PPUSTATUS struct {
	SpriteOverflow bool
	Sprite0Hit     bool
	VerticalBlank  bool
}

type PPUTileInfo struct {
	LowByte       uint8
	HighByte      uint8
	PaletteOffset uint8
	TileAddr      uint16
}

type PPUSprite struct {
	ID    uint8
	X     uint8
	Y     uint8
	Tile  uint8
	Attr  uint8
	DataL uint8
	DataH uint8
}

type PPUSpriteInfo struct {
	SpriteX            uint8
	LowByte            uint8
	HighByte           uint8
	PaletteOffset      uint8
	HorizontalMirror   bool
	BackgroundPriority bool
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

type APU struct {
	Square1      APUSquare
	Square2      APUSquare
	Triangle     APUTriangle
	Noise        APUNoise
	DMC          APUDMC
	FrameCounter APUFrameCounter
}

type APUTimer struct {
	Timer      uint16
	Period     uint16
	LastOutput int8
}

type APUEnveloppe struct {
	LengthCounter APULengthCounter
	ConstVolume   bool
	Vol           uint8
	Start         bool
	Divider       int8
	Counter       uint8
}

type APULengthCounter struct {
	Enabled   bool
	Halt      bool
	NewHalt   bool
	Counter   uint8
	PrevVal   uint8
	ReloadVal uint8
}

type APUSquare struct {
	SweepTargetPeriod uint32
	RealPeriod        uint16
	Timer             APUTimer
	Envelope          APUEnveloppe
	SweepEnabled      bool
	SweepPeriod       uint8
	SweepNegate       bool
	SweepShift        uint8
	SweepDivider      uint8

	ReloadSweep bool
	Duty        uint8
	DutyPos     uint8

	DutyRegVal   uint8
	SweepRegVal  uint8
	TimerRegVal  uint8
	LengthRegVal uint8
}

type APUTriangle struct {
	LengthCounter       APULengthCounter
	Timer               APUTimer
	LinearCounter       uint8
	LinearCounterReload uint8
	LinearReload        bool
	LinearCtrl          bool
	Pos                 uint8

	LinearRegVal uint8
	UnusedRegVal uint8
	TimerRegVal  uint8
	LengthRegVal uint8
}

type APUNoise struct {
	Envelope       APUEnveloppe
	Timer          APUTimer
	ShitftRegister uint16
	Mode           bool

	VolumeRegVal uint8
	UnusedRegVal uint8
	PeriodRegVal uint8
	LengthRegVal uint8
}

type APUDMC struct {
	Timer APUTimer

	SampleAddr  uint16
	SampleLen   uint16
	CurrentAddr uint16
	Remaining   uint16

	OutputLevel  uint8
	ReadBuf      uint8
	BitsLeft     uint8
	StartDelay   uint8
	DisableDelay uint8

	IRQEnabled bool
	Loop       bool
	BufEmpty   bool
	ShiftReg   uint8
	Silence    bool
	NeedToRun  bool

	FLAGSRegVal      uint8
	LOADRegVal       uint8
	SAMPLEADDRRegVal uint8
	SAMPLELENRegVal  uint8
}

type APUFrameCounter struct {
	PrevCycle  int32
	CurStep    uint32
	StepMode   uint32
	InhibitIRQ bool

	BlockTick         uint8
	WriteDelayCounter int8
	NewVal            int16
}

type APUMixer struct {
	ClockRate           uint32
	SampleRate          uint32
	CurrentOutput       [hwdefs.NumAudioChannels]int16
	PreviousOutputLeft  int16
	PreviousOutputRight int16
}

type MapperState struct {
	Num  uint16   `msg:"num"`
	Data msgp.Raw `msg:"data"`
}
