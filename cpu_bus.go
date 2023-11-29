package main

import (
	"nestor/emu/hwio"
)

type cpuBus struct {
	mmap hwio.MemMap
	name string
}

func newCpuBus(name string) *cpuBus {
	bus := &cpuBus{name: name}
	bus.Reset()
	return bus
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

func (b *cpuBus) MapReg8(addr uint16, reg *hwio.Reg8) {
	b.mmap.MapReg8(addr, reg)
}
