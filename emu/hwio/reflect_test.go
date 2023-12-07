package hwio

import "testing"

type test1 struct {
	Reg1   Reg8 `hwio:"offset=0x111,reset=0x23,rwmask=0x1,wcb"`
	Reg2   Reg8 `hwio:"offset=0x444,bank=1,rcb"`
	called bool
}

func (t *test1) WriteREG1(old, val uint8) {
	t.called = true
}

func (t *test1) ReadREG2(val uint8) uint8 {
	return val | 1
}

func TestReflect(t *testing.T) {
	ts := &test1{}

	err := InitRegs(ts)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ts)
	if ts.Reg1.Name != "Reg1" || ts.Reg2.Name != "Reg2" {
		t.Error("invalid names:", ts.Reg1, ts.Reg2)
	}

	if ts.Reg2.Read8(0) != 1 {
		t.Error("invalid read8:", ts.Reg2.Read8(0))
	}

	val := ts.Reg1.Read8(0)
	if val != 0x23 {
		t.Error("invalid read8", val)
	}

	ts.Reg1.Write8(0, 0)
	if ts.Reg1.Value != 0x22 {
		t.Error("invalid read after rwmask", ts.Reg1.Value)
	}
	if !ts.called {
		t.Error("callback not called")
	}
}

func TestParseBank(t *testing.T) {
	ts := &test1{}
	info, err := bankGetRegs(ts, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(info) != 1 {
		t.Fatal("wrong number of regs in bank:", len(info))
	}
	if info[0].offset != 0x111 {
		t.Errorf("invalid reg offset: %x", info[0].offset)
	}

	rptr, ok := info[0].regPtr.(*Reg8)
	if !ok {
		t.Errorf("invalid reg ptr type: %T", info[0].regPtr)
	} else if rptr != &ts.Reg1 {
		t.Errorf("invalid reg ptr")
	}

	info, err = bankGetRegs(ts, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(info) != 1 {
		t.Fatal("wrong number of regs in bank:", len(info))
	}
	if info[0].offset != 0x444 {
		t.Errorf("invalid reg offset: %x", info[0].offset)
	}
}

func TestReadWriteOnly(t *testing.T) {
	type test2 struct {
		Reg1 Reg8 `hwio:"reset=0x23,readonly"`
		Reg2 Reg8 `hwio:"writeonly"`
	}

	ts := &test2{}
	err := InitRegs(ts)
	if err != nil {
		t.Fatal(err)
	}

	ts.Reg1.Write8(0, 0) // this should be ignored
	if ts.Reg1.Read8(0) != 0x23 {
		t.Error("invalid reg1 read:", ts.Reg1.Read8(0))
	}

	ts.Reg2.Write8(0, 0x23)
	if ts.Reg2.Read8(0) != 0 {
		t.Error("invalid reg2 read:", ts.Reg2.Read8(0))
	}
}

func TestValuesTooBig(t *testing.T) {
	type test3 struct {
		R Reg8 `hwio:"reset=0x123"`
	}
	type test4 struct {
		R Reg8 `hwio:"rwmask=0x123"`
	}

	ts := &test3{}
	err := InitRegs(ts)
	if err == nil {
		t.Fatal("initregs should fail")
	}

	ts2 := &test4{}
	err = InitRegs(ts2)
	if err == nil {
		t.Fatal("initregs should fail")
	}
}
