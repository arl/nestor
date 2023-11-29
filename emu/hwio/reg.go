package hwio

import (
	"fmt"

	log "nestor/emu/logger"
)

type RegFlags uint8

const (
	RegFlagReadOnly RegFlags = (1 << iota)
	RegFlagWriteOnly
)

type Reg8 struct {
	Name   string
	Value  uint8
	RoMask uint8

	Flags   RegFlags
	ReadCb  func(val uint8) uint8
	WriteCb func(old uint8, val uint8)
}

func (reg Reg8) String() string {
	s := fmt.Sprintf("%s{%02x", reg.Name, reg.Value)
	if reg.ReadCb != nil {
		s += ",r!"
	}
	if reg.WriteCb != nil {
		s += ",w!"
	}
	return s + "}"
}

func (reg *Reg8) write(val uint8, romask uint8) {
	romask = romask | reg.RoMask
	old := reg.Value
	reg.Value = (reg.Value & romask) | (val &^ romask)
	if reg.WriteCb != nil {
		reg.WriteCb(old, reg.Value)
	}
}

func hex16(val uint16) string {
	return fmt.Sprintf("%04x", val)
}

func (reg *Reg8) Write8(addr uint16, val uint8) {
	if reg.Flags&RegFlagReadOnly != 0 {
		log.ModHwIo.ErrorZ("invalid Write8 from readonly reg").
			String("name", reg.Name).
			Hex16("addr", addr)
		return
	}
	reg.write(val, 0)
}

func (reg *Reg8) Read8(addr uint16) uint8 {
	if reg.Flags&RegFlagWriteOnly != 0 {
		log.ModHwIo.ErrorZ("invalid Read8 from writeonly reg").
			String("name", reg.Name).
			Hex16("addr", addr)
		return 0
	}
	if reg.ReadCb != nil {
		return reg.ReadCb(reg.Value)
	}
	return reg.Value
}
