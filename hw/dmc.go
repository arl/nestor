package hw

import (
	"nestor/emu/log"
	"nestor/hw/apu"
	"nestor/hw/hwio"
)

var dmcPeriodLUT = [16]uint16{428, 380, 340, 320, 286, 254, 226, 214, 190, 160, 142, 128, 106, 84, 72, 54}

type DMCChannel struct {
	APU   *APU
	timer apu.Timer

	sampleAddr   uint16
	sampleLength uint16
	outputLevel  uint8
	irqEnabled   bool
	loopFlag     bool

	currentAddr    uint16
	bytesRemaining uint16
	readBuffer     uint8
	bufferEmpty    bool

	shiftRegister      uint8
	bitsRemaining      uint8
	silenceFlag        bool
	needToRun          bool
	disableDelay       uint8
	transferStartDelay uint8

	lastValue4011 uint8

	FLAGS      hwio.Reg8 `hwio:"offset=0x10,writeonly,wcb"`
	LOAD       hwio.Reg8 `hwio:"offset=0x11,writeonly,wcb"`
	SAMPLEADDR hwio.Reg8 `hwio:"offset=0x12,writeonly,wcb"`
	SAMPLELEN  hwio.Reg8 `hwio:"offset=0x13,writeonly,wcb"`
}

func NewDMCChannel(APU *APU, mixer *AudioMixer) DMCChannel {
	return DMCChannel{
		APU:         APU,
		silenceFlag: true,
		timer:       *apu.NewTimer(apu.DMC, mixer),
	}
}

func (dc *DMCChannel) WriteFLAGS(_, val uint8) {
	dc.APU.Run()

	dc.irqEnabled = (val & 0x80) == 0x80
	dc.loopFlag = (val & 0x40) == 0x40

	// The rate determines for how many CPU cycles happen between changes in the
	// output level during automatic delta-encoded sample playback. Because
	// BaseApuChannel does not decrement when setting _timer, we need to
	// actually set the value to 1 less than the lookup table
	period := dmcPeriodLUT[val&0x0F] - 1
	dc.timer.SetPeriod(period)

	if !dc.irqEnabled {
		dc.APU.cpu.clearIrqSource(dmc)
	}

	log.ModSound.InfoZ("write dmc FLAGS").
		Uint8("reg", val).
		Bool("irq enabled", dc.irqEnabled).
		Bool("loop", dc.loopFlag).
		Uint16("period", period).
		End()
}

func abs[T ~int | ~int8 | ~int16 | ~int32 | ~int64](x T) T {
	xmask := x >> 7
	return ((x + xmask) ^ xmask)
}

func (dc *DMCChannel) WriteLOAD(_, val uint8) {
	dc.APU.Run()

	newval := val & 0x7F
	previousLevel := dc.outputLevel
	dc.outputLevel = newval

	if abs(int8(dc.outputLevel)-int8(previousLevel)) > 50 {
		// Reduce popping sounds for 4011 writes
		dc.outputLevel -= (dc.outputLevel - previousLevel) / 2
	}

	// 4011 applies new output right away, not on the timer's reload. This
	// fixes bad DMC sound when playing through 4011.
	dc.timer.AddOutput(int8(dc.outputLevel))

	dc.lastValue4011 = newval

	log.ModSound.InfoZ("write dmc LOAD").
		Uint8("reg", val).
		Uint8("out lvl", dc.outputLevel).
		End()
}

func (dc *DMCChannel) WriteSAMPLEADDR(_, val uint8) {
	dc.APU.Run()
	dc.sampleAddr = 0xC000 | uint16(val)<<6

	log.ModSound.InfoZ("write dmc SAMPLEADDR").
		Uint8("val", val).
		Uint16("addr", dc.sampleAddr).
		End()
}

func (dc *DMCChannel) WriteSAMPLELEN(_, val uint8) {
	dc.APU.Run()
	dc.sampleLength = uint16(val)<<4 | 0x1

	log.ModSound.InfoZ("write dmc SAMPLELEN").
		Uint8("val", val).
		Uint16("len", dc.sampleLength).
		End()
}

func (dc *DMCChannel) Reset(soft bool) {
	dc.timer.Reset(soft)

	if !soft {
		// At power on, the sample address is set to $C000 and the sample length
		// is set to 1 Resetting does not reset their value
		dc.sampleAddr = 0xC000
		dc.sampleLength = 1
	}

	dc.outputLevel = 0
	dc.irqEnabled = false
	dc.loopFlag = false

	dc.currentAddr = 0
	dc.bytesRemaining = 0
	dc.readBuffer = 0
	dc.bufferEmpty = true

	dc.shiftRegister = 0
	dc.bitsRemaining = 8
	dc.silenceFlag = true
	dc.needToRun = false
	dc.transferStartDelay = 0
	dc.disableDelay = 0

	dc.lastValue4011 = 0

	// Not sure if this is accurate, but it seems to make things better rather
	// than worse (for dpcmletterbox) "On the real thing, I think the power-on
	// value is 428 (or the equivalent at least - it uses a linear feedback
	// shift register), though only the even/oddness should matter for this
	// test."
	dc.timer.SetPeriod(dmcPeriodLUT[0] - 1)

	// Make sure the DMC doesn't tick on the first cycle - this is part of what
	// keeps Sprite/DMC DMA tests working while fixing dmcdc.pitch.
	dc.timer.SetTimer(dc.timer.Period())
}

