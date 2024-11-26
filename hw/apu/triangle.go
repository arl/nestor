package apu

import (
	"nestor/emu/log"
	"nestor/hw/hwio"
)

// The TriangleChannel contains the following: Timer, 32-step sequencer, Length
// Counter, Linear Counter, 4-bit DAC.
//
//  +---------+    +---------+
//  |LinearCtr|    | Length  |
//  +---------+    +---------+
//       |              |
//       v              v
//  +---------+        |\             |\         +---------+    +---------+
//  |  Timer  |------->| >----------->| >------->|Sequencer|--->|   DAC   |
//  +---------+        |/             |/         +---------+    +---------+
//

type TriangleChannel struct {
	apu        apu
	lenCounter lengthCounter
	timer      Timer

	linearCounter       uint8
	linearCounterReload uint8
	linearReload        bool
	linearCtrl          bool

	pos uint8 // current position on "triangleSequence".

	Linear hwio.Reg8 `hwio:"offset=0x08,writeonly,wcb"`
	Timer  hwio.Reg8 `hwio:"offset=0x0A,writeonly,wcb"`
	Length hwio.Reg8 `hwio:"offset=0x0B,writeonly,wcb"`
}

func NewTriangleChannel(apu apu, mixer mixer) TriangleChannel {
	return TriangleChannel{
		apu: apu,
		lenCounter: lengthCounter{
			channel: Triangle,
			apu:     apu,
		},
		timer: Timer{
			Channel: Triangle,
			Mixer:   mixer,
		},
	}
}

var triangleSequence = [32]int8{
	15, 14, 13, 12, 11, 10, 9, 8,
	7, 6, 5, 4, 3, 2, 1, 0,
	0, 1, 2, 3, 4, 5, 6, 7,
	8, 9, 10, 11, 12, 13, 14, 15,
}

func (tc *TriangleChannel) Run(targetCycle uint32) {
	for tc.timer.Run(targetCycle) {
		// The sequencer is clocked by the timer as long as both the linear
		// counter and the length counter are nonzero.
		if tc.lenCounter.status() && tc.linearCounter > 0 {
			tc.pos = (tc.pos + 1) & 0x1F

			if tc.timer.Period() >= 2 {
				// Disabling the triangle channel when period is < 2 removes
				// "pops" in the audio that are caused by the ultrasonic
				// frequencies
				tc.timer.AddOutput(triangleSequence[tc.pos])
			}
		}
	}
}

func (tc *TriangleChannel) Reset(soft bool) {
	tc.timer.Reset(soft)
	tc.lenCounter.reset(soft)

	tc.linearCounter = 0
	tc.linearCounterReload = 0
	tc.linearReload = false
	tc.linearCtrl = false
	tc.pos = 0
}

func (tc *TriangleChannel) WriteLINEAR(_, val uint8) {
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

func (tc *TriangleChannel) WriteTIMER(_, val uint8) {
	tc.apu.Run()

	period := (tc.timer.Period() & 0xFF00) | uint16(val)
	tc.timer.SetPeriod(period)

	log.ModSound.InfoZ("write triangle timer").
		Uint8("reg", val).
		Uint8("period", val).
		End()
}

func (tc *TriangleChannel) WriteLENGTH(_, val uint8) {
	tc.apu.Run()

	tc.lenCounter.load(val >> 3)

	period := (tc.timer.Period() & 0xFF) | (uint16(val&0x07) << 8)
	tc.timer.SetPeriod(period)

	// Sets the linear counter reload flag (side effect).
	tc.linearReload = true
	log.ModSound.InfoZ("write triangle length").
		Uint8("reg", val).
		Uint16("period", period).
		Uint8("length", val>>3).
		End()
}

func (tc *TriangleChannel) TickLinearCounter() {
	if tc.linearReload {
		tc.linearCounter = tc.linearCounterReload
	} else if tc.linearCounter > 0 {
		tc.linearCounter--
	}

	if !tc.linearCtrl {
		tc.linearReload = false
	}
}

func (tc *TriangleChannel) TickLengthCounter() {
	tc.lenCounter.tick()
}

func (tc *TriangleChannel) ReloadLengthCounter() {
	tc.lenCounter.reload()
}

func (tc *TriangleChannel) EndFrame() {
	tc.timer.EndFrame()
}

func (tc *TriangleChannel) SetEnabled(enabled bool) {
	tc.lenCounter.setEnabled(enabled)
}

func (tc *TriangleChannel) Status() bool {
	return tc.lenCounter.status()
}

func (tc *TriangleChannel) Output() uint8 {
	return uint8(tc.timer.LastOutput())
}
