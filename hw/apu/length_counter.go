package apu

import "nestor/hw/snapshot"

// The length counter allows automatic duration control. Counting can be halted
// and the counter can be disabled by clearing the appropriate bit in the status
// register, which immediately sets the counter to 0 and keeps it there.
type lengthCounter struct {
	apu *APU

	channel Channel
	newHalt bool

	enabled   bool
	halt      bool
	counter   uint8
	reloadVal uint8
	prevVal   uint8
}

func (lc *lengthCounter) init(halt bool) {
	lc.apu.SetNeedToRun()
	lc.newHalt = halt
}

var lenCounterLUT = [32]uint8{
	10, 254, 20, 2, 40, 4, 80, 6,
	160, 8, 60, 10, 14, 12, 26, 14,
	12, 16, 24, 18, 48, 20, 96, 22,
	192, 24, 72, 26, 16, 28, 32, 30,
}

func (lc *lengthCounter) load(val uint8) {
	if lc.enabled {
		lc.reloadVal = lenCounterLUT[val]
		lc.prevVal = lc.counter
		lc.apu.SetNeedToRun()
	}
}

func (lc *lengthCounter) reset(soft bool) {
	if soft {
		lc.enabled = false
		if lc.channel != Triangle {
			// At reset, length counters should be enabled, triangle unaffected
			lc.halt = false
			lc.counter = 0
			lc.newHalt = false
			lc.reloadVal = 0
			lc.prevVal = 0
		}
	} else {
		lc.enabled = false
		lc.halt = false
		lc.counter = 0
		lc.newHalt = false
		lc.reloadVal = 0
		lc.prevVal = 0
	}
}

func (lc *lengthCounter) status() bool {
	return lc.counter > 0
}

func (lc *lengthCounter) isHalted() bool {
	return lc.halt
}

func (lc *lengthCounter) reload() {
	if lc.reloadVal != 0 {
		if lc.counter == lc.prevVal {
			lc.counter = lc.reloadVal
		}
		lc.reloadVal = 0
	}

	lc.halt = lc.newHalt
}

func (lc *lengthCounter) tick() {
	if lc.counter > 0 && !lc.halt {
		lc.counter--
	}
}

func (lc *lengthCounter) setEnabled(enabled bool) {
	if !enabled {
		lc.counter = 0
	}
	lc.enabled = enabled
}

func (lc *lengthCounter) isEnabled() bool {
	return lc.enabled
}

func (lc *lengthCounter) saveState(state *snapshot.APULengthCounter) {
	state.Enabled = lc.enabled
	state.Halt = lc.halt
	state.NewHalt = lc.newHalt
	state.Counter = lc.counter
	state.PrevVal = lc.prevVal
	state.ReloadVal = lc.reloadVal
}

func (lc *lengthCounter) setState(state *snapshot.APULengthCounter) {
	lc.enabled = state.Enabled
	lc.halt = state.Halt
	lc.newHalt = state.NewHalt
	lc.counter = state.Counter
	lc.prevVal = state.PrevVal
	lc.reloadVal = state.ReloadVal
}
