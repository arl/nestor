package hw

/* Extra methods for the processor status register */

func (p *P) checkNZ(v uint8) {
	p.setNegative(v&0x80 != 0)
	p.setZero(v == 0)
}

// sets N flag if bit 7 of v is set, clears it otherwise.
func (p *P) checkN(v uint8) {
	p.setNegative(v&(1<<7) != 0)
}

// sets Z flag if v == 0, clears it otherwise.
func (p *P) checkZ(v uint8) {
	p.setZero(v == 0)
}

func (p *P) checkCV(x, y uint8, sum uint16) {

	// forward carry or unsigned overflow.
	p.setCarry(sum > 0xFF)

	// signed overflow, can only happen if the sign of the sum differs
	// from that of both operands.
	v := (uint16(x) ^ sum) & (uint16(y) ^ sum) & 0x80
	p.setOverflow(v != 0)
}

func (p P) String() string {
	const bits = "nvubdizcNVUBDIZC"

	s := make([]byte, 8)
	for i := 0; i < 8; i++ {
		ibit := (uint8(p) & (1 << (7 - i))) >> (7 - i)
		s[i] = bits[i+int(8*ibit)]
	}
	return string(s)
}
