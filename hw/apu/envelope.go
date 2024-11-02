package apu

type Envelope struct {
	constantVolume bool
	volume         uint8

	start   bool
	divider int8
	counter uint8

	LengthCounter LengthCounter
}

func (env *Envelope) InitializeEnvelope(regValue uint8) {
	env.LengthCounter.Init((regValue & 0x20) == 0x20)
	env.constantVolume = (regValue & 0x10) == 0x10
	env.volume = regValue & 0x0F
}

func (env *Envelope) ResetEnvelope() {
	env.start = true
}

func (env *Envelope) Volume() uint32 {
	if env.LengthCounter.Status() {
		if env.constantVolume {
			return uint32(env.volume)
		}
		return uint32(env.counter)
	}
	return 0
}

func (env *Envelope) Reset(soft bool) {
	env.LengthCounter.Reset(soft)
	env.constantVolume = false
	env.volume = 0
	env.start = false
	env.divider = 0
	env.counter = 0
}

func (env *Envelope) Tick() {
	if !env.start {
		env.divider--
		if env.divider < 0 {
			env.divider = int8(env.volume)
			if env.counter > 0 {
				env.counter--
			} else if env.LengthCounter.IsHalted() {
				env.counter = 15
			}
		}
	} else {
		env.start = false
		env.counter = 15
		env.divider = int8(env.volume)
	}
}