func (dc *DMCChannel) initSample() {
	dc.currentAddr = dc.sampleAddr
	dc.bytesRemaining = dc.sampleLength
	dc.needToRun = dc.needToRun || dc.bytesRemaining > 0
}

func (dc *DMCChannel) startDmcTransfer() {
	if dc.bufferEmpty && dc.bytesRemaining > 0 {
		dc.APU.cpu.startDmcTransfer()
	}
}

func (dc *DMCChannel) getReadAddress() uint16 {
	return dc.currentAddr
}

func (dc *DMCChannel) setReadBuffer(value uint8) {
	if dc.bytesRemaining > 0 {
		dc.readBuffer = value
		dc.bufferEmpty = false

		// The address is incremented; if it exceeds $FFFF,
		// it is wrapped around to $8000.
		dc.currentAddr++
		if dc.currentAddr == 0 {
			dc.currentAddr = 0x8000
		}

		dc.bytesRemaining--

		if dc.bytesRemaining == 0 {
			if dc.loopFlag {
				// Looped sample should never set IRQ flag
				dc.initSample()
			} else if dc.irqEnabled {
				dc.APU.cpu.setIrqSource(dmc)
			}
		}
	}

	if dc.sampleLength == 1 && !dc.loopFlag {
		if dc.bitsRemaining == 1 && dc.timer.Timer() < 2 {
			// When the DMA ends on the APU cycle before the bit counter resets
			// If it this happens right before the bit counter resets,
			// a DMA is triggered and aborted 1 cycle later (causing one halted CPU cycle)
			dc.shiftRegister = dc.readBuffer
			dc.bufferEmpty = false
			dc.initSample()
			dc.disableDelay = 3
		}
	}
}

func (dc *DMCChannel) Run(targetCycle uint32) {
	for dc.timer.Run(targetCycle) {
		if !dc.silenceFlag {
			if dc.shiftRegister&0x01 != 0 {
				if dc.outputLevel <= 125 {
					dc.outputLevel += 2
				}
			} else {
				if dc.outputLevel >= 2 {
					dc.outputLevel -= 2
				}
			}
			dc.shiftRegister >>= 1
		}

		dc.bitsRemaining--
		if dc.bitsRemaining == 0 {
			dc.bitsRemaining = 8
			if dc.bufferEmpty {
				dc.silenceFlag = true
			} else {
				dc.silenceFlag = false
				dc.shiftRegister = dc.readBuffer
				dc.bufferEmpty = true
				dc.needToRun = true
				dc.startDmcTransfer()
			}
		}

		dc.timer.AddOutput(int8(dc.outputLevel))
	}
}

func (dc *DMCChannel) IRQPending(cyclesToRun uint32) bool {
	if dc.irqEnabled && dc.bytesRemaining > 0 {
		cyclesToEmptyBuffer := (uint16(dc.bitsRemaining) + (dc.bytesRemaining-1)*8) * dc.timer.Period()
		if cyclesToRun >= uint32(cyclesToEmptyBuffer) {
			return true
		}
	}
	return false
}

func (dc *DMCChannel) Status() bool {
	return dc.bytesRemaining > 0
}

func (dc *DMCChannel) EndFrame() {
	dc.timer.EndFrame()
}

func (dc *DMCChannel) SetEnabled(enabled bool) {
	if !enabled {
		if dc.disableDelay == 0 {
			// Disabling takes effect with a 1 apu cycle delay
			// If a DMA starts during this time, it gets cancelled
			// but this will still cause the CPU to be halted for 1 cycle
			if (dc.APU.cpu.Cycles & 0x01) == 0 {
				dc.disableDelay = 2
			} else {
				dc.disableDelay = 3
			}
		}
		dc.needToRun = true
	} else if dc.bytesRemaining == 0 {
		dc.initSample()

		// Delay a number of cycles based on odd/even cycles
		// Allows behavior to match dmc_dma_start_test
		if (dc.APU.cpu.Cycles & 0x01) == 0 {
			dc.transferStartDelay = 2
		} else {
			dc.transferStartDelay = 3
		}
		dc.needToRun = true
	}
}

func (dc *DMCChannel) processClock() {
	if dc.disableDelay > 0 {
		dc.disableDelay--
		if dc.disableDelay == 0 {
			dc.bytesRemaining = 0

			// Abort any on-going transfer that hasn't fully started
			dc.APU.cpu.stopDmcTransfer()
		}
	}

	if dc.transferStartDelay > 0 {
		dc.transferStartDelay--
		if dc.transferStartDelay == 0 {
			dc.startDmcTransfer()
		}
	}

	dc.needToRun = dc.disableDelay != 0 || dc.transferStartDelay != 0 || dc.bytesRemaining != 0
}

func (dc *DMCChannel) NeedToRun() bool {
	if dc.needToRun {
		dc.processClock()
	}
	return dc.needToRun
}

func (dc *DMCChannel) Output() uint8 {
	return uint8(dc.timer.LastOutput())
}
