package ppu

type TVStandard struct {
	FramesPerSecond   int
	Frameµs           int
	Scanlines         int
	VBlank            int
	CyclesPerScanline float64
	Width, Height     int
	CPUSpeed          float64
}

var NTSC = TVStandard{
	FramesPerSecond:   60,
	Frameµs:           16670,
	Scanlines:         262,
	VBlank:            20,
	CyclesPerScanline: 113.33,
	Width:             256,
	Height:            224,
	CPUSpeed:          1.79,
}

var PAL = TVStandard{
	FramesPerSecond:   50,
	Frameµs:           20000,
	Scanlines:         312,
	VBlank:            70,
	CyclesPerScanline: 106.56,
	Width:             256,
	Height:            240,
	CPUSpeed:          1.66,
}
