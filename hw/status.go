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

func (p P) checkFlag(flag uint8) bool {
	return uint8(p)&flag == flag
}

func (p *P) setCarry(val bool) {
	var ival P
	if val {
		ival = 1
	}
	*p &^= 0x1
	*p |= ival
}

func (p P) zero() bool {
	return p&0x2 != 0
}

func (p *P) setZero(val bool) {
	var ival P
	if val {
		ival = 1
	}
	*p &^= 0x2
	*p |= ival << 1
}

func (p P) intDisable() bool {
	return p&0x4 != 0
}

func (p *P) setIntDisable(val bool) {
	var ival P
	if val {
		ival = 1
	}
	*p &^= 0x4
	*p |= ival << 2
}

func (p P) decimal() bool {
	return p&0x8 != 0
}

func (p *P) setDecimal(val bool) {
	var ival P
	if val {
		ival = 1
	}
	*p &^= 0x8
	*p |= ival << 3
}

func (p P) brk() bool {
	return p&0x10 != 0
}

func (p *P) setBrk(val bool) {
	var ival P
	if val {
		ival = 1
	}
	*p &^= 0x10
	*p |= ival << 4
}

func (p P) unused() bool {
	return p&0x20 != 0
}

func (p *P) setUnused(val bool) {
	var ival P
	if val {
		ival = 1
	}
	*p &^= 0x20
	*p |= ival << 5
}

func (p P) overflow() bool {
	return p&0x40 != 0
}

func (p *P) setOverflow(val bool) {
	var ival P
	if val {
		ival = 1
	}
	*p &^= 0x40
	*p |= ival << 6
}

func (p P) negative() bool {
	return p&0x80 != 0
}

func (p *P) setNegative(val bool) {
	var ival P
	if val {
		ival = 1
	}
	*p &^= 0x80
	*p |= ival << 7
}
