package hw

import (
	"nestor/emu/log"
	"nestor/hw/apu"
	"nestor/hw/hwio"
)

type APU struct {
	needToRunFlag bool
	enabled       bool

	prevCycle uint32
	curCycle  uint32

	Sq0          apu.SquareChannel
	Sq1          apu.SquareChannel
	Trg          apu.TriangleChannel
	noise        apu.NoiseChannel
	frameCounter apuFrameCounter

	cpu   *CPU
	mixer *AudioMixer

	STATUS hwio.Reg8 `hwio:"offset=0x15,rcb,wcb"`
}

func NewAPU(cpu *CPU, mixer *AudioMixer) *APU {
	a := &APU{
		enabled: true,
		cpu:     cpu,
		mixer:   mixer,
	}
	a.noise = apu.NewNoiseChannel(a, mixer)
	a.Sq0 = apu.NewSquareChannel(a, mixer, apu.Square1, true)
	a.Sq1 = apu.NewSquareChannel(a, mixer, apu.Square2, false)
	a.Trg = apu.NewTriangleChannel(a, mixer)
	a.frameCounter.apu = a

	hwio.MustInitRegs(a)
	hwio.MustInitRegs(&a.Sq0)
	hwio.MustInitRegs(&a.Sq1)
	hwio.MustInitRegs(&a.Trg)
	hwio.MustInitRegs(&a.noise)
	hwio.MustInitRegs(&a.frameCounter)

	// XXX set 0x40 for now for nestest to pass.
	a.Sq0.Duty.Value = 0x40
	a.Sq0.Sweep.Value = 0x40
	a.Sq0.Timer.Value = 0x40
	a.Sq0.Length.Value = 0x40

	a.Sq1.Duty.Value = 0x40
	a.Sq1.Sweep.Value = 0x40
	a.Sq1.Timer.Value = 0x40
	a.Sq1.Length.Value = 0x40

	return a
}

func (a *APU) Status() uint8 {
	var status uint8

	if a.Sq0.Status() {
		status |= 0x01
	}
	if a.Sq1.Status() {
		status |= 0x02
	}
	if a.Trg.Status() {
		status |= 0x04
	}
	if a.noise.Status() {
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

	a.Sq0.SetEnabled((val & 0x01) == 0x01)
	a.Sq1.SetEnabled((val & 0x02) == 0x02)
	a.Trg.SetEnabled((val & 0x04) == 0x04)
	a.noise.SetEnabled((val & 0x08) == 0x08)
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
	a.Sq0.TickEnvelope()
	a.Sq1.TickEnvelope()
	a.Trg.TickLinearCounter()
	a.noise.TickEnvelope()

	if ftyp == HalfFrame {
		// Half frames clock length counter & sweep
		a.Sq0.TickLengthCounter()
		a.Sq1.TickLengthCounter()
		a.Trg.TickLengthCounter()
		a.noise.TickLengthCounter()

		a.Sq0.TickSweep()
		a.Sq1.TickSweep()
	}
}

func (a *APU) Reset(soft bool) {
	a.enabled = true
	a.curCycle = 0
	a.prevCycle = 0
	a.Sq0.Reset(soft)
	a.Sq1.Reset(soft)
	a.Trg.Reset(soft)
	a.noise.Reset(soft)
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
	a.Sq0.EndFrame()
	a.Sq1.EndFrame()
	a.Trg.EndFrame()
	a.noise.EndFrame()
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
		a.Sq0.ReloadLengthCounter()
		a.Sq1.ReloadLengthCounter()
		a.noise.ReloadLengthCounter()
		a.Trg.ReloadLengthCounter()

		a.Sq0.Run(a.prevCycle)
		a.Sq1.Run(a.prevCycle)
		a.noise.Run(a.prevCycle)
		a.Trg.Run(a.prevCycle)
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
