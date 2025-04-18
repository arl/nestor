package emu

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"

	"nestor/hw"
)

type TestingOutputConfig struct {
	// Framebuffer dimensions
	Width, Height int32

	// SaveFrameNum is the frame number to save as a PNG file.
	// The output will stop once that frame has been saved.
	//
	// NOTE: to let the emulator run indefinitely, let SaveFrameNum to 0.
	SaveFrameNum int64
	// SaveFrameFile is the filename to save the PNG file to.
	SaveFrameFile string
	// SaveFrameDir is the directory to save the PNG file to.
	SaveFrameDir string
}

type TestingOutput struct {
	framebuf []byte

	framecounter int

	cfg TestingOutputConfig
}

func newTestingOutput(cfg TestingOutputConfig) *TestingOutput {
	if cfg.SaveFrameNum == 0 {
		cfg.SaveFrameNum = math.MaxInt64
	}
	return &TestingOutput{
		framebuf: make([]byte, cfg.Width*cfg.Height*4),
		cfg:      cfg,
	}
}

func (to *TestingOutput) Close() {}

func (to *TestingOutput) BeginFrame() (frame hw.Frame) {
	return hw.Frame{Video: to.framebuf}
}

func (to *TestingOutput) framePath(isGolden bool) string {
	golden := ""
	if isGolden {
		golden = "golden."
	}
	fn := fmt.Sprintf("%s.%03d.%spng", to.cfg.SaveFrameFile, to.cfg.SaveFrameNum, golden)
	return filepath.Join(to.cfg.SaveFrameDir, fn)
}

func (to *TestingOutput) Screenshot() *image.RGBA {
	return hw.FramebufImage(to.framebuf, to.cfg.Width, to.cfg.Height)
}

func (to *TestingOutput) EndFrame(_ hw.Frame) {
	if to.framecounter == int(to.cfg.SaveFrameNum) {
		if err := hw.SaveAsPNG(to.Screenshot(), to.framePath(false)); err != nil {
			panic("failed to save frame: " + err.Error())
		}
	}

	to.framecounter++
}

func (to *TestingOutput) Poll() bool {
	return to.framecounter <= int(to.cfg.SaveFrameNum)
}

func (to *TestingOutput) CompareFrame(t *testing.T) {
	t.Helper()

	framePath := to.framePath(false)
	got, err := os.ReadFile(framePath)
	if err != nil {
		t.Fatal(err)
	}
	goldenPath := to.framePath(true)
	if *updateGolden {
		writeGolden(t, goldenPath, got)
	} else {
		want := readGolden(t, goldenPath)
		if !bytes.Equal(want, got) {
			temp := tempfilename()
			filecopy(t, temp, framePath)
			t.Logf("current frame saved for investigation at %s", temp)
			t.Errorf("%s: mismatch", goldenPath)
		}
	}

	os.Remove(framePath)
}
