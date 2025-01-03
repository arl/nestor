package hwio_test

import (
	"testing"

	"nestor/hw/hwio"
)

type testTable struct {
	t testing.TB
	*hwio.Table
	RAM  hwio.Mem  `hwio:"bank=0,offset=0x0,size=0x800,vsize=0x2000"`
	Reg1 hwio.Reg8 `hwio:"bank=1,offset=0x1,rwmask=0xF0,rcb,reset=0x99"`
}

// $2001
func (tbl *testTable) ReadREG1(val uint8) uint8 {
	tbl.Reg1.Value++
	return tbl.Reg1.Value
}

func newTestTable(tb testing.TB) *testTable {
	tbl := &testTable{t: tb, Table: hwio.NewTable("bus")}
	hwio.MustInitRegs(tbl)
	tbl.Table.MapBank(0x0000, tbl, 0)
	tbl.Table.MapBank(0x2000, tbl, 1)
	return tbl
}

func (tbl *testTable) wantRead8(addr uint16, want uint8) {
	if got := tbl.Read8(addr); got != want {
		tbl.t.Errorf("Read8(%04X) = %02X, want %02X", addr, got, want)
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
	tbl.wantRead8(0x2001, 0x9A)
	tbl.wantRead8(0x2001, 0x9B)
	tbl.Write8(0x2001, 0xFF)
	tbl.wantRead8(0x2001, 0xA0)
}
