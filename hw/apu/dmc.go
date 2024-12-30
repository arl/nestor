package apu

import (
	"nestor/emu/log"
	"nestor/hw/hwdefs"
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
	APU   apu
	CPU   cpu
	timer Timer

	sampleAddr uint16
	sampleLen  uint16
	outlvl     uint8
	irqEnabled bool
	loop       bool

	curaddr   uint16
	remaining uint16
	readbuf   uint8
	bufEmpty  bool

	shiftReg     uint8
	bitsLeft     uint8
	silence      bool
	needToRun    bool
	disableDelay uint8
	startDelay   uint8 // delay before transfer starts

	last4011 uint8

	FLAGS      hwio.Reg8 `hwio:"offset=0x10,writeonly,wcb"`
	LOAD       hwio.Reg8 `hwio:"offset=0x11,writeonly,wcb"`
	SAMPLEADDR hwio.Reg8 `hwio:"offset=0x12,writeonly,wcb"`
	SAMPLELEN  hwio.Reg8 `hwio:"offset=0x13,writeonly,wcb"`
}

func NewDMC(apu apu, cpu cpu, mixer mixer) DMC {
	return DMC{
		APU:     apu,
		CPU:     cpu,
		silence: true,
		timer: Timer{
			Channel: DPCM,
			Mixer:   mixer,
		},
	}
}

func (dc *DMC) initSample() {
	dc.curaddr = dc.sampleAddr
	dc.remaining = dc.sampleLen
	dc.needToRun = dc.needToRun || dc.remaining > 0
}

func (dc *DMC) Reset(soft bool) {
	dc.timer.Reset(soft)

	if !soft {
		dc.sampleAddr = 0xC000
		dc.sampleLen = 1
	}

	dc.outlvl = 0
	dc.irqEnabled = false
	dc.loop = false

	dc.curaddr = 0
	dc.remaining = 0
	dc.readbuf = 0
	dc.bufEmpty = true

	dc.shiftReg = 0
	dc.bitsLeft = 8
	dc.silence = true
	dc.needToRun = false
	dc.startDelay = 0
	dc.disableDelay = 0

	dc.last4011 = 0

	period := dmcPeriodLUT[0] - 1
	dc.timer.SetPeriod(period)

	// Prevent DMC to tick on first cycle (so that sprite DMC/DMA test pass).
	dc.timer.SetTimer(dc.timer.Period())
}

var dmcPeriodLUT = [16]uint16{428, 380, 340, 320, 286, 254, 226, 214, 190, 160, 142, 128, 106, 84, 72, 54}

// $4010
func (dc *DMC) WriteFLAGS(_, val uint8) {
	dc.APU.Run()

	dc.irqEnabled = (val & 0x80) == 0x80
	dc.loop = (val & 0x40) == 0x40

	period := dmcPeriodLUT[val&0x0F] - 1
	dc.timer.SetPeriod(period)

	if !dc.irqEnabled {
		dc.CPU.ClearIrqSource(hwdefs.DMC)
	}

	log.ModSound.InfoZ("write dmc FLAGS").
		Uint8("reg", val).
		Bool("irq enabled", dc.irqEnabled).
		Bool("loop", dc.loop).
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
	previousLevel := dc.outlvl
	dc.outlvl = newval

	if abs(int8(dc.outlvl)-int8(previousLevel)) > 50 {
		// Reduce popping sounds for 4011 writes
		dc.outlvl -= (dc.outlvl - previousLevel) / 2
	}

	// 4011 applies new output right away, not on the timer's reload. This
	// fixes bad DMC sound when playing through 4011.
	dc.timer.AddOutput(int8(dc.outlvl))

	dc.last4011 = newval

	log.ModSound.InfoZ("write dmc LOAD").
		Uint8("reg", val).
		Uint8("out lvl", dc.outlvl).
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
	dc.sampleLen = uint16(val)<<4 | 0x1

	log.ModSound.InfoZ("write dmc SAMPLELEN").
		Uint8("val", val).
		Uint16("len", dc.sampleLen).
		End()
}

