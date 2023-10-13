package main

import "log"

// Range8 is an 8-bit adressable component, such as an hardware component, a
// memory region, etc.
type Region8 interface {
	Read8(addr uint32) uint8
	Write8(addr uint32, val uint8)
}

type MemMap struct {
	addrs radixTree
}

func (mmap *MemMap) Read8(addr uint32) uint8 {
	r := mmap.addrs.Search(addr)
	if r == nil {
		log.Printf("read at unmapped address 0x%x", addr)
		return 0
	}
	return r.(Region8).Read8(addr)
}

func (mmap *MemMap) Write8(addr uint32, val uint8) {
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
func (mmap *MemMap) MapSlice(addr, end uint32, buf []byte) {
	if len(buf)&(len(buf)-1) != 0 {
		panic("mapped buffer size must be a power of 2")
	}
	if err := mmap.addrs.InsertRange(addr, end, &MemRegion{
		Buf:   buf,
		mask:  uint32(len(buf) - 1),
		VSize: int(end - addr + 1),
	}); err != nil {
		panic(err)
	}
}

type MemRegion struct {
	Buf   []byte // mapped buffer
	VSize int    // virtual size (size of the mapped range)
	mask  uint32
}

func (mr *MemRegion) Read8(addr uint32) uint8 {
	return mr.Buf[addr&mr.mask]
}

func (mr *MemRegion) Write8(addr uint32, val uint8) {
	mr.Buf[addr&mr.mask] = val
}
