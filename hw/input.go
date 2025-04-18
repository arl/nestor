package hw

import (
	"nestor/hw/hwio"
)

type inputStateLoader interface {
	LoadState() (uint8, uint8)
}

// InputPorts handles I/O with an InputDevice (such as standard NES controller
// for example).
type InputPorts struct {
	In hwio.Reg8 `hwio:"offset=0x16,pcb,rcb,wcb"`

	provider           inputStateLoader
	prevStrobe, strobe bool     // to observe strobe falling edge.
	state              [2]uint8 // state shift registers.
}

func (ip *InputPorts) initBus() {
	hwio.MustInitRegs(ip)
}

func (ip *InputPorts) regval(port uint8) uint8 {
	ret := ip.state[port] & 1
	ip.state[port] >>= 1

	// After 8 bits are read, all subsequent bits will report 1 on a standard
	// NES controller, but third party and other controllers may report other
	// values here.
	ip.state[port] |= 0x80

	// Emulate open bus behavior.
	return 0x40 | ret
}

// like regval but without side effects.
func (ip *InputPorts) regvalPeek(port uint8) uint8 {
	ret := ip.state[port] & 1

	// Emulate open bus behavior.
	return 0x40 | ret
}

// capture state of all connected input devices.
func (ip *InputPorts) loadstate() {
	if ip.provider == nil {
		// No controller is connected.
		ip.state[0] = 0x00
		ip.state[1] = 0x00
		return
	}

	ip.state[0], ip.state[1] = ip.provider.LoadState()
}

// In: $4016
func (ip *InputPorts) WriteIN(old, val uint8) {
	ip.prevStrobe = ip.strobe
	ip.strobe = val&1 == 1
	if ip.prevStrobe && !ip.strobe {
		ip.loadstate()
	}
}

func (ip *InputPorts) PeekIN(_ uint8) uint8 {
	return ip.regvalPeek(0)
}

func (ip *InputPorts) ReadIN(_ uint8) uint8 {
	if ip.strobe {
		ip.loadstate()
	}
	return ip.regval(0)
}

// Out: $4017
func (ip *InputPorts) PeekOUT(_ uint8) uint8 {
	return ip.regvalPeek(1)
}

func (ip *InputPorts) ReadOUT(_ uint8) uint8 {
	if ip.strobe {
		ip.loadstate()
	}

	return ip.regval(1)
}
