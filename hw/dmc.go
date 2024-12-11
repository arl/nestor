package hw

import (
	"nestor/emu/log"
	"nestor/hw/apu"
	"nestor/hw/hwio"
)

// The DMC (Delta Modulation Channel) can output samples composed of 1-bit
// deltas and its DAC can be directly changed. It contains the following: DMA
// reader, interrupt flag, sample buffer, Timer, output unit, 7-bit counter tied
// to 7-bit DAC.
//
//	+----------+    +---------+
//	|DMA Reader|    |  Timer  |
//	+----------+    +---------+
//	     |               |
//	     |               v
//	+----------+    +---------+     +---------+     +---------+
//	|  Buffer  |----| Output  |---->| Counter |---->|   DAC   |
//	+----------+    +---------+     +---------+     +---------+
type DMC struct {
	APU   *APU
	timer apu.Timer

	sampleAddr   uint16
	sampleLength uint16
	outputLevel  uint8
	irqEnabled   bool
	loopFlag     bool

	curaddr     uint16
	remaining   uint16
	readBuffer  uint8
	bufferEmpty bool

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

func NewDMC(APU *APU, mixer *AudioMixer) DMC {
	return DMC{
		APU:         APU,
		silenceFlag: true,
		timer: apu.Timer{
			Channel: apu.DMC,
			Mixer:   mixer,
		},
	}
}

func (dc *DMC) initSample() {
	dc.curaddr = dc.sampleAddr
	dc.remaining = dc.sampleLength
	dc.needToRun = dc.needToRun || dc.remaining > 0
}

func (dc *DMC) Reset(soft bool) {
	dc.timer.Reset(soft)

	if !soft {
		// At power on, the sample address is set to $C000 and the sample length
		// is set to 1. Resetting does not reset their value
		dc.sampleAddr = 0xC000
		dc.sampleLength = 1
	}

	dc.outputLevel = 0
	dc.irqEnabled = false
	dc.loopFlag = false

	dc.curaddr = 0
	dc.remaining = 0
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
	period := dmcPeriodLUT[0] - 1
	dc.timer.SetPeriod(period)

	// Make sure the DMC doesn't tick on the first cycle - this is part of what
	// keeps Sprite/DMC DMA tests working while fixing dmcdc.pitch.
	dc.timer.SetTimer(dc.timer.Period())
}

var dmcPeriodLUT = [16]uint16{428, 380, 340, 320, 286, 254, 226, 214, 190, 160, 142, 128, 106, 84, 72, 54}

// $4010
func (dc *DMC) WriteFLAGS(_, val uint8) {
	dc.APU.Run()

	dc.irqEnabled = (val & 0x80) == 0x80
	dc.loopFlag = (val & 0x40) == 0x40

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

// $4011
func (dc *DMC) WriteLOAD(_, val uint8) {
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

// $4012 start of DMC sample is at address $C000 + $40*$xx
func (dc *DMC) WriteSAMPLEADDR(_, val uint8) {
	dc.APU.Run()
	dc.sampleAddr = 0xC000 | uint16(val)<<6

	log.ModSound.InfoZ("write dmc SAMPLEADDR").
		Uint8("val", val).
		Uint16("addr", dc.sampleAddr).
		End()
}

// $4013 Length of DMC waveform is $10*$xx + 1 bytes (128*$xx + 8 samples)
func (dc *DMC) WriteSAMPLELEN(_, val uint8) {
	dc.APU.Run()
	dc.sampleLength = uint16(val)<<4 | 0x1

	log.ModSound.InfoZ("write dmc SAMPLELEN").
		Uint8("val", val).
		Uint16("len", dc.sampleLength).
		End()
}

func (dc *DMC) startDMCTransfer() {
	if dc.bufferEmpty && dc.remaining > 0 {
		dc.APU.cpu.startDmcTransfer()
	}
}

func (dc *DMC) getReadAddress() uint16 {
	return dc.curaddr
}

func (dc *DMC) setReadBuffer(val uint8) {
	log.ModSound.DebugZ("set DMC read buffer").
		Uint8("value", val).
		End()

	if dc.remaining > 0 {
		dc.readBuffer = val
		dc.bufferEmpty = false

		// address wraps around to $8000, not $0000.
		dc.curaddr++
		if dc.curaddr == 0 {
			dc.curaddr = 0x8000
		}

		dc.remaining--

		if dc.remaining == 0 {
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

func (dc *DMC) Run(targetCycle uint32) {
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
				dc.startDMCTransfer()
			}
		}

		dc.timer.AddOutput(int8(dc.outputLevel))
	}
}

func (dc *DMC) IRQPending(cyclesToRun uint32) bool {
	if dc.irqEnabled && dc.remaining > 0 {
		cyclesToEmptyBuffer := (uint16(dc.bitsRemaining) + (dc.remaining-1)*8) * dc.timer.Period()
		if cyclesToRun >= uint32(cyclesToEmptyBuffer) {
			return true
		}
	}
	return false
}

func (dc *DMC) Status() bool {
	return dc.remaining > 0
}

func (dc *DMC) EndFrame() {
	dc.timer.EndFrame()
}

func (dc *DMC) SetEnabled(enabled bool) {
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
	} else if dc.remaining == 0 {
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

func (dc *DMC) processClock() {
	if dc.disableDelay != 0 {
		dc.disableDelay--
		if dc.disableDelay == 0 {
			dc.remaining = 0
			// Abort any on-going transfer that hasn't fully started
			dc.APU.cpu.stopDmcTransfer()
		}
	}

	if dc.transferStartDelay != 0 {
		dc.transferStartDelay--
		if dc.transferStartDelay == 0 {
			dc.startDMCTransfer()
		}
	}

	dc.needToRun = dc.disableDelay != 0 || dc.transferStartDelay != 0 || dc.remaining != 0
}

func (dc *DMC) NeedToRun() bool {
	if dc.needToRun {
		dc.processClock()
	}
	return dc.needToRun
}

func (dc *DMC) Output() uint8 {
	return uint8(dc.timer.LastOutput())
}
