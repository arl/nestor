package main

import (
	"errors"
	"testing"
)

func TestMemMap(t *testing.T) {
	mm := MemMap{}
	buf := make([]byte, 0x0800)
	buf[0x12] = 0x34
	if err := mm.MapSlice(0x0000, 0x07FF, buf); err != nil {
		t.Fatal(err)
	}

	if val := mm.Read8(0x0012); val != 0x34 {
		t.Fatalf("Read8() = 0x%x, want 0x%x", val, 0x34)
	}

	mm.Write8(0x0012, 0x23)
	if val := mm.Read8(0x0012); val != 0x23 {
		t.Fatalf("Read8() = 0x%x, want 0x%x", val, 0x23)
	}
}

func TestMemMapOverlappingRange(t *testing.T) {
	mm := MemMap{}
	if err := mm.MapSlice(0x0000, 0x07FF, make([]byte, 0x0800)); err != nil {
		t.Fatal(err)
	}
	if err := mm.MapSlice(0x0050, 0x0051, make([]byte, 0x1)); !errors.Is(err, ErrOverlappingRange) {
		t.Fatal(err)
	}
}

func TestMirroredRam(t *testing.T) {
	mm := MemMap{}
	buf := make([]byte, 0x0800)
	if err := mm.MapSlice(0x0000, 0x07FF, buf); err != nil {
		t.Fatal(err)
	}

	// Add 3 mirrors.
	if err := mm.MapSlice(0x0800, 0x0FFF, buf); err != nil {
		t.Fatal(err)
	}
	if err := mm.MapSlice(0x1000, 0x17FF, buf); err != nil {
		t.Fatal(err)
	}
	if err := mm.MapSlice(0x1800, 0x1FFF, buf); err != nil {
		t.Fatal(err)
	}

	mm.Write8(0x0000, 0xFF)
	if val := mm.Read8(0x0000); val != 0xFF {
		t.Fatalf("Read8() = 0x%x, want 0x%x", val, 0xFF)
	}
	if val := mm.Read8(0x0800); val != 0xFF {
		t.Fatalf("Read8() = 0x%x, want 0x%x", val, 0xFF)
	}
	if val := mm.Read8(0x1000); val != 0xFF {
		t.Fatalf("Read8() = 0x%x, want 0x%x", val, 0xFF)
	}
	if val := mm.Read8(0x1800); val != 0xFF {
		t.Fatalf("Read8() = 0x%x, want 0x%x", val, 0xFF)
	}
}
