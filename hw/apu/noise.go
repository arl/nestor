package apu

import (
	"nestor/emu/log"
	"nestor/hw/hwio"
)

// NoiseChannel generates pseudo-random 1-bit noise at 16 different frequencies.
//
//	      Timer --> Shift Register   Length Counter
//	                    |                |
//	                    v                v
//	Envelope -------> Gate ----------> Gate --> (to mixer)
type NoiseChannel struct {
	Volume hwio.Reg8 `hwio:"offset=0x0C,wcb"`
	Unused hwio.Reg8 `hwio:"offset=0x0D,wcb"`
	Period hwio.Reg8 `hwio:"offset=0x0E,wcb"`
	Length hwio.Reg8 `hwio:"offset=0x0F,wcb"`

	shiftReg uint16
	mode     bool // mode flag.
	timer    timer
	env      envelope
	apu      apu
}

func NewNoiseChannel(apu apu, mixer mixer) NoiseChannel {
	return NoiseChannel{
		apu: apu,
		env: envelope{
			lenCounter: lengthCounter{
				channel: Noise,
				apu:     apu,
			},
		},
		timer: timer{
			Channel: Noise,
			Mixer:   mixer,
		},
	}
}

func (nc *NoiseChannel) WriteVOLUME(old, val uint8) {
	log.ModSound.InfoZ("write noise volume").Uint8("val", val).End()
	nc.apu.Run()
	nc.env.init(val)
}

func (nc *NoiseChannel) WriteUNUSED(_, _ uint8) {
	nc.apu.Run()
}

var noisePeriodLUT = [16]uint16{4, 8, 16, 32, 64, 96, 128, 160, 202, 254, 380, 508, 762, 1016, 2034, 4068}

func (nc *NoiseChannel) WritePERIOD(old, val uint8) {
	log.ModSound.InfoZ("write noise period").Uint8("val", val).End()

	nc.apu.Run()
	nc.timer.period = noisePeriodLUT[val&0x0F] - 1
	nc.mode = val&0x80 != 0
}

func (nc *NoiseChannel) WriteLENGTH(old, val uint8) {
	log.ModSound.InfoZ("write noise length").Uint8("val", val).End()
	nc.apu.Run()
	nc.env.lenCounter.load(val >> 3)
	nc.env.restart()
}

func (nc *NoiseChannel) Run(targetCycle uint32) {
	for nc.timer.run(targetCycle) {
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
			nc.timer.addOutput(0)
		} else {
			nc.timer.addOutput(int8(nc.env.volume()))
		}
	}
}

func (nc *NoiseChannel) Output() uint8 {
	return uint8(nc.timer.lastOutput)
}

func (nc *NoiseChannel) isMuted() bool {
	// The mixer receives the current envelope volume except when bit 0 of the
	// shift register is set, or the length counter is zero.
	return (nc.shiftReg & 0x01) == 0x01
}

func (nc *NoiseChannel) TickEnvelope() {
	nc.env.tick()
}

func (nc *NoiseChannel) TickLengthCounter() {
	nc.env.lenCounter.tick()
}

func (nc *NoiseChannel) ReloadLengthCounter() {
	nc.env.lenCounter.reload()
}

func (nc *NoiseChannel) EndFrame() {
	nc.timer.endFrame()
}

func (nc *NoiseChannel) SetEnabled(enabled bool) {
	nc.env.lenCounter.setEnabled(enabled)
}

func (nc *NoiseChannel) Status() bool {
	return nc.env.lenCounter.status()
}

func (nc *NoiseChannel) Reset(soft bool) {
	nc.env.reset(soft)
	nc.timer.reset(soft)

	nc.timer.period = noisePeriodLUT[0] - 1
	nc.shiftReg = 1
	nc.mode = false
}
