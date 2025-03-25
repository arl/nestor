package rpc

import (
	"net"

	"nestor/emu/log"
)

var modRPC = log.NewModule("rpc")

func UnusedPort() int {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic("pickUnusedPort failed: " + err.Error())
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		panic("pickUnusedPort failed: " + err.Error())
	}
	return port
}
