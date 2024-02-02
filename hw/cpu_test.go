package hw

import "testing"

func TestPflag(t *testing.T) {
	p := P(0x40)
	p = p.SetIntDisable(true)
	if p != 0x44 {
		t.Errorf("got P = %q, want %q", p.String(), P(0x44))
	}

	p = p.SetBreak(true)
	if p != 0x54 {
		t.Errorf("got P = %q, want %q", p.String(), P(0x54))
	}

	// Negative flag
	p.checkN(0xff)
	if !p.Negative() {
		t.Error("N bit should be set")
	}
	p.checkN(0x7f)
	if p.Negative() {
		t.Error("N bit should not be set")
	}
	p.checkN(0x80)
	if !p.Negative() {
		t.Error("N bit should be set")
	}

	// Zero flag
	p.checkZ(0)
	if !p.Zero() {
		t.Error("Z bit should be set")
	}

	p.checkZ(1)
	if p.Zero() {
		t.Error("Z bit should not be set")
	}

	p.checkZ(0xff)
	if p.Zero() {
		t.Error("Z bit should not be set")
	}
}

func TestPString(t *testing.T) {
	p := P(0b00110100)
	if got := p.String(); got != "nvUBdIzc" {
		t.Errorf("got P = %s, want %s", got, "nvUBdIzc")
	}
	p = P(0b00000100)
	if p.String() != "nvubdIzc" {
		t.Errorf("got P = %s, want %s", p.String(), "nvubdIzc")
	}
}
