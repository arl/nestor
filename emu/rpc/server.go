package rpc

import (
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
	Stop()

	SetTempDir(path string)
}

type emuProxy struct {
	emu    Emu
	tmpdir string
}

func (ep *emuProxy) SetTempDir(path string, _ *struct{}) error { ep.emu.SetTempDir(path); return nil }
func (ep *emuProxy) Reset(_, _ *struct{}) error                { ep.emu.Reset(); return nil }
func (ep *emuProxy) Restart(_, _ *struct{}) error              { ep.emu.Restart(); return nil }
func (ep *emuProxy) SetPause(pause bool, _ *struct{}) error    { ep.emu.SetPause(pause); return nil }
func (ep *emuProxy) Stop(_ *struct{}, _ *struct{}) error       { ep.emu.Stop(); return nil }

func (ep *emuProxy) IsReady(_ *struct{}, reply *bool) error {
	*reply = true
	return nil
}

type Server struct {
	io.Closer
}

func NewServer(port int, emu Emu) (*Server, error) {
	proxy := &emuProxy{emu: emu}
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
