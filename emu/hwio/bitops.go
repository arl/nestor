package hwio

func GetBit16(v uint16, n uint) bool {
	return GetBiti16(v, n) != 0
}

func GetBiti16(v uint16, n uint) uint16 {
	return v >> (n) & 0x01
}

func SetBit(v *uint16, n uint) {
	*v |= (1 << n)
}

func ClearBit(v *uint16, n uint) {
	*v &= ^(1 << n)
}

func FlipBit(v *uint16, n uint) {
	*v ^= (1 << n)
}

func ClearBits(v *uint16, mask uint16) {
	*v &= ^mask
}
