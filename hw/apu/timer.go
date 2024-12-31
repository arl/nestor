package apu

// timer is a divider driven by the ~1.79 MHz clock and is used by all APU
// channels.
type timer struct {
	Mixer mixer

	prevCycle  uint32
	timer      uint16
	period     uint16
	lastOutput int8

	Channel Channel
}

func (t *timer) reset(_ bool) {
	t.timer = 0
	t.period = 0
	t.prevCycle = 0
	t.lastOutput = 0
}

func (t *timer) addOutput(output int8) {
	if output != t.lastOutput {
		t.Mixer.AddDelta(t.Channel, t.prevCycle, int16(output-t.lastOutput))
		t.lastOutput = output
	}
}

func (t *timer) run(targetCycle uint32) bool {
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

func (t *timer) endFrame() {
	t.prevCycle = 0
}
