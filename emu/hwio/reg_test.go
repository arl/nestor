package hwio

import "testing"

func TestReg8(t *testing.T) {
	r := Reg8{Value: 0x11, RoMask: 0xF0}

	if r.Read8(0) != 0x11 {
		t.Errorf("invalid read: %x", r.Read8(0))
	}
	if r.Read8(9999) != 0x11 {
		t.Errorf("invalid read with offset: %x", r.Read8(9999))
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

func TestReg8FliptBit(t *testing.T) {
	tests := []struct {
		v    uint8
		n    uint
		want uint8
	}{
		{0b00010000, 0, 0b00010001},
		{0b00010001, 0, 0b00010000},
		{0b00010000, 1, 0b00010010},
		{0b00010000, 4, 0b00000000},
		{0b00010000, 7, 0b10010000},
		{0b10010000, 7, 0b00010000},
	}

	for _, tt := range tests {
		reg := Reg8{Value: tt.v}
		if reg.FlipBit(tt.n); reg.Value != tt.want {
			t.Errorf("FlipBit(0b%08b, %d) = 0b%08b, want 0b%08b", tt.v, tt.n, reg.Value, tt.want)
		}
	}
}

func TestReg8GetBit(t *testing.T) {
	tests := []struct {
		v     uint8
		n     uint
		want  bool
		wanti uint8
	}{
		{0b00010000, 0, false, 0},
		{0b00010001, 0, true, 1},
		{0b00010000, 1, false, 0},
		{0b00010000, 4, true, 1},
		{0b00010000, 7, false, 0},
		{0b10010000, 7, true, 1},
	}

	for _, tt := range tests {
		reg := Reg8{Value: tt.v}
		if v := reg.GetBit(tt.n); v != tt.want {
			t.Errorf("GetBit(0b%08b, %d) = %t, want %t", tt.v, tt.n, v, tt.want)
		}
		if v := reg.GetBiti(tt.n); v != tt.wanti {
			t.Errorf("GetBiti(0b%08b, %d) = %d, want %d", tt.v, tt.n, v, tt.wanti)
		}
	}
}

func TestReg8SetBit(t *testing.T) {
	tests := []struct {
		v    uint8
		n    uint
		want uint8
	}{
		{0b00010000, 0, 0b00010001},
		{0b00010001, 0, 0b00010001},
		{0b00010000, 4, 0b00010000},
		{0b10000001, 4, 0b10010001},
		{0b00010000, 7, 0b10010000},
		{0b10010000, 7, 0b10010000},
	}

	for _, tt := range tests {
		reg := Reg8{Value: tt.v}
		if reg.SetBit(tt.n); reg.Value != tt.want {
			t.Errorf("SetBit(0b%08b, %d) = 0b%08b, want 0b%08b", tt.v, tt.n, reg.Value, tt.want)
		}
	}
}

func TestReg8ClearBit(t *testing.T) {
	tests := []struct {
		v    uint8
		n    uint
		want uint8
	}{
		{0b00010000, 0, 0b00010000},
		{0b00010001, 0, 0b00010000},
		{0b00010000, 4, 0b00000000},
		{0b10000001, 4, 0b10000001},
		{0b00010000, 7, 0b00010000},
		{0b10010000, 7, 0b00010000},
	}

	for _, tt := range tests {
		reg := Reg8{Value: tt.v}
		if reg.ClearBit(tt.n); reg.Value != tt.want {
			t.Errorf("ClearBit(0b%08b, %d) = 0b%08b, want 0b%08b", tt.v, tt.n, reg.Value, tt.want)
		}
	}
}
