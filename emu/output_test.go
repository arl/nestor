package emu

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type FrameSaver struct {
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

func (fs *FrameSaver) Drive(ctx context.Context, framech <-chan *Frame) {
	var framecounter int64

	for {
		select {
		case <-ctx.Done():
			return

		case f := <-framech:

			if framecounter == fs.SaveFrameNum {
				img := FramebufImage(f.Video, fs.Width, fs.Height)
				if err := SaveAsPNG(img, fs.framePath(false)); err != nil {
					panic("failed to save frame: " + err.Error())
				}
				return
			}

			framecounter++
		}
	}
}

func (fs FrameSaver) framePath(isGolden bool) string {
	golden := ""
	if isGolden {
		golden = "golden."
	}
	fn := fmt.Sprintf("%s.%03d.%spng", fs.SaveFrameFile, fs.SaveFrameNum, golden)
	return filepath.Join(fs.SaveFrameDir, fn)
}

func (fs *FrameSaver) CompareFrame(t *testing.T) {
	t.Helper()

	framePath := fs.framePath(false)
	got, err := os.ReadFile(framePath)
	if err != nil {
		t.Fatal(err)
	}
	goldenPath := fs.framePath(true)
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
