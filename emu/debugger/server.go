package debugger

import (
	"fmt"
	"io/fs"
	"net"
	"net/http"

	"github.com/gorilla/websocket"

	"nestor/emu/log"
	nestor_dbg "nestor/nestor-dbg"
)

func runServer(hostport string, dbg *debugger) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleWebsocket(dbg))
	mux.HandleFunc("/", handleIndex())

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

// handleWebsocket returns the WebSocket handler for the debugger to connect.
func handleWebsocket(dbg *debugger) http.HandlerFunc {
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

// handleIndex returns the handler serving the debugger embedded assets (the
// content of the frontend /build directory).
func handleIndex() http.HandlerFunc {
	build, err := fs.Sub(nestor_dbg.Assets, "build")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(build)).ServeHTTP
}
