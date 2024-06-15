package emu

import (
	"sync/atomic"

	"nestor/emu/log"
)

type UserInputReader interface {
	ReadUserInput(<-chan uint16)
}

type StdPadButton byte

const (
	PadA StdPadButton = iota
	PadB
	PadSelect
	PadStart
	PadUp
	PadDown
	PadLeft
	PadRight
)

type StdControllerPair struct {
	Pad1Connected bool
	Pad2Connected bool

	state atomic.Uint32
}

func (c *StdControllerPair) ReadUserInput(ch <-chan uint16) {
	go func() {
		for v := range ch {
			c.state.Store(uint32(v))
			log.ModInput.DebugZ("input state update").
				Hex8("ctrl-1", uint8(v&0xff)).
				Hex8("ctrl-2", uint8(v>>8)).
				End()
		}
	}()
}

func (c *StdControllerPair) LoadState() (uint8, uint8) {
	var s1, s2 uint8

	cur := c.state.Load()
	if c.Pad1Connected {
		s1 = uint8(cur & 0xff)
	}
	if c.Pad2Connected {
		s2 = uint8(cur >> 8)
	}
	return s1, s2
}
