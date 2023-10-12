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
func (mmap *MemMap) MapSlice(addr, end uint32, buf []byte) error {
	return mmap.addrs.InsertRange(addr, end, &memRegion{
		buf:   buf,
		start: addr,
		vsize: end - addr + 1,
	})
}

type memRegion struct {
	buf   []byte // mapped buffer
	start uint32 // start address
	vsize uint32 // virtual size (size of the mapped range)
}

func (mr *memRegion) Read8(addr uint32) uint8 {
	return mr.buf[addr-mr.start]
}

func (mr *memRegion) Write8(addr uint32, val uint8) {
	mr.buf[addr-mr.start] = val
}
