package snapshot

import "nestor/hw/hwdefs"

//go:generate go tool msgp -tests=false -marshal=false

type NES struct {
	Version int
	CPU     *CPU
	RAM     [0x800]uint8
	DMA     *DMA
	PPU     *PPU
	APU     *APU
	Mixer   *APUMixer
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
}

type APUTriangle struct {
	LengthCounter       APULengthCounter
	Timer               APUTimer
	LinearCounter       uint8
	LinearCounterReload uint8
	LinearReload        bool
	LinearCtrl          bool
	Pos                 uint8
}

type APUNoise struct {
	Envelope       APUEnveloppe
	Timer          APUTimer
	ShitftRegister uint16
	Mode           bool
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
