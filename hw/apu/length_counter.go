package apu

type LengthCounter struct {
	channel Channel
	newHalt bool

	enabled       bool
	halt          bool
	counter       uint8
	reloadValue   uint8
	previousValue uint8

	apu apu
}

func (lc *LengthCounter) Init(halt bool) {
	lc.apu.SetNeedToRun()
	lc.newHalt = halt
}

func (lc *LengthCounter) Load(val uint8) {
	var lut = [32]uint8{10, 254, 20, 2, 40, 4, 80, 6, 160, 8, 60, 10, 14, 12, 26, 14, 12, 16, 24, 18, 48, 20, 96, 22, 192, 24, 72, 26, 16, 28, 32, 30}

	if lc.enabled {
		lc.reloadValue = lut[val]
		lc.previousValue = lc.counter
		lc.apu.SetNeedToRun()
	}
}

func (lc *LengthCounter) Reset(soft bool) {
	if soft {
		lc.enabled = false
		if lc.channel != Triangle {
			// At reset, length counters should be enabled, triangle unaffected
			lc.halt = false
			lc.counter = 0
			lc.newHalt = false
			lc.reloadValue = 0
			lc.previousValue = 0
		}
	} else {
		lc.enabled = false
		lc.halt = false
		lc.counter = 0
		lc.newHalt = false
		lc.reloadValue = 0
		lc.previousValue = 0
	}
}

func (lc *LengthCounter) Status() bool {
	return lc.counter > 0
}

func (lc *LengthCounter) IsHalted() bool {
	return lc.halt
}

func (lc *LengthCounter) Reload() {
	if lc.reloadValue != 0 {
		if lc.counter == lc.previousValue {
			lc.counter = lc.reloadValue
		}
		lc.reloadValue = 0
	}

	lc.halt = lc.newHalt
}

func (lc *LengthCounter) Tick() {
	if lc.counter > 0 && !lc.halt {
		lc.counter--
	}
}

func (lc *LengthCounter) SetEnabled(enabled bool) {
	if !enabled {
		lc.counter = 0
	}
	lc.enabled = enabled
}

func (lc *LengthCounter) IsEnabled() bool {
	return lc.enabled
}
