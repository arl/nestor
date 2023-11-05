package emu

type Bus interface {
	Reset()
	Read8(addr uint16) uint8
	Write8(addr uint16, val uint8)
	MapSlice(addr, end uint16, buf []byte)
}
