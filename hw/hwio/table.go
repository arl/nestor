package hwio

import (
	"fmt"

	"nestor/emu/log"
)

type BankIO8 interface {
	// Read8 reads a byte from the given address.
	Read8(addr uint16) uint8

	// Peek8  performs a read without side effects, for debugging/tracking.
	Peek8(addr uint16) uint8

	// Write8 writes a byte to the given address.
	Write8(addr uint16, val uint8)
}

func Write16(b BankIO8, addr uint16, val uint16) {
	lo := uint8(val & 0xff)
	hi := uint8(val >> 8)
	b.Write8(addr, lo)
	b.Write8(addr+1, hi)
}

func Read16(b BankIO8, addr uint16) uint16 {
	lo := b.Read8(addr)
	hi := b.Read8(addr + 1)
	return uint16(hi)<<8 | uint16(lo)
}

func Peek16(b BankIO8, addr uint16) uint16 {
	lo := b.Peek8(addr)
	hi := b.Peek8(addr + 1)
	return uint16(hi)<<8 | uint16(lo)
}

// LogUnmapped is a BankIO8 that logs all accesses to the console.
// meant to be used as Table.Unmapped for debugging.
type LogUnmapped struct {
	Name string
}

func (b *LogUnmapped) Peek8(addr uint16) uint8 {
	return 0
}

func (b *LogUnmapped) Read8(addr uint16) uint8 {
	log.ModHwIo.WarnZ("unmapped Read8").
		String("name", b.Name).
		Hex16("addr", addr).
		End()
	return 0
}

func (b *LogUnmapped) Write8(addr uint16, val uint8) {
	log.ModHwIo.WarnZ("unmapped Write8").
		String("name", b.Name).
		Hex16("addr", addr).
		Hex8("val", val).
		End()
}

type Table struct {
	table8 radixTree

	// if non-nil, unmapped accesses are forwarded to this device, allows for
	// easy open bus implementations.
	Unmapped BankIO8

	Name string
}

func NewTable(name string) *Table {
	t := new(Table)
	t.Name = name
	t.Reset()
	return t
}

func (t *Table) Reset() {
	t.table8 = radixTree{}
}

// Map a register bank (that is, a structure containing mulitple IoReg* fields).
// For this function to work, registers must have a struct tag "hwio", containing
// the following fields:
//
//	offset=0x12     Byte-offset within the register bank at which this
//	                register is mapped. There is no default value: if this
//	                option is missing, the register is assumed not to be
//	                part of the bank, and is ignored by this call.
//
//	bank=NN         Ordinal bank number (if not specified, default to zero).
//	                This option allows for a structure to expose multiple
//	                banks, as regs can be grouped by bank by specified the
//	                bank number.
func (t *Table) MapBank(addr uint16, bank any, bankNum int) {
	regs, err := bankGetRegs(bank, bankNum)
	if err != nil {
		panic(err)
	}

	for _, reg := range regs {
		switch r := reg.regPtr.(type) {
		case *Mem:
			t.MapMem(addr+reg.offset, r)
		case *Reg8:
			t.MapReg8(addr+reg.offset, r)
		case *Device:
			t.MapDevice(addr+reg.offset, r)
		default:
			panic(fmt.Errorf("invalid reg type: %T", r))
		}
	}
}

func (t *Table) UnmapBank(addr uint16, bank any, bankNum int) {
	regs, err := bankGetRegs(bank, bankNum)
	if err != nil {
		panic(err)
	}

	for _, reg := range regs {
		switch r := reg.regPtr.(type) {
		case *Mem:
			t.Unmap(addr+reg.offset, addr+reg.offset+uint16(r.VSize)-1)
		case *Reg8:
			t.Unmap(addr+reg.offset, addr+reg.offset+0)
		default:
			panic(fmt.Errorf("invalid reg type: %T", r))
		}
	}
}

func (t *Table) mapBus8(addr, size uint16, io BankIO8, allowremap bool) {
	_ = allowremap
	err := t.table8.InsertRange(addr, addr+size-1, io)
	if err != nil {
		panic(err)
	}
}

func (t *Table) MapReg8(addr uint16, io *Reg8) {
	t.mapBus8(addr, 1, io, false)
}

func (t *Table) MapDevice(addr uint16, io *Device) {
	t.mapBus8(addr, uint16(io.Size), io, false)
}

func (t *Table) MapMem(addr uint16, mem *Mem) {
	log.ModHwIo.DebugZ("mapping mem").
		Hex16("addr", addr).
		Hex16("size", uint16(mem.VSize)).
		String("area", mem.Name).
		String("bus", t.Name).
		End()

	if len(mem.Data)&(len(mem.Data)-1) != 0 {
		panic("memory buffer size is not pow2")
	}

	t.mapBus8(addr, uint16(mem.VSize), mem.BankIO8(), false)
}

func (t *Table) MapMemorySlice(addr, end uint16, mem []uint8, readonly bool) {
	log.ModHwIo.DebugZ("mapping slice").
		Hex16("addr", addr).
		Hex16("end", end).
		String("bus", t.Name).
		Bool("ro", readonly).
		End()

	var flags MemFlags
	if readonly {
		flags |= MemFlag8ReadOnly
	}
	t.MapMem(addr, &Mem{
		Data:  mem,
		Flags: flags,
		VSize: int(end - addr + 1),
	})
}

func (t *Table) Unmap(begin, end uint16) {
	t.table8.RemoveRange(begin, end)
}

// Read8 searches in the table for the device mapped at the given address and
// forward the read to it. Accesses to unmapped addresses are logged as errors
// if peek is false.
func (t *Table) Read8(addr uint16) uint8 {
	io := t.table8.Search(addr)
	if io == nil {
		if t.Unmapped != nil {
			return t.Unmapped.Read8(addr)
		}
		return 0
	}
	return io.(BankIO8).Read8(addr)
}

func (t *Table) Peek8(addr uint16) uint8 {
	io := t.table8.Search(addr)
	if io == nil {
		if t.Unmapped != nil {
			return t.Unmapped.Peek8(addr)
		}
		return 0
	}

	return io.(BankIO8).Peek8(addr)
}

func (t *Table) Write8(addr uint16, val uint8) {
	io := t.table8.Search(addr)
	if io == nil {
		if t.Unmapped != nil {
			t.Unmapped.Write8(addr, val)
		}
		return
	}
	if mem, ok := io.(*mem); ok {
		// NOTE: we use the CheckRO format so that the success codepath
		// (that is, when the memory is read-write) is fully inlined and
		// requires no function call.
		ok := mem.Write8CheckRO(addr, val)
		if !ok {
			log.ModHwIo.ErrorZ("Write8 to read-only address").
				String("name", t.Name).
				Hex16("addr", addr).
				Hex8("val", val).
				End()
		}
		return
	}
	io.(BankIO8).Write8(addr, val)
}

func (t *Table) FetchPointer(addr uint16) []uint8 {
	io := t.table8.Search(addr)
	if mem, ok := io.(*mem); ok {
		return mem.FetchPointer(addr)
	}
	return nil
}
