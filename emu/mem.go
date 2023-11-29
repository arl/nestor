package emu

import (
	"log"
	"nestor/emu/hwio"
)

type Region8 interface {
	Read8(addr uint16) uint8
	Write8(addr uint16, val uint8)
}

type MemMap struct {
	addrs radixTree
}

func (mmap *MemMap) Read8(addr uint16) uint8 {
	r := mmap.addrs.Search(addr)
	if r == nil {
		log.Printf("read at unmapped address 0x%x", addr)
		return 0
	}
	return r.(Region8).Read8(addr)
}

func (mmap *MemMap) Write8(addr uint16, val uint8) {
	r := mmap.addrs.Search(addr)
	if r == nil {
		log.Printf("write at unmapped address 0x%x", addr)
		return
	}
	r.(Region8).Write8(addr, val)
}

func (mmap *MemMap) Reset() {
	mmap.addrs = radixTree{}
}

// MapSlice maps a slice at a given range.
func (mmap *MemMap) MapSlice(addr, end uint16, buf []byte) {
	if len(buf)&(len(buf)-1) != 0 {
		panic("mapped buffer size must be a power of 2")
	}
	if err := mmap.addrs.InsertRange(addr, end, &MemRegion{
		Buf:   buf,
		mask:  uint16(len(buf) - 1),
		VSize: int(end - addr + 1),
	}); err != nil {
		panic(err)
	}
}

// MapReg8 maps an 8-bit register at a given address.
func (mmap *MemMap) MapReg8(addr uint16, r8 *hwio.Reg8) {
	mmap.mapBus8(addr, 1, r8)
}

func (mmap *MemMap) mapBus8(addr uint16, size uint16, r8 Region8) {
	err := mmap.addrs.InsertRange(addr, addr+size-1, r8)
	if err != nil {
		panic(err)
	}
}

type MemRegion struct {
	Buf   []byte // mapped buffer
	VSize int    // virtual size (size of the mapped range)
	mask  uint16
}

func (mr *MemRegion) Read8(addr uint16) uint8 {
	return mr.Buf[addr&mr.mask]
}

func (mr *MemRegion) Write8(addr uint16, val uint8) {
	mr.Buf[addr&mr.mask] = val
}
