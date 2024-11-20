package hw

import (
	"nestor/emu/log"
	"nestor/hw/hwio"
)

// DMA handles DMA transfer of OAM (sprites attributes) to the PPU
// and DMC samples to the APU.
type DMA struct {
	cpuBus *hwio.Table
	cpu    *CPU

	inProgress    bool
	oamInProgress bool
	dmcInProgress bool

	OAMDMA  hwio.Reg8 `hwio:"offset=0x00,writeonly,wcb"`
	oamPage uint8

	// DMA can only be started on even CPU cycles. We use
	// a dummy cycle, when necessary, to align the transfer
	// with an even cycle.
	dummy bool

	// TODO
	// s/dmcDmaRunning/dmaInProgress/
	// s/abortDmaDma/abortDmc/
	dmcDmaRunning bool
	abortDmcDma   bool
}

func (dma *DMA) InitBus(cpu *CPU) {
	hwio.MustInitRegs(dma)
	dma.cpuBus = cpu.Bus
	dma.cpu = cpu
	dma.reset()
}

func (dma *DMA) reset() {
	dma.oamPage = 0x00
	dma.dummy = true
	dma.inProgress = false
	dma.dmcInProgress = false
	dma.oamInProgress = false
	dma.dmcDmaRunning = false
	dma.abortDmcDma = false
}

func (dma *DMA) WriteOAMDMA(_, val uint8) {
	log.ModDMA.InfoZ("Write to OAMDMA reg").Hex8("val", val).End()
	dma.oamPage = val
	dma.oamInProgress = true
	dma.inProgress = true
}

func (dma *DMA) startDMCTransfert() {
	dma.dmcDmaRunning = true
	dma.dummy = true
	dma.inProgress = true
}

func (dma *DMA) stopDmcTransfer() {
	if dma.dmcDmaRunning {
		if dma.inProgress {
			// If interrupted before the halt cycle starts, cancel DMA
			// completely This can happen when a write prevents the DMA from
			// starting after being queued
			dma.dmcDmaRunning = false
			dma.dummy = false
			dma.inProgress = false
		} else {
			// Abort DMA if possible (this only appears to be possible if done
			// within the first cycle of DMA)
			dma.abortDmcDma = true
		}
	}
}

