package emu

import (
	"bytes"
	"fmt"
	"image/png"
	"sync/atomic"

	"nestor/emu/log"
	"nestor/hw"
	"nestor/ines"
)

type Emulator struct {
	NES *NES
	out *Output
	cfg EmulationConfig

	// Loop can be concurrently controlled by the emulator and ui.
	loopstate atomic.Uint64
	blockch   chan struct{}

	// Use channel to store load/save state result across goroutines boundary.
	savedstatech   chan savestateResult
	loadstateerrch chan error
	// contain the state to load on loopstateLoadSnapshot.
	statetoload []byte
}

type savestateResult struct {
	state ExecState
	err   error
}

const (
	loopstateRunning uint64 = iota
	loopstateQuit
	loopstateBlock
	loopstateReset
	loopstateRestart
	loopstateSaveSnapshot
	loopstateLoadSnapshot
)

// Launch starts the various hardware subsystems, shows the window, setups the
// video and audio streams and plugs controllers. It doesn't start the emulation
// loop, call Run() for that.
func Launch(rom *ines.Rom, cfg Config, out *Output, inp hw.InputStateLoader) (*Emulator, error) {
	nes, err := powerUp(rom)
	if err != nil {
		return nil, fmt.Errorf("power up failed: %s", err)
	}

	nes.CPU.PlugInputDevice(inp)

	// CPU execution trace setup.
	if cfg.TraceOut != nil {
		nes.CPU.SetTraceOutput(cfg.TraceOut)
	}

	return &Emulator{
		NES:            nes,
		out:            out,
		cfg:            cfg.Emulation,
		blockch:        make(chan struct{}, 1),
		savedstatech:   make(chan savestateResult, 1),
		loadstateerrch: make(chan error, 1),
	}, nil
}

func (e *Emulator) RunOneFrame() {
	if e.cfg.RunAheadFrames > 0 {
		e.RunFrameWithRunAhead()
	} else {
		frame := e.out.BeginFrame()
		e.NES.RunOneFrame(&frame)
		e.out.EndFrame(&frame)
	}
}

func (e *Emulator) RunFrameWithRunAhead() {
	frames := e.cfg.RunAheadFrames

	// Run a single frame, make a snapshot, but do not render video nor play
	// audio out of it.
	e.NES.isRunAheadFrame = true
	e.NES.CPU.EnableTrace(true)
	e.NES.CPU.Run(29781)
	e.NES.APU.EndFrame(nil)
	e.NES.CPU.EnableTrace(false)

	buf, err := e.NES.Snapshot()
	if err != nil {
		log.ModEmu.PanicZ("failed run-ahead frame snapshot").Error("err", err).End()
	}

	for frames > 1 {
		e.NES.CPU.Run(29781)
		e.NES.APU.EndFrame(nil)
		frames--
	}
	e.NES.isRunAheadFrame = false

	// Run one frame normally.
	frame := e.out.BeginFrame()
	e.NES.RunOneFrame(&frame)
	e.out.EndFrame(&frame)

	e.NES.isRunAheadFrame = true
	if err := e.NES.LoadSnapshot(buf); err != nil {
		log.ModEmu.PanicZ("failed to load snapshot").Error("err", err).End()
	}
	e.NES.isRunAheadFrame = false
}

type ExecState struct {
	PNGBytes   []byte // last frame in PNG format.
	Snapshot   []byte // emulator state snapshot (i.e savestate).
	BatteryRAM []byte // battery-backed RAM data.
}

func (e *Emulator) Run() (ExecState, error) {
	for !e.NES.CPU.IsHalted() {
		switch e.loopstate.Load() {
		case loopstateRunning:
			e.RunOneFrame()
		case loopstateQuit:
			log.ModEmu.InfoZ("Emulation loop exited").End()
			return e.SavestateUnsafe()
		case loopstateBlock:
			<-e.blockch
		case loopstateSaveSnapshot:
			log.ModEmu.InfoZ("Savestate requested").End()
			e.loopstate.Store(loopstateRunning)
			state, err := e.SavestateUnsafe()
			e.savedstatech <- savestateResult{state: state, err: err}
		case loopstateLoadSnapshot:
			log.ModEmu.InfoZ("Loading savestate").End()
			e.loopstate.Store(loopstateRunning)
			e.loadstateerrch <- e.LoadstateUnsafe(e.statetoload)
		case loopstateReset:
			e.loopstate.Store(loopstateRunning)
			log.ModEmu.InfoZ("Issueing soft reset").End()
			frame := e.out.BeginFrame()
			e.NES.RunResetFrame(&frame, true)
			e.out.EndFrame(&frame)
		case loopstateRestart:
			e.loopstate.Store(loopstateRunning)
			log.ModEmu.InfoZ("Issueing hard reset").End()
			frame := e.out.BeginFrame()
			e.NES.RunResetFrame(&frame, false)
			e.out.EndFrame(&frame)
		}
	}

	return ExecState{}, fmt.Errorf("emulation ended: CPU halted")
}

// Block, Unblock, Reset, Restart, Stop and Savestate are all safe for concurrent use.
func (e *Emulator) Reset()   { e.loopstate.Store(loopstateReset) }
func (e *Emulator) Restart() { e.loopstate.Store(loopstateRestart) }
func (e *Emulator) Stop()    { e.loopstate.Store(loopstateQuit) }
func (e *Emulator) Block()   { e.loopstate.Store(loopstateBlock) }
func (e *Emulator) Unblock() {
	e.loopstate.Store(loopstateRunning)
	select {
	case e.blockch <- struct{}{}:
	default:
		// Avoid deadlock if we were not blocked.
	}
}

func (e *Emulator) Loadstate(snapshot []byte) error {
	e.statetoload = snapshot
	e.loopstate.Store(loopstateLoadSnapshot)
	return <-e.loadstateerrch
}

func (e *Emulator) LoadstateUnsafe(snapshot []byte) error {
	return e.NES.LoadSnapshot(snapshot)
}

// Savestate serializes SavestateUnsafe call with the emulator loop to avoid
// race conditions.
func (e *Emulator) Savestate() (ExecState, error) {
	e.loopstate.Store(loopstateSaveSnapshot)

	res := <-e.savedstatech
	return res.state, res.err
}

// Savestate is not safe for concurrent use, hence it must be called when the
// emulator loop is already blocked.
func (e *Emulator) SavestateUnsafe() (ExecState, error) {
	// Get a state snapshot.
	savestate, err := e.NES.Snapshot()
	if err != nil {
		return ExecState{}, fmt.Errorf("failed to save state: %w", err)
	}

	// Make screenshot from the last frame.
	var screenshot bytes.Buffer
	if err := png.Encode(&screenshot, e.out.Screenshot()); err != nil {
		return ExecState{}, fmt.Errorf("failed to encode screenshot to png: %w", err)
	}

	return ExecState{
		PNGBytes:   screenshot.Bytes(),
		Snapshot:   savestate,
		BatteryRAM: e.NES.Mapper.BatteryPackedRAM(),
	}, nil
}
