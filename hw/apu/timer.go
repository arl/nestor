package apu

type Timer struct {
	previousCycle uint32
	timer         uint16
	period        uint16
	lastOutput    int8

	channel Channel
	mixer   mixer
}

func NewTimer(channel Channel, mixer mixer) *Timer {
	t := &Timer{
		channel: channel,
		mixer:   mixer,
	}
	t.Reset(false)
	return t
}

func (t *Timer) Reset(_ bool) {
	t.timer = 0
	t.period = 0
	t.previousCycle = 0
	t.lastOutput = 0
}

func (t *Timer) AddOutput(output int8) {
	if output != t.lastOutput {
		t.mixer.AddDelta(t.channel, t.previousCycle, int16(output-t.lastOutput))
		t.lastOutput = output
	}
}

func (t *Timer) LastOutput() int8 {
	return t.lastOutput
}

func (t *Timer) Run(targetCycle uint32) bool {
	cyclesToRun := uint16(targetCycle - t.previousCycle)

	if cyclesToRun > t.timer {
		t.previousCycle += uint32(t.timer) + 1
		t.timer = t.period
		return true
	}

	t.timer -= cyclesToRun
	t.previousCycle = targetCycle
	return false
}

func (t *Timer) EndFrame() {
	t.previousCycle = 0
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
