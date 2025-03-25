package blip

import (
	"math"
)

const (
	Stereo = true
	Mono   = false
)

const (
	preShift        = 32
	timeBits        = preShift + 20
	timeUnit uint64 = 1 << timeBits
)

const (
	bassShift     = 9 // affects high-pass filter breakpoint frequency
	endFrameExtra = 2 // allows deltas slightly after frame length
)

const (
	halfWidth  = 8
	bufExtra   = halfWidth*2 + endFrameExtra
	phaseBits  = 5
	phaseCount = 1 << phaseBits
	deltaBits  = 15
	deltaUnit  = 1 << deltaBits
	fracBits   = timeBits - preShift
)

const (
	maxSample = math.MaxInt16
	minSample = math.MinInt16
)

const (
	// Maximum clockRate/sampleRate ratio. For a given sampleRate,
	// lockRate must not be greater than sampleRate*MaxRatio.
	MaxRatio = 1 << 20

	// Maximum number of samples that can be generated from one time frame.
	MaxFrame = 4000
)

// Unsigned is a constraint that permits any unsigned integer type.
// If future releases of Go add new predeclared unsigned integer types,
// this constraint will be modified to include them.
type unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func clamp[T ~int | ~int32 | ~int64](n T) T {
	if T(int16(n)) != n {
		n = (n >> 16) ^ T(maxSample)
	}
	return n
}

// Buffer is a sample buffer that resamples to output rate and accumulates
// samples until they're read out.
type Buffer struct {
	factor     uint64
	offset     uint64
	avail      int
	size       int
	integrator int

	samples []int32
}

// NewBuffer creates a Buffer that can hold at most nsamples samples. Sets
// rates so that there are [MaxRatio] clocks per sample.
func NewBuffer(nsamples int) *Buffer {
	buf := &Buffer{
		samples: make([]int32, nsamples+bufExtra),
		factor:  timeUnit / MaxRatio,
		size:    nsamples,
	}
	buf.Clear()
	return buf
}

// Clear clears the entire buffer. Afterwards, SamplesAvailable() returns 0.
func (b *Buffer) Clear() {
	// We could set offset to 0, factor/2, or factor-1. 0 is suitable if factor
	// is rounded up. factor-1 is suitable if factor is rounded down. Since we
	// don't know rounding direction, factor/2 accommodates either, with the
	// slight loss of showing an error in half the time. Since for a 64-bit
	// factor this is years, the halving isn't a problem.
	b.offset = b.factor / 2
	b.avail = 0
	b.integrator = 0
	clear(b.samples)
}

// SetRates sets approximate input clock rate and output sample rate. For every
// clockRate input clocks, approximately sampleRate samples are generated.
func (b *Buffer) SetRates(clockRate, sampleRate float64) {
	factor := float64(timeUnit) * sampleRate / clockRate
	b.factor = uint64(factor)

	// Fails if clockRate exceeds maximum, relative to sampleRate
	if !(0 <= factor-float64(b.factor) && factor-float64(b.factor) < 1) {
		panic("clock rate exceeds maximum")
	}

	if float64(b.factor) < factor {
		b.factor++
	}

	// At this point, factor is most likely rounded up, but could still have
	// been rounded down in the floating-point calculation.
}

// Length of time frame, in clocks, needed to make nsamples additional samples
// available.
func (b *Buffer) ClocksNeeded(nsamples int) int {
	var needed uint64

	// Fails if buffer can't hold that many more samples
	if nsamples < 0 || b.avail+nsamples > b.size {
		panic("buffer can't hold that many samples")
	}

	needed = uint64(nsamples) * timeUnit
	if needed < uint64(b.offset) {
		return 0
	}

	return int((needed - b.offset + b.factor - 1) / b.factor)
}

// EndFrame makes input clocks before clockDuration available for reading as
// output samples. Also begins new time frame at clockDuration, so that clock
// time 0 in the new time frame specifies the same clock as clockDuration in the
// old time frame specified. Deltas can have been added slightly past
// clockDuration (up to however many clocks there are in two output samples).
func (b *Buffer) EndFrame(clockDuration int) {
	off := uint64(clockDuration)*b.factor + b.offset
	b.avail += int(off >> timeBits)
	b.offset = off & (timeUnit - 1)

	// Fails if buffer size was exceeded
	if b.avail > b.size {
		panic("buffer size exceeded")
	}
}

// SamplesAvailable reports the number of buffered samples available for
// reading.
func (b *Buffer) SamplesAvailable() int {
	return b.avail
}

func (b *Buffer) removeSamples(count int) {
	remain := b.avail + bufExtra - count
	b.avail -= count

	copy(b.samples[:remain], b.samples[count:])
	clear(b.samples[remain : remain+count])
}

// ReadSamples reads and removes at most count samples and writes them to 'out'.
// If stereo is true, writes output to every other element of 'out', allowing
// easy interleaving of two buffers into a stereo sample stream. Outputs 16-bit
// signed samples. Returns number of samples actually read.
func (b *Buffer) ReadSamples(out []int16, count int, stereo bool) int {
	if count < 0 {
		panic("count must be positive")
	}

	if count > b.avail {
		count = b.avail
	}

	if count != 0 {
		step := 2
		if !stereo {
			step = 1
		}

		sum := b.integrator
		for idx := range b.samples[:count] {
			// Eliminate fraction
			s := sum >> deltaBits
			sum += int(b.samples[idx])
			out[idx*step] = int16(clamp(s))

			// High-pass filter
			sum -= s << (deltaBits - bassShift)
		}
		b.integrator = sum
		b.removeSamples(count)
	}

	return count
}

