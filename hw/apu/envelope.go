package apu

type Envelope struct {
	constantVol bool
	vol         uint8

	start   bool
	divider int8
	counter uint8

	lenCounter lengthCounter
}

func (env *Envelope) init(regValue uint8) {
	env.lenCounter.init((regValue & 0x20) == 0x20)
	env.constantVol = (regValue & 0x10) == 0x10
	env.vol = regValue & 0x0F
}

func (env *Envelope) resetEnvelope() {
	env.start = true
}

func (env *Envelope) volume() uint32 {
	if env.lenCounter.status() {
		if env.constantVol {
			return uint32(env.vol)
		}
		return uint32(env.counter)
	}
	return 0
}

func (env *Envelope) reset(soft bool) {
	env.lenCounter.reset(soft)
	env.constantVol = false
	env.vol = 0
	env.start = false
	env.divider = 0
	env.counter = 0
}

func (env *Envelope) tick() {
	if !env.start {
		env.divider--
		if env.divider < 0 {
			env.divider = int8(env.vol)
			if env.counter > 0 {
				env.counter--
			} else if env.lenCounter.isHalted() {
				env.counter = 15
			}
		}
	} else {
		env.start = false
		env.counter = 15
		env.divider = int8(env.vol)
	}
}
