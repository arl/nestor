package hw

import (
	"slices"
	"unsafe"

	"github.com/arl/blip"
	"github.com/veandco/go-sdl2/sdl"

	"nestor/emu/log"
	"nestor/hw/apu"
)

const MaxSampleRate = 96000
const MaxSamplesPerFrame = MaxSampleRate / 60 * 4 * 2 //x4 to allow CPU overclocking up to 10x, x2 for panning stereo
const MaxChannelCount = 11

const CycleLength = 10000
const BitsPerSample = 16

const (
	AudioFormat     = sdl.AUDIO_S16LSB
	AudioChannels   = 2
	AudioBufferSize = 4096 // TODO: adjust based on latency.
)

type AudioMixer struct {
	clockRate  uint32
	sampleRate uint32

	outbuf            [MaxSamplesPerFrame]int16
	bufleft, bufright *blip.Buffer

	prevOutleft  int16
	prevOutright int16

	sampleCount int
	hasPanning  bool

	volumes [MaxChannelCount]float64
	panning [MaxChannelCount]float64

	timestamps []uint32
	chanoutput [MaxChannelCount][CycleLength]int16
	curoutput  [MaxChannelCount]int16
}

func NewAudioMixer() *AudioMixer {
	am := &AudioMixer{
		bufleft:    blip.NewBuffer(MaxSamplesPerFrame),
		bufright:   blip.NewBuffer(MaxSamplesPerFrame),
		sampleRate: MaxSampleRate,
	}

	return am
}

func (am *AudioMixer) Reset() {
	am.sampleCount = 0

	am.prevOutleft = 0
	am.prevOutright = 0
	am.bufleft.Clear()
	am.bufright.Clear()
	am.timestamps = am.timestamps[:0]

	for i := range MaxChannelCount {
		am.volumes[i] = 1.0
		am.panning[i] = 0
	}
	clear(am.chanoutput[:])
	clear(am.curoutput[:])

	am.updateRates(true)
}

func (am *AudioMixer) PlayAudioBuffer(time uint32) {
	am.EndFrame(time)

	out := am.outbuf[am.sampleCount*2:]
	sampleCount := am.bufleft.ReadSamples(out, MaxSamplesPerFrame, blip.Stereo)

	if am.hasPanning {
		am.bufright.ReadSamples(out[1:], MaxSamplesPerFrame, blip.Stereo)
	} else {
		// Copy left channel to right channel (optimization - when no panning is used)
		copy(out[1:], out[:sampleCount*2])
	}

	am.sampleCount += sampleCount

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

	am.sampleCount = 0

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
	for i := range MaxChannelCount {
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

func (am *AudioMixer) channelOutput(channel apu.Channel, right bool) float64 {
	if right {
		return float64(am.curoutput[channel]) * am.volumes[channel] * am.panning[channel]
	}
	return float64(am.curoutput[channel]) * am.volumes[channel] * (2.0 - am.panning[channel])
}

func (am *AudioMixer) outputVolume(isRight bool) int16 {
	squareOutput := am.channelOutput(apu.Square1, isRight) + am.channelOutput(apu.Square2, isRight)
	tndOutput := /*am.channelOutput(apu.DMC, right) + */
		2.7516713261*am.channelOutput(apu.Triangle, isRight) +
			1.8493587125*am.channelOutput(apu.Noise, isRight)

	squareVolume := uint16(((95.88 * 5000.0) / (8128.0/squareOutput + 100.0)))
	tndVolume := uint16(((159.79 * 5000.0) / (22638.0/tndOutput + 100.0)))

	return int16(squareVolume + tndVolume) /* +
	am.channelOutput(apu.FDS, right)*20 +
	am.channelOutput(apu.MMC5, right)*43 +
	am.channelOutput(apu.Namco163, right)*20 +
	am.channelOutput(apu.Sunsoft5B, right)*15 +
	am.channelOutput(apu.VRC6, right)*75 +
	am.channelOutput(apu.VRC7, right))*/
}

func (am *AudioMixer) AddDelta(ch apu.Channel, time uint32, delta int16) {
	if delta != 0 {
		am.timestamps = append(am.timestamps, time)
		am.chanoutput[ch][time] += delta
	}
}

func (am *AudioMixer) EndFrame(time uint32) {
	// Remove consecutive duplicates.
	slices.Sort(am.timestamps)
	am.timestamps = slices.Compact(am.timestamps)

	for _, stamp := range am.timestamps {
		for j := range MaxChannelCount {
			am.curoutput[j] += am.chanoutput[j][stamp]
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
