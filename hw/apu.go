package hw

import (
	"nestor/emu/log"
	"nestor/hw/apu"
	"nestor/hw/hwdefs"
	"nestor/hw/hwio"
	"nestor/hw/snapshot"
)

type APU struct {
	cpu   *CPU
	mixer *AudioMixer

	Square1  apu.SquareChannel
	Square2  apu.SquareChannel
	Triangle apu.TriangleChannel
	Noise    apu.NoiseChannel
	DMC      apu.DMC

	frameCounter apu.FrameCounter

	prevCycle  uint32
	curCycle   uint32
	needToRun_ bool
	enabled    bool

	STATUS hwio.Reg8 `hwio:"offset=0x15,pcb,rcb,wcb"`
	DAC0   hwio.Reg8 `hwio:"offset=0x18,rcb,readonly"` // current instant DAC value of B=pulse2 and A=pulse1 (either 0 or current volume)
	DAC1   hwio.Reg8 `hwio:"offset=0x19,rcb,readonly"` // current instant DAC value of N=noise (either 0 or current volume) and T=triangle (anywhere from 0 to 15)
	DAC2   hwio.Reg8 `hwio:"offset=0x1A,rcb,readonly"` // current instant DAC value of DPCM channel (same as value written to $4011)
}

func NewAPU(cpu *CPU, mixer *AudioMixer) *APU {
	a := &APU{
		enabled: true,
		cpu:     cpu,
		mixer:   mixer,
	}
	a.Noise = apu.NewNoiseChannel(a, mixer)
	a.Square1 = apu.NewSquareChannel(a, mixer, apu.Square1, true)
	a.Square2 = apu.NewSquareChannel(a, mixer, apu.Square2, false)
	a.Triangle = apu.NewTriangleChannel(a, mixer)
	a.DMC = apu.NewDMC(a, cpu, mixer)

	a.frameCounter.Init(a, cpu)

	hwio.MustInitRegs(a)
	hwio.MustInitRegs(&a.Square1)
	hwio.MustInitRegs(&a.Square2)
	hwio.MustInitRegs(&a.Triangle)
	hwio.MustInitRegs(&a.Noise)
	hwio.MustInitRegs(&a.frameCounter)
	hwio.MustInitRegs(&a.DMC)

	return a
}

func (a *APU) Status() uint8 {
	var status uint8

	if a.Square1.Status() {
		status |= 0x01
	}
	if a.Square2.Status() {
		status |= 0x02
	}
	if a.Triangle.Status() {
		status |= 0x04
	}
	if a.Noise.Status() {
		status |= 0x08
	}
	if a.DMC.Status() {
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

	a.Square1.SetEnabled((val & 0x01) == 0x01)
	a.Square2.SetEnabled((val & 0x02) == 0x02)
	a.Triangle.SetEnabled((val & 0x04) == 0x04)
	a.Noise.SetEnabled((val & 0x08) == 0x08)
	a.DMC.SetEnabled((val & 0x10) == 0x10)
}

func (a *APU) ReadDAC0(val uint8) uint8 {
	a.Run()
	return a.Square1.Output() | a.Square2.Output()<<4
}

func (a *APU) ReadDAC1(val uint8) uint8 {
	a.Run()
	return a.Triangle.Output() | a.Noise.Output()<<4
}

func (a *APU) ReadDAC2(val uint8) uint8 {
	a.Run()
	return a.DMC.Output()
}

func (a *APU) FrameCounterTick(ftyp apu.FrameType) {
	// Quarter & half frame clock envelope & linear counter
	a.Square1.TickEnvelope()
	a.Square2.TickEnvelope()
	a.Triangle.TickLinearCounter()
	a.Noise.TickEnvelope()

	if ftyp == apu.HalfFrame {
		// Half frames clock length counter & sweep
		a.Square1.TickLengthCounter()
		a.Square2.TickLengthCounter()
		a.Triangle.TickLengthCounter()
		a.Noise.TickLengthCounter()

		a.Square1.TickSweep()
		a.Square2.TickSweep()
	}
}

func (a *APU) Reset(soft bool) {
	a.enabled = true
	a.curCycle = 0
	a.prevCycle = 0

	a.Square1.Reset(soft)
	a.Square2.Reset(soft)
	a.Triangle.Reset(soft)
	a.Noise.Reset(soft)
	a.DMC.Reset(soft)
	a.frameCounter.Reset(soft)
}

func (a *APU) Tick() {
	a.curCycle++
	if a.curCycle == CycleLength-1 {
		a.EndFrame()
	} else if a.needToRun(a.curCycle) {
		a.Run()
	}
}

func (a *APU) EndFrame() {
	a.DMC.ProcessClock()
	a.Run()
	a.Square1.EndFrame()
	a.Square2.EndFrame()
	a.Triangle.EndFrame()
	a.Noise.EndFrame()
	a.DMC.EndFrame()

	a.mixer.PlayAudioBuffer(a.curCycle)

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
		a.prevCycle += a.frameCounter.Run(&cyclesToRun)

		// Reload counters set by writes to 4003/4008/400B/400F after running
		// the frame counter to allow the length counter to be clocked first.
		// This fixes the test "len_reload_timing" (tests 4 & 5)
		a.Square1.ReloadLengthCounter()
		a.Square2.ReloadLengthCounter()
		a.Noise.ReloadLengthCounter()
		a.Triangle.ReloadLengthCounter()

		a.Square1.Run(a.prevCycle)
		a.Square2.Run(a.prevCycle)
		a.Noise.Run(a.prevCycle)
		a.Triangle.Run(a.prevCycle)
		a.DMC.Run(a.prevCycle)
	}
}

func (a *APU) SetNeedToRun() {
	a.needToRun_ = true
}

func (a *APU) setEnabled(enabled bool) {
	a.enabled = enabled
}

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
	return a.frameCounter.NeedToRun(cyclesToRun) || a.DMC.IRQPending(cyclesToRun)
}

func (a *APU) State() *snapshot.APU {
	var state snapshot.APU
	a.Square1.SaveState(&state.Square1)
	a.Square2.SaveState(&state.Square2)
	a.Triangle.SaveState(&state.Triangle)
	a.Noise.SaveState(&state.Noise)
	a.DMC.SaveState(&state.DMC)
	a.frameCounter.SaveState(&state.FrameCounter)
	return &state
}

func (a *APU) SetState(state *snapshot.APU) {
	a.Square1.SetState(&state.Square1)
	a.Square2.SetState(&state.Square2)
	a.Triangle.SetState(&state.Triangle)
	a.Noise.SetState(&state.Noise)
	a.DMC.SetState(&state.DMC)
	a.frameCounter.SetState(&state.FrameCounter)

	// Reset the cycle counters to ensure that the APU is in sync with the CPU.
	// This is important for accurate emulation of the DMC channel.
	a.prevCycle = 0
	a.curCycle = 0
}
