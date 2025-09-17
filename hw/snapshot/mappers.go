package snapshot

//go:generate go tool msgp -tests=false -marshal=true

type BaseState struct {
	// Currently mapped ROM banks (we only need bank numbers, not data)
	// The actual ROM data is immutable and stored in b.rom
	// PRGROM [0x8000]byte

	// RAM data (always save complete contents as they're writable)
	PRGRAM     []byte
	PRGNVRAM   []byte
	Nametables [0x800]byte

	// CHR data - could be ROM or RAM
	CHRROM []byte // Always save as it could be CHR-RAM
}

type NROMState struct {
	*BaseState
}

type CNROMState struct {
	*BaseState
	CHRBank      uint32
	BusConflicts bool
}

type AxROMState struct {
	*BaseState
	NTM          uint8
	PRGBank      uint32
	BusConflicts bool
}

type MMC1State struct {
	*BaseState
	PrevCycle   int64
	Serial      uint8
	Counter     uint8
	NT          uint8
	PRGMode     uint8
	CHRMode     uint8
	CHRBank0    uint32
	CHRBank1    uint32
	LastCHR     uint16
	PRGBank     uint32
	DisableWRAM bool
}

type GxROMState struct {
	*BaseState
	CHRBank uint32
	PRGBank uint32
}

type UxROMState struct {
	*BaseState
	PRGBank      uint32
	BankMask     uint8
	BusConflicts bool
}