// Sinc_Generator( 0.9, 0.55, 4.5 )
var blStep = [(phaseCount + 1) * halfWidth]int16{
	43, -115, 350, -488, 1136, -914, 5861, 21022,
	44, -118, 348, -473, 1076, -799, 5274, 21001,
	45, -121, 344, -454, 1011, -677, 4706, 20936,
	46, -122, 336, -431, 942, -549, 4156, 20829,
	47, -123, 327, -404, 868, -418, 3629, 20679,
	47, -122, 316, -375, 792, -285, 3124, 20488,
	47, -120, 303, -344, 714, -151, 2644, 20256,
	46, -117, 289, -310, 634, -17, 2188, 19985,
	46, -114, 273, -275, 553, 117, 1758, 19675,
	44, -108, 255, -237, 471, 247, 1356, 19327,
	43, -103, 237, -199, 390, 373, 981, 18944,
	42, -98, 218, -160, 310, 495, 633, 18527,
	40, -91, 198, -121, 231, 611, 314, 18078,
	38, -84, 178, -81, 153, 722, 22, 17599,
	36, -76, 157, -43, 80, 824, -241, 17092,
	34, -68, 135, -3, 8, 919, -476, 16558,
	32, -61, 115, 34, -60, 1006, -683, 16001,
	29, -52, 94, 70, -123, 1083, -862, 15422,
	27, -44, 73, 106, -184, 1152, -1015, 14824,
	25, -36, 53, 139, -239, 1211, -1142, 14210,
	22, -27, 34, 170, -290, 1261, -1244, 13582,
	20, -20, 16, 199, -335, 1301, -1322, 12942,
	18, -12, -3, 226, -375, 1331, -1376, 12293,
	15, -4, -19, 250, -410, 1351, -1408, 11638,
	13, 3, -35, 272, -439, 1361, -1419, 10979,
	11, 9, -49, 292, -464, 1362, -1410, 10319,
	9, 16, -63, 309, -483, 1354, -1383, 9660,
	7, 22, -75, 322, -496, 1337, -1339, 9005,
	6, 26, -85, 333, -504, 1312, -1280, 8355,
	4, 31, -94, 341, -507, 1278, -1205, 7713,
	3, 35, -102, 347, -506, 1238, -1119, 7082,
	1, 40, -110, 350, -499, 1190, -1021, 6464,
	0, 43, -115, 350, -488, 1136, -914, 5861,
}

// AddDelta adds positive/negative delta into buffer at specified clock time.
func (bl Buffer) AddDelta(time uint64, delta int32) {
	fixed := ((time*bl.factor + bl.offset) >> preShift)

	const phaseShift = fracBits - phaseBits
	phase := fixed >> phaseShift & (phaseCount - 1)

	interp := fixed >> (phaseShift - deltaBits) & (deltaUnit - 1)
	delta2 := (delta * int32(interp)) >> deltaBits
	delta -= delta2

	// Fails if buffer size was exceeded
	if uint64(bl.avail)+(fixed>>fracBits) > uint64(bl.size)+endFrameExtra {
		panic("buffer exceeded")
	}

	out := bl.samples[uint64(bl.avail)+(fixed>>fracBits):]

	idx := phase * halfWidth

	out[0] += int32(blStep[idx+0])*delta + int32(blStep[idx+halfWidth+0])*delta2
	out[1] += int32(blStep[idx+1])*delta + int32(blStep[idx+halfWidth+1])*delta2
	out[2] += int32(blStep[idx+2])*delta + int32(blStep[idx+halfWidth+2])*delta2
	out[3] += int32(blStep[idx+3])*delta + int32(blStep[idx+halfWidth+3])*delta2
	out[4] += int32(blStep[idx+4])*delta + int32(blStep[idx+halfWidth+4])*delta2
	out[5] += int32(blStep[idx+5])*delta + int32(blStep[idx+halfWidth+5])*delta2
	out[6] += int32(blStep[idx+6])*delta + int32(blStep[idx+halfWidth+6])*delta2
	out[7] += int32(blStep[idx+7])*delta + int32(blStep[idx+halfWidth+7])*delta2

	rev := (phaseCount - phase) * halfWidth

	out[8] += int32(blStep[rev+7])*delta + int32(blStep[rev-1])*delta2
	out[9] += int32(blStep[rev+6])*delta + int32(blStep[rev-2])*delta2
	out[10] += int32(blStep[rev+5])*delta + int32(blStep[rev-3])*delta2
	out[11] += int32(blStep[rev+4])*delta + int32(blStep[rev-4])*delta2
	out[12] += int32(blStep[rev+3])*delta + int32(blStep[rev-5])*delta2
	out[13] += int32(blStep[rev+2])*delta + int32(blStep[rev-6])*delta2
	out[14] += int32(blStep[rev+1])*delta + int32(blStep[rev-7])*delta2
	out[15] += int32(blStep[rev+0])*delta + int32(blStep[rev-8])*delta2
}

// AddDeltaFast is like AddDelta but uses faster, lower-quality synthesis.
func (bl Buffer) AddDeltaFast(time uint64, delta int32) {
	fixed := ((time*bl.factor + bl.offset) >> preShift)

	interp := fixed >> (fracBits - deltaBits) & (deltaUnit - 1)
	delta2 := (delta * int32(interp))

	// Fails if buffer size was exceeded
	if uint64(bl.avail)+(fixed>>fracBits) > uint64(bl.size)+endFrameExtra {
		panic("buffer exceeded")
	}

	out := bl.samples[uint64(bl.avail)+(fixed>>fracBits):]
	out[7] += delta*deltaUnit - delta2
	out[8] += delta2
}
