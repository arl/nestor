package debugger

import (
	"fmt"

	"encoding/json"
	"nestor/emu/log"

	"github.com/gorilla/websocket"
)

// Emulator and debugger communicates via a websocket connection, following this
// simple protocol. The first ever exchanged message is sent by the emulator
// sending its current state. After which, the emulator just waits for the
// debugger requests (WSRequest) to which it always responds with a WSResponse..

/* Debugger -> Emulator requests */

// WSRequest is a debugger->emulator request.
type WSRequest struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// data for the 'set-cpu-state' request.
type setCPUStateData string

/* Emulator -> Debugger responses */

// WSResponse is an emulator->debugger response.
type WSResponse struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// data for the 'state' event.
type stateData struct {
	PC     uint16 `json:"pc"`
	Status string `json:"status"`
}

type wsdriver struct {
	dbg *reactDebugger
	ws  *websocket.Conn

	handlers map[string]wsHandlerFunc
}

type wsHandlerFunc func(data []byte) (*WSResponse, error)

func newWsDriver(dbg *reactDebugger, ws *websocket.Conn) *wsdriver {
	drv := &wsdriver{
		dbg:      dbg,
		ws:       ws,
		handlers: make(map[string]wsHandlerFunc),
	}
	drv.handlers["set-cpu-state"] = drv.handleSetCPUState
	return drv
}

func (d *wsdriver) drive() error {
	log.ModDbg.DebugZ("debugger connection initated").End()

	d.initMsg()

	for {
		// Wait for next request from the debugger.
		var req WSRequest
		if err := d.ws.ReadJSON(&req); err != nil {
			return err
		}

		log.ModDbg.DebugZ("received message from debugger").
			String("event", req.Event).
			String("data", string(req.Data)).
			End()

		handler, ok := d.handlers[req.Event]
		if !ok {
			log.ModDbg.WarnZ("received unknown debugger event").
				String("event", req.Event).
				String("data", string(req.Data)).
				End()

			// TODO: is it necessary to inform the debugger?
			continue
		}

		resp, err := handler(req.Data)
		if err != nil {
			log.ModDbg.ErrorZ("error handling debugger event").
				String("event", req.Event).
				String("data", string(req.Data)).
				Error("err", err).
				End()
		}

		if err := d.ws.WriteJSON(resp); err != nil {
			log.ModDbg.ErrorZ("error responding to debugger event").
				String("event", resp.Event).
				Error("err", err).
				End()
		}
	}
}

func (d *wsdriver) initMsg() {
	initmsg := WSResponse{
		Event: "state",
		Data: stateData{
			Status: running.String(),
			PC:     0,
		},
	}

	if err := d.ws.WriteJSON(initmsg); err != nil {
		log.ModDbg.FatalZ("failed to send initial state to debugger").Error("err", err).End()
	}

	log.ModDbg.DebugZ("initmsg sent").End()
}

// TODO(arl): consider converting this as a method of the debugger.
func (d *wsdriver) handleSetCPUState(data []byte) (*WSResponse, error) {
	var state setCPUStateData
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	var stateResp stateData

	switch state {
	case "run":
		// Changes the CPU state and unblock it.
		stateResp.Status = running.String()
		d.dbg.setStatus(running)
		<-d.dbg.blockAcks
	case "pause":
		// Changes CPU state then wait for it be effectively blocked.
		d.dbg.setStatus(paused)

		// TODO: this channel should probably directly return the emulator state.
		// Wait
		rds := <-d.dbg.cpuBlock
		stateResp.Status = rds.stat.String()
		stateResp.PC = rds.pc
	case "step":
		// Changes the CPU state and unblock it.
		d.dbg.setStatus(stepping)
		<-d.dbg.blockAcks
		rds := <-d.dbg.cpuBlock
		stateResp.Status = rds.stat.String()
		stateResp.PC = rds.pc

	default:
		return nil, fmt.Errorf("unexpected cpu state: %s", state)
	}

	return &WSResponse{Event: "state", Data: stateResp}, nil
}
