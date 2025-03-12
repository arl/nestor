[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=round-square)](https://pkg.go.dev/github.com/arl/blip)

# Blip


Blip is a low-level audio Go library to resample audio waveforms from input clock rate to output sample rate.

Examples:

| Package                                       | Description                                                           |
|-----------------------------------------------|-----------------------------------------------------------------------|
| [demo_basic](./examples/demo_basic/main.go)   | Generates square wave sweep                                           |
| [demo_stereo](./examples/demo_stereo/main.go) | Generates stereo sound using two blip buffers                         |
| [demo_fixed](./examples/demo_fixed/main.go)   | Works in fixed-point time rather than clocks                          |
| [demo_sdl](./examples/demo_sdl/main.go)       | Plays sound live using SDL multimedia library                         |
| [demo_chip](./examples/demo_chip/main.go)     | Emulates sound hardware and plays back log.txt                        |
| [wave](./wave/wave.go)                        | Simple package demonstrating a wave sound file write, used by demos   |



This library is a **pure-go** port of the C library called **blip_buf** by Shay Green (blargg).  
Here's the original README from the C library, from which blip also kept the
original license (function names are those of the Go library though):

blip_buf
--------------
Author  : Shay Green <gblargg@gmail.com>
Website : http://www.slack.net/~ant/
License : GNU Lesser General Public License (LGPL)


Contents
--------
* Overview
* Buffer creation
* Waveform generation
* Time frames
* Complex waveforms
* Sample buffering
* Thanks


Overview
--------
This library resamples audio waveforms from input clock rate to output
sample rate. Usage follows this general pattern:

* Create buffer with blip.NewBuffer().
* Set clock rate and sample rate with buf.SetRates().
* Waveform generation loop:
	- Generate several clocks of waveform with buf.AddDelta().
	- End time frame with buf.EndFrame().
	- Read samples from buffer with buf.ReadSamples().


Buffer creation
---------------
Before synthesis, a buffer must be created with blip.NewBuffer(). Its
size is the maximum number of unread samples it can hold. For most uses,
this can be 1/10 the sample rate or less, since samples will usually be
read out immediately after being generated.

After the buffer is created, the input clock rate and output sample rate
must be set with buf.SetRates(). This determines how many input clocks
there are per second, and how many output samples are generated per
second.

If the compiler supports a 64-bit integer type, then the input-output
ratio is stored very accurately. If the compiler only supports a 32-bit
integer type, then the ratio is stored with only 20 fraction bits, so
some ratios cannot be represented exactly (for example, sample
rate=48000 and clock rate=48001). The ratio is internally rounded up, so
there will never be fewer than 'sample rate' samples per second. Having
too many per second is generally better than having too few.


Waveform generation
-------------------
Waveforms are generated at the input clock rate. Consider a simple
square wave with 8 clocks per cycle (4 clocks high, 4 clocks low):

                   |<-- 8 clocks ->|
        +5|        ._._._._        ._._._._        ._._._._        ._._
          |        |       |       |       |       |       |       |
    Amp  0|._._._._        |       |       |       |       |       |
          |                |       |       |       |       |       |
        -5|                ._._._._        ._._._._        ._._._._ 
           * . . . * . . . * . . . * . . . * . . . * . . . * . . . * .
    Time   0       4       8      12      16      20      24      28

The wave changes amplitude at time points 0, 4, 8, 12, 16, etc.

The following generates the amplitude at every clock of above waveform
at the input clock rate:

	int wave [30];
	
	for ( int i = 4; i < 30; ++i )
	{
		if ( i % 8 < 4 )
			wave [i] = -5;
		else
			wave [i] = +5;
	}

Without this library, the wave array would then need to be resampled
from the input clock rate to the output sample rate. This library does
this resampling internally, so it won't be discussed further; waveform
generation code can focus entirely on the input clocks.

Rather than specify the amplitude at every clock, this library merely
needs to know the points where the amplitude CHANGES, referred to as a
delta. The time of a delta is specified with a clock count. The deltas
for this square wave are shown below the time points they occur at:

        +5|        ._._._._        ._._._._        ._._._._        ._._
          |        |       |       |       |       |       |       |
    Amp  0|._._._._        |       |       |       |       |       |
          |                |       |       |       |       |       |
        -5|                ._._._._        ._._._._        ._._._._ 
           * . . . * . . . * . . . * . . . * . . . * . . . * . . . * .
    Time   0       4       8      12      16      20      24      28
    Delta         +5     -10     +10     -10     +10     -10     +10

The following calls generate the above waveform:

	buf.AddDelta(  4,  +5 );
	buf.AddDelta(  8, -10 );
	buf.AddDelta( 12, +10 );
	buf.AddDelta( 16, -10 );
	buf.AddDelta( 20, +10 );
	buf.AddDelta( 24, -10 );
	buf.AddDelta( 28, +10 );

In the examples above, the amplitudes are small for clarity. The 16-bit
sample range is -32768 to +32767, so actual waveform amplitudes would
need to be in the thousands to be audible (for example, -5000 to +5000).

This library allows waveform generation code to pay NO attention to the
output sample rate. It can focus ENTIRELY on the essence of the
waveform: the points where its amplitude changes. Since these points can
be efficiently generated in a loop, synthesis is efficient. Sound chip
emulation code can be structured to allow full accuracy down to a single
clock, with the emulated CPU being able to simply tell the sound chip to
"emulate from wherever you left off, up to clock time T within the
current time frame".


Time frames
-----------
Since time keeps increasing, if left unchecked, at some point it would
overflow the range of an integer. This library's solution to the problem
is to break waveform generation into time frames of moderate length.
Clock counts within a time frame are thus relative to the beginning of
the frame, where 0 is the beginning of the frame. When a time frame of
length T is ended, what was at time T in the old time frame is now at
time 0 in the new time frame. Breaking the above waveform into time
frames of 10 clocks each looks like this:

        +5|        ._._._._        ._._._._        ._._._._        ._._
          |        |       |       |       |       |       |       |
    Amp  0|._._._._        |       |       |       |       |       |
          |                |       |       |       |       |       |
        -5|                ._._._._        ._._._._        ._._._._ 
           * . . . * . . . * . . . * . . . * . . . * . . . * . . . * .
    Time  |0       4       8  |    2       6      |0       4       8  |
          | first time frame  | second time frame | third time frame  |
          |<--- 10 clocks --->|<--- 10 clocks --->|<--- 10 clocks --->|

The following calls generate the above waveform. After they execute, the
first 30 clocks of the waveform will have been resampled and be
available as output samples for reading with buf.ReadSamples().

	buf.AddDelta( 4,  +5 );
	buf.AddDelta( 8, -10 );
	buf.EndFrame( 10 );
	
	buf.AddDelta( 2, +10 );
	buf.AddDelta( 6, -10 );
	buf.EndFrame( 10 );
	
	buf.AddDelta( 0, +10 );
	buf.AddDelta( 4, -10 );
	buf.AddDelta( 8, +10 );
	buf.EndFrame( 10 );
	...

Time frames can be a convenient length, and the length can vary from one
frame to the next. Once a time frame is ended, the resulting output
samples become available for reading immediately, and no more deltas can
be added to it.

There is a limit of about 4000 output samples per time frame. The number
of clocks depends on the clock rate. At common sample rates, this allows
time frames of at least 1/15 second, plenty for most uses. This limit
allows increased resampling ratio accuracy.

In an emulator, it is usually convenient to have audio time frames
correspond to video frames, where the CPU's clock counter is reset at
the beginning of each video frame and thus can be used directly as the
relative clock counts for audio time frames.


Complex waveforms
-----------------
Any sort of waveform can be generated, not just a square wave. For
example, a saw-like wave:

        +5|        ._._._._                ._._._._                ._._
          |        |       |               |       |               |
    Amp  0|._._._._        |       ._._._._        |       ._._._._
          |                |       |               |       |
        -5|                ._._._._                ._._._._
           * . . . * . . . * . . . * . . . * . . . * . . . * . . . * .
    Time   0       4       8      12      16      20      24      28
    Delta         +5     -10      +5      +5     -10      +5      +5

Code to generate above waveform:

	buf.AddDelta(  4,  +5 );
	buf.AddDelta(  8, -10 );
	buf.AddDelta( 12,  +5 );
	buf.AddDelta( 16,  +5 );
	buf.AddDelta( 20, +10 );
	buf.AddDelta( 24,  +5 );
	buf.AddDelta( 28,  +5 );

Similarly, multiple waveforms can be added within a time frame without
problem. It doesn't matter what order they're added, because all the
library needs are the deltas. The synthesis code doesn't need to know
all the waveforms at once either; it can calculate and add the deltas
for each waveform individually. Deltas don't need to be added in
chronological order either.


Sample buffering
----------------
Sample buffering is very flexible. Once a time frame is ended, the
resampled waveforms become output samples that are immediately made
available for reading with buf.ReadSamples(). They don't have to be
read immediately; they can be allowed to accumulate in the buffer, with
each time frame appending more samples to the buffer. When reading, some
or all of the samples in can be read out, with the remaining unread
samples staying in the buffer for later. Usually a program will
immediately read all available samples after ending a time frame and
play them immediately. In some systems, a program needs samples in
fixed-length blocks; in that case, it would keep generating time frames
until some number of samples are available, then read only that many,
even if slightly more were available in the buffer.

In some systems, one wants to run waveform generation for exactly the
number of clocks necessary to generate some desired number of output
samples, and no more. In that case, use buf.ClocksNeeded( N ) to
find out how many clocks are needed to generate N additional samples.
Ending a time frame with this value will result in exactly N more
samples becoming available for reading.


Thanks
------
Thanks to Jsr (FamiTracker author), the Mednafen team (multi-system
emulator), ShizZie (Nhes GMB author), Marcel van Tongeren, Luke Molnar
(UberNES author), Fredrick Meunier (Fuse contributor) for using and
giving feedback for another similar library. Thanks to Disch for his
interest and discussions about the synthesis algorithm itself, and for
writing his own implementation of it (Schpune) rather than just using
mine. Thanks to Xodnizel for Festalon, whose sound quality got me
interested in video game sound emulation in the first place, and where I
first came up with the algorithm while optimizing its brute-force
filter.

-- 
Shay Green <gblargg@gmail.com>
