// Package mappers provides the interface and implementations for NES mappers.
package mappers

import (
	"github.com/arl/nestor/emu/log"
)

var modMapper = log.NewModule("mapper")

var All = map[uint16]mapperDesc{
	0:  NROM,
	1:  MMC1,
	2:  UxROM,
	3:  CNROM,
	7:  AxROM,
	34: BNROM,
	66: GxROM,
}

type mapperDesc struct {
	Name            string
	Load            func(*base) (Mapper, error)
	PRGBankSize     uint32
	CHRBankSize     uint32
	HasBusConflicts func(*base) bool

	RegisterStart uint16 // defaults to 0x8000 if not set
	RegisterEnd   uint16 // defaults to 0xFFFF if not set
}
