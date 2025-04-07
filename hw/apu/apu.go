package apu

import (
	"nestor/emu/log"
	"nestor/hw/hwdefs"
	"nestor/hw/hwio"
	"nestor/hw/snapshot"
)

type APU struct {
	cpu   cpu
	mixer *Mixer

	Square1  squareChannel
	Square2  squareChannel
	Triangle triangleChannel
	Noise    noiseChannel
	DMC      DMC

	frameCounter frameCounter

	prevCycle  uint32
	curCycle   uint32
	needToRun_ bool
	enabled    bool

	STATUS hwio.Reg8 `hwio:"offset=0x15,pcb,rcb,wcb"`
	DAC0   hwio.Reg8 `hwio:"offset=0x18,rcb,readonly"` // current instant DAC value of B=pulse2 and A=pulse1 (either 0 or current volume)
	DAC1   hwio.Reg8 `hwio:"offset=0x19,rcb,readonly"` // current instant DAC value of N=noise (either 0 or current volume) and T=triangle (anywhere from 0 to 15)
	DAC2   hwio.Reg8 `hwio:"offset=0x1A,rcb,readonly"` // current instant DAC value of DPCM channel (same as value written to $4011)
}

func New(cpu cpu, mixer *Mixer) *APU {
	a := &APU{
		enabled: true,
		cpu:     cpu,
		mixer:   mixer,
	}
	a.Noise = newNoiseChannel(a, mixer)
	a.Square1 = newSquareChannel(a, mixer, Square1, true)
	a.Square2 = newSquareChannel(a, mixer, Square2, false)
	a.Triangle = newTriangleChannel(a, mixer)
	a.DMC = newDMC(a, cpu, mixer)

	a.frameCounter.init(a, cpu)

	hwio.MustInitRegs(a)
	hwio.MustInitRegs(&a.Square1)
	hwio.MustInitRegs(&a.Square2)
	hwio.MustInitRegs(&a.Triangle)
	hwio.MustInitRegs(&a.Noise)
	hwio.MustInitRegs(&a.frameCounter)
	hwio.MustInitRegs(&a.DMC)

	return a
}

func (a *APU) WriteFrameCounterReg(old, val uint8) {
	a.frameCounter.WriteFRAMECOUNTER(old, val)
}

func (a *APU) Status() uint8 {
	var status uint8

	if a.Square1.status() {
		status |= 0x01
	}
	if a.Square2.status() {
		status |= 0x02
	}
	if a.Triangle.status() {
		status |= 0x04
	}
	if a.Noise.status() {
		status |= 0x08
	}
	if a.DMC.status() {
		status |= 0x10
	}

	if a.cpu.HasIRQSource(hwdefs.FrameCounter) {
		status |= 0x40
	}
	if a.cpu.HasIRQSource(hwdefs.DMC) {
		status |= 0x80
	}

	return status
}

// STATUS: $4015
func (a *APU) PeekSTATUS(val uint8) uint8 {
	return a.Status()
}

func (a *APU) ReadSTATUS(val uint8) uint8 {
	a.Run()
	status := a.Status()

	// Reading $4015 clears the Frame Counter interrupt flag.
	a.cpu.ClearIRQSource(hwdefs.FrameCounter)

	log.ModSound.InfoZ("read status").Uint8("status", status).End()
	return status
}

func (a *APU) WriteSTATUS(old, val uint8) {
	log.ModSound.InfoZ("write status").Uint8("val", val).End()

	a.Run()

	// Writing to $4015 clears the DMC interrupt flag. This needs to be done
	// before setting the enabled flag for the DMC (because doing so can trigger
	// an IRQ).
	a.cpu.ClearIRQSource(hwdefs.DMC)

	a.Square1.setEnabled((val & 0x01) == 0x01)
	a.Square2.setEnabled((val & 0x02) == 0x02)
	a.Triangle.setEnabled((val & 0x04) == 0x04)
	a.Noise.setEnabled((val & 0x08) == 0x08)
	a.DMC.setEnabled((val & 0x10) == 0x10)
}

func (a *APU) ReadDAC0(val uint8) uint8 {
	a.Run()
	return a.Square1.output() | a.Square2.output()<<4
}

func (a *APU) ReadDAC1(val uint8) uint8 {
	a.Run()
	return a.Triangle.output() | a.Noise.output()<<4
}

