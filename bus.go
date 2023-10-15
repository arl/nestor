package main

type Bus struct {
	mmap MemMap

	name string
}

func NewBus(name string) *Bus {
	bus := &Bus{name: name}
	bus.Reset()
	return bus
}

func (b *Bus) Reset() {
	b.mmap.Reset()
}

func (b *Bus) Read8(addr uint16) uint8 {
	return b.mmap.Read8(addr)
}

func (b *Bus) Write8(addr uint16, val uint8) {
	b.mmap.Write8(addr, val)
}

func (b *Bus) MapSlice(addr, end uint16, buf []byte) {
	b.mmap.MapSlice(addr, end, buf)
}
