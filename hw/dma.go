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

	needHalt bool

	oamInProgress bool
	// dmcInProgress bool

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
	dma.needHalt = false
	dma.oamInProgress = false
	dma.dmcDmaRunning = false
	dma.abortDmcDma = false
}

func (dma *DMA) WriteOAMDMA(_, val uint8) {
	log.ModDMA.DebugZ("start OAM DMA transfer").Hex8("page", val).End()
	dma.oamPage = val
	dma.oamInProgress = true
	dma.needHalt = true
}

func (dma *DMA) startDMCTransfer() {
	log.ModDMA.DebugZ("start DMC DMA transfer").End()
	dma.dmcDmaRunning = true
	dma.dummy = true
	dma.needHalt = true
}

func (dma *DMA) stopDmcTransfer() {
	log.ModDMA.DebugZ("stop DMC DMA transfer").End()
	if dma.dmcDmaRunning {
		if dma.needHalt {
			// If interrupted before the halt cycle starts, cancel DMA
			// completely This can happen when a write prevents the DMA from
			// starting after being queued
			dma.dmcDmaRunning = false
			dma.dummy = false
			dma.needHalt = false
		} else {
			// Abort DMA if possible (this only appears to be possible if done
			// within the first cycle of DMA)
			dma.abortDmcDma = true
		}
	}
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func mustT[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func (dma *DMA) process(readAddr uint16) {
	if !dma.needHalt {
		return
	}

	dmc := &dma.cpu.APU.DMC

	prevReadAddress := readAddr
	enableInternalRegReads := (readAddr & 0xFFE0) == 0x4000
	skipFirstInputClock := false
	if enableInternalRegReads && dma.dmcDmaRunning && (readAddr == 0x4016 || readAddr == 0x4017) {
		dmcAddress := dmc.getReadAddress()
		if (dmcAddress & 0x1F) == (readAddr & 0x1F) {
			// DMC will cause a read on the same address as the CPU was reading
			// from This will hide the reads from the controllers because /OE
			// will be active the whole time
			skipFirstInputClock = true
		}
	}
	// On PAL, the dummy/idle reads done by the DMA don't appear to be done on
	// the address that the CPU was about to read. This prevents the 2+x reads
	// on registers issues. The exact specifics of where the CPU reads instead
	// aren't known yet - so just disable read side-effects entirely on PAL
	const isNtscInputBehavior = true

	// On Famicom, each dummy/idle read to 4016/4017 is intepreted as a read of
	// the joypad registers On NES (or AV Famicom), only the first dummy/idle
	// read causes side effects (e.g only a single bit is lost)
	const isNesBehavior = true
	skipDummyReads := !isNtscInputBehavior || (isNesBehavior && (readAddr == 0x4016 || readAddr == 0x4017))

	dma.needHalt = false
	cpu := dma.cpu

	cpu.cycleBegin(true)
	if dma.abortDmcDma && isNesBehavior && (readAddr == 0x4016 || readAddr == 0x4017) {
		// Skip halt cycle dummy read on 4016/4017 The DMA was aborted, and the
		// CPU will read 4016/4017 next If 4016/4017 is read here, the
		// controllers will see 2 separate reads even though they would only see
		// a single read on hardware (except the original Famicom)
	} else if isNesBehavior && !skipFirstInputClock {
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
			dma.needHalt = false
		} else if dma.needHalt {
			dma.needHalt = false
		} else if dma.dummy {
			dma.dummy = false
		}
		cpu.cycleBegin(true)
	}

	for dma.dmcDmaRunning || dma.oamInProgress {
		getCycle := (cpu.Cycles & 0x01) == 0
		if getCycle {
			if dma.dmcDmaRunning && !dma.needHalt && !dma.dummy {
				// DMC DMA is ready to read a byte (both halt and dummy read
				// cycles were performed before this)
				processCycle()
				val = dma.processRead(dmc.getReadAddress(), &prevReadAddress, enableInternalRegReads, true)
				cpu.cycleEnd(true)
				dma.dmcDmaRunning = false
				dma.abortDmcDma = false
				dmc.setReadBuffer(val)
			} else if dma.oamInProgress {
				// DMC DMA is not running, or not ready, run sprite DMA
				processCycle()
				val = dma.processRead(uint16(dma.oamPage)*0x100+uint16(spriteAddr), &prevReadAddress, enableInternalRegReads, true)
				cpu.cycleEnd(true)
				spriteAddr++
				counter++
			} else {
				// DMC DMA is running, but not ready (need halt/dummy read) and
				// sprite DMA isn't runnnig, perform a dummy read
				if !dma.needHalt && !dma.dummy {
					panic("unexpected")
				}
				processCycle()
				if !skipDummyReads {
					dma.cpuBus.Read8(readAddr, false)
				}
				cpu.cycleEnd(true)
			}
		} else {
			if dma.oamInProgress && (counter&0x01 != 0) {
				// Sprite DMA write cycle (only do this if a sprite dma read was
				// performed last cycle).
				processCycle()
				dma.cpuBus.Write8(0x2004, val)
				cpu.cycleEnd(true)
				counter++
				if counter == 0x200 {
					dma.oamInProgress = false
				}
			} else {
				// Align to read cycle before starting sprite DMA (or align to
				// perform DMC read)
				processCycle()
				if !skipDummyReads {
					dma.cpuBus.Read8(readAddr, false)
				}
				cpu.cycleEnd(true)
			}
		}
	}
}

// TODO: do not use pointers for prevReadAddress
func (dma *DMA) processRead(addr uint16, prevReadAddress *uint16, enableInternalRegReads bool, isNesBehavior bool) uint8 {
	// This is to reproduce a CPU bug that can occur during DMA which can cause
	// the 2A03 to read from its internal registers (4015, 4016, 4017) at the
	// same time as the DMA unit reads a byte from the bus. This bug occurs if
	// the CPU is halted while it's reading a value in the $4000-$401F range.
	//
	// This has a number of side effects:
	//  - It can cause a read of $4015 to occur without the program's knowledge,
	//    which would clear the frame counter's IRQ flag
	//  - It can cause additional bit deletions while reading the input (e.g more
	//    than the DMC glitch usually causes)
	//  - It can also *prevent* bit deletions from occurring at all in another scenario
	//  - It can replace/corrupt the byte that the DMA is reading, causing DMC to
	//    play the wrong sample

	var val uint8
	if !enableInternalRegReads {
		if addr >= 0x4000 && addr <= 0x401F {
			// Nothing will respond on $4000-$401F on the external bus - return
			// open bus value
			//
			// TODO: should read openbus here
			val = 0x00
		} else {
			val = dma.cpuBus.Read8(addr, false)
		}
		*prevReadAddress = addr
		return val
	}

	// This glitch causes the CPU to read from the internal APU/Input registers
	// regardless of the address the DMA unit is trying to read
	internalAddr := 0x4000 | (addr & 0x1F)
	isSameAddress := internalAddr == addr

	switch internalAddr {
	case 0x4015:
		val = dma.cpu.Bus.Read8(internalAddr, false)
		if !isSameAddress {
			// Also trigger a read from the actual address the CPU was
			// supposed to read from (external bus)
			dma.cpu.Bus.Read8(addr, false)
		}

	case 0x4016, 0x4017:
		if isNesBehavior && *prevReadAddress == internalAddr {
			// Reading from the same input register twice in a row, skip the
			// read entirely to avoid triggering a bit loss from the read, since
			// the controller won't react to this read Return the same value as
			// the last read, instead On PAL, the behavior is unknown - for now,
			// don't cause any bit deletions.
			//
			// TODO: get value from openbus
			val = 0x00
		} else {
			val = dma.cpu.Bus.Read8(internalAddr, false)
		}

		if !isSameAddress {
			// The DMA unit is reading from a different address, read from it
			// too (external bus).
			//
			const obMask = uint8(0xE0)
			externalValue := dma.cpu.Bus.Read8(addr, false)

			// Merge values, keep the external value for all open bus pins on
			// the 4016/4017 port AND all other bits together (bus conflict)
			val = (externalValue & obMask) | ((val & ^obMask) & (externalValue & ^obMask))
		}

	default:
		val = dma.cpu.Bus.Read8(addr, false)
	}

	*prevReadAddress = internalAddr
	return val
}
