package apu

type Channel uint8

const (
	Square1 Channel = iota
	Square2
	Triangle
	Noise
	DMC
)

type mixer interface {
	AddDelta(ch Channel, time uint32, delta int16)
}

type apu interface {
	SetNeedToRun()
	Run()
}
