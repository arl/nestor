package hwdefs

type IRQSource uint8

const (
	External IRQSource = 1 << iota
	FrameCounter
	DMC

	numSources = 3
)

var irqSrcNames = [numSources]string{
	"external",
	"frameCounter",
	"dmc",
}

func (irq IRQSource) String() string {
	if irq == 0 {
		return ""
	}

	str := ""
	append := func(i int) {
		if str != "" {
			str += "|"
		}
		str += irqSrcNames[i]
	}

	for i := range numSources {
		if irq&(1<<i) != 0 {
			append(i)
		}
	}

	return str
}
