package hwio

import "nestor/emu/log"

// Device is a BankIO8 implementation that allows manual management of an entire
// range of memory.
type Device struct {
	Name  string // name of the memory area (for debugging)
	Size  int    // size of the memory area
	Flags RWFlags

	ReadCb  func(addr uint16) uint8
	PeekCb  func(addr uint16) uint8
	WriteCb func(addr uint16, val uint8)
}

func (m *Device) Read8(addr uint16) uint8 {
	switch {
	case m.Flags&WriteOnlyFlag != 0:
		log.ModHwIo.ErrorZ("invalid Read8 from writeonly device").
			String("name", m.Name).
			Hex16("addr", addr).
			End()
		fallthrough
	case m.ReadCb == nil:
		return 0
	}
	return m.ReadCb(addr)
}

func (m *Device) Peek8(addr uint16) uint8 {
	if m.PeekCb != nil {
		return m.PeekCb(addr)
	}
	return 0
}

func (m *Device) Write8(addr uint16, val uint8) {
	switch {
	case m.Flags&ReadOnlyFlag != 0:
		log.ModHwIo.ErrorZ("invalid Write8 to readonly device").
			String("name", m.Name).
			Hex16("addr", addr).
			End()
		fallthrough
	case m.WriteCb == nil:
		return
	}

	m.WriteCb(addr, val)
}

func nopRead8(_ uint16) uint8     { return 0 }
func nopPeek8(_ uint16) uint8     { return 0 }
func nopWrite8(_ uint16, _ uint8) {}
