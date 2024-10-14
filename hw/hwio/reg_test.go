package hwio

import "testing"

func TestReg8(t *testing.T) {
	r := Reg8{Value: 0x11, RoMask: 0xF0}

	if got := r.Read8(0, false); got != 0x11 {
		t.Errorf("invalid read: %x", got)
	}
	if got := r.Read8(9999, false); got != 0x11 {
		t.Errorf("invalid read with offset: %x", got)
	}

	r.Write8(0, 0x77)
	if r.Value != 0x17 {
		t.Errorf("writemask not respected: %x", r.Value)
	}
	r.Write8(9999, 0x88)
	if r.Value != 0x18 {
		t.Errorf("writemask with offset not respected: %x", r.Value)
	}
}
