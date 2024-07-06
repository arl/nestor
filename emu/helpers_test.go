package emu

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func diffFrames(t *testing.T, paths []string) {
	for _, path := range paths {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		ext := filepath.Ext(path)
		golden := path[:len(path)-len(ext)] + ".golden.png"

		if *updateGolden {
			if err := os.WriteFile(golden, got, 0644); err != nil {
				t.Fatal(err)
			}
			return
		}

		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("frame differs. check %s", path)
		}

		if err := os.Remove(path); err != nil {
			t.Logf("failed to remove %s: %v", path, err)
		}
	}
}

// saveFrames reads the frames from the channel and saves the frames with the
// provided indices as png files into "testdata/orgpath-<index>.png".
func saveFrames(frames <-chan image.RGBA, path string, indices ...int) ([]string, error) {
	var paths []string
	slices.Sort(indices)
	for i := range indices[len(indices)-1] + 1 {
		frame := <-frames

		if !slices.Contains(indices, i) {
			// skip frame
			continue
		}

		_, file := filepath.Split(path)
		ext := filepath.Ext(file)
		fn := fmt.Sprintf("%s.%03d.png", file[:len(file)-len(ext)], i)
		fn = filepath.Join("testdata", fn)

		f, err := os.Create(fn)
		if err != nil {
			return nil, err
		}
		if err := png.Encode(f, &frame); err != nil {
			return nil, fmt.Errorf("error encoding frame %d: %v", i, err)
		}
		if err := f.Close(); err != nil {
			return nil, err
		}

		paths = append(paths, fn)
	}

	return paths, nil
}
