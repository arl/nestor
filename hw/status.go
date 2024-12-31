package hw

type P uint8

const (
	Carry = 1 << iota
	Zero
	Interrupt
	Decimal
	Break
	Reserved
	Overflow
	Negative
)

func (p P) String() string {
	const bits = "nvubdizcNVUBDIZC"

	s := make([]byte, 8)
	for i := 0; i < 8; i++ {
		ibit := (uint8(p) & (1 << (7 - i))) >> (7 - i)
		s[i] = bits[i+int(8*ibit)]
	}
	return string(s)
}

func (p *P) setFlags(flags uint8) {
	*p |= P(flags)
}

func (p *P) clearFlags(flags uint8) {
	*p &= ^P(flags)
}

func (p P) hasFlag(flag P) bool {
	return p&flag == flag
}

func (p *P) setNZ(val uint8) {
	if val == 0 {
		p.setFlags(Zero)
	} else if val&0x80 != 0 {
		p.setFlags(Negative)
	}
}
