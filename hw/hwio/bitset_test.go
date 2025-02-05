package hwio

import (
	"math/rand/v2"
	"testing"
)

func TestBitset(t *testing.T) {
	var b Bitset
	for i := range NumBits {
		if b.Test(uint(i)) {
			t.Fatalf("Bit %d is set", i)
		}
	}

	b.SetAll()
	for i := range NumBits {
		if !b.Test(uint(i)) {
			t.Fatalf("Bit %d is not set", i)
		}
	}

	b.Reset()
	for i := range NumBits {
		if b.Test(uint(i)) {
			t.Fatalf("Bit %d is set", i)
		}
	}

	for i := range NumBits {
		b.Set(uint(i))
		if !b.Test(uint(i)) {
			t.Fatalf("Bit %d is not set", i)
		}
		b.Clear(uint(i))
		if b.Test(uint(i)) {
			t.Fatalf("Bit %d is set", i)
		}
	}
}

func TestBitsetRanges(t *testing.T) {
	var b Bitset

	for range 10000 {
		start := rand.UintN(NumBits)
		end := rand.UintN(NumBits)
		if start > end {
			start, end = end, start
		}
		if start == end {
			if start == 0 {
				end++
			} else {
				start--
			}
		}

		b.Reset()
		b.SetRange(start, end)
		for i := range NumBits {
			ui := uint(i)
			if ui >= start && ui < end {
				if !b.Test(ui) {
					t.Fatalf("SetRange(%d, %d) but bit %d is not set", start, end, i)
				}
			} else {
				if b.Test(ui) {
					t.Fatalf("SetRange(%d, %d) but bit %d is set", start, end, i)
				}
			}
		}

		b.SetAll()
		b.ClearRange(start, end)
		for i := range NumBits {
			ui := uint(i)
			if ui >= start && ui < end {
				if b.Test(ui) {
					t.Fatalf("ClearRange(%d, %d) but bit %d is set", start, end, i)
				}
			} else {
				if !b.Test(ui) {
					t.Fatalf("ClearRange(%d, %d) but bit %d is not set", start, end, i)
				}
			}
		}
	}
}
