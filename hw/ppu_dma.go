package hw

import (
	"nestor/emu/hwio"
	"nestor/emu/log"
)

// ppuDMA handles the DMA transfer of OAM (sprites attributes) to the PPU.
type ppuDMA struct {
	oam    []byte
	cpuBus hwio.BankIO8

	page       uint8
	addr       uint8
	data       uint8
	inProgress bool

	OAMDMA hwio.Reg8 `hwio:"offset=0x00,writeonly,wcb"`

	// Since DMA can only be started on an even CPU cycle, we use a dummy cycle
	// to align the transfer with an even cycle.
	dummy bool
}

func (dma *ppuDMA) InitBus(cpubus hwio.BankIO8, oam []byte) {
	hwio.MustInitRegs(dma)
	dma.cpuBus = cpubus
	dma.oam = oam
	dma.reset()
}

func (dma *ppuDMA) reset() {
	dma.page = 0x00
	dma.addr = 0x00
	dma.data = 0x00
	dma.dummy = true
	dma.inProgress = false
}

func (dma *ppuDMA) WriteOAMDMA(_, val uint8) {
	log.ModDMA.InfoZ("Write to OAMDMA reg").Hex8("val", val).End()
	dma.page = val
	dma.addr = 0x00
	dma.inProgress = true
}

func (dma *ppuDMA) process(cpuTicks int64) {
	if !dma.inProgress {
		return
	}

	// Start DMA transfer
	const (
		even = 0
		odd  = 1
	)

	// The first cycle is always idle.
	// On odd cycle count we add an extrac idle cycle.
	if dma.dummy {
		if cpuTicks%2 == odd {
			dma.dummy = false
			log.ModDMA.InfoZ("Begin PPU DMA transfer").
				Hex8("page", dma.page).
				Int64("ticks", cpuTicks).
				End()
		}
		return
	}

	switch cpuTicks % 2 {
	case even:
		// Read from CPU bus
		addr := uint16(dma.page)<<8 | uint16(dma.addr)
		dma.data = dma.cpuBus.Read8(addr)

	case odd:
		// Write to PPU OAM
		dma.oam[dma.addr] = dma.data
		dma.addr++
		// When this wraps around we know that 256 bytes have been written.
		if dma.addr == 0x00 {
			log.ModDMA.InfoZ("Ending PPU DMA transfer").
				Blob("bytes", dma.oam).
				Hex8("page", dma.page).
				Int64("ticks", cpuTicks).
				End()
			dma.inProgress = false
			dma.dummy = true
		}
	}
}
