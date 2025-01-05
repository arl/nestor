package hwio_test

import (
	"bytes"
	"testing"

	"nestor/hw/hwio"
)

// Unmapped
type openbus struct{}

func (ob *openbus) Read8(addr uint16) uint8       { return 0xD3 }
func (ob *openbus) Peek8(addr uint16) uint8       { return 0xD4 }
func (ob *openbus) Write8(addr uint16, val uint8) {}

type testTable struct {
	t          testing.TB
	Bus        *hwio.Table
	RAM        hwio.Mem    `hwio:"bank=0,offset=0x0,size=0x800,vsize=0x2000"`
	Reg1       hwio.Reg8   `hwio:"bank=1,offset=0x1,rwmask=0xF0,rcb,reset=0x99"`
	Reg2       hwio.Reg8   `hwio:"bank=1,offset=0x2,rwmask=0xF0,readonly,pcb=PeekReg2"`
	DefaultDev hwio.Device `hwio:"bank=2,offset=0x0,size=0x100"`
	DEV        hwio.Device `hwio:"bank=2,offset=0x100,size=0x100,rcb,wcb"` // no peek-callback
	RoDEV      hwio.Device `hwio:"bank=2,offset=0x200,size=0x100,rcb,pcb,readonly"`
	WoDEV      hwio.Device `hwio:"bank=2,offset=0x300,size=0x100,wcb,writeonly"` // no peek-callback

	devval uint8
}

// $2001
func (tbl *testTable) ReadREG1(val uint8) uint8 {
	return tbl.Reg1.Value + 1
}

// $2002
func (tbl *testTable) PeekReg2(val uint8) uint8 {
	return 0x12
}

func (tbl *testTable) ReadDEV(addr uint16) uint8         { return 0xE1 }
func (tbl *testTable) WriteDEV(addr uint16, val uint8)   { tbl.devval = uint8(addr) & val }
func (tbl *testTable) ReadRODEV(addr uint16) uint8       { return 0xC5 }
func (tbl *testTable) PeekRODEV(addr uint16) uint8       { return 0xC8 }
func (tbl *testTable) WriteWODEV(addr uint16, val uint8) { tbl.devval = uint8(addr) & ^val }

func newTestTable(tb testing.TB) *testTable {
	tbl := &testTable{t: tb}
	hwio.MustInitRegs(tbl)

	tbl.Bus = hwio.NewTable("bus")
	tbl.Bus.MapBank(0x0000, tbl, 0)
	tbl.Bus.MapBank(0x2000, tbl, 1)
	tbl.Bus.MapBank(0x4000, tbl, 2)
	tbl.Bus.Unmapped = &openbus{}
	return tbl
}

func (tbl *testTable) wantRead8(addr uint16, want uint8) {
	tbl.t.Helper()

	if got := tbl.Bus.Read8(addr); got != want {
		tbl.t.Errorf("Read8(%04X) = %02X, want %02X", addr, got, want)
	}
}

func (tbl *testTable) Write8(addr uint16, val uint8) {
	tbl.Bus.Write8(addr, val)
}

func (tbl *testTable) wantPeek8(addr uint16, want uint8) {
	tbl.t.Helper()

	if got := tbl.Bus.Peek8(addr); got != want {
		tbl.t.Errorf("Peek8(%04X) = %02X, want %02X", addr, got, want)
	}
}

func TestTableMapMem(t *testing.T) {
	tbl := newTestTable(t)

	// Mem
	tbl.wantRead8(0x00, 0)
	tbl.Write8(0x00, 0x12)
	tbl.wantRead8(0x00, 0x12)
	tbl.wantRead8(0x800, 0x12)

	// Reg1
	tbl.wantRead8(0x2001, 0x9a)
	tbl.Write8(0x2001, 0xff)
	tbl.wantRead8(0x2001, 0xfa)
	tbl.Write8(0x2001, 0xF0)
	tbl.wantRead8(0x2001, 0xfa)
	tbl.Write8(0x2001, 0x0F)
	tbl.wantRead8(0x2001, 0x0A)

	// Reg2
	tbl.wantRead8(0x2002, 0x00)
	tbl.wantPeek8(0x2002, 0x12)
	tbl.Write8(0x2002, 0x9b)
	tbl.wantRead8(0x2002, 0x00)
	tbl.wantPeek8(0x2002, 0x12)

	// Unmapped
	tbl.wantRead8(0x2020, 0xd3)
	tbl.wantPeek8(0x2020, 0xd4)

	// MapMemorySlice
	rom := bytes.Repeat([]byte("\x12\x34"), 0x100)
	tbl.Bus.MapMemorySlice(0x3000, 0x3199, rom, true)

	tbl.wantRead8(0x3000, 0x12)
	tbl.wantRead8(0x3001, 0x34)
	tbl.wantRead8(0x3199, 0x34)
	tbl.wantRead8(0x3200, 0xd3) // unmapped

	// MapDevice
	tbl.wantRead8(0x4000, 0x00)
	tbl.wantPeek8(0x4000, 0x00)
	tbl.Write8(0x4000, 0xff)

	tbl.wantRead8(0x4100, 0xe1)
	tbl.wantPeek8(0x4100, 0x00)
	tbl.Write8(0x4120, 0x27)
	if tbl.devval != 0x20 {
		t.Errorf("devval = %02X, want 0x20", tbl.devval)
	}

	tbl.wantRead8(0x4200, 0xc5)
	tbl.wantPeek8(0x4200, 0xc8)
	tbl.Write8(0x4200, 0xff) // readonly
	if tbl.devval != 0x20 {
		t.Errorf("devval = %02X, want 0x27", tbl.devval)
	}

	tbl.wantRead8(0x4300, 0x00) // writeonly
	tbl.wantPeek8(0x4300, 0x00) // writeonly
	tbl.Write8(0x4355, 0x0f)
	if tbl.devval != 0x50 {
		t.Errorf("devval = %02X, want 0x27", tbl.devval)
	}
}
