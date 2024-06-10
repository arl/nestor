package debugger

import (
	"fmt"

	"encoding/json"
	"nestor/emu/log"

	"github.com/gorilla/websocket"
)

// Emulator and debugger communicates via a websocket connection, following this
// simple protocol.
//
// The first ever exchanged message is sent by the emulator sending its current
// state. After which, the emulator just waits for the debugger requests (i.e
// WSRequest) to which it always has to respond (with a WSResponse).

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
		Data:  d.dbg.computeState(),
	}

	if err := d.ws.WriteJSON(initmsg); err != nil {
		log.ModDbg.FatalZ("failed to send initial state to debugger").Error("err", err).End()
	}

	log.ModDbg.DebugZ("initmsg sent").End()
}

func (d *wsdriver) handleSetCPUState(data []byte) (*WSResponse, error) {
	var cpuState setCPUStateData
	if err := json.Unmarshal(data, &cpuState); err != nil {
		return nil, err
	}

	switch cpuState {
	case "run":
		d.dbg.setStatus(running)
		<-d.dbg.blockAcks

	case "pause":
		d.dbg.setStatus(paused)
		<-d.dbg.cpuBlock

	case "step":
		d.dbg.setStatus(stepping)
		<-d.dbg.blockAcks
		<-d.dbg.cpuBlock

	default:
		return nil, fmt.Errorf("unexpected cpu state: %s", cpuState)
	}

	return &WSResponse{
		Event: "state",
		Data:  d.dbg.computeState(),
	}, nil
}