func (a *APU) ReadDAC2(val uint8) uint8 {
	a.Run()
	return a.DMC.output()
}

func (a *APU) FrameCounterTick(ftyp FrameType) {
	// Quarter & half frame clock envelope & linear counter
	a.Square1.tickEnvelope()
	a.Square2.tickEnvelope()
	a.Triangle.tickLinearCounter()
	a.Noise.tickEnvelope()

	if ftyp == HalfFrame {
		// Half frames clock length counter & sweep
		a.Square1.tickLengthCounter()
		a.Square2.tickLengthCounter()
		a.Triangle.tickLengthCounter()
		a.Noise.tickLengthCounter()

		a.Square1.tickSweep()
		a.Square2.tickSweep()
	}
}

func (a *APU) Reset(soft bool) {
	a.enabled = true
	a.curCycle = 0
	a.prevCycle = 0

	a.Square1.reset(soft)
	a.Square2.reset(soft)
	a.Triangle.reset(soft)
	a.Noise.reset(soft)
	a.DMC.reset(soft)
	a.frameCounter.reset(soft)
}

func (a *APU) Tick() {
	a.curCycle++
	if a.curCycle == cycleLength-1 {
		panic("frame overflow")
	} else if a.needToRun(a.curCycle) {
		a.Run()
	}
}

func (a *APU) EndFrame(buf *AudioBuffer) {
	a.DMC.processClock()
	a.Run()
	a.Square1.endFrame()
	a.Square2.endFrame()
	a.Triangle.endFrame()
	a.Noise.endFrame()
	a.DMC.endFrame()

	a.mixer.playAudioBuffer(a.curCycle, buf)

	a.curCycle = 0
	a.prevCycle = 0
}

func (a *APU) Run() {
	// Update framecounter and all channels
	// This is called:
	// - At the end of a frame
	// - Before APU registers are read/written to
	// - When a DMC or FrameCounter interrupt needs to be fired
	cyclesToRun := int32(a.curCycle - a.prevCycle)

	for cyclesToRun > 0 {
		a.prevCycle += a.frameCounter.run(&cyclesToRun)

		// Reload counters set by writes to 4003/4008/400B/400F after running
		// the frame counter to allow the length counter to be clocked first.
		// This fixes the test "len_reload_timing" (tests 4 & 5)
		a.Square1.reloadLengthCounter()
		a.Square2.reloadLengthCounter()
		a.Noise.reloadLengthCounter()
		a.Triangle.teloadLengthCounter()

		a.Square1.run(a.prevCycle)
		a.Square2.run(a.prevCycle)
		a.Noise.run(a.prevCycle)
		a.Triangle.run(a.prevCycle)
		a.DMC.run(a.prevCycle)
	}
}

func (a *APU) Enabled() bool           { return a.enabled }
func (a *APU) setEnabled(enabled bool) { a.enabled = enabled }
func (a *APU) SetNeedToRun()           { a.needToRun_ = true }

func (a *APU) needToRun(curCycle uint32) bool {
	if a.DMC.NeedToRun() || a.needToRun_ {
		// Need to run:
		//  - whenever we alter the length counters
		//  - need to run every cycle when DMC is running to get accurate
		//    emulation (CPU stalling, interaction with sprite DMA, etc.)
		a.needToRun_ = false
		return true
	}

	cyclesToRun := curCycle - a.prevCycle
	return a.frameCounter.needToRun(cyclesToRun) || a.DMC.irqPending(cyclesToRun)
}

func (a *APU) State() *snapshot.APU {
	var state snapshot.APU
	a.Square1.saveState(&state.Square1)
	a.Square2.saveState(&state.Square2)
	a.Triangle.saveState(&state.Triangle)
	a.Noise.saveState(&state.Noise)
	a.DMC.saveState(&state.DMC)
	a.frameCounter.saveState(&state.FrameCounter)
	return &state
}

func (a *APU) SetState(state *snapshot.APU) {
	a.Square1.setState(&state.Square1)
	a.Square2.setState(&state.Square2)
	a.Triangle.setState(&state.Triangle)
	a.Noise.setState(&state.Noise)
	a.DMC.setState(&state.DMC)
	a.frameCounter.setState(&state.FrameCounter)
}
