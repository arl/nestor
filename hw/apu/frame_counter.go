package apu

import (
	"nestor/emu/log"
	"nestor/hw/hwdefs"
)

var stepCycles = [2][6]int32{
	{7457, 14913, 22371, 29828, 29829, 29830},
	{7457, 14913, 22371, 29829, 37281, 37282},
}

var frameType = [2][6]FrameType{
	{QuarterFrame, HalfFrame, QuarterFrame, NoFrame, HalfFrame, NoFrame},
	{QuarterFrame, HalfFrame, QuarterFrame, NoFrame, HalfFrame, NoFrame},
}

type FrameCounter struct {
	APU apu
	CPU cpu

	stepCycles        [2][6]int32
	prevCycle         int32
	curStep           uint32
	stepMode          uint32 //0: 4-step mode, 1: 5-step mode
	inhibitIRQ        bool
	blockTick         uint8
	newval            int16
	writeDelayCounter int8
}

func (afc *FrameCounter) Init(apu apu, cpu cpu) {
	afc.APU = apu
	afc.CPU = cpu
}

func (afc *FrameCounter) Reset(soft bool) {
	afc.prevCycle = 0

	// After reset: APU mode in $4017 was unchanged, so we need to keep
	// whatever value stepMode has for soft resets
	if !soft {
		afc.stepMode = 0
	}

	afc.curStep = 0

	// After reset or power-up, APU acts as if $4017 were written with $00
	// from 9 to 12 clocks before first instruction begins. This is
	// emulated in the cpu.reset. function Reset acts as if $00 was written
	// to $4017
	afc.newval = 0
	if afc.stepMode != 0 {
		afc.newval = 0x80
	}
	afc.writeDelayCounter = 3
	afc.inhibitIRQ = false

	afc.blockTick = 0
}

func (afc *FrameCounter) WriteFRAMECOUNTER(old, val uint8) {
	log.ModSound.InfoZ("write framecounter").Uint8("val", val).End()
	afc.APU.Run()
	afc.newval = int16(val)

	// Reset sequence after $4017 is written to
	if afc.CPU.CurrentCycle()&0x01 != 0 {
		// If the write occurs between APU cycles, the effects occur 4 CPU
		// cycles after the write cycle.
		afc.writeDelayCounter = 4
	} else {
		// If the write occurs during an APU cycle, the effects occur 3 CPU
		// cycles after the $4017 write cycle
		afc.writeDelayCounter = 3
	}

	afc.inhibitIRQ = (val & 0x40) == 0x40
	if afc.inhibitIRQ {
		afc.CPU.ClearIRQSource(hwdefs.FrameCounter)
	}
}

// TODO: use return value instead of pointer?
func (afc *FrameCounter) Run(cyclesToRun *int32) uint32 {
	var cyclesRan int32

	if afc.prevCycle+*cyclesToRun >= stepCycles[afc.stepMode][afc.curStep] {
		if !afc.inhibitIRQ && afc.stepMode == 0 && afc.curStep >= 3 {
			// Set irq on the last 3 cycles for 4-step mode
			afc.CPU.SetIRQSource(hwdefs.FrameCounter)
		}

		ftyp := frameType[afc.stepMode][afc.curStep]
		if ftyp != NoFrame && afc.blockTick == 0 {
			afc.APU.FrameCounterTick(ftyp)

			// Do not allow writes to 4017 to clock the frame counter for the
			// next cycle (i.e this odd cycle + the following even cycle)
			afc.blockTick = 2
		}

		if stepCycles[afc.stepMode][afc.curStep] < afc.prevCycle {
			// This can happen when switching from PAL to NTSC, which can cause
			// a freeze (endless loop in APU)
			cyclesRan = 0
		} else {
			cyclesRan = stepCycles[afc.stepMode][afc.curStep] - afc.prevCycle
		}

		*cyclesToRun -= cyclesRan

		afc.curStep++
		if afc.curStep == 6 {
			afc.curStep = 0
			afc.prevCycle = 0
		} else {
			afc.prevCycle += cyclesRan
		}
	} else {
		cyclesRan = *cyclesToRun
		*cyclesToRun = 0
		afc.prevCycle += cyclesRan
	}

	if afc.newval >= 0 {
		afc.writeDelayCounter--
		if afc.writeDelayCounter == 0 {
			// Apply new value after the appropriate number of cycles has elapsed
			if (afc.newval & 0x80) == 0x80 {
				afc.stepMode = 1
			} else {
				afc.stepMode = 0
			}

			afc.writeDelayCounter = -1
			afc.curStep = 0
			afc.prevCycle = 0
			afc.newval = -1

			if afc.stepMode != 0 && afc.blockTick == 0 {
				// Writing to $4017 with bit 7 set will immediately generate
				// a clock for both the quarter frame and the half frame
				// units, regardless of what the sequencer is doing.
				afc.APU.FrameCounterTick(HalfFrame)
				afc.blockTick = 2
			}
		}
	}

	if afc.blockTick > 0 {
		afc.blockTick--
	}

	return uint32(cyclesRan)
}

func (afc *FrameCounter) NeedToRun(cyclesToRun uint32) bool {
	// Run APU when:
	// - A new value is pending
	// - The "blockTick" process is running
	// - We're at the before-last or last tick of the current step
	return afc.newval >= 0 ||
		afc.blockTick > 0 ||
		(afc.prevCycle+int32(cyclesToRun) >= stepCycles[afc.stepMode][afc.curStep]-1)
}
