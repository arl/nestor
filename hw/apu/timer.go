package apu

// Timer is a divider driven by the ~1.79 MHz clock and is used by all APU
// channels.
type Timer struct {
	Mixer mixer

	prevCycle  uint32
	timer      uint16
	period     uint16
	lastOutput int8

	Channel Channel
}

func (t *Timer) Reset(_ bool) {
	t.timer = 0
	t.period = 0
	t.prevCycle = 0
	t.lastOutput = 0
}

func (t *Timer) AddOutput(output int8) {
	if output != t.lastOutput {
		t.Mixer.AddDelta(t.Channel, t.prevCycle, int16(output-t.lastOutput))
		t.lastOutput = output
	}
}

func (t *Timer) LastOutput() int8 {
	return t.lastOutput
}

func (t *Timer) Run(targetCycle uint32) bool {
	cyclesToRun := uint16(targetCycle - t.prevCycle)

	if cyclesToRun > t.timer {
		t.prevCycle += uint32(t.timer) + 1
		t.timer = t.period
		return true
	}

	t.timer -= cyclesToRun
	t.prevCycle = targetCycle
	return false
}

func (t *Timer) EndFrame() {
	t.prevCycle = 0
}

func (t *Timer) SetPeriod(period uint16) {
	t.period = period
}

func (t *Timer) Period() uint16 {
	return t.period
}

func (t *Timer) Timer() uint16 {
	return t.timer
}

func (t *Timer) SetTimer(timer uint16) {
	t.timer = timer
}
