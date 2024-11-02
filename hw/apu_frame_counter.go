package hw

type FrameType uint8

const (
	NoFrame FrameType = iota
	QuarterFrame
	HalfFrame
)

var stepCycles = [2][6]int32{
	{7457, 14913, 22371, 29828, 29829, 29830},
	{7457, 14913, 22371, 29829, 37281, 37282},
}

var frameType = [2][6]FrameType{
	{QuarterFrame, HalfFrame, QuarterFrame, NoFrame, HalfFrame, NoFrame},
	{QuarterFrame, HalfFrame, QuarterFrame, NoFrame, HalfFrame, NoFrame},
}

type apuFrameCounter struct {
	apu *APU

	stepCycles            [2][6]int32
	previousCycle         int32
	currentStep           uint32
	stepMode              uint32 //0: 4-step mode, 1: 5-step mode
	inhibitIRQ            bool
	blockFrameCounterTick uint8
	newValue              int16
	writeDelayCounter     int8
}

func newAPUFrameCounter() *apuFrameCounter {
	// _console = console;
	afc := &apuFrameCounter{}
	afc.reset(false)
	return afc
}

func (afc *apuFrameCounter) reset(soft bool) {
	afc.previousCycle = 0

	// After reset: APU mode in $4017 was unchanged, so we need to keep
	// whatever value stepMode has for soft resets
	if !soft {
		afc.stepMode = 0
	}

	afc.currentStep = 0

	// After reset or power-up, APU acts as if $4017 were written with $00
	// from 9 to 12 clocks before first instruction begins. This is
	// emulated in the cpu.reset. function Reset acts as if $00 was written
	// to $4017
	afc.newValue = 0
	if afc.stepMode != 0 {
		afc.newValue = 0x80
	}
	afc.writeDelayCounter = 3
	afc.inhibitIRQ = false

	afc.blockFrameCounterTick = 0
}

// TODO: use return value instead of pointer?
func (afc *apuFrameCounter) run(cyclesToRun *int32) uint32 {
	var cyclesRan int32

	if afc.previousCycle+*cyclesToRun >= stepCycles[afc.stepMode][afc.currentStep] {
		if !afc.inhibitIRQ && afc.stepMode == 0 && afc.currentStep >= 3 {
			// Set irq on the last 3 cycles for 4-step mode
			afc.apu.cpu.setIrqSource(frameCounter)
		}

		ftyp := frameType[afc.stepMode][afc.currentStep]
		if ftyp != NoFrame && afc.blockFrameCounterTick == 0 {
			afc.apu.FrameCounterTick(ftyp)

			// Do not allow writes to 4017 to clock the frame counter for the
			// next cycle (i.e this odd cycle + the following even cycle)
			afc.blockFrameCounterTick = 2
		}

		if stepCycles[afc.stepMode][afc.currentStep] < afc.previousCycle {
			// This can happen when switching from PAL to NTSC, which can cause
			// a freeze (endless loop in APU)
			cyclesRan = 0
		} else {
			cyclesRan = stepCycles[afc.stepMode][afc.currentStep] - afc.previousCycle
		}

		*cyclesToRun -= cyclesRan

		afc.currentStep++
		if afc.currentStep == 6 {
			afc.currentStep = 0
			afc.previousCycle = 0
		} else {
			afc.previousCycle += cyclesRan
		}
	} else {
		cyclesRan = *cyclesToRun
		*cyclesToRun = 0
		afc.previousCycle += cyclesRan
	}

	if afc.newValue >= 0 {
		afc.writeDelayCounter--
		if afc.writeDelayCounter == 0 {
			// Apply new value after the appropriate number of cycles has elapsed
			if (afc.newValue & 0x80) == 0x80 {
				afc.stepMode = 1
			} else {
				afc.stepMode = 0
			}

			afc.writeDelayCounter = -1
			afc.currentStep = 0
			afc.previousCycle = 0
			afc.newValue = -1

			if afc.stepMode != 0 && afc.blockFrameCounterTick == 0 {
				// Writing to $4017 with bit 7 set will immediately generate
				// a clock for both the quarter frame and the half frame
				// units, regardless of what the sequencer is doing.
				afc.apu.FrameCounterTick(HalfFrame)
				afc.blockFrameCounterTick = 2
			}
		}
	}

	if afc.blockFrameCounterTick > 0 {
		afc.blockFrameCounterTick--
	}

	return uint32(cyclesRan)
}

func (afc *apuFrameCounter) needToRun(cyclesToRun uint32) bool {
	// Run APU when:
	// - A new value is pending
	// - The "blockFrameCounterTick" process is running
	// - We're at the before-last or last tick of the current step
	return afc.newValue >= 0 ||
		afc.blockFrameCounterTick > 0 ||
		(afc.previousCycle+int32(cyclesToRun) >= stepCycles[afc.stepMode][afc.currentStep]-1)
}
