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
	for i := range paths {
		got, err := os.ReadFile(paths[i])
		if err != nil {
			t.Fatal(err)
		}

		if *updateGolden {
			if err := os.WriteFile(paths[i]+".golden", got, 0644); err != nil {
				t.Fatal(err)
			}
			return
		}

		want, err := os.ReadFile(paths[i] + ".golden")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("frame %d differs. check %s", i, paths[i])
		}

		if err := os.Remove(paths[i]); err != nil {
			t.Logf("failed to remove %s: %v", paths[i], err)
		}
	}
}

func goldenPathIndex(idx int, path, ext string) string {
	_, fn := filepath.Split(path)
	return fmt.Sprintf("%s-%02d.%s", fn[:len(fn)-len(filepath.Ext(fn))], idx, ext)
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

		fn := filepath.Join("testdata", goldenPathIndex(i, path, "png"))
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
