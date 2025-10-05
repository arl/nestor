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

type cpu interface {
	SetIRQSource(src hwdefs.IRQSource)
	HasIRQSource(src hwdefs.IRQSource) bool
	ClearIRQSource(src hwdefs.IRQSource)

	CurrentCycle() int64

	StartDMCTransfer()
	StopDMCTransfer()
}
