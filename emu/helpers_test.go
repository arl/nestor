package emu

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func CompareFileWithGolden(t *testing.T, gotfile, filename string, update bool) {
	got, err := os.ReadFile(gotfile)
	if err != nil {
		t.Fatal(err)
	}
	CompareWithGolden(t, string(got), filename, update)
}

func CompareWithGolden(t *testing.T, got, filename string, update bool) {
	t.Helper()
	if update {
		writeGolden(t, filename, got)
	} else {
		want := readGolden(t, filename)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("%s: mismatch (-want, +got):\n%s", filename, diff)
		}
	}
}

func writeGolden(t *testing.T, name string, data string) {
	if err := os.WriteFile(name, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s", name)
}

func readGolden(t *testing.T, name string) string {
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func diffFrames(t *testing.T, paths []string) {
	t.Helper()
	for _, path := range paths {
		ext := filepath.Ext(path)
		want := path[:len(path)-len(ext)] + ".golden.png"
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		CompareWithGolden(t, string(got), want, *updateGolden)
		os.Remove(path)
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
