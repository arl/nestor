package debugger

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"

	"nestor/emu/log"
	"nestor/hw"
)

var modDbg = log.NewModule("debugger")

// a debugger holds the state of the CPU debugger. In order to be able to debug
// a program at any moment, the debugger has to keep track of the CPU state,
// even when inactive. The state to keep track of is kept to the minimum, that
// is the current PC and stack frames.
type reactDebugger struct {
	active atomic.Bool
	status atomic.Int32

	cpuBlock  chan debuggerState
	blockAcks chan struct{}

	cpu *hw.CPU

	prevPC     uint16
	prevOpcode uint8
	resetPC    uint16

	cstack callStack
}

type reactDebuggerState struct {
	stat status
	pc   uint16
}

func (s reactDebuggerState) String() string {
	str := ""
	switch s.stat {
	case running:
		str = "running"
	case paused:
		str = "paused"
	case stepping:
		str = "stepping"
	}
	return fmt.Sprintf("pc: $%04X, stat: %s", s.pc, str)
}

// Ws returns the WebSocket handler the debugger will connect to.
func (dbg *reactDebugger) Ws() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			modDbg.FatalZ("failed to perform websocket handshake").Error("err", err).End()
			return
		}
		defer ws.Close()

		modDbg.WarnZ("websocket handshake success").End()

		wc, err := ws.NextWriter(websocket.TextMessage)
		if err != nil {
			modDbg.FatalZ("failed to get writer").Error("err", err).End()
		}

		// Send the initial CPU state to the debugger.
		str := fmt.Sprintf(`{"event": "state", "data": {"cpu": "%s"}}`, dbg.getStatus().String())
		modDbg.WarnZ("Sending").String("msg", str).End()
		fmt.Fprintf(wc, `%s`, str)
		wc.Close()

		for {
			mt, wsr, err := ws.NextReader()
			_ = mt
			if err != nil {
				modDbg.ErrorZ("failed to get writer").Error("err", err).End()
				return
			}
			buf, err := io.ReadAll(wsr)
			if err != nil {
				modDbg.ErrorZ("failed to get writer").Error("err", err).End()
				return
			}
			modDbg.Infof("received %s", buf)
		}
	}
}

func NewReactDebugger(cpu *hw.CPU, addr string) *reactDebugger {
	dbg := &reactDebugger{
		cpu:       cpu,
		cpuBlock:  make(chan debuggerState),
		blockAcks: make(chan struct{}),
	}
	cpu.SetDebugger(dbg)
	dbg.setStatus(running)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", dbg.Ws())

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		// Start HTTP server
		modDbg.Infof("Debugger server listening on %s", addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			modDbg.Fatalf("failed to start server: %s", err)
		}
	}()

	return dbg
}

func (dbg *reactDebugger) Reset() {
	// Reads PC at reset vector.
	dbg.resetPC = dbg.cpu.PC
}

func (dbg *reactDebugger) getStatus() status {
	return status(dbg.status.Load())
}

func (dbg *reactDebugger) setStatus(s status) {
	dbg.status.Store(int32(s))
}

func (dbg *reactDebugger) detach() {
	dbg.active.Store(false)
	if dbg.getStatus() != running {
		dbg.blockAcks <- struct{}{}
		dbg.setStatus(running)
	}
}

// Trace must be called before each opcode is executed. This is the main entry
// point for debugging activity, as the debug can stop the CPU execution by
// making this function blocking until user interaction finishes.
func (d *reactDebugger) Trace(pc uint16) {
	d.updateStack(pc, sffNone)
	if !d.active.Load() {
		return
	}

	// disasm := d.cpu.Disasm(pc)
	// d.prevOpSize = uint8(len(disasm.Bytes))
	opcode := d.cpu.Bus.Read8(pc)
	d.prevPC = pc
	d.prevOpcode = opcode

	switch st := d.getStatus(); st {
	case running:
		break
	case paused, stepping:
		d.cpuBlock <- debuggerState{
			pc:   pc,
			stat: st,
		}
		<-d.blockAcks
	}
}

func (d *reactDebugger) updateStack(dstPc uint16, sff stackFrameFlag) {
	switch d.prevOpcode {
	case 0x20: // JSR
		d.cstack.push(d.prevPC, dstPc, (d.prevPC+3)&0xFFFF, sff)
	case 0x40, 0x60: // RTS RTI
		d.cstack.pop()
	}
}

func (d *reactDebugger) FrameEnd() {
	d.cstack.reset()
}

func (d *reactDebugger) Interrupt(prevpc, curpc uint16, isNMI bool) {
	flag := sffIRQ
	if isNMI {
		flag = sffNMI
	}
	d.updateStack(prevpc, flag)
	d.prevOpcode = 0xFF

	d.cstack.push(d.prevPC, curpc, prevpc, flag)
}

func (d *reactDebugger) WatchRead(addr uint16) {

}

func (d *reactDebugger) WatchWrite(addr uint16, val uint16) {

}

// Break can be called by the CPU core to force breaking into the debugger.
func (d *reactDebugger) Break(msg string) {
	d.setStatus(paused)
}
