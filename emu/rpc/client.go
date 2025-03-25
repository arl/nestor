package rpc

import (
	"fmt"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

type Client struct {
	client *rpc.Client
	tmpdir string
}

func NewClient(port int) (*Client, error) {
	var (
		rpcclient *rpc.Client
		err       error
	)

	const maxretries = 20
	for i := range maxretries {
		if rpcclient, err = rpc.DialHTTP("tcp", ":"+strconv.Itoa(port)); err != nil {
			modRPC.DebugZ("dial tcp failed").Error("err", err).Int("retry", i).End()
		} else {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if rpcclient == nil {
		return nil, fmt.Errorf("dial failed max retries: %v", err)
	}

	tmpdir, err := os.MkdirTemp("", "nestor.out.*")
	if err != nil {
		return nil, fmt.Errorf("failed to create nestor out temp directory: %v", err)
	}

	client := &Client{client: rpcclient, tmpdir: tmpdir}
	client.setTempDir(tmpdir)
	return client, nil
}

func (c *Client) Close() error {
	modRPC.DebugZ("closing rpc client").End()
	return c.client.Close()
}

func (c *Client) TempDir() string { return c.tmpdir }

func (c *Client) setTempDir(path string) { call(c.client, "emu.SetTempDir", path) }
func (c *Client) IsReady() bool          { return request[bool](c.client, "emu.IsReady", nil) }
func (c *Client) Reset()                 { call(c.client, "emu.Reset", nil) }
func (c *Client) Restart()               { call(c.client, "emu.Restart", nil) }
func (c *Client) SetPause(pause bool)    { call(c.client, "emu.SetPause", pause) }
func (c *Client) Stop() {
	call(c.client, "emu.Stop", nil)
}

func request[T any](client *rpc.Client, funcname string, args any) T {
	if args == nil {
		args = &struct{}{}
	}
	var reply T
	if err := client.Call(funcname, args, &reply); err != nil {
		modRPC.FatalZ("RPC call failed").String("func", funcname).Error("err", err).End()
	}
	return reply
}

func call(client *rpc.Client, funcname string, args any) {
	request[struct{}](client, funcname, args)
}
