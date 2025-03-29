package hw

import (
	"nestor/emu/log"
	"nestor/hw/hwio"
	"nestor/hw/snapshot"
)

// DMA handles DMA transfer of OAM (sprite attributes) to the PPU and DMC (audio
// samples) to the APU. On real hardware, there are 2 separate DMA units, but
// here we handle both of them in the same place to ensure correct timing and
// synchronization when both DMC and OAM transfers occur simultaneously.
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

func (dma *DMA) stopDMCTransfer() {
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
	cpu := dma.cpu

	skipInitClock := false

	isInternalReg := (addr & 0xFFE0) == 0x4000
	if isInternalReg && dma.dmcRunning && (addr == 0x4016 || addr == 0x4017) {
		dmcAddress := dmc.CurrentAddr()
		if (dmcAddress & 0x1F) == (addr & 0x1F) {
			// DMC causes a read on the same address as the CPU was reading
			// from. This hides reads from controllers because /OE will be
			// active the whole time.
			skipInitClock = true
		}
	}

	dma.needHalt = false
	skipDummyReads := addr == 0x4016 || addr == 0x4017

	cpu.cycleBegin(true)
	if dma.abortDMC && skipDummyReads {
		// Skip halt cycle dummy read on 4016/4017. The DMA was aborted, and the
		// CPU will read 4016/4017 next if 4016/4017 is read here, the
		// controllers will see 2 separate reads even though they would only see
		// a single read on hardware.
	} else if !skipInitClock {
		cpu.Bus.Read8(addr)
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

	for dma.dmcRunning || dma.oamRunning {
		if (cpu.Cycles & 0x01) == 0 {
			// Read cycle.
			switch {
			case dma.dmcRunning && !dma.needHalt && !dma.dummy:
				// DMC DMA is ready to read a byte (both halt and dummy read
				// cycles were performed before this)
				processCycle()
				val = dma.processRead(dmc.CurrentAddr(), isInternalReg)
				cpu.cycleEnd(true)
				dma.dmcRunning = false
				dma.abortDMC = false
				dmc.SetReadBuffer(val)
			case dma.oamRunning:
				// DMC DMA is not running, or not ready, run sprite DMA
				processCycle()
				addr := uint16(dma.oamPage)*0x100 + uint16(spriteAddr)
				val = dma.processRead(addr, isInternalReg)
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
					cpu.Bus.Read8(addr)
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
				// perform DMC read).
				processCycle()
				if !skipDummyReads {
					cpu.Bus.Read8(addr)
				}
				cpu.cycleEnd(true)
			}
		}
	}
}

func (dma *DMA) processRead(addr uint16, isInternalReg bool) uint8 {
	if !isInternalReg && addr >= 0x4000 && addr <= 0x401F {
		// Nothing will respond on $4000-$401F on the external bus - return
		// open bus value
		//
		// TODO: should read openbus here
		return 0x00
	}
	val := dma.cpu.Bus.Read8(addr)
	log.ModDMA.DebugZ("read").Hex16("addr", addr).Hex8("val", val).End()
	return val
}

func (dma *DMA) State() *snapshot.DMA {
	return &snapshot.DMA{
		DMCRunning: dma.dmcRunning,
		AbortDMC:   dma.abortDMC,
		OAMRunning: dma.oamRunning,
		DummyCycle: dma.dummy,
		NeedHalt:   dma.needHalt,
	}
}

func (dma *DMA) SetState(state *snapshot.DMA) {
	dma.dmcRunning = state.DMCRunning
	dma.abortDMC = state.AbortDMC
	dma.oamRunning = state.OAMRunning
	dma.dummy = state.DummyCycle
	dma.needHalt = state.NeedHalt
}
