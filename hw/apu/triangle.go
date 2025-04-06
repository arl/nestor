package apu

import (
	"nestor/emu/log"
	"nestor/hw/hwio"
	"nestor/hw/snapshot"
)

// The triangleChannel contains the following: Timer, 32-step sequencer, Length
// Counter, Linear Counter, 4-bit DAC.
//
//	+---------+    +---------+
//	|LinearCtr|    | Length  |
//	+---------+    +---------+
//	     |              |
//	     v              v
//	+---------+        |\             |\         +---------+    +---------+
//	|  Timer  |------->| >----------->| >------->|Sequencer|--->|   DAC   |
//	+---------+        |/             |/         +---------+    +---------+
type triangleChannel struct {
	apu        *APU
	lenCounter lengthCounter
	timer      timer

	linearCounter       uint8
	linearCounterReload uint8
	linearReload        bool
	linearCtrl          bool

	pos uint8 // current position on "triangleSequence".

	Linear hwio.Reg8 `hwio:"offset=0x08,wcb"`
	Unused hwio.Reg8 `hwio:"offset=0x09,wcb"`
	Timer  hwio.Reg8 `hwio:"offset=0x0A,wcb"`
	Length hwio.Reg8 `hwio:"offset=0x0B,wcb"`
}

func newTriangleChannel(apu *APU, mixer *Mixer) triangleChannel {
	return triangleChannel{
		apu: apu,
		lenCounter: lengthCounter{
			channel: Triangle,
			apu:     apu,
		},
		timer: timer{
			Channel: Triangle,
			mixer:   mixer,
		},
	}
}

var triangleSequence = [32]int8{
	15, 14, 13, 12, 11, 10, 9, 8,
	7, 6, 5, 4, 3, 2, 1, 0,
	0, 1, 2, 3, 4, 5, 6, 7,
	8, 9, 10, 11, 12, 13, 14, 15,
}

func (tc *triangleChannel) run(targetCycle uint32) {
	for tc.timer.run(targetCycle) {
		// The sequencer is clocked by the timer as long as both the linear
		// counter and the length counter are nonzero.
		if tc.lenCounter.status() && tc.linearCounter > 0 {
			tc.pos = (tc.pos + 1) & 0x1F

			if tc.timer.period >= 2 {
				// Disabling the triangle channel when period is < 2 removes
				// "pops" in the audio that are caused by the ultrasonic
				// frequencies
				tc.timer.addOutput(triangleSequence[tc.pos])
			}
		}
	}
}

func (tc *triangleChannel) reset(soft bool) {
	tc.timer.reset(soft)
	tc.lenCounter.reset(soft)

	tc.linearCounter = 0
	tc.linearCounterReload = 0
	tc.linearReload = false
	tc.linearCtrl = false
	tc.pos = 0
}

func (tc *triangleChannel) WriteLINEAR(_, val uint8) {
	tc.apu.Run()
	tc.linearCtrl = (val & 0x80) == 0x80
	tc.linearCounterReload = val & 0x7F

	tc.lenCounter.init(tc.linearCtrl)

	log.ModSound.InfoZ("write triangle linear").
		Uint8("reg", val).
		Bool("ctrl", tc.linearCtrl).
		Uint8("reload", val).
		End()
}

func (tc *triangleChannel) WriteUNUSED(_, _ uint8) {
	tc.apu.Run()
}

func (tc *triangleChannel) WriteTIMER(_, val uint8) {
	tc.apu.Run()

	period := (tc.timer.period & 0xFF00) | uint16(val)
	tc.timer.period = period

	log.ModSound.InfoZ("write triangle timer").
		Uint8("reg", val).
		Uint8("period", val).
		End()
}

func (tc *triangleChannel) WriteLENGTH(_, val uint8) {
	tc.apu.Run()

	tc.lenCounter.load(val >> 3)

	period := (tc.timer.period & 0xFF) | (uint16(val&0x07) << 8)
	tc.timer.period = period

	// Sets the linear counter reload flag (side effect).
	tc.linearReload = true
	log.ModSound.InfoZ("write triangle length").
		Uint8("reg", val).
		Uint16("period", period).
		Uint8("length", val>>3).
		End()
}

func (tc *triangleChannel) tickLinearCounter() {
	if tc.linearReload {
		tc.linearCounter = tc.linearCounterReload
	} else if tc.linearCounter > 0 {
		tc.linearCounter--
	}

	if !tc.linearCtrl {
		tc.linearReload = false
	}
}

func (tc *triangleChannel) tickLengthCounter() {
	tc.lenCounter.tick()
}

func (tc *triangleChannel) teloadLengthCounter() {
	tc.lenCounter.reload()
}

func (tc *triangleChannel) endFrame() {
	tc.timer.endFrame()
}

func (tc *triangleChannel) setEnabled(enabled bool) {
	tc.lenCounter.setEnabled(enabled)
}

func (tc *triangleChannel) status() bool {
	return tc.lenCounter.status()
}

func (tc *triangleChannel) output() uint8 {
	return uint8(tc.timer.lastOutput)
}

func (tc *triangleChannel) saveState(state *snapshot.APUTriangle) {
	tc.lenCounter.saveState(&state.LengthCounter)
	tc.timer.saveState(&state.Timer)
	state.LinearCounter = tc.linearCounter
	state.LinearCounterReload = tc.linearCounterReload
	state.LinearReload = tc.linearReload
	state.LinearCtrl = tc.linearCtrl
	state.Pos = tc.pos
}

func (tc *triangleChannel) setState(state *snapshot.APUTriangle) {
	tc.lenCounter.setState(&state.LengthCounter)
	tc.timer.setState(&state.Timer)
	tc.linearCounter = state.LinearCounter
	tc.linearCounterReload = state.LinearCounterReload
	tc.linearReload = state.LinearReload
	tc.linearCtrl = state.LinearCtrl
	tc.pos = state.Pos
}