func (dma *DMA) process(readAddr uint16) {
	if !dma.inProgress {
		return
	}

	prevReadAddress := readAddr
	enableInternalRegReads := (readAddr & 0xFFE0) == 0x4000
	skipFirstInputClock := false
	if enableInternalRegReads && dma.dmcDmaRunning && (readAddr == 0x4016 || readAddr == 0x4017) {
		dmcAddress := dma.cpu.APU.DMC.getReadAddress()
		if (dmcAddress & 0x1F) == (readAddr & 0x1F) {
			// DMC will cause a read on the same address as the CPU was reading
			// from This will hide the reads from the controllers because /OE
			// will be active the whole time
			skipFirstInputClock = true
		}
	}

	// On Famicom, each dummy/idle read to 4016/4017 is intepreted as a read of the joypad registers
	// On NES (or AV Famicom), only the first dummy/idle read causes side effects (e.g only a single bit is lost)
	skipDummyReads := (readAddr == 0x4016 || readAddr == 0x4017)

	dma.inProgress = false
	cpu := dma.cpu
	cpu.cycleBegin(true)

	if dma.abortDmcDma && (readAddr == 0x4016 || readAddr == 0x4017) {
		// Skip halt cycle dummy read on 4016/4017 The DMA was aborted, and the
		// CPU will read 4016/4017 next If 4016/4017 is read here, the
		// controllers will see 2 separate reads even though they would only see
		// a single read on hardware (except the original Famicom)
	} else if !skipFirstInputClock {
		// _memoryManager->Read(readAddress, MemoryOperationType::DmaRead);
		dma.cpuBus.Read8(readAddr, false)

	}
	cpu.cycleEnd(true)

	if dma.abortDmcDma {
		dma.dmcDmaRunning = false
		dma.abortDmcDma = false

		if !dma.oamInProgress {
			// If DMC DMA was cancelled and OAM DMA isn't about to start,
			// stop processing DMA entirely. Otherwise, OAM DMA needs to run,
			// so the DMA process has to continue.
			dma.dummy = false
			return
		}
	}

	counter := 0
	spriteAddr := uint8(0)
	val := uint8(0)

	processCycle := func() {
		// Sprite DMA cycles count as halt/dummy cycles for the DMC DMA when
		// both run at the same time
		if dma.abortDmcDma {
			dma.dmcDmaRunning = false
			dma.abortDmcDma = false
			dma.dummy = false
			dma.inProgress = false
		} else if dma.inProgress {
			dma.inProgress = false
		} else if dma.dummy {
			dma.dummy = false
		}
		cpu.cycleBegin(true)
	}

	for dma.dmcDmaRunning || dma.oamInProgress {
		getCycle := (cpu.Cycles & 0x01) == 0
		if getCycle {
			if dma.dmcDmaRunning && !dma.inProgress && !dma.dummy {
				// DMC DMA is ready to read a byte (both halt and dummy read
				// cycles were performed before this)
				processCycle()
				dma.dmcInProgress = true // used by debugger to distinguish between dmc and oam/dummy dma reads
				val = dma.processRead(dma.cpu.APU.DMC.getReadAddress(), prevReadAddress, enableInternalRegReads, true)
				dma.dmcInProgress = false
				cpu.cycleEnd(true)
				dma.dmcDmaRunning = false
				dma.abortDmcDma = false
				dma.cpu.APU.DMC.setReadBuffer(val)
			} else if dma.oamInProgress {
				// DMC DMA is not running, or not ready, run sprite DMA
				processCycle()
				val = dma.processRead(uint16(dma.oamPage)*0x100+uint16(spriteAddr), prevReadAddress, enableInternalRegReads, true)
				cpu.cycleEnd(true)
				spriteAddr++
				counter++
			} else {
				// DMC DMA is running, but not ready (need halt/dummy read) and sprite DMA isn't runnnig, perform a dummy read
				if dma.inProgress || dma.dummy {

				} else {
					panic("unexpected")
				}
				processCycle()
				if !skipDummyReads {
					dma.cpuBus.Write8(readAddr, val)
				}
				cpu.cycleEnd(true)
			}
		} else {
			if dma.oamInProgress && (counter&0x01 != 0) {
				// Sprite DMA write cycle (only do this if a sprite dma read was performed last cycle)
				processCycle()
				dma.cpuBus.Write8(0x2004, val)
				cpu.cycleEnd(true)
				counter++
				if counter == 0x200 {
					dma.oamInProgress = false
				}
			} else {
				// Align to read cycle before starting sprite DMA (or align to perform DMC read)
				processCycle()
				if !skipDummyReads {
					dma.cpuBus.Read8(readAddr, false)
				}
				cpu.cycleEnd(true)
			}
		}
	}

	// cpu.cycleEnd(true)

	/*
		for dma.inProgress {
			if (cpu.Cycles & 0x01) == 0 {
				// read cycle.
				cpu.cycleBegin(true)
				addr := uint16(dma.oamPage)<<8 | uint16(spriteAddr)
				val = dma.cpuBus.Read8(addr, false)
				cpu.cycleEnd(true)
				spriteAddr++
				counter++
			} else {
				// write cycle.
				if counter&0x01 != 0 {
					cpu.cycleBegin(true)
					dma.cpuBus.Write8(0x2004, val)
					cpu.cycleEnd(true)
					counter++
					if counter == 0x200 {
						dma.inProgress = false
					}
				} else {
					cpu.cycleBegin(true)
					cpu.cycleEnd(true)
				}
			}
		}

	*/
}

func (dma *DMA) processRead(addr uint16, prevReadAddress uint16, enableInternalRegReads bool, isNesBehavior bool) uint8 {
	// This is to reproduce a CPU bug that can occur during DMA which can cause the 2A03 to read from
	// its internal registers (4015, 4016, 4017) at the same time as the DMA unit reads a byte from
	// the bus. This bug occurs if the CPU is halted while it's reading a value in the $4000-$401F range.
	//
	// This has a number of side effects:
	//  -It can cause a read of $4015 to occur without the program's knowledge, which would clear the frame counter's IRQ flag
	//  -It can cause additional bit deletions while reading the input (e.g more than the DMC glitch usually causes)
	//  -It can also *prevent* bit deletions from occurring at all in another scenario
	//  -It can replace/corrupt the byte that the DMA is reading, causing DMC to play the wrong sample
	var val uint8
	if !enableInternalRegReads {
		if addr >= 0x4000 && addr <= 0x401F {
			// Nothing will respond on $4000-$401F on the external bus - return open bus value
			// TODO: should read openbus here
			val = dma.cpuBus.Read8(addr, false)
		} else {
			val = dma.cpuBus.Read8(addr, false)
		}
		prevReadAddress = addr
		return val
	}
	// TODO
	return 0
}
