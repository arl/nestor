package main

import (
	"testing"
)

func TestPflag(t *testing.T) {
	p := P(0x34)
	p.clear()
	if p != 0x40 {
		t.Errorf("got P = %q, want %q", p.String(), P(0x40))
	}

	p |= 1 << pbitI
	if p != 0x44 {
		t.Errorf("got P = %q, want %q", p.String(), P(0x44))
	}

	p |= 1 << pbitB
	if p != 0x54 {
		t.Errorf("got P = %q, want %q", p.String(), P(0x54))
	}

	// Negative flag
	p.maybeSetN(0xff)
	if !p.N() {
		t.Error("N bit should be set")
	}
	p.maybeSetN(0x7f)
	if p.N() {
		t.Error("N bit should not be set")
	}
	p.maybeSetN(0x80)
	if !p.N() {
		t.Error("N bit should be set")
	}

	// Zero flag
	p.maybeSetZ(0)
	if !p.Z() {
		t.Error("Z bit should be set")
	}

	p.maybeSetZ(1)
	if p.Z() {
		t.Error("Z bit should not be set")
	}

	p.maybeSetZ(0xff)
	if p.Z() {
		t.Error("Z bit should not be set")
	}
}
