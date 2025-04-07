package hwdefs

import "strings"

type IRQSource uint8

const (
	External IRQSource = 1 << iota
	FrameCounter
	DMC

	numSources = 3
)

var irqSrcNames = [numSources]string{
	"ext",
	"fcnt",
	"dmc",
}

func (irq IRQSource) String() string {
	var names []string
	for i := range numSources {
		if irq&(1<<i) != 0 {
			names = append(names, irqSrcNames[i])
		}
	}
	return strings.Join(names, "|")
}

const (
	SoftReset = true
	HardReset = false
)

const NumAudioChannels = 5 // Square1, Square2, Triangle, Noise, DMC
