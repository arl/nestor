package apu

import (
	"nestor/emu/log"
	"nestor/hw/hwio"
	"nestor/hw/snapshot"
)

// There are two square channels beginning at registers $4000 and $4004. Each
// contains the following: Envelope Generator, Sweep Unit, Timer with
// divide-by-two on the output, 8-step sequencer, Length Counter.
//
//	               +---------+    +---------+
//	               |  Sweep  |--->|Timer / 2|
//	               +---------+    +---------+
//	                    |              |
//	                    |              v
//	                    |         +---------+    +---------+
//	                    |         |Sequencer|    | Length  |
//	                    |         +---------+    +---------+
//	                    |              |              |
//	                    v              v              v
//	+---------+        |\             |\             |\          +---------+
//	|Envelope |------->| >----------->| >----------->| >-------->|   DAC   |
//	+---------+        |/             |/             |/          +---------+
type squareChannel struct {
	apu      *APU
	envelope envelope
	timer    timer

	isChannel1 bool

	duty    uint8
	dutyPos uint8

	sweepEnabled      bool
	sweepPeriod       uint8
	sweepNegate       bool
	sweepShift        uint8
	reloadSweep       bool
	sweepDivider      uint8
	sweepTargetPeriod uint32
	realPeriod        uint16

	Duty   hwio.Reg8 `hwio:"offset=0x00,wcb"`
	Sweep  hwio.Reg8 `hwio:"offset=0x01,wcb"`
	Timer  hwio.Reg8 `hwio:"offset=0x02,wcb"`
	Length hwio.Reg8 `hwio:"offset=0x03,wcb"`
}

func newSquareChannel(apu *APU, mixer *Mixer, channel Channel, isChannel1 bool) squareChannel {
	return squareChannel{
		isChannel1: isChannel1,
		apu:        apu,
		envelope: envelope{
			lenCounter: lengthCounter{
				channel: channel,
				apu:     apu,
			},
		},
		timer: timer{
			Channel: channel,
			mixer:   mixer,
		},
	}
}

func (sc *squareChannel) WriteDUTY(_, val uint8) {
	sc.apu.Run()

	sc.envelope.init(val)
	sc.duty = (val & 0xC0) >> 6

	log.ModSound.InfoZ("write pulse duty").
		Uint8("reg", val).
		Uint8("duty", sc.duty).
		End()
}

func (sc *squareChannel) WriteSWEEP(_, val uint8) {
	sc.apu.Run()
	sc.initSweep(val)

	log.ModSound.InfoZ("write pulse sweep").
		Uint8("reg", val).
		End()
}

func (sc *squareChannel) WriteTIMER(_, val uint8) {
	sc.apu.Run()
	period := (sc.realPeriod & 0x0700) | uint16(val)
	sc.setPeriod(period)

	log.ModSound.InfoZ("write pulse timer").
		Uint8("reg", val).
		Uint16("period", period).
		End()
}

func (sc *squareChannel) WriteLENGTH(_, val uint8) {
	sc.apu.Run()

	envlen := val >> 3
	sc.envelope.lenCounter.load(envlen)
	period := (sc.realPeriod & 0xFF) | (uint16(val&0x07) << 8)
	sc.setPeriod(period)

	// sequencer is restarted at the first value of the current sequence.
	sc.dutyPos = 0

	// envelope is also restarted.
	sc.envelope.restart()

	log.ModSound.InfoZ("write pulse length").
		Uint8("reg", val).
		Uint8("env len", envlen).
		Uint16("period", period).
		End()
}

func (sc *squareChannel) isMuted() bool {
	// A period of t < 8, either set explicitly or via a sweep period update,
	// silences the corresponding pulse channel.
	return sc.realPeriod < 8 || (!sc.sweepNegate && sc.sweepTargetPeriod > 0x7FF)
}

func (sc *squareChannel) initSweep(regValue uint8) {
	sc.sweepEnabled = (regValue & 0x80) == 0x80
	sc.sweepNegate = (regValue & 0x08) == 0x08

	// The divider's period is set to P + 1
	sc.sweepPeriod = ((regValue & 0x70) >> 4) + 1
	sc.sweepShift = (regValue & 0x07)

	sc.updateTargetPeriod()

	// Side effects: Sets the reload flag
	sc.reloadSweep = true
}

