package hw

import (
	"slices"
	"unsafe"

	"github.com/arl/blip"
	"github.com/veandco/go-sdl2/sdl"

	"nestor/emu/log"
	"nestor/hw/apu"
)

const numChannels = 5 // Square1, Square2, Triangle, Noise, DMC

const maxSampleRate = 96000
const maxSamplesPerFrame = maxSampleRate / 60 * 4 * 2 //x4 to allow CPU overclocking up to 10x, x2 for panning stereo

const CycleLength = 10000
const BitsPerSample = 16

const (
	AudioFormat     = sdl.AUDIO_S16LSB
	AudioChannels   = 2
	AudioBufferSize = 4096 // TODO: adjust based on latency.
)

type AudioMixer struct {
	outbuf   [maxSamplesPerFrame]int16
	bufleft  *blip.Buffer
	bufright *blip.Buffer

	prevOutleft  int16
	prevOutright int16

	nsamples   int
	hasPanning bool

	volumes [numChannels]float64
	panning [numChannels]float64

	timestamps []uint32
	chanoutput [numChannels][CycleLength]int16
	curOutput  [numChannels]int16

	clockRate  uint32
	sampleRate uint32
}

func NewAudioMixer() *AudioMixer {
	am := &AudioMixer{
		bufleft:    blip.NewBuffer(maxSamplesPerFrame),
		bufright:   blip.NewBuffer(maxSamplesPerFrame),
		sampleRate: maxSampleRate,
	}

	return am
}

func (am *AudioMixer) Reset() {
	am.nsamples = 0

	am.prevOutleft = 0
	am.prevOutright = 0
	am.bufleft.Clear()
	am.bufright.Clear()
	am.timestamps = am.timestamps[:0]

	for i := range numChannels {
		am.volumes[i] = 1.0
		am.panning[i] = 0
	}
	clear(am.chanoutput[:])
	clear(am.curOutput[:])

	am.updateRates(true)
}

func (am *AudioMixer) PlayAudioBuffer(time uint32) {
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

	// Actuall play this with SDL2
	// copy the buffer
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&out[0])), sampleCount*2*2)
	cpy := make([]byte, len(buf))
	copy(cpy, buf)

	// play the buffer
	if err := sdl.QueueAudio(audioDeviceID, cpy); err != nil {
		log.ModSound.DebugZ("failed to queue audio buffer").Error("err", err).End()
	}

	am.nsamples = 0
	am.updateRates(false)
}

const ntscClockRate uint32 = 1789773

func (am *AudioMixer) updateRates(forceUpdate bool) {
	clockRate := ntscClockRate
	if forceUpdate || am.clockRate != clockRate {
		am.clockRate = clockRate

		am.bufleft.SetRates(float64(am.clockRate), float64(am.sampleRate))
		am.bufright.SetRates(float64(am.clockRate), float64(am.sampleRate))
	}

	// TODO: apply general volume
	// TODO: handle panning

	hasPanning := false
	for i := range numChannels {
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

func (am *AudioMixer) channelOutput(ch apu.Channel, right bool) float64 {
	if right {
		return float64(am.curOutput[ch]) * am.volumes[ch] * am.panning[ch]
	}
	return float64(am.curOutput[ch]) * am.volumes[ch] * (2.0 - am.panning[ch])
}

func (am *AudioMixer) outputVolume(isRight bool) int16 {
	squareOutput := am.channelOutput(apu.Square1, isRight) + am.channelOutput(apu.Square2, isRight)
	tndOutput := am.channelOutput(apu.DPCM, isRight) +
		2.7516713261*am.channelOutput(apu.Triangle, isRight) +
		1.8493587125*am.channelOutput(apu.Noise, isRight)

	squareVolume := uint16(((95.88 * 5000.0) / (8128.0/squareOutput + 100.0)))
	tndVolume := uint16(((159.79 * 5000.0) / (22638.0/tndOutput + 100.0)))

	return int16(squareVolume + tndVolume)
}

func (am *AudioMixer) AddDelta(ch apu.Channel, time uint32, delta int16) {
	if delta != 0 {
		am.timestamps = append(am.timestamps, time)
		am.chanoutput[ch][time] += delta
	}
}

func (am *AudioMixer) EndFrame(time uint32) {
	// Remove duplicates.
	slices.Sort(am.timestamps)
	am.timestamps = slices.Compact(am.timestamps)

	for _, stamp := range am.timestamps {
		for j := range numChannels {
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
