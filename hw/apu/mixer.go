package apu

import (
	"slices"
	"unsafe"

	"github.com/arl/blip"
	"github.com/veandco/go-sdl2/sdl"

	"nestor/emu/log"
	"nestor/hw/hwdefs"
	"nestor/hw/snapshot"
)

const MaxSampleRate = 96000
const maxSamplesPerFrame = MaxSampleRate / 60 * 4 * 2 //x4 to allow CPU overclocking up to 10x, x2 for panning stereo

const cycleLength = 10000
const bitsPerSample = 16

const (
	AudioFormat     = sdl.AUDIO_S16LSB
	AudioChannels   = 2
	AudioBufferSize = 4096 // TODO: adjust based on latency.
)

type Mixer struct {
	outbuf   [maxSamplesPerFrame]int16
	bufleft  *blip.Buffer
	bufright *blip.Buffer

	prevOutleft  int16
	prevOutright int16

	nsamples   int
	hasPanning bool

	volumes [hwdefs.NumAudioChannels]float64
	panning [hwdefs.NumAudioChannels]float64

	timestamps []uint32
	chanoutput [hwdefs.NumAudioChannels][cycleLength]int16
	curOutput  [hwdefs.NumAudioChannels]int16

	clockRate  uint32
	sampleRate uint32

	console console
}

type console interface {
	IsRunAheadFrame() bool
}

func NewMixer(c console) *Mixer {
	am := &Mixer{
		bufleft:    blip.NewBuffer(maxSamplesPerFrame),
		bufright:   blip.NewBuffer(maxSamplesPerFrame),
		sampleRate: MaxSampleRate,
		console:    c,
	}

	return am
}

func (am *Mixer) Reset() {
	am.nsamples = 0

	am.prevOutleft = 0
	am.prevOutright = 0
	am.bufleft.Clear()
	am.bufright.Clear()
	am.timestamps = am.timestamps[:0]

	for i := range hwdefs.NumAudioChannels {
		am.volumes[i] = 1.0
		am.panning[i] = 0
	}
	clear(am.chanoutput[:])
	clear(am.curOutput[:])

	am.updateRates(true)
}

func (am *Mixer) playAudioBuffer(time uint32) {
	am.EndFrame(time)

	out := am.outbuf[am.nsamples*2:]
	sampleCount := am.bufleft.ReadSamples(out, maxSamplesPerFrame, blip.Stereo)

	if am.hasPanning {
		am.bufright.ReadSamples(out[1:], maxSamplesPerFrame, blip.Stereo)
	} else {
		// When no panning, just copy the left channel to the right one.
		for i := 0; i < sampleCount*2; i += 2 {
			out[i+1] = out[i]
		}
	}

	am.nsamples += sampleCount

	// TODO: apply stereo filters

	if !am.console.IsRunAheadFrame() {
		// Actuall play this with SDL2
		// copy the buffer
		buf := unsafe.Slice((*byte)(unsafe.Pointer(&out[0])), sampleCount*2*2)
		cpy := make([]byte, len(buf))
		copy(cpy, buf)

		if err := sdl.QueueAudio(AudioDeviceID, cpy); err != nil {
			log.ModSound.DebugZ("failed to queue audio buffer").Error("err", err).End()
		}
	}

	am.nsamples = 0
	am.updateRates(false)
}

const ntscClockRate uint32 = 1789773

func (am *Mixer) updateRates(forceUpdate bool) {
	clockRate := ntscClockRate
	if forceUpdate || am.clockRate != clockRate {
		am.clockRate = clockRate

		am.bufleft.SetRates(float64(am.clockRate), float64(am.sampleRate))
		am.bufright.SetRates(float64(am.clockRate), float64(am.sampleRate))
	}

	// TODO: apply general volume
	// TODO: handle panning

	hasPanning := false
	for i := range hwdefs.NumAudioChannels {
		am.volumes[i] = 0.8
		am.panning[i] = 1.0
		if am.panning[i] != 1.0 {
			if !am.hasPanning {
				am.bufleft.Clear()
				am.bufright.Clear()
			}
			am.hasPanning = true
		}
	}
	am.hasPanning = hasPanning
}

func (am *Mixer) channelOutput(ch Channel, right bool) float64 {
	if right {
		return float64(am.curOutput[ch]) * am.volumes[ch] * am.panning[ch]
	}
	return float64(am.curOutput[ch]) * am.volumes[ch] * (2.0 - am.panning[ch])
}

func (am *Mixer) outputVolume(isRight bool) int16 {
	squareOutput := am.channelOutput(Square1, isRight) + am.channelOutput(Square2, isRight)
	tndOutput := am.channelOutput(DPCM, isRight) +
		2.7516713261*am.channelOutput(Triangle, isRight) +
		1.8493587125*am.channelOutput(Noise, isRight)

	squareVolume := uint16(((95.88 * 5000.0) / (8128.0/squareOutput + 100.0)))
	tndVolume := uint16(((159.79 * 5000.0) / (22638.0/tndOutput + 100.0)))

	return int16(squareVolume + tndVolume)
}

func (am *Mixer) addDelta(ch Channel, time uint32, delta int16) {
	if delta != 0 {
		am.timestamps = append(am.timestamps, time)
		am.chanoutput[ch][time] += delta
	}
}

func (am *Mixer) EndFrame(time uint32) {
	// Remove duplicates.
	slices.Sort(am.timestamps)
	am.timestamps = slices.Compact(am.timestamps)

	for _, stamp := range am.timestamps {
		for j := range hwdefs.NumAudioChannels {
			am.curOutput[j] += am.chanoutput[j][stamp]
		}

		currentOut := am.outputVolume(false) * 4
		am.bufleft.AddDelta(uint64(stamp), int32(currentOut-am.prevOutleft))
		am.prevOutleft = currentOut

		if am.hasPanning {
			currentOut = am.outputVolume(true) * 4
			am.bufright.AddDelta(uint64(stamp), int32(currentOut-am.prevOutright))
			am.prevOutright = currentOut
		}
	}

	am.bufleft.EndFrame(int(time))
	if am.hasPanning {
		am.bufright.EndFrame(int(time))
	}

	// Reset everything.
	am.timestamps = am.timestamps[:0]
	for i := range am.chanoutput {
		clear(am.chanoutput[i][:])
	}
}

func (am *Mixer) State() *snapshot.APUMixer {
	var state snapshot.APUMixer
	state.ClockRate = am.clockRate
	state.SampleRate = am.sampleRate

	state.PreviousOutputLeft = am.prevOutleft
	state.PreviousOutputRight = am.prevOutright
	for i := range hwdefs.NumAudioChannels {
		state.CurrentOutput[i] = am.curOutput[i]
	}

	return &state
}

func (am *Mixer) SetState(state *snapshot.APUMixer) {
	am.clockRate = state.ClockRate
	am.sampleRate = state.SampleRate

	am.Reset()
	am.updateRates(true)

	am.prevOutleft = state.PreviousOutputLeft
	am.prevOutright = state.PreviousOutputRight

	for i := range hwdefs.NumAudioChannels {
		am.curOutput[i] = state.CurrentOutput[i]
	}
}
