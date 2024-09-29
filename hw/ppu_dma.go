package hw

import (
	"nestor/emu/hwio"
	"nestor/emu/log"
)

// ppuDMA handles the DMA transfer of OAM (sprites attributes) to the PPU.
type ppuDMA struct {
	// oam    []byte
	cpuBus hwio.BankIO8
	cpu    *CPU

	page       uint8
	inProgress bool

	OAMDMA hwio.Reg8 `hwio:"offset=0x00,writeonly,wcb"`

	// Since DMA can only be started on an even CPU cycle, we use a dummy cycle
	// to align the transfer with an even cycle.
	dummy bool
}

func (dma *ppuDMA) InitBus(cpubus hwio.BankIO8) {
	hwio.MustInitRegs(dma)
	dma.cpuBus = cpubus
	dma.reset()
}

func (dma *ppuDMA) reset() {
	dma.page = 0x00
	dma.dummy = true
	dma.inProgress = false
}

func (dma *ppuDMA) WriteOAMDMA(_, val uint8) {
	log.ModDMA.InfoZ("Write to OAMDMA reg").Hex8("val", val).End()
	dma.page = val
	dma.inProgress = true
}

func (dma *ppuDMA) process() {
	if !dma.inProgress {
		return
	}

	counter := 0
	spriteAddr := uint8(0)
	val := uint8(0)

	dma.cpu.cycleBegin(true)
	dma.cpu.cycleEnd(true)

	for dma.inProgress {
		if (dma.cpu.Cycles & 0x01) == 0 {
			// read cycle.
			dma.cpu.cycleBegin(true)
			addr := uint16(dma.page)<<8 | uint16(spriteAddr)
			val = dma.cpuBus.Read8(addr)
			dma.cpu.cycleEnd(true)
			spriteAddr++
			counter++
		} else {
			// write cycle.
			if counter&0x01 != 0 {
				dma.cpu.cycleBegin(true)
				dma.cpuBus.Write8(0x2004, val)
				dma.cpu.cycleEnd(true)
				counter++
				if counter == 0x200 {
					dma.inProgress = false
				}
			} else {
				dma.cpu.cycleBegin(true)
				dma.cpu.cycleEnd(true)
			}
		}
	}
}
