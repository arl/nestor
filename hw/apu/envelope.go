package apu

type envelope struct {
	constVolume bool
	vol         uint8

	start   bool
	divider int8
	counter uint8

	lenCounter lengthCounter
}

func (env *envelope) init(regValue uint8) {
	env.lenCounter.init((regValue & 0x20) == 0x20)
	env.constVolume = (regValue & 0x10) == 0x10
	env.vol = regValue & 0x0F
}

func (env *envelope) resetEnvelope() {
	env.start = true
}

func (env *envelope) volume() uint32 {
	if env.lenCounter.status() {
		if env.constVolume {
			return uint32(env.vol)
		}
		return uint32(env.counter)
	}
	return 0
}

func (env *envelope) reset(soft bool) {
	env.lenCounter.reset(soft)
	env.constVolume = false
	env.vol = 0
	env.start = false
	env.divider = 0
	env.counter = 0
}

func (env *envelope) tick() {
	if env.start {
		env.start = false
		env.counter = 15
		env.divider = int8(env.vol)
	} else {
		env.divider--
		if env.divider < 0 {
			env.divider = int8(env.vol)
			if env.counter > 0 {
				env.counter--
			} else if env.lenCounter.isHalted() {
				env.counter = 15
			}
		}
	}
}
