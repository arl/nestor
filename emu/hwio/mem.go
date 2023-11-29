package hwio

import (
	"fmt"

	log "nestor/emu/logger"
)

type Region8 interface {
	Read8(addr uint16) uint8
	Write8(addr uint16, val uint8)
}

type MemMap struct {
	Name  string
	addrs radixTree
}

func (mmap *MemMap) Read8(addr uint16) uint8 {
	r := mmap.addrs.Search(addr)
	if r == nil {
		log.ModHwIo.ErrorZ("unmapped Read8").
			String("name", mmap.Name).
			Hex16("addr", addr).
			End()
		return 0
	}
	return r.(Region8).Read8(addr)
}

func (mmap *MemMap) Write8(addr uint16, val uint8) {
	r := mmap.addrs.Search(addr)
	if r == nil {
		log.ModHwIo.ErrorZ("unmapped Write8").
			String("name", mmap.Name).
			Hex16("addr", addr).
			Hex8("val", val).
			End()
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
func (mmap *MemMap) MapReg8(addr uint16, r8 *Reg8) {
	mmap.mapBus8(addr, 1, r8)
}

func (mmap *MemMap) mapBus8(addr uint16, size uint16, r8 Region8) {
	err := mmap.addrs.InsertRange(addr, addr+size-1, r8)
	if err != nil {
		panic(err)
	}
}

func (mmap *MemMap) MapBank(addr uint16, bank any, bankNum int) {
	regs, err := bankGetRegs(bank, bankNum)
	if err != nil {
		panic(err)
	}

	for _, reg := range regs {
		switch r := reg.regPtr.(type) {
		case *MemRegion:
			mmap.mapBus8(addr, uint16(r.VSize), r)
		case *Reg8:
			mmap.MapReg8(addr+uint16(reg.offset), r)
		default:
			panic(fmt.Errorf("invalid reg type: %T", r))
		}
	}
}

func (mmap *MemMap) UnmapBank(addr uint16, bank any, bankNum int) {
	regs, err := bankGetRegs(bank, bankNum)
	if err != nil {
		panic(err)
	}

	for _, reg := range regs {
		switch r := reg.regPtr.(type) {
		case *MemRegion:
			mmap.Unmap(addr+uint16(reg.offset), addr+uint16(reg.offset)+uint16(r.VSize)-1)
		case *Reg8:
			mmap.Unmap(addr+uint16(reg.offset), addr+uint16(reg.offset)+0)
		default:
			panic(fmt.Errorf("invalid reg type: %T", r))
		}
	}
}

func (mmap *MemMap) Unmap(begin uint16, end uint16) {
	mmap.addrs.RemoveRange(begin, end)
}

type MemFlags int

const (
	MemFlag8             MemFlags = (1 << iota) // 8-bit access is allowed
	MemFlag16Unaligned                          // 16-bit access is allowed, even if unaligned
	MemFlag16ForceAlign                         // 16-bit access is allowed, and it is forcibly aligned to 16-bit boundary
	MemFlag16Byteswapped                        // 16-bit access is allowed, and if not aligned the data is byteswapped
	MemFlag32Unaligned                          // 32-bit access is allowed, even if unaligned
	MemFlag32ForceAlign                         // 32-bit access is allowed, and it is forcibly aligned to 32-bit boundary
	MemFlag32Byteswapped                        // 32-bit access is allowed, and if not aligned the data is byteswapped
	MemFlag8ReadOnly                            // 8-bit accesses are read-only (requires MemFlag8)
	MemFlag16ReadOnly                           // 16-bit accesses are read-only (requires one of MemFlag16*)
	MemFlag32ReadOnly                           // 32-bit accesses are read-only (requires one of MemFlag32*)
	MemFlagNoROLog                              // skip logging attempts to write when configured to readonly

	MemFlagReadOnly = MemFlag8ReadOnly | MemFlag16ReadOnly | MemFlag32ReadOnly // all writes are forbidden
)

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
