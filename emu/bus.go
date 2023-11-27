package emu

type Bus interface {
	Reset()
	Read8(addr uint16) uint8
	Write8(addr uint16, val uint8)
	MapSlice(addr, end uint16, buf []byte)
}

func Read16(b Bus, addr uint16) uint16 {
	lo := b.Read8(addr)
	hi := b.Read8(addr + 1)
	return uint16(hi)<<8 | uint16(lo)
}
