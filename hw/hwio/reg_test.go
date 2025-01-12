package hwio

import "testing"

func TestReg8(t *testing.T) {
	r := Reg8{Value: 0x31, RoMask: 0x1F}

	if got := r.Read8(0); got != 0x31 {
		t.Errorf("invalid read: %x", got)
	}
	if got := r.Read8(9999); got != 0x31 {
		t.Errorf("invalid read with offset: %x", got)
	}

	r.Write8(0, 0x17)
	if r.Value != 0x11 {
		t.Errorf("romask not respected: %x", r.Value)
	}
	r.Write8(9999, 0x3f)
	if r.Value != 0x31 {
		t.Errorf("romask not respected: %x", r.Value)
	}
}
