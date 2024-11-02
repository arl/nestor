package apu

import (
	"nestor/emu/log"
	"nestor/hw/hwio"
)

var noisePeriodLUT = [16]uint16{4, 8, 16, 32, 64, 96, 128, 160, 202, 254, 380, 508, 762, 1016, 2034, 4068}

// NoiseChannel generates pseudo-random 1-bit noise at 16 different frequencies.
//
//	      Timer --> Shift Register   Length Counter
//	                    |                |
//	                    v                v
//	Envelope -------> Gate ----------> Gate --> (to mixer)
type NoiseChannel struct {
	Volume hwio.Reg8 `hwio:"offset=0x0C,writeonly,wcb"`
	Period hwio.Reg8 `hwio:"offset=0x0E,writeonly,wcb"`
	Length hwio.Reg8 `hwio:"offset=0x0F,writeonly,wcb"`

	shiftReg uint16
	mode     bool // mode flag.
	timer    Timer
	envelope Envelope
	apu      apu
}

func NewNoiseChannel(apu apu, mixer mixer) NoiseChannel {
	return NoiseChannel{
		apu: apu,
		envelope: Envelope{
			LengthCounter: LengthCounter{
				channel: Noise,
				apu:     apu,
			},
		},
		timer: *NewTimer(Noise, mixer),
	}
}

func (nc *NoiseChannel) WriteVOLUME(old, val uint8) {
	log.ModSound.InfoZ("write noise volume").Uint8("val", val).End()
	nc.apu.Run()
	nc.envelope.InitializeEnvelope(val)
}

func (nc *NoiseChannel) WritePERIOD(old, val uint8) {
	log.ModSound.InfoZ("write noise period").Uint8("val", val).End()

	nc.apu.Run()
	nc.timer.SetPeriod(noisePeriodLUT[val&0x0F] - 1)
	nc.mode = val&0x80 != 0
}

func (nc *NoiseChannel) WriteLENGTH(old, val uint8) {
	log.ModSound.InfoZ("write noise length").Uint8("val", val).End()
	nc.apu.Run()
	nc.envelope.LengthCounter.Load(val >> 3)
	nc.envelope.ResetEnvelope()
}

func (nc *NoiseChannel) Run(targetCycle uint32) {
	for nc.timer.Run(targetCycle) {
		// Feedback is calculated as the exclusive-OR of bit 0 and one other
		// bit: bit 6 if Mode flag is set, otherwise bit 1.
		modebit := 1
		if nc.mode {
			modebit = 6
		}

		feedback := (nc.shiftReg & 0x01) ^ ((nc.shiftReg >> modebit) & 0x01)
		nc.shiftReg >>= 1
		nc.shiftReg |= (feedback << 14)

		if nc.isMuted() {
			nc.timer.AddOutput(0)
		} else {
			nc.timer.AddOutput(int8(nc.envelope.Volume()))
		}
	}
}

func (nc *NoiseChannel) Output() uint8 {
	return uint8(nc.timer.LastOutput())
}

func (nc *NoiseChannel) isMuted() bool {
	// The mixer receives the current envelope volume except when bit 0 of the
	// shift register is set, or the length counter is zero.
	return (nc.shiftReg & 0x01) == 0x01
}

func (nc *NoiseChannel) TickEnvelope() {
	nc.envelope.Tick()
}

func (nc *NoiseChannel) TickLengthCounter() {
	nc.envelope.LengthCounter.Tick()
}

func (nc *NoiseChannel) ReloadLengthCounter() {
	nc.envelope.LengthCounter.Reload()
}

func (nc *NoiseChannel) EndFrame() {
	nc.timer.EndFrame()
}

func (nc *NoiseChannel) SetEnabled(enabled bool) {
	nc.envelope.LengthCounter.SetEnabled(enabled)
}

func (nc *NoiseChannel) Status() bool {
	return nc.envelope.LengthCounter.Status()
}

func (nc *NoiseChannel) Reset(soft bool) {
	nc.envelope.Reset(soft)
	nc.timer.Reset(soft)

	nc.timer.SetPeriod(noisePeriodLUT[0] - 1)
	nc.shiftReg = 1
	nc.mode = false
}
