package hw

import "unsafe"

func nthbit8(val uint8, n uint8) uint8    { return (val >> n) & 1 }
func nthbit16(val uint16, n uint8) uint16 { return (val >> n) & 1 }

func u8tob(v uint8) bool { return v != 0 }

// Avoid branches. In the SSA compiler, this compiles to
// exactly what you would want it to.

func b2u8(x bool) uint8   { return *(*uint8)(unsafe.Pointer(&x)) }
func b2u16(x bool) uint16 { return uint16(*(*uint8)(unsafe.Pointer(&x))) }
func b2i(x bool) int      { return int(*(*uint8)(unsafe.Pointer(&x))) }

// 8-bit operations
func GetBit8(v uint8, n uint) bool {
	return GetBiti8(v, n) != 0
}

func GetBiti8(v uint8, n uint) uint8 {
	return v >> (n) & 0x01
}

func SetBit8(v *uint8, n uint) {
	*v |= (1 << n)
}

func ClearBit8(v *uint8, n uint) {
	*v &= ^(1 << n)
}

func FlipBit8(v *uint8, n uint) {
	*v ^= (1 << n)
}

func ClearBits8(v *uint8, mask uint8) {
	*v &= ^mask
}

// 16-bit operations
func GetBit16(v uint16, n uint) bool {
	return GetBiti16(v, n) != 0
}

func GetBiti16(v uint16, n uint) uint16 {
	return v >> (n) & 0x01
}

func SetBit16(v *uint16, n uint) {
	*v |= (1 << n)
}

func ClearBit16(v *uint16, n uint) {
	*v &= ^(1 << n)
}

func FlipBit16(v *uint16, n uint) {
	*v ^= (1 << n)
}

func ClearBits16(v *uint16, mask uint16) {
	*v &= ^mask
}
