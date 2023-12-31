package hwio

import (
	"fmt"

	log "nestor/emu/logger"
)

type BankIO8 interface {
	Read8(addr uint16) uint8
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

type Table struct {
	Name string
	ws   int

	table8 radixTree
}

func NewTable(name string) *Table {
	t := new(Table)
	t.Name = name
	t.Reset()
	return t
}

func (t *Table) SetWaitStates(ws int) {
	t.ws = ws
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
	err := t.table8.InsertRange(addr, addr+size-1, io)
	if err != nil {
		panic(err)
	}
}

func (t *Table) MapReg8(addr uint16, io *Reg8) {
	t.mapBus8(addr, 1, io, false)
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

	b8 := mem.BankIO8()
	if b8 != nil {
		t.mapBus8(addr, uint16(mem.VSize), b8, false)
	}
}

func (t *Table) MapMemorySlice(addr, end uint16, mem []uint8, readonly bool) {
	log.ModHwIo.DebugZ("mapping slice").
		Hex16("addr", addr).
		Hex16("end", end).
		String("bus", t.Name).
		Bool("ro", readonly).
		End()

	flags := MemFlag8
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

func (t *Table) Read8(addr uint16) uint8 {
	io := t.table8.Search(addr)
	if io == nil {
		log.ModHwIo.ErrorZ("unmapped Read8").
			String("name", t.Name).
			Hex16("addr", addr).
			End()
		return 0
	}
	if mem, ok := io.(*memUnalignedLE); ok {
		return mem.Read8(addr)
	}
	return io.(BankIO8).Read8(addr)
}

func (t *Table) Write8(addr uint16, val uint8) {
	io := t.table8.Search(addr)
	if io == nil {
		log.ModHwIo.ErrorZ("unmapped Write8").
			String("name", t.Name).
			Hex16("addr", addr).
			Hex8("val", val).
			End()
		return
	}
	if mem, ok := io.(*memUnalignedLE); ok {
		// NOTE: we use the CheckRO format so that the success codepath
		// (that is, when the memory is read-write) is fully inlined and
		// requires no function call.
		ok := mem.Write8CheckRO(addr, val)
		if !ok {
			log.ModHwIo.ErrorZ("Write8 to ROM").
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
	if mem, ok := io.(*memUnalignedLE); ok {
		return mem.FetchPointer(addr)
	}
	return nil
}

func (t *Table) WaitStates() int {
	return t.ws
}
