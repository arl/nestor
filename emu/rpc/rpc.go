package rpc

import (
	"net"

	"nestor/emu/log"
)

var modRPC = log.NewModule("rpc")

// PickUnusedPort finds an unused port by trying to listen on port 0 and letting
// the OS pick a port, then closing that connection and returning that port
// number. This is inherently racy.
func PickUnusedPort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		return 0, err
	}
	return port, nil
}