func (dc *DMC) startDMCTransfer() {
	if dc.bufEmpty && dc.remaining > 0 {
		dc.CPU.StartDMCTransfer()
	}
}

func (dc *DMC) CurrentAddr() uint16 {
	return dc.curaddr
}

func (dc *DMC) SetReadBuffer(val uint8) {
	log.ModSound.DebugZ("set DMC read buffer").
		Uint8("value", val).
		End()

	if dc.remaining > 0 {
		dc.readbuf = val
		dc.bufEmpty = false

		// Address wraps around to $8000, not $0000.
		dc.curaddr++
		if dc.curaddr == 0 {
			dc.curaddr = 0x8000
		}

		dc.remaining--

		if dc.remaining == 0 {
			if dc.loop {
				// Looped sample should never set IRQ flag
				dc.initSample()
			} else if dc.irqEnabled {
				dc.CPU.SetIrqSource(hwdefs.DMC)
			}
		}
	}

	if dc.sampleLen == 1 && !dc.loop {
		if dc.bitsLeft == 1 && dc.timer.Timer() < 2 {
			// When the DMA ends on the APU cycle before the bit, counter
			// resets. If it this happens right before the bit counter resets, a
			// DMA is triggered and aborted 1 cycle later (causing one halted
			// CPU cycle).
			dc.shiftReg = dc.readbuf
			dc.bufEmpty = false
			dc.initSample()
			dc.disableDelay = 3
		}
	}
}

func (dc *DMC) Run(targetCycle uint32) {
	for dc.timer.Run(targetCycle) {
		if !dc.silence {
			if dc.shiftReg&0x01 != 0 {
				if dc.outlvl <= 125 {
					dc.outlvl += 2
				}
			} else {
				if dc.outlvl >= 2 {
					dc.outlvl -= 2
				}
			}
			dc.shiftReg >>= 1
		}

		dc.bitsLeft--
		if dc.bitsLeft == 0 {
			dc.bitsLeft = 8
			if dc.bufEmpty {
				dc.silence = true
			} else {
				dc.silence = false
				dc.shiftReg = dc.readbuf
				dc.bufEmpty = true
				dc.needToRun = true
				dc.startDMCTransfer()
			}
		}

		dc.timer.AddOutput(int8(dc.outlvl))
	}
}

func (dc *DMC) IRQPending(cyclesToRun uint32) bool {
	if dc.irqEnabled && dc.remaining > 0 {
		// IRQ is set when the sample buffer is emptied.
		ncycles := (uint16(dc.bitsLeft) + (dc.remaining-1)*8) * dc.timer.Period()
		if cyclesToRun >= uint32(ncycles) {
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
			if (dc.CPU.CurrentCycle() & 0x01) == 0 {
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
		if (dc.CPU.CurrentCycle() & 0x01) == 0 {
			dc.startDelay = 2
		} else {
			dc.startDelay = 3
		}
		dc.needToRun = true
	}
}

func (dc *DMC) ProcessClock() {
	if dc.disableDelay != 0 {
		dc.disableDelay--
		if dc.disableDelay == 0 {
			dc.remaining = 0
			// Abort any on-going transfer that hasn't fully started.
			dc.CPU.StopDMCTransfer()
		}
	}

	if dc.startDelay != 0 {
		dc.startDelay--
		if dc.startDelay == 0 {
			dc.startDMCTransfer()
		}
	}

	dc.needToRun = dc.disableDelay != 0 || dc.startDelay != 0 || dc.remaining != 0
}

func (dc *DMC) NeedToRun() bool {
	if dc.needToRun {
		dc.ProcessClock()
	}
	return dc.needToRun
}

func (dc *DMC) Output() uint8 {
	return uint8(dc.timer.LastOutput())
}
