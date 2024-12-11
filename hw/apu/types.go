package apu

import "nestor/hw/hwdefs"

type Channel uint8

const (
	Square1 Channel = iota
	Square2
	Triangle
	Noise
	DPCM
)

type mixer interface {
	AddDelta(ch Channel, time uint32, delta int16)
}

type apu interface {
	SetNeedToRun()
	Run()
}

type cpu interface {
	SetIrqSource(src hwdefs.IRQSource)
	HasIrqSource(src hwdefs.IRQSource) bool
	ClearIrqSource(src hwdefs.IRQSource)

	CurrentCycle() int64

	StartDmcTransfer()
	StopDmcTransfer()
}
