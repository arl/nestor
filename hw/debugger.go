package hw

// A Debugger controls and monitors a CPU.
type Debugger interface {
	// Reset informs the debugger the CPU has been reset.
	Reset()

	// Trace must be called before each opcode is executed. This is the main
	// entry point for debugging activity, as the debug can stop the CPU
	// execution by making this function blocking until user interaction
	// finishes.
	Trace(pc uint16)

	// Interrupt is called when an interrupt is about to be executed. prevpc is
	// the address of the instruction that was about to be executed, curpc is
	// the address of the interrupt handler, and isNMI is true if the interrupt
	// is a non-maskable interrupt.
	Interrupt(prevpc, curpc uint16, isNMI bool)

	// WatchRead/WatchWrite must be called before each memory access. They can
	// be used by the debugger to implement watchpoints and thus intercept
	// memory accesses
	WatchRead(addr uint16)
	WatchWrite(addr uint16, val uint16)

	// Break can be called by the CPU core to force breaking into the debugger.
	Break(msg string)

	// FrameEnd signals the debugger the end of the current frame.
	FrameEnd()
}

// NopDebugger is a no-op Debugger implementation, the default when no debugger
// has been configured.
type NopDebugger struct{}

func (NopDebugger) Reset()                                     {}
func (NopDebugger) Trace(pc uint16)                            {}
func (NopDebugger) Interrupt(prevpc, curpc uint16, isNMI bool) {}
func (NopDebugger) WatchRead(addr uint16)                      {}
func (NopDebugger) WatchWrite(addr uint16, val uint16)         {}
func (NopDebugger) Break(msg string)                           {}
func (NopDebugger) FrameEnd()                                  {}
