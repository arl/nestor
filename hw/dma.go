package hw

import (
	"nestor/emu/log"
	"nestor/hw/hwio"
)

// DMA handles DMA transfer of OAM (sprites attributes) to the PPU
// and DMC samples to the APU.
type DMA struct {
	cpu *CPU

	needHalt bool
	// DMA transfer can only be started on even CPU cycles. We use a dummy
	// cycle, when necessary, to align the transfer with an even cycle.
	dummy bool

	dmcRunning bool
	abortDMC   bool

	OAMDMA     hwio.Reg8 `hwio:"offset=0x00,writeonly,wcb"`
	oamPage    uint8
	oamRunning bool // OAM DMA in progress
}

func (dma *DMA) InitBus(cpu *CPU) {
	hwio.MustInitRegs(dma)
	dma.cpu = cpu
	dma.reset()
}

func (dma *DMA) reset() {
	dma.oamPage = 0x00
	dma.dummy = true
	dma.needHalt = false
	dma.oamRunning = false
	dma.dmcRunning = false
	dma.abortDMC = false
}

func (dma *DMA) WriteOAMDMA(_, val uint8) {
	log.ModDMA.DebugZ("start OAM DMA transfer").Hex8("page", val).End()
	dma.oamPage = val
	dma.oamRunning = true
	dma.needHalt = true
}

func (dma *DMA) startDMCTransfer() {
	log.ModDMA.DebugZ("start DMC DMA transfer").End()
	dma.dmcRunning = true
	dma.dummy = true
	dma.needHalt = true
}

func (dma *DMA) stopDmcTransfer() {
	log.ModDMA.DebugZ("stop DMC DMA transfer").End()
	if dma.dmcRunning {
		if dma.needHalt {
			// If interrupted before the halt cycle starts, cancel DMA
			// completely. This can happen when a write prevents the DMA from
			// starting after being queued.
			dma.dmcRunning = false
			dma.dummy = false
			dma.needHalt = false
		} else {
			// Abort DMA if possible (this only appears to be possible if done
			// within the first cycle of DMA).
			dma.abortDMC = true
		}
	}
}

func (dma *DMA) processPending(addr uint16) {
	if !dma.needHalt {
		return
	}

	dmc := &dma.cpu.APU.DMC

	isInternalReg := (addr & 0xFFE0) == 0x4000
	skipFirstInputClock := false
	if isInternalReg && dma.dmcRunning && (addr == 0x4016 || addr == 0x4017) {
		dmcAddress := dmc.CurrentAddress()
		if (dmcAddress & 0x1F) == (addr & 0x1F) {
			// DMC will cause a read on the same address as the CPU was reading
			// from This will hide the reads from the controllers because /OE
			// will be active the whole time
			skipFirstInputClock = true
		}
	}

	// On Famicom, each dummy/idle read to 4016/4017 is intepreted as a read of
	// the joypad registers On NES (or AV Famicom), only the first dummy/idle
	// read causes side effects (e.g only a single bit is lost).
	const isNesBehavior = true
	skipDummyReads := (isNesBehavior && (addr == 0x4016 || addr == 0x4017))

	dma.needHalt = false
	cpu := dma.cpu

	cpu.cycleBegin(true)
	if dma.abortDMC && isNesBehavior && (addr == 0x4016 || addr == 0x4017) {
		// Skip halt cycle dummy read on 4016/4017 The DMA was aborted, and the
		// CPU will read 4016/4017 next If 4016/4017 is read here, the
		// controllers will see 2 separate reads even though they would only see
		// a single read on hardware (except the original Famicom)
	} else if isNesBehavior && !skipFirstInputClock {
		cpu.Bus.Read8(addr, false)
	}
	cpu.cycleEnd(true)

	if dma.abortDMC {
		dma.dmcRunning = false
		dma.abortDMC = false

		if !dma.oamRunning {
			// If DMC DMA was cancelled and OAM DMA isn't about to start,
			// stop processing DMA entirely. Otherwise, OAM DMA needs to run,
			// so the DMA process has to continue.
			dma.dummy = false
			return
		}
	}

	processCycle := func() {
		// Sprite DMA cycles count as halt/dummy cycles for the DMC DMA when
		// both run at the same time
		if dma.abortDMC {
			dma.dmcRunning = false
			dma.abortDMC = false
			dma.dummy = false
			dma.needHalt = false
		} else if dma.needHalt {
			dma.needHalt = false
		} else if dma.dummy {
			dma.dummy = false
		}
		cpu.cycleBegin(true)
	}

	oamCounter := 0
	spriteAddr := uint8(0)
	val := uint8(0)

	prevAddr := addr

	for dma.dmcRunning || dma.oamRunning {
		if (cpu.Cycles & 0x01) == 0 {
			// Read cycle.
			switch {
			case dma.dmcRunning && !dma.needHalt && !dma.dummy:
				// DMC DMA is ready to read a byte (both halt and dummy read
				// cycles were performed before this)
				processCycle()
				val, prevAddr = dma.processRead(dmc.CurrentAddress(), prevAddr, isInternalReg)
				cpu.cycleEnd(true)
				dma.dmcRunning = false
				dma.abortDMC = false
				dmc.SetReadBuffer(val)
			case dma.oamRunning:
				// DMC DMA is not running, or not ready, run sprite DMA
				processCycle()
				addr := uint16(dma.oamPage)*0x100 + uint16(spriteAddr)
				val, prevAddr = dma.processRead(addr, prevAddr, isInternalReg)
				cpu.cycleEnd(true)
				spriteAddr++
				oamCounter++
			default:
				// DMC DMA is running, but not ready (need halt/dummy read) and
				// sprite DMA isn't runnnig, perform a dummy read
				if !dma.needHalt && !dma.dummy {
					panic("unexpected")
				}
				processCycle()
				if !skipDummyReads {
					cpu.Bus.Read8(addr, false)
				}
				cpu.cycleEnd(true)
			}
		} else {
			// Write cycle.
			if dma.oamRunning && (oamCounter&0x01 != 0) {
				// Sprite DMA write cycle (only do this if a sprite dma read was
				// performed last cycle).
				processCycle()
				cpu.Bus.Write8(0x2004, val)
				cpu.cycleEnd(true)
				oamCounter++
				if oamCounter == 0x200 {
					dma.oamRunning = false
				}
			} else {
				// Align to read cycle before starting sprite DMA (or align to
				// perform DMC read)
				processCycle()
				if !skipDummyReads {
					cpu.Bus.Read8(addr, false)
				}
				cpu.cycleEnd(true)
			}
		}
	}
}

func (dma *DMA) processRead(addr uint16, prevAddr uint16, isInternalReg bool) (val uint8, readAddr uint16) {
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

	if !isInternalReg {
		if addr >= 0x4000 && addr <= 0x401F {
			// Nothing will respond on $4000-$401F on the external bus - return
			// open bus value
			//
			// TODO: should read openbus here
			val = 0x00
		} else {
			val = dma.cpu.Bus.Read8(addr, false)
		}
		return val, addr
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
		if prevAddr == internalAddr {
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
			const openbusMask = uint8(0xE0)
			extval := dma.cpu.Bus.Read8(addr, false)

			// Merge values, keep the external value for all open bus pins on
			// the 4016/4017 port AND all other bits together (bus conflict).
			val = (extval & openbusMask) | ((val & ^openbusMask) & (extval & ^openbusMask))
		}

	default:
		val = dma.cpu.Bus.Read8(addr, false)
	}

	return val, internalAddr
}
