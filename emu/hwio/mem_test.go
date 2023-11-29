package hwio

import (
	"errors"
	"testing"
)

func mustntPanic(t *testing.T, f func()) {
	defer func() {
		t.Helper()

		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %q", r)
		}
	}()
	f()
}

func mustPanicWith(t *testing.T, err error, f func()) {
	defer func() {
		t.Helper()

		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		rerr, ok := r.(error)
		if !ok {
			t.Fatalf("expected error, but panicked with %q", r)
		}
		if !errors.Is(rerr, err) {
			t.Fatalf("expected error %q, but panicked with %q", err, rerr)
		}
	}()
	f()
}

func TestMemMap(t *testing.T) {
	mm := MemMap{}
	buf := make([]byte, 0x0800)
	buf[0x12] = 0x34

	mustntPanic(t, func() { mm.MapSlice(0x0000, 0x07FF, buf) })

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
	mustntPanic(t, func() {
		mm.MapSlice(0x0000, 0x07FF, make([]byte, 0x0800))
	})
	mustPanicWith(t, ErrOverlappingRange, func() {
		mm.MapSlice(0x0050, 0x0051, make([]byte, 0x1))
	})
}

func TestMirroredRam(t *testing.T) {
	mm := MemMap{}
	buf := make([]byte, 0x0800)
	mustntPanic(t, func() { mm.MapSlice(0x0000, 0x07FF, buf) })

	// Add 3 mirrors.
	mustntPanic(t, func() { mm.MapSlice(0x0800, 0x0FFF, buf) })
	mustntPanic(t, func() { mm.MapSlice(0x1000, 0x17FF, buf) })
	mustntPanic(t, func() { mm.MapSlice(0x1800, 0x1FFF, buf) })

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
