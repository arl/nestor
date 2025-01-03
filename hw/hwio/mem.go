package hwio

import (
	"unsafe"

	"nestor/emu/log"
)

// mem is the main structure used for linear memory access.
//
// We use this structure by pointer rather than by value because it is stored as
// BankIO interface within Table, and checking if a concrete pointer type is
// behind the interface is faster than checking a non-pointer type.
type mem struct {
	ptr  unsafe.Pointer
	mask uint16
	wcb  func(uint16, int)
	ro   uint8 // 0: read/write, 1: readonly, 2: silent readonly (no log)
}

func newMem(buf []byte, wcb func(uint16, int), roflag uint8) *mem {
	if len(buf)&(len(buf)-1) != 0 {
		panic("memory buffer size is not pow2")
	}
	return &mem{
		ptr:  unsafe.Pointer(&buf[0]),
		mask: uint16(len(buf) - 1),
		wcb:  wcb,
		ro:   roflag,
	}
}

func (m *mem) FetchPointer(addr uint16) []uint8 {
	off := uintptr(addr & m.mask)
	buf := (*[1 << 30]uint8)(unsafe.Pointer(uintptr(m.ptr) + off))
	len := m.mask + 1 - uint16(off)
	return buf[:len:len]
}

func (m *mem) Read8(addr uint16) uint8 {
	off := uintptr(addr & m.mask)
	return *(*uint8)(unsafe.Pointer(uintptr(m.ptr) + off))
}

func (m *mem) Peek8(addr uint16) uint8 {
	off := uintptr(addr & m.mask)
	return *(*uint8)(unsafe.Pointer(uintptr(m.ptr) + off))
}

func (m *mem) Write8CheckRO(addr uint16, val uint8) bool {
	off := uintptr(addr & m.mask)
	if m.ro == 0 {
		*(*uint8)(unsafe.Pointer(uintptr(m.ptr) + off)) = val
		if m.wcb != nil {
			m.wcb(addr, 1)
		}
		return true
	}
	return m.ro == 2 // fake success if we're in silent mode
}

func (m *mem) Write8(addr uint16, val uint8) {
	if !m.Write8CheckRO(addr, val) {
		log.ModHwIo.ErrorZ("Write8 to readonly memory").
			Hex8("val", val).
			Hex16("addr", addr).
			End()
	}
}

type MemFlags int

const (
	MemFlag8ReadOnly MemFlags = (1 << iota) // read-only accesses
	MemFlagNoROLog                          // skip logging attempts to write when configured to readonly
)

// Linear memory area that can be mapped into a Table.
//
// NOTE: this structure does not directly implement the BankIO interface for
// performance reasons. In fact, it would be inefficient to parse all the flags
// at runtime for each memory access to correctly implement it; so, clients must
// call the BankIO8 method to create adaptors that implement memory access
// depending on the memory bank configuration.
type Mem struct {
	Name    string            // name of the memory area (for debugging)
	Data    []byte            // actual memory buffer
	VSize   int               // virtual size of the memory (can be bigger than physical size)
	Flags   MemFlags          // flags determining how the memory can be accessed
	WriteCb func(uint16, int) // optional write callback (receives full address and number of bytes written)
}

func (m *Mem) roFlag(robit MemFlags) uint8 {
	var roflag uint8
	if m.Flags&robit != 0 {
		if m.Flags&MemFlagNoROLog != 0 {
			roflag = 2
		} else {
			roflag = 1
		}
	}
	return roflag
}

func (m *Mem) BankIO8() BankIO8 {
	roflag := m.roFlag(MemFlag8ReadOnly)
	return newMem(m.Data, m.WriteCb, roflag)
}

// Manual implements the BankIO8 interface for a manually managed memory area.
// TODO: finx a better name: Area maybe?
type Manual struct {
	Name    string // name of the memory area (for debugging)
	Size    int    // size of the memory area
	ReadCb  func(addr uint16) uint8
	PeekCb  func(addr uint16) uint8
	WriteCb func(addr uint16, val uint8)
}

func (m *Manual) Read8(addr uint16) uint8 {
	return m.ReadCb(addr)
}

func (m *Manual) Peek8(addr uint16) uint8 {
	return m.PeekCb(addr)
}

func (m *Manual) Write8(addr uint16, val uint8) {
	m.WriteCb(addr, val)
}
