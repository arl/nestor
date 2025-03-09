package rpc

import (
	"image"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

type Emu interface {
	Reset()
	Restart()
	SetPause(pause bool)
	Stop() *image.RGBA
}

type emuProxy struct{ Emu }

func (ep *emuProxy) Reset(_, _ *struct{}) error             { ep.Emu.Reset(); return nil }
func (ep *emuProxy) Restart(_, _ *struct{}) error           { ep.Emu.Restart(); return nil }
func (ep *emuProxy) SetPause(pause bool, _ *struct{}) error { ep.Emu.SetPause(pause); return nil }
func (ep *emuProxy) Stop(_ *struct{}, reply *image.RGBA) error {
	rep := ep.Emu.Stop()
	*reply = *rep
	return nil
}

type Server struct {
	io.Closer
}

func NewServer(port int, emu Emu) (*Server, error) {
	proxy := &emuProxy{Emu: emu}
	if err := rpc.RegisterName("emu", proxy); err != nil {
		panic("failed to register RPC server: " + err.Error())
	}
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	modRPC.InfoZ("rpc server listening").Int("port", port).End()
	go http.Serve(l, nil)
	return &Server{Closer: l}, nil
}
