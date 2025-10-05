package apu

import (
	"slices"

	"github.com/arl/blip"

	"nestor/hw/hwdefs"
	"nestor/hw/snapshot"
)

const MaxSampleRate = 96_000
const maxSamplesPerFrame = MaxSampleRate / 60 * 2

const ntscClockRate uint32 = 1_789_773

const cycleLength = 32768

// ensure cycleLength is strictly greater than the number of clocks per frame.
const _ uint32 = (cycleLength - (ntscClockRate / 60))

const (
	AudioChannels = 2 // stereo
)

type AudioBuffer struct {
	Samples []int16
}

type Mixer struct {
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

func (am *Mixer) playAudioBuffer(time uint32, buf *AudioBuffer) {
	am.EndFrame(time)

	if am.console.IsRunAheadFrame() {
		// Discard the audio samples we produced during run-ahead frame.
		am.bufleft.Clear()
		am.bufright.Clear()
	} else {
		sampleCount := am.bufleft.ReadSamples(buf.Samples, maxSamplesPerFrame, blip.Stereo)
		if am.hasPanning {
			am.bufright.ReadSamples(buf.Samples[1:], maxSamplesPerFrame, blip.Stereo)
		} else {
			// When no panning, just copy the left channel to the right one.
			for i := 0; i < sampleCount*2; i += 2 {
				buf.Samples[i+1] = buf.Samples[i]
			}
		}
		am.nsamples += sampleCount

		// TODO: apply stereo filters
		buf.Samples = buf.Samples[:sampleCount*2]
	}

	am.nsamples = 0
	am.updateRates(false)
}

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
