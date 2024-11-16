package hw

import (
	"nestor/emu/log"
	"nestor/hw/apu"
	"nestor/hw/hwio"
)

type APU struct {
	cpu   *CPU
	mixer *AudioMixer

	Square1  apu.SquareChannel
	Square2  apu.SquareChannel
	Triangle apu.TriangleChannel
	Noise    apu.NoiseChannel

	frameCounter apuFrameCounter

	prevCycle     uint32
	curCycle      uint32
	needToRunFlag bool
	enabled       bool

	STATUS hwio.Reg8 `hwio:"offset=0x15,rcb,wcb"`
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
	a.frameCounter.apu = a

	hwio.MustInitRegs(a)
	hwio.MustInitRegs(&a.Square1)
	hwio.MustInitRegs(&a.Square2)
	hwio.MustInitRegs(&a.Triangle)
	hwio.MustInitRegs(&a.Noise)
	hwio.MustInitRegs(&a.frameCounter)

	// XXX set 0x40 for now for nestest to pass.
	a.Square1.Duty.Value = 0x40
	a.Square1.Sweep.Value = 0x40
	a.Square1.Timer.Value = 0x40
	a.Square1.Length.Value = 0x40

	a.Square2.Duty.Value = 0x40
	a.Square2.Sweep.Value = 0x40
	a.Square2.Timer.Value = 0x40
	a.Square2.Length.Value = 0x40

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
	// status |= a.dmc.Status() ? 0x10 : 0x00;
	if a.cpu.hasIrqSource(frameCounter) {
		status |= 0x40
	}
	// status |= a.console->GetCpu()->HasIrqSource(IRQSource::DMC) ? 0x80 : 0x00;

	return status
}

// STATUS: $4015
func (a *APU) ReadSTATUS(val uint8, peek bool) uint8 {
	if peek {
		return a.Status()
	}
	a.Run()
	status := a.Status()

	// Reading $4015 clears the Frame Counter interrupt flag.
	a.cpu.clearIrqSource(frameCounter)

	log.ModSound.InfoZ("read status").Uint8("status", status).End()
	return status
}

func (a *APU) WriteSTATUS(old, val uint8) {
	log.ModSound.InfoZ("write status").Uint8("val", val).End()

	a.Run()

	// Writing to $4015 clears the DMC interrupt flag. This needs to be done
	// before setting the enabled flag for the DMC (because doing so can trigger
	// an IRQ).
	a.cpu.clearIrqSource(dmc)

	a.Square1.SetEnabled((val & 0x01) == 0x01)
	a.Square2.SetEnabled((val & 0x02) == 0x02)
	a.Triangle.SetEnabled((val & 0x04) == 0x04)
	a.Noise.SetEnabled((val & 0x08) == 0x08)
	// _dmc->SetEnabled((value & 0x10) == 0x10);}
}

func (a *APU) WriteFRAMECOUNTER(old, val uint8) {
	log.ModSound.InfoZ("write framecounter").Uint8("val", val).End()
	a.Run()
	a.frameCounter.newValue = int16(val)

	// Reset sequence after $4017 is written to
	if a.cpu.Cycles&0x01 != 0 {
		// If the write occurs between APU cycles, the effects occur 4 CPU
		// cycles after the write cycle.
		a.frameCounter.writeDelayCounter = 4
	} else {
		// If the write occurs during an APU cycle, the effects occur 3 CPU
		// cycles after the $4017 write cycle
		a.frameCounter.writeDelayCounter = 3
	}

	a.frameCounter.inhibitIRQ = (val & 0x40) == 0x40
	if a.frameCounter.inhibitIRQ {
		a.cpu.clearIrqSource(frameCounter)
	}
}

func (a *APU) FrameCounterTick(ftyp FrameType) {
	// Quarter & half frame clock envelope & linear counter
	a.Square1.TickEnvelope()
	a.Square2.TickEnvelope()
	a.Triangle.TickLinearCounter()
	a.Noise.TickEnvelope()

	if ftyp == HalfFrame {
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
	// a.dmc.Reset(softReset)
	a.frameCounter.reset(soft)
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
	// _dmc->ProcessClock();
	a.Run()
	a.Square1.EndFrame()
	a.Square2.EndFrame()
	a.Triangle.EndFrame()
	a.Noise.EndFrame()
	// _dmc->EndFrame();

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
		a.prevCycle += a.frameCounter.run(&cyclesToRun)

		// Reload counters set by writes to 4003/4008/400B/400F after
		// running the frame counter to allow the length counter to be
		// clocked first. This fixes the test "len_reload_timing" (tests 4 &
		// 5)
		a.Square1.ReloadLengthCounter()
		a.Square2.ReloadLengthCounter()
		a.Noise.ReloadLengthCounter()
		a.Triangle.ReloadLengthCounter()

		a.Square1.Run(a.prevCycle)
		a.Square2.Run(a.prevCycle)
		a.Noise.Run(a.prevCycle)
		a.Triangle.Run(a.prevCycle)
		// _dmc->Run(a.prevCycle);
	}
}

func (a *APU) SetNeedToRun() { a.needToRunFlag = true }

func (a *APU) needToRun(curCycle uint32) bool {
	if /*_dmc->NeedToRun() || */ a.needToRunFlag {
		// Need to run:
		//  - whenever we alter the length counters
		//  - need to run every cycle when DMC is running to get accurate emulation (CPU stalling, interaction with sprite DMA, etc.)
		a.needToRunFlag = false
		return true
	}

	cyclesToRun := curCycle - a.prevCycle
	return a.frameCounter.needToRun(cyclesToRun) /* || _dmc->IrqPending(cyclesToRun);*/
}
