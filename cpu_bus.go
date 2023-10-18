package main

type cpuBus struct {
	mmap MemMap
	name string
}

func newcpuBus(name string) *cpuBus {
	bus := &cpuBus{name: name}
	bus.Reset()
	return bus
}

func (b *cpuBus) MapMemory() {
	// RAM is 0x800 bytes, mirrored.
	ram := make([]byte, 0x0800)
	b.MapSlice(0x0000, 0x07FF, ram)
	b.MapSlice(0x0800, 0x0FFF, ram)
	b.MapSlice(0x1000, 0x17FF, ram)
	b.MapSlice(0x1800, 0x1FFF, ram)
}

func (b *cpuBus) Reset() {
	b.mmap.Reset()
}

func (b *cpuBus) Read8(addr uint16) uint8 {
	return b.mmap.Read8(addr)
}

func (b *cpuBus) Write8(addr uint16, val uint8) {
	b.mmap.Write8(addr, val)
}

func (b *cpuBus) MapSlice(addr, end uint16, buf []byte) {
	b.mmap.MapSlice(addr, end, buf)
}
