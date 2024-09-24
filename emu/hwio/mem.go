package hwio

import (
	"unsafe"

	"nestor/emu/log"
)

// 16-bit / 32-bit access to memory with correct unalignment
// This is the main structure used for linear memory access, and
// should be used by default for most memory areas.
// We use this structure by pointer rather than by value because
// it is stored as BankIO interface within Table, and checking if
// a concrete pointer type is behind the interface is faster than
// checking a non-pointer type.
type memUnalignedLE struct {
	ptr  unsafe.Pointer
	mask uint16
	wcb  func(uint16, int)
	ro   uint8 // 0: read/write, 1: readonly, 2: silent readonly (no log)
}

func newMemUnalignedLE(mem []byte, wcb func(uint16, int), roflag uint8) *memUnalignedLE {
	if len(mem)&(len(mem)-1) != 0 {
		panic("memory buffer size is not pow2")
	}
	return &memUnalignedLE{
		ptr:  unsafe.Pointer(&mem[0]),
		mask: uint16(len(mem) - 1),
		wcb:  wcb,
		ro:   roflag,
	}
}

func (m *memUnalignedLE) FetchPointer(addr uint16) []uint8 {
	off := uintptr(addr & m.mask)
	buf := (*[1 << 30]uint8)(unsafe.Pointer(uintptr(m.ptr) + off))
	len := m.mask + 1 - uint16(off)
	return buf[:len:len]
}

func (m *memUnalignedLE) Read8(addr uint16) uint8 {
	off := uintptr(addr & m.mask)
	return *(*uint8)(unsafe.Pointer(uintptr(m.ptr) + off))
}

func (m *memUnalignedLE) Peek8(addr uint16) uint8 {
	return m.Read8(addr)
}

func (m *memUnalignedLE) Write8CheckRO(addr uint16, val uint8) bool {
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

func (m *memUnalignedLE) Write8(addr uint16, val uint8) {
	if !m.Write8CheckRO(addr, val) {
		log.ModHwIo.ErrorZ("Write8 to readonly memory").
			Hex8("val", val).
			Hex16("addr", addr).
			End()
	}
}

type MemFlags int

const (
	// TODO(arl) remove since it's now the default
	MemFlag8         MemFlags = (1 << iota) // 8-bit access is allowed
	MemFlag8ReadOnly                        // 8-bit accesses are read-only (requires MemFlag8)
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

func (mem *Mem) roFlag(robit MemFlags) uint8 {
	var roflag uint8
	if mem.Flags&robit != 0 {
		if mem.Flags&MemFlagNoROLog != 0 {
			roflag = 2
		} else {
			roflag = 1
		}
	}
	return roflag
}

func (mem *Mem) BankIO8() BankIO8 {
	if mem.Flags&MemFlag8 == 0 {
		return nil
	}
	roflag := mem.roFlag(MemFlag8ReadOnly)
	return newMemUnalignedLE(mem.Data, mem.WriteCb, roflag)
}
