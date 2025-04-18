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

type FrameType uint8

const (
	NoFrame FrameType = iota
	QuarterFrame
	HalfFrame
)

type apu interface {
	SetNeedToRun()
	Run()
	FrameCounterTick(FrameType)
}

type cpu interface {
	SetIRQSource(src hwdefs.IRQSource)
	HasIRQSource(src hwdefs.IRQSource) bool
	ClearIRQSource(src hwdefs.IRQSource)

	CurrentCycle() int64

	StartDMCTransfer()
	StopDMCTransfer()
}