func (sc *squareChannel) updateTargetPeriod() {
	shiftResult := (sc.realPeriod >> sc.sweepShift)
	if sc.sweepNegate {
		sc.sweepTargetPeriod = uint32(sc.realPeriod - shiftResult)
		if sc.isChannel1 {
			// As a result, a negative sweep on pulse channel 1 will subtract
			// the shifted period value minus 1
			sc.sweepTargetPeriod--
		}
	} else {
		sc.sweepTargetPeriod = uint32(sc.realPeriod + shiftResult)
	}
}

func (sc *squareChannel) setPeriod(newPeriod uint16) {
	sc.realPeriod = newPeriod
	sc.timer.period = (sc.realPeriod * 2) + 1
	sc.updateTargetPeriod()
}

// duty cycle sequences for the square channels.
var squareDuty = [4][8]uint8{
	{0, 0, 0, 0, 0, 0, 0, 1},
	{0, 0, 0, 0, 0, 0, 1, 1},
	{0, 0, 0, 0, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 0, 0},
}

func (sc *squareChannel) updateOutput() {
	if sc.isMuted() {
		sc.timer.addOutput(0)
	} else {
		out := squareDuty[sc.duty][sc.dutyPos] * uint8(sc.envelope.volume())
		sc.timer.addOutput(int8(out))
	}
}

func (sc *squareChannel) run(targetCycle uint32) {
	for sc.timer.run(targetCycle) {
		sc.dutyPos = (sc.dutyPos - 1) & 0x07
		sc.updateOutput()
	}
}

func (sc *squareChannel) reset(soft bool) {
	sc.envelope.reset(soft)
	sc.timer.reset(soft)

	sc.duty = 0
	sc.dutyPos = 0

	sc.realPeriod = 0

	sc.sweepEnabled = false
	sc.sweepPeriod = 0
	sc.sweepNegate = false
	sc.sweepShift = 0
	sc.reloadSweep = false
	sc.sweepDivider = 0
	sc.sweepTargetPeriod = 0
	sc.updateTargetPeriod()
}

func (sc *squareChannel) tickSweep() {
	sc.sweepDivider--
	if sc.sweepDivider == 0 {
		if sc.sweepShift > 0 && sc.sweepEnabled && sc.realPeriod >= 8 && sc.sweepTargetPeriod <= 0x7FF {
			sc.setPeriod(uint16(sc.sweepTargetPeriod))
		}
		sc.sweepDivider = sc.sweepPeriod
	}

	if sc.reloadSweep {
		sc.sweepDivider = sc.sweepPeriod
		sc.reloadSweep = false
	}
}

func (sc *squareChannel) tickEnvelope() {
	sc.envelope.tick()
}

func (sc *squareChannel) tickLengthCounter() {
	sc.envelope.lenCounter.tick()
}

func (sc *squareChannel) reloadLengthCounter() {
	sc.envelope.lenCounter.reload()
}

func (sc *squareChannel) endFrame() {
	sc.timer.endFrame()
}

func (sc *squareChannel) setEnabled(enabled bool) {
	sc.envelope.lenCounter.setEnabled(enabled)
}

func (sc *squareChannel) status() bool {
	return sc.envelope.lenCounter.status()
}

func (sc *squareChannel) output() uint8 {
	return uint8(sc.timer.lastOutput)
}

func (sc *squareChannel) saveState(state *snapshot.APUSquare) {
	state.SweepTargetPeriod = sc.sweepTargetPeriod
	state.RealPeriod = sc.realPeriod
	sc.timer.saveState(&state.Timer)
	sc.envelope.saveState(&state.Envelope)
	state.SweepEnabled = sc.sweepEnabled
	state.SweepPeriod = sc.sweepPeriod
	state.SweepNegate = sc.sweepNegate
	state.SweepShift = sc.sweepShift
	state.SweepDivider = sc.sweepDivider
	state.ReloadSweep = sc.reloadSweep
	state.Duty = sc.duty
	state.DutyPos = sc.dutyPos
}

func (sc *squareChannel) setState(state *snapshot.APUSquare) {
	sc.sweepTargetPeriod = state.SweepTargetPeriod
	sc.realPeriod = state.RealPeriod
	sc.timer.setState(&state.Timer)
	sc.envelope.setState(&state.Envelope)
	sc.sweepEnabled = state.SweepEnabled
	sc.sweepPeriod = state.SweepPeriod
	sc.sweepNegate = state.SweepNegate
	sc.sweepShift = state.SweepShift
	sc.sweepDivider = state.SweepDivider
	sc.reloadSweep = state.ReloadSweep
	sc.duty = state.Duty
	sc.dutyPos = state.DutyPos
}
