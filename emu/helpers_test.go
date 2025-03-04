package emu

import (
	"flag"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func CompareFileWithGolden(t *testing.T, gotfile, filename string, update bool) {
	got, err := os.ReadFile(gotfile)
	if err != nil {
		t.Fatal(err)
	}
	CompareWithGolden(t, got, filename, update)
}

func CompareWithGolden(t *testing.T, got []byte, pathGolden string, update bool) {
	t.Helper()
	if update {
		writeGolden(t, pathGolden, got)
	} else {
		want := readGolden(t, pathGolden)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("%s: mismatch (-want, +got):\n%s", pathGolden, diff)
		}
	}
}

func writeGolden(t *testing.T, path string, data []byte) {
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("error writing golden file %s: %s", path, err)
	}
	t.Logf("wrote %s", path)
}

func readGolden(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("error reading golden file %s: %s", path, err)
	}
	return data
}

func filecopy(tb testing.TB, dst, src string) {
	srcf, err := os.Open(src)
	if err != nil {
		tb.Fatalf("failed to open %s: %v", src, err)
	}
	defer srcf.Close()

	dstf, err := os.Create(dst)
	if err != nil {
		tb.Fatalf("failed to create %s: %v", dst, err)
	}
	defer dstf.Close()

	if _, err = io.Copy(dstf, srcf); err != nil {
		tb.Fatalf("failed to copy %s to %s: %v", src, dst, err)
	}
}

func tempfilename() string {
	f, err := os.CreateTemp("", "nestor")
	if err != nil {
		panic("failed to create temp file: " + err.Error())
	}
	f.Close()
	return f.Name()
}
