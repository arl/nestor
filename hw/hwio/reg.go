package hwio

import (
	"fmt"

	"nestor/emu/log"
)

type RWFlags uint8

const (
	ReadWriteFlag RWFlags = 0
	ReadOnlyFlag  RWFlags = (1 << iota)
	WriteOnlyFlag
)

type Reg8 struct {
	Name   string
	Value  uint8
	RoMask uint8

	Flags   RWFlags
	ReadCb  func(val uint8) uint8
	PeekCb  func(val uint8) uint8
	WriteCb func(old uint8, val uint8)
}

func (reg Reg8) String() string {
	s := fmt.Sprintf("%s{%02x", reg.Name, reg.Value)
	if reg.ReadCb != nil {
		s += ",r!"
	}
	if reg.PeekCb != nil {
		s += ",p!"
	}
	if reg.WriteCb != nil {
		s += ",w!"
	}
	return s + "}"
}

func (reg *Reg8) write(val uint8) {
	old := reg.Value
	reg.Value = (reg.Value & reg.RoMask) | (val &^ reg.RoMask)
	if reg.WriteCb != nil {
		reg.WriteCb(old, reg.Value)
	}
}

func (reg *Reg8) Write8(addr uint16, val uint8) {
	if reg.Flags&ReadOnlyFlag != 0 {
		log.ModHwIo.ErrorZ("invalid Write8 to readonly reg").
			String("name", reg.Name).
			Hex16("addr", addr).
			End()
		return
	}
	reg.write(val)
}

func (reg *Reg8) Read8(addr uint16) uint8 {
	if reg.Flags&WriteOnlyFlag != 0 {
		log.ModHwIo.ErrorZ("invalid Read8 from writeonly reg").
			String("name", reg.Name).
			Hex16("addr", addr).
			End()
		return 0
	}
	if reg.ReadCb != nil {
		return reg.ReadCb(reg.Value)
	}
	return reg.Value
}

func (reg *Reg8) Peek8(addr uint16) uint8 {
	if reg.PeekCb != nil {
		return reg.PeekCb(reg.Value)
	}
	return reg.Value
}
