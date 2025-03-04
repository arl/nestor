package hwio

import "fmt"

const (
	NumBits  = 0x10000            // NES addressing space is 64K
	wordSize = 64                 // using 64-bit words
	numWords = NumBits / wordSize // 1024 words exactly
)

// Bitset is a 64Kbit set. Zero value is an empty set (all bits cleared).
type Bitset struct {
	words [numWords]uint64
}

// Set sets the bit at index i.
func (b *Bitset) Set(i uint) {
	b.words[i/wordSize] |= 1 << (i % wordSize)
}

// Clear clears the bit at index i.
func (b *Bitset) Clear(i uint) {
	b.words[i/wordSize] &^= 1 << (i % wordSize)
}

// Test returns true if the bit at index i is set.
func (b *Bitset) Test(i uint) bool {
	return (b.words[i/wordSize] & (1 << (i % wordSize))) != 0
}

// SetRange sets all bits in the half-open interval [start, end).
// It panics if start >= end or end > NumBits.
func (b *Bitset) SetRange(start, end uint) {
	if start >= end || end > NumBits {
		panic(fmt.Sprintf("invalid range [%d, %d)", start, end))
	}
	startWord := start / wordSize
	endWord := (end - 1) / wordSize
	startBit := start % wordSize
	endBit := (end - 1) % wordSize

	if startWord == endWord {
		mask := ((uint64(1) << (endBit - startBit + 1)) - 1) << startBit
		b.words[startWord] |= mask
		return
	}

	// First word.
	b.words[startWord] |= ^uint64(0) << startBit

	// Middle full words.
	for i := startWord + 1; i < endWord; i++ {
		b.words[i] = ^uint64(0)
	}

	// Last word.
	mask := (uint64(1) << (endBit + 1)) - 1
	b.words[endWord] |= mask
}

// ClearRange clears all bits in the half-open interval [start, end).
// It panics if start >= end or end > NumBits.
func (b *Bitset) ClearRange(start, end uint) {
	if start >= end || end > NumBits {
		panic(fmt.Sprintf("invalid range [%d, %d)", start, end))
	}
	startWord := start / wordSize
	endWord := (end - 1) / wordSize
	startBit := start % wordSize
	endBit := (end - 1) % wordSize

	if startWord == endWord {
		mask := ((uint64(1) << (endBit - startBit + 1)) - 1) << startBit
		b.words[startWord] &^= mask
		return
	}

	// First word.
	b.words[startWord] &^= ^uint64(0) << startBit

	// Middle full words.
	for i := startWord + 1; i < endWord; i++ {
		b.words[i] = 0
	}

	// Last word
	mask := (uint64(1) << (endBit + 1)) - 1
	b.words[endWord] &^= mask
}

// Reset clears all bits in the Bitset.
func (b *Bitset) Reset() {
	for i := range b.words {
		b.words[i] = 0
	}
}

// SetAll sets all bits in the Bitset.
func (b *Bitset) SetAll() {
	for i := range b.words {
		b.words[i] = ^uint64(0)
	}
}
