package debugger

import (
	"fmt"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"

	"nestor/emu/log"
	"nestor/hw"
)

// a debugger holds the state of the CPU debugger. In order to be able to debug
// a program at any moment, the debugger has to keep track of the CPU state,
// even when inactive. The state to keep track of is kept to the minimum, that
// is the current PC and stack frames.
type debugger struct {
	active atomic.Bool
	status atomic.Int32

	cpuBlock  chan struct{}
	blockAcks chan struct{}

	cpu *hw.CPU

	prevPC     uint16
	prevOpcode uint8
	resetPC    uint16

	cstack callStack
}

func NewDebugger(cpu *hw.CPU, addr string) (*debugger, error) {
	dbg := &debugger{
		cpu:       cpu,
		cpuBlock:  make(chan struct{}),
		blockAcks: make(chan struct{}),
	}
	dbg.setStatus(running)
	dbg.active.Store(true)

	if err := runServer(addr, dbg); err != nil {
		return nil, fmt.Errorf("failed to start debugger server: %w", err)
	}

	cpu.SetDebugger(dbg)
	return dbg, nil
}

// computeState returns the current debugger state.
func (dbg *debugger) computeState() stateData {
	var sdata stateData

	switch status := dbg.getStatus(); status {
	case running:
		// If we're running, we have nothing else to pass to the debugger.
		sdata = stateData{
			Status: status.String(),
		}
	default:
		sdata = stateData{
			Status: status.String(),
			PC:     dbg.cpu.PC,
		}
	}

	return sdata
}

// Ws returns the WebSocket handler the debugger will connect to.
func Ws(dbg *debugger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.ModDbg.FatalZ("failed to perform websocket handshake").Error("err", err).End()
			return
		}
		defer ws.Close()

		log.ModDbg.DebugZ("websocket handshake success").End()

		if err := newWsDriver(dbg, ws).drive(); err != nil {
			log.ModDbg.ErrorZ("connection to debugger ended").Error("err", err).End()
		}
	}
}

func runServer(hostport string, dbg *debugger) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", Ws(dbg))

	server := http.Server{
		Addr:    hostport,
		Handler: mux,
	}

	ln, err := net.Listen("tcp", hostport)
	if err != nil {
		return err
	}

	go func() {
		log.ModDbg.InfoZ(fmt.Sprintf("Debugger server listening on %s", hostport)).End()
		server.Serve(ln)
	}()
	return nil
}

func (dbg *debugger) Reset() {
	// Reads PC at reset vector.
	dbg.resetPC = dbg.cpu.PC
}

func (dbg *debugger) getStatus() status {
	return status(dbg.status.Load())
}

func (dbg *debugger) setStatus(s status) {
	dbg.status.Store(int32(s))
}

// Trace must be called before each opcode is executed. This is the main entry
// point for debugging activity, as the debug can stop the CPU execution by
// making this function blocking until user interaction finishes.
func (d *debugger) Trace(pc uint16) {
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
		d.cpuBlock <- struct{}{}
		d.blockAcks <- struct{}{}
	}
}

func (d *debugger) updateStack(dstPc uint16, sff stackFrameFlag) {
	switch d.prevOpcode {
	case 0x20: // JSR
		d.cstack.push(d.prevPC, dstPc, (d.prevPC+3)&0xFFFF, sff)
	case 0x40, 0x60: // RTS RTI
		d.cstack.pop()
	}
}

func (d *debugger) FrameEnd() {
	d.cstack.reset()
}

func (d *debugger) Interrupt(prevpc, curpc uint16, isNMI bool) {
	flag := sffIRQ
	if isNMI {
		flag = sffNMI
	}
	d.updateStack(prevpc, flag)
	d.prevOpcode = 0xFF

	d.cstack.push(d.prevPC, curpc, prevpc, flag)
}

func (d *debugger) WatchRead(addr uint16) {

}

func (d *debugger) WatchWrite(addr uint16, val uint16) {

}

// Break can be called by the CPU core to force breaking into the debugger.
func (d *debugger) Break(msg string) {
	d.setStatus(paused)
}
