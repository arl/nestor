package apu

import (
	"nestor/emu/log"
	"nestor/hw/hwio"
)

type SquareChannel struct {
	apu      apu
	envelope Envelope
	timer    Timer

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

	Duty   hwio.Reg8 `hwio:"offset=0x00,writeonly,wcb"`
	Sweep  hwio.Reg8 `hwio:"offset=0x01,writeonly,wcb"`
	Timer  hwio.Reg8 `hwio:"offset=0x02,writeonly,wcb"`
	Length hwio.Reg8 `hwio:"offset=0x03,writeonly,wcb"`
}

func NewSquareChannel(apu apu, mixer mixer, channel Channel, isChannel1 bool) SquareChannel {
	return SquareChannel{
		apu: apu,
		envelope: Envelope{
			LengthCounter: LengthCounter{
				channel: channel,
				apu:     apu,
			},
		},
		timer:      *NewTimer(channel, mixer),
		isChannel1: isChannel1,
	}
}

func (sc *SquareChannel) WriteDUTY(_, val uint8) {
	sc.apu.Run()

	sc.envelope.InitializeEnvelope(val)
	sc.duty = (val & 0xC0) >> 6

	log.ModSound.InfoZ("write pulse duty").
		Uint8("reg", val).
		Uint8("duty", sc.duty).
		End()
}

func (sc *SquareChannel) WriteSWEEP(_, val uint8) {
	sc.apu.Run()
	sc.initSweep(val)

	log.ModSound.InfoZ("write pulse sweep").
		Uint8("reg", val).
		End()
}

func (sc *SquareChannel) WriteTIMER(_, val uint8) {
	sc.apu.Run()
	period := (sc.realPeriod & 0x0700) | uint16(val)
	sc.setPeriod(period)

	log.ModSound.InfoZ("write pulse timer").
		Uint8("reg", val).
		Uint16("period", period).
		End()
}

func (sc *SquareChannel) WriteLENGTH(_, val uint8) {
	sc.apu.Run()

	envlen := val >> 3
	sc.envelope.LengthCounter.Load(envlen)
	period := (sc.realPeriod & 0xFF) | (uint16(val&0x07) << 8)
	sc.setPeriod(period)

	// The sequencer is restarted at the first value of the current sequence.
	sc.dutyPos = 0

	//The envelope is also restarted.
	sc.envelope.ResetEnvelope()

	log.ModSound.InfoZ("write pulse length").
		Uint8("reg", val).
		Uint8("env len", envlen).
		Uint16("period", period).
		End()
}

func (sc *SquareChannel) isMuted() bool {
	// A period of t < 8, either set explicitly or via a sweep period update,
	// silences the corresponding pulse channel.
	return sc.realPeriod < 8 || (!sc.sweepNegate && sc.sweepTargetPeriod > 0x7FF)
}

func (sc *SquareChannel) initSweep(regValue uint8) {
	sc.sweepEnabled = (regValue & 0x80) == 0x80
	sc.sweepNegate = (regValue & 0x08) == 0x08

	// The divider's period is set to P + 1
	sc.sweepPeriod = ((regValue & 0x70) >> 4) + 1
	sc.sweepShift = (regValue & 0x07)

	sc.updateTargetPeriod()

	// Side effects: Sets the reload flag
	sc.reloadSweep = true
}

func (sc *SquareChannel) updateTargetPeriod() {
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

func (sc *SquareChannel) setPeriod(newPeriod uint16) {
	sc.realPeriod = newPeriod
	sc.timer.SetPeriod((sc.realPeriod * 2) + 1)
	sc.updateTargetPeriod()
}

func (sc *SquareChannel) updateOutput() {
	var dutySequences = [4][8]uint8{
		{0, 0, 0, 0, 0, 0, 0, 1},
		{0, 0, 0, 0, 0, 0, 1, 1},
		{0, 0, 0, 0, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 0, 0},
	}

	if sc.isMuted() {
		sc.timer.AddOutput(0)
	} else {
		out := dutySequences[sc.duty][sc.dutyPos] * uint8(sc.envelope.Volume())
		sc.timer.AddOutput(int8(out))
	}
}

func (sc *SquareChannel) Run(targetCycle uint32) {
	for sc.timer.Run(targetCycle) {
		sc.dutyPos = (sc.dutyPos - 1) & 0x07
		sc.updateOutput()
	}
}

func (sc *SquareChannel) Reset(soft bool) {
	sc.envelope.Reset(soft)
	sc.timer.Reset(soft)

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

func (sc *SquareChannel) TickSweep() {
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

func (sc *SquareChannel) TickEnvelope() {
	sc.envelope.Tick()
}

func (sc *SquareChannel) TickLengthCounter() {
	sc.envelope.LengthCounter.Tick()
}

func (sc *SquareChannel) ReloadLengthCounter() {
	sc.envelope.LengthCounter.Reload()
}

func (sc *SquareChannel) EndFrame() {
	sc.timer.EndFrame()
}

func (sc *SquareChannel) SetEnabled(enabled bool) {
	sc.envelope.LengthCounter.SetEnabled(enabled)
}

func (sc *SquareChannel) Status() bool {
	return sc.envelope.LengthCounter.Status()
}

func (sc *SquareChannel) Output() uint8 {
	return uint8(sc.timer.LastOutput())
}
