package rpc

import (
	"fmt"
	"image"
	"net/rpc"
	"strconv"
	"time"
)

type Client struct {
	client *rpc.Client
}

func NewClient(port int) (*Client, error) {
	var (
		client *rpc.Client
		err    error
	)
	const maxretries = 5
	for i := range maxretries {
		if client, err = rpc.DialHTTP("tcp", ":"+strconv.Itoa(port)); err != nil {
			client = nil
			modRPC.WarnZ("dial tcp failed").Error("err", err).Int("retry", i).End()
		}
		time.Sleep(250 * time.Millisecond)
	}

	if client == nil {
		return nil, fmt.Errorf("dial failed max retries: %v", err)
	}

	return &Client{client: client}, nil
}

func (c *Client) Close() error {
	modRPC.DebugZ("closing rpc client").End()
	return c.client.Close()
}

func (c *Client) Reset()              { call(c.client, "emu.Reset", nil) }
func (c *Client) Restart()            { call(c.client, "emu.Restart", nil) }
func (c *Client) SetPause(pause bool) { call(c.client, "emu.SetPause", pause) }
func (c *Client) Stop() *image.RGBA   { return request[*image.RGBA](c.client, "emu.Stop", nil) }

func call(client *rpc.Client, funcname string, args any) {
	request[struct{}](client, funcname, args)
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
